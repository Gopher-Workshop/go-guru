// Package main defines entry point of the application.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/cbrgm/githubevents/githubevents"
	"github.com/google/go-github/v62/github"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	port              = os.Getenv("PORT")
	applicationID     = os.Getenv("GITHUB_APP_ID")
	appPrivateKeyPath = os.Getenv("GITHUB_APP_PRIVATE_KEY_PATH")
	webhookSecretKey  = os.Getenv("GITHUB_WEBHOOK_SECRET")
)

func main() {
	_ = context.Background()

	if port == "" {
		port = "8080"
	}
	if appPrivateKeyPath == "" {
		appPrivateKeyPath = "../../certs/private-key.pem"
	}

	_ = loadPrivateKey(appPrivateKeyPath)

	e := echo.New()
	e.Use(middleware.Recover())

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	whHandler := githubevents.New(webhookSecretKey)

	whHandler.OnPullRequestEventOpened(
		func(deliveryID, eventName string, event *github.PullRequestEvent) error {
			logger.With(
				slog.String("delivery_id", deliveryID),
				slog.String("event_name", eventName),
			).Info("Received pull request event")
			return nil
		},
	)

	e.POST("/github/event", func(c echo.Context) error {
		eventLogger := logger.WithGroup("github").With(
			slog.String("event", c.Request().Header.Get("X-GitHub-Event")),
			slog.String("delivery_id", c.Request().Header.Get("X-GitHub-Delivery")),
		)

		eventLogger.Info("Received GitHub event")

		if err := whHandler.HandleEventRequest(c.Request()); err != nil {
			eventLogger.With(slog.Any("error", err)).Error("Error handling event request")
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Error handling event request: %v", err))
		}

		return c.String(http.StatusOK, "")
	})

	e.Logger.Fatal(e.Start(":" + port))
}

func loadPrivateKey(path string) []byte {
	key, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read private key: %v", err)
	}
	return key
}
