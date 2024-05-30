package github

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"golang.org/x/oauth2"
)

func Test_installationTokenSource_Token(t *testing.T) {
	now := time.Now().UTC()
	expiration := now.Add(10 * time.Minute)

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.PostAppInstallationsAccessTokensByInstallationId,
			github.InstallationToken{
				Token: github.String("mocked-installation-token"),
				ExpiresAt: &github.Timestamp{
					Time: expiration,
				},
				Permissions: &github.InstallationPermissions{
					PullRequests: github.String("read"),
				},
				Repositories: []*github.Repository{
					{
						Name: github.String("mocked-repo-1"),
						ID:   github.Int64(1),
					},
				},
			},
		),
	)

	privateKey, err := os.ReadFile("testdata/private-key.pem")
	if err != nil {
		t.Fatal(err)
	}

	appSrc, err := NewApplicationTokenSource("app-id", privateKey)
	if err != nil {
		t.Fatal(err)
	}

	type fields struct {
		id   int64
		src  oauth2.TokenSource
		opts []InstallationTokenSourceOpt
	}
	tests := []struct {
		name    string
		fields  fields
		want    *oauth2.Token
		wantErr bool
	}{
		{
			name: "generate a new installation token",
			fields: fields{
				id:  1,
				src: appSrc,
				opts: []InstallationTokenSourceOpt{
					WithInstallationTokenOptions(&github.InstallationTokenOptions{}),
					WithHTTPClient(mockedHTTPClient),
				},
			},
			want: &oauth2.Token{
				AccessToken: "mocked-installation-token",
				TokenType:   "Bearer",
				Expiry:      expiration,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewInstallationTokenSource(tt.fields.id, tt.fields.src, tt.fields.opts...)
			got, err := tr.Token()
			if (err != nil) != tt.wantErr {
				t.Errorf("installationTokenSource.Token() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("installationTokenSource.Token() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewApplicationTokenSource(t *testing.T) {
	type args struct {
		appID      string
		privateKey []byte
	}
	tests := []struct {
		name    string
		args    args
		want    oauth2.TokenSource
		wantErr bool
	}{
		{
			name:    "application id is not provided",
			args:    args{},
			wantErr: true,
		},
		{
			name:    "private key is not provided",
			args:    args{appID: "app-id"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewApplicationTokenSource(tt.args.appID, tt.args.privateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewApplicationTokenSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewApplicationTokenSource() = %v, want %v", got, tt.want)
			}
		})
	}
}
