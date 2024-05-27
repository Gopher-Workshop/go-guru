// Package github provides ..
package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

// ApplicationToken represents a GitHub App token.
type ApplicationToken struct {
	ApplicationID int
	PrivateKey    []byte
	expiresAt     time.Time
	token         string
}

// Expired returns true if the token is expired.
func (t *ApplicationToken) Expired() bool {
	return t.token == "" || t.expiresAt.Before(time.Now())
}

// Token returns the string representation of the token.
func (t *ApplicationToken) Token() (string, error) {
	if t.ApplicationID == 0 {
		return "", errors.New("ApplicationID is required")
	}
	if len(t.PrivateKey) == 0 {
		return "", errors.New("PrivateKey is required")
	}

	if t.Expired() {
		if err := t.refresh(); err != nil {
			return "", err
		}
	}

	return t.token, nil

}

func (t *ApplicationToken) refresh() error {
	now := time.Now()

	t.expiresAt = now.Add(10 * time.Minute)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now.Unix(),
		"exp": t.expiresAt.Unix(),
		"iss": t.ApplicationID,
	})

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(t.PrivateKey)
	if err != nil {
		return err
	}

	t.token, err = token.SignedString(privKey)
	return err
}

// InstallationToken represents a GitHub App installation token.
type InstallationToken struct {
	ApplicationToken *ApplicationToken
	InstallationID   int64
	expiresAt        time.Time
	token            string
}

// Expired returns true if the token is expired.
func (t *InstallationToken) Expired() bool {
	return t.token == "" || t.expiresAt.Before(time.Now())
}

// Token generates a new GitHub installation token.
func (t *InstallationToken) Token() (string, error) {
	if t.InstallationID == 0 {
		return "", errors.New("InstallationID is required")
	}
	if t.ApplicationToken == nil {
		return "", errors.New("ApplicationToken is required")
	}

	if t.Expired() {
		if err := t.refresh(); err != nil {
			return "", err
		}
	}

	return t.token, nil
}

func (t *InstallationToken) refresh() error {
	appToken, err := t.ApplicationToken.Token()
	if err != nil {
		return err
	}

	client := &http.Client{}

	reqURL := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", t.InstallationID)

	req, err := http.NewRequest(http.MethodPost, reqURL, http.NoBody)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", appToken))
	req.Header.Set("Accept", "application/vnd.github+json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var resBody struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}

	if err := json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return err
	}

	t.token = resBody.Token
	t.expiresAt = resBody.ExpiresAt

	return nil
}
