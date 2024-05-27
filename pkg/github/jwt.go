// Package github provides ..
package github

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v62/github"
)

const (
	acceptHeader = "application/vnd.github.v3+json"
)

// ApplicationToken represents a GitHub App token.
type ApplicationToken struct {
	appID      string
	privateKey *rsa.PrivateKey
}

// NewApplicationToken creates a new GitHub App token.
// An application token is used to authenticate as a GitHub App.
func NewApplicationToken(appID string, privateKey []byte) (*ApplicationToken, error) {
	if appID == "" {
		return nil, errors.New("applicationID is required")
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil, err
	}

	return &ApplicationToken{
		appID:      appID,
		privateKey: privKey,
	}, nil
}

// Token creates a new GitHub App token.
// The token is used to authenticate as a GitHub App.
// Each token is valid for 10 minutes.
func (t *ApplicationToken) Token() (string, error) {
	now := time.Now()
	expiresAt := now.Add(10 * time.Minute)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Issuer:    t.appID,
	})

	return token.SignedString(t.privateKey)
}

// InstallationAccessToken represents a GitHub App installation access token.
type InstallationAccessToken struct {
	Token        string                         `json:"token"`
	ExpiresAt    time.Time                      `json:"expires_at"`
	Permissions  github.InstallationPermissions `json:"permissions,omitempty"`
	Repositories []github.Repository            `json:"repositories,omitempty"`
}

// InstallationToken represents a GitHub App installation token.
type InstallationToken struct {
	ApplicationToken *ApplicationToken
	TokenOptions     *github.InstallationTokenOptions
	InstallationID   int64
	accessToken      InstallationAccessToken

	mu sync.Mutex
}

// IsExpired returns true if the token is expired.
func (t *InstallationToken) IsExpired() bool {
	return t.accessToken.Token == "" || t.accessToken.ExpiresAt.Before(time.Now())
}

// Token generates a new GitHub installation token.
func (t *InstallationToken) Token(ctx context.Context) (string, error) {
	if t.InstallationID == 0 {
		return "", errors.New("InstallationID is required")
	}
	if t.ApplicationToken == nil {
		return "", errors.New("ApplicationToken is required")
	}

	if t.IsExpired() {
		if err := t.refresh(ctx); err != nil {
			return "", err
		}
	}

	return t.accessToken.Token, nil
}

func (t *InstallationToken) refresh(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	appToken, err := t.ApplicationToken.Token()
	if err != nil {
		return err
	}

	client := &http.Client{}

	var reqBody io.Reader
	if t.TokenOptions != nil {
		body, err := json.Marshal(t.TokenOptions)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(body)
	}

	reqURL := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", t.InstallationID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", appToken))
	req.Header.Set("Accept", acceptHeader)

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&t.accessToken); err != nil {
		return err
	}

	return nil
}
