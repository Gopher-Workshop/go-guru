package github

import (
	"context"
	"log/slog"

	"github.com/google/go-github/v62/github"
)

// PullRequestOpenedEvent represents a pull request opened event.
type PullRequestOpenedEvent struct {
	Logger *slog.Logger
}

// Handle handles the pull request opened event.
func (e *PullRequestOpenedEvent) Handle(ctx context.Context, event *github.PullRequestEvent) error {
	// TODO(jferrl): maybe we should check if the event data is ok.

	e.Logger.Info("pull request opened", slog.String("repository", event.GetRepo().GetFullName()), slog.Int("PR", event.GetPullRequest().GetNumber()))

	githubClient := github.NewClient(nil)

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
