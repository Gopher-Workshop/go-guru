package github

import (
	"context"
	"crypto/rsa"
	"errors"
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/google/go-github/v62/github"
)

// TokenSource is the interface that wraps the Token method.
type TokenSource interface {
	Token() (*Token, error)
}

// applicationTokenSource represents a GitHub App token.
type applicationTokenSource struct {
	appID      string
	privateKey *rsa.PrivateKey
}

// NewApplicationTokenSource creates a new GitHub App token source.
// An application token is used to authenticate as a GitHub App.
func NewApplicationTokenSource(appID string, privateKey []byte) (TokenSource, error) {
	if appID == "" {
		return nil, errors.New("applicationID is required")
	}

	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil, err
	}

	return &applicationTokenSource{
		appID:      appID,
		privateKey: privKey,
	}, nil
}

// Token creates a new GitHub App token.
// The token is used to authenticate as a GitHub App.
// Each token is valid for 10 minutes.
func (t *applicationTokenSource) Token() (*Token, error) {
	now := time.Now()
	expiresAt := now.Add(10 * time.Minute)

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Issuer:    t.appID,
	})

	tokenString, err := token.SignedString(t.privateKey)
	if err != nil {
		return nil, err
	}

	return &Token{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		Expiry:      expiresAt,
	}, nil
}

// InstallationTokenSourceOpt is a functional option for InstallationTokenSource.
type InstallationTokenSourceOpt func(*installationTokenSource)

// WithInstallationTokenOptions sets the options for the GitHub App installation token.
func WithInstallationTokenOptions(opts *github.InstallationTokenOptions) InstallationTokenSourceOpt {
	return func(i *installationTokenSource) {
		i.opts = opts
	}
}

// WithHTTPClient sets the HTTP client for the GitHub App installation token source.
func WithHTTPClient(c *http.Client) InstallationTokenSourceOpt {
	return func(i *installationTokenSource) {
		c.Transport = &Transport{
			Source: i.src,
			Base:   c.Transport,
		}
		i.github = github.NewClient(c)
	}
}

// InstallationTokenSource represents a GitHub App installation token source.
type installationTokenSource struct {
	id     int64
	src    TokenSource
	github *github.Client
	opts   *github.InstallationTokenOptions
}

// NewInstallationTokenSource creates a new GitHub App installation token source.
func NewInstallationTokenSource(id int64, src TokenSource, opts ...InstallationTokenSourceOpt) TokenSource {
	client := &http.Client{
		Transport: &Transport{
			Source: src,
		},
	}

	i := &installationTokenSource{
		id:     id,
		src:    src,
		github: github.NewClient(client),
	}

	for _, opt := range opts {
		opt(i)
	}

	return i
}

// Token creates a new GitHub App installation token.
// The token is used to authenticate as a GitHub App installation.
func (t *installationTokenSource) Token() (*Token, error) {
	ctx := context.Background()

	token, _, err := t.github.Apps.CreateInstallationToken(ctx, t.id, t.opts)
	if err != nil {
		return nil, err
	}

	return &Token{
		AccessToken: token.GetToken(),
		TokenType:   "Bearer",
		Expiry:      token.GetExpiresAt().Time,
	}, nil
}
