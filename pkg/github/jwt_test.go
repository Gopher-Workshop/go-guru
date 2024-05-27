// Package github provides ..
package github

import (
	"os"
	"testing"
)

func TestApplicationToken_Token(t *testing.T) {
	privateKey, err := os.ReadFile("testdata/private-key.pem")
	if err != nil {
		t.Fatalf("Error reading private key: %v", err)
	}

	type fields struct {
		appID      string
		privateKey []byte
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "application id is required",
			fields:  fields{},
			wantErr: true,
		},
		{
			name:    "invalid private key",
			fields:  fields{appID: "123"},
			wantErr: true,
		},
		{
			name: "generate token",
			fields: fields{
				appID:      "123",
				privateKey: privateKey,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := NewApplicationToken(tt.fields.appID, tt.fields.privateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewApplicationToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tr == nil {
				return
			}

			got, err := tr.Token()
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplicationToken.Token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == "" {
				t.Error("ApplicationToken.Token() = empty token, want a token")
			}
		})
	}
}
