package github

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/sashabaranov/go-openai"
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

// PullRequestFirstReviewEvent is a handler for the pull request first review event.
type PullRequestFirstReviewEvent struct {
	Logger                      *slog.Logger
	AssistantID                 string
	OpenAI                      *openai.Client
	InstallationClientRetriever InstallationClientRetriever
}

// Handle handles the pull request first review event.
func (e *PullRequestFirstReviewEvent) Handle(deliveryID string, eventName string, event *github.PullRequestEvent) error {
	ctx := context.Background()

	logger := e.Logger.With(
		slog.String("delivery_id", deliveryID),
		slog.String("event", eventName),
		slog.String("repository", event.GetRepo().GetFullName()),
		slog.Int("PR", event.GetPullRequest().GetNumber()),
	)

	logger.Info("Handling pull request first review event")

	githubClient := e.InstallationClientRetriever.Client(event.GetInstallation().GetID())

	owner := event.GetRepo().GetOwner().GetLogin()
	repo := event.GetRepo().GetName()
	number := event.GetPullRequest().GetNumber()

	files, _, err := githubClient.PullRequests.ListFiles(ctx, owner, repo, number, nil)
	if err != nil {
		e.Logger.With(slog.Any("error", err)).Error("error getting pull request")
		return err
	}

	type fileRun struct {
		file *github.CommitFile
		run  openai.Run
	}

	runs := make([]fileRun, 0, len(files))

	for _, file := range files {
		logger.Info("File changed", slog.String("file", file.GetFilename()), slog.String("content", file.GetPatch()))

		createdRun, err := e.OpenAI.CreateThreadAndRun(ctx, openai.CreateThreadAndRunRequest{
			Thread: openai.ThreadRequest{
				Messages: []openai.ThreadMessage{
					{
						Role:    openai.ThreadMessageRoleUser,
						Content: file.GetPatch(),
					},
				},
			},
			RunRequest: openai.RunRequest{
				AssistantID: e.AssistantID,
			},
		})
		if err != nil {
			logger.With(slog.Any("error", err)).Error("error creating thread and running assistant")
			continue
		}

		runs = append(runs, fileRun{
			file: file,
			run:  createdRun,
		})
	}

	for _, run := range runs {
		logger.Info("Runs created", slog.Any("runs", runs))
		go func(r fileRun) {
			for {
				steps, err := e.OpenAI.ListRunSteps(ctx, r.run.ThreadID, r.run.ID, openai.Pagination{})
				if err != nil {
					logger.With(slog.Any("error", err)).Error("error listing run steps")
					return
				}

				for _, step := range steps.RunSteps {
					if step.Status == openai.RunStepStatusCompleted {
						message, err := e.OpenAI.RetrieveMessage(ctx, r.run.ThreadID, step.StepDetails.MessageCreation.MessageID)
						if err != nil {
							logger.With(slog.Any("error", err)).Error("error retrieving message")
							return
						}

						for _, content := range message.Content {
							if content.Text != nil {
								_, _, err := githubClient.PullRequests.CreateComment(ctx, owner, repo, number, &github.PullRequestComment{
									Body:      github.String(content.Text.Value),
									Path:      r.file.Filename,
									CommitID:  event.GetPullRequest().GetHead().SHA,
									StartLine: github.Int(1),
									Line:      github.Int(2),
								})
								if err != nil {
									logger.With(slog.Any("error", err)).Error("error creating comment")
								}

							}
						}

						return
					}
				}
				time.Sleep(2 * time.Second)
			}
		}(run)
	}
	return nil
}
