package github

import (
	"context"
	"log/slog"

	githubpkg "github.com/Gopher-Workshop/guru/pkg/github"
	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

// PullRequestWelcomeEvent is a handler for the pull request opened event.
// It will create the first comment on the pull request with a welcome message.
type PullRequestWelcomeEvent struct {
	Logger                 *slog.Logger
	ApplicationTokenSource oauth2.TokenSource
}

// Handle handles the pull request opened event.
func (e *PullRequestWelcomeEvent) Handle(deliveryID string, eventName string, event *github.PullRequestEvent) error {
	ctx := context.Background()

	logger := e.Logger.With(
		slog.String("delivery_id", deliveryID),
		slog.String("event", eventName),
		slog.String("repository", event.GetRepo().GetFullName()),
		slog.Int("PR", event.GetPullRequest().GetNumber()),
	)

	logger.Info("Handling pull request welcome event")

	ins := githubpkg.NewInstallationTokenSource(event.Installation.GetID(), e.ApplicationTokenSource)

	githubClient := github.NewClient(oauth2.NewClient(ctx, ins))

	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	number := event.GetPullRequest().GetNumber()

	_, _, err := githubClient.Issues.CreateComment(ctx, owner, repo, number, &github.IssueComment{
		Body: github.String(botWelcomeMessage),
	})
	if err != nil {
		e.Logger.With(slog.Any("error", err)).Error("error creating comment")
		return err
	}

	return nil
}
