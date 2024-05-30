package github

import (
	"net/http"
	"strings"
	"time"
)

const (
	// defaultExpiryDelta determines how earlier a token should be considered
	// expired than its actual expiration time. It is used to avoid late
	// expirations due to client-server time mismatches.
	defaultExpiryDelta = 10 * time.Second
)

// Token represents an Github token.
type Token struct {
	// AccessToken is the token that authorizes and authenticates
	// the requests.
	AccessToken string `json:"access_token"`
	// TokenType is the type of token.
	// The Type method returns either this or "Bearer", the default.
	TokenType string `json:"token_type,omitempty"`
	// Expiry is the optional expiration time of the access token.
	Expiry time.Time `json:"expiry,omitempty"`
	// expiryDelta is used to calculate when a token is considered
	// expired, by subtracting from Expiry. If zero, defaultExpiryDelta
	// is used.
	expiryDelta time.Duration

	// now is used to override time.Now in tests.
	now func() time.Time
}

// SetAuthHeader sets the Authorization header on the request.
func (t Token) SetAuthHeader(req *http.Request) {
	req.Header.Set("Authorization", t.Type()+" "+t.AccessToken)
}

// Type returns the token type.
func (t *Token) Type() string {
	if strings.EqualFold(t.TokenType, "bearer") {
		return "Bearer"
	}
	if strings.EqualFold(t.TokenType, "mac") {
		return "MAC"
	}
	if strings.EqualFold(t.TokenType, "basic") {
		return "Basic"
	}
	if t.TokenType != "" {
		return t.TokenType
	}
	return "Bearer"
}

// expired reports whether the token is expired.
// t must be non-nil.
func (t *Token) expired() bool {
	if t.Expiry.IsZero() {
		return false
	}

	expiryDelta := defaultExpiryDelta
	if t.expiryDelta != 0 {
		expiryDelta = t.expiryDelta
	}
	return t.Expiry.Round(0).Add(-expiryDelta).Before(time.Now())
}

// Valid reports whether t is non-nil, has an AccessToken, and is not expired.
func (t *Token) Valid() bool {
	return t != nil && t.AccessToken != "" && !t.expired()
}
