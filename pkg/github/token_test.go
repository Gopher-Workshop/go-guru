package github

import (
	"testing"
	"time"
)

func TestToken_Type(t *testing.T) {
	tests := []struct {
		name     string
		token    Token
		expected string
	}{
		{
			name: "default token type",
			token: Token{
				TokenType: "",
			},
			expected: "Bearer",
		},
		{
			name: "Bearer token type",
			token: Token{
				TokenType: "Bearer",
			},
			expected: "Bearer",
		},
		{
			name: "MAC token type",
			token: Token{
				TokenType: "MAC",
			},
			expected: "MAC",
		},
		{
			name: "Basic token type",
			token: Token{
				TokenType: "Basic",
			},
			expected: "Basic",
		},
		{
			name: "custom token type",
			token: Token{
				TokenType: "Custom",
			},
			expected: "Custom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.Type()
			if got != tt.expected {
				t.Errorf("Token.Type() = %s, expected %s", got, tt.expected)
			}
		})
	}
}

func TestToken_Valid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		token Token
		valid bool
	}{
		{
			name: "valid token",
			token: Token{
				AccessToken: "abc123",
				Expiry:      time.Now().Add(time.Hour),
			},
			valid: true,
		},
		{
			name: "without expiry",
			token: Token{
				AccessToken: "ghi789",
			},
			valid: true,
		},
		{
			name: "expired token",
			token: Token{
				AccessToken: "def456",
				Expiry:      time.Now().Add(-time.Hour),
			},
		},
		{
			name: "empty access token",
			token: Token{
				AccessToken: "",
				Expiry:      time.Now().Add(time.Hour),
			},
		},
		{
			name: "cusom delta",
			token: Token{
				AccessToken: "abc123",
				Expiry:      now.Add(time.Minute),
				expiryDelta: time.Minute,
				now: func() time.Time {
					return now.Add(30 * time.Second)
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.token.Valid()
			if got != tt.valid {
				t.Errorf("Token.Valid() = %v, expected %v", got, tt.valid)
			}
		})
	}
}
