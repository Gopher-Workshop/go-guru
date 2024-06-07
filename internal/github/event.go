package github

import (
	"context"
	"log/slog"

	"github.com/google/go-github/v62/github"
)

// InstallationClientRetriever is an interface for retrieving a GitHub client for a given installation ID.
type InstallationClientRetriever interface {
	Client(id int64) *github.Client
}

// PullRequestWelcomeEvent is a handler for the pull request opened event.
// It will create the first comment on the pull request with a welcome message.
type PullRequestWelcomeEvent struct {
	Logger                      *slog.Logger
	InstallationClientRetriever InstallationClientRetriever
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

	githubClient := e.InstallationClientRetriever.Client(event.GetInstallation().GetID())

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
