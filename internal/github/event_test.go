package github

import (
	"log/slog"
	"testing"

	"github.com/google/go-github/v62/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

type installationClientRetriever struct {
	client *github.Client
}

func (r *installationClientRetriever) Client(_ int64) *github.Client {
	return r.client
}

func TestPullRequestWelcomeEvent_Handle(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.PostReposIssuesCommentsByOwnerByRepoByIssueNumber,
			github.IssueComment{
				ID: github.Int64(0),
			},
		),
	)

	type fields struct {
		InstallationClientRetriever InstallationClientRetriever
	}
	type args struct {
		deliveryID string
		eventName  string
		event      *github.PullRequestEvent
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "error creating comment",
			fields: fields{
				InstallationClientRetriever: &installationClientRetriever{
					client: github.NewClient(mockedHTTPClient),
				},
			},
			args:    args{},
			wantErr: true,
		},
		{
			name: "create welcome comment on pull request",
			fields: fields{
				InstallationClientRetriever: &installationClientRetriever{
					client: github.NewClient(mockedHTTPClient),
				},
			},
			args: args{
				deliveryID: "123",
				eventName:  "pull_request",
				event: &github.PullRequestEvent{
					Repo: &github.Repository{
						Owner: &github.User{
							Login: github.String("owner"),
						},
						Name: github.String("repo"),
					},
					Installation: &github.Installation{
						ID: github.Int64(1),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &PullRequestWelcomeEvent{
				Logger:                      slog.Default(),
				InstallationClientRetriever: tt.fields.InstallationClientRetriever,
			}
			if err := e.Handle(tt.args.deliveryID, tt.args.eventName, tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("PullRequestWelcomeEvent.Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
