// Package github provides ..
package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

// ApplicationToken represents a GitHub App token.
type ApplicationToken struct {
	ApplicationID string
	PrivateKey    []byte
	ExpiresAt     time.Time
}

// Generate generates a new GitHub App token.
func (t *ApplicationToken) Generate() (string, error) {
	if t.ApplicationID == "" {
		return "", errors.New("ApplicationID is required")
	}

	if len(t.PrivateKey) == 0 {
		return "", errors.New("PrivateKey is required")
	}

	now := time.Now()

	if t.ExpiresAt.IsZero() {
		t.ExpiresAt = now.Add(10 * time.Minute)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Unix(),
		"exp": t.ExpiresAt.Unix(),
		"iss": t.ApplicationID,
	})

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(t.PrivateKey)
	if err != nil {
		return "", err
	}

	tokenStr, err := token.SignedString(privKey)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

// InstallationToken represents a GitHub App installation token.
type InstallationToken struct {
	ApplicationToken

	InstallationID string
}

// Generate generates a new GitHub App installation token.
func (t *InstallationToken) Generate() (string, error) {
	if t.InstallationID == "" {
		return "", errors.New("InstallationID is required")
	}

	appToken, err := t.ApplicationToken.Generate()
	if err != nil {
		return "", err
	}

	client := &http.Client{}

	reqURL := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", t.InstallationID)

	req, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", appToken))
	req.Header.Set("Accept", "application/vnd.github+json")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var resBody struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}

	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return "", err
	}

	return resBody.Token, nil
}
