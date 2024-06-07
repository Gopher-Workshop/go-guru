// Package main defines entry point of the application.
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	githubguru "github.com/Gopher-Workshop/guru/internal/github"
	"github.com/google/uuid"
	"github.com/jferrl/go-githubauth"

	"github.com/cbrgm/githubevents/githubevents"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/oauth2"
)

var (
	port              = os.Getenv("PORT")
	applicationID     = os.Getenv("GITHUB_APP_ID")
	appPrivateKeyPath = os.Getenv("GITHUB_APP_PRIVATE_KEY_PATH")
	webhookSecretKey  = os.Getenv("GITHUB_WEBHOOK_SECRET")
)

func main() {
	ctx := context.Background()

	if port == "" {
		port = "8080"
	}
	if appPrivateKeyPath == "" {
		appPrivateKeyPath = "../../certs/private-key.pem"
	}

	appID, err := strconv.ParseInt(applicationID, 10, 64)
	if err != nil {
		log.Fatalf("Unable to parse application ID: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(ctx, slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.LogAttrs(ctx, slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))
	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return uuid.NewString()
		},
	}))
	e.Use(middleware.Secure())

	whHandler := githubevents.New(webhookSecretKey)

	appTokenSrc, err := githubauth.NewApplicationTokenSource(appID, loadPrivateKey(appPrivateKeyPath))
	if err != nil {
		log.Fatalf("Unable to create application token source: %v", err)
	}

	// Reuse the token source to avoid creating a new token for each request.
	appTokenSrc = oauth2.ReuseTokenSource(nil, appTokenSrc)

	installations := githubguru.NewInstallations(appTokenSrc)

	welcomeEvent := &githubguru.PullRequestWelcomeEvent{
		InstallationClientRetriever: installations,
		Logger:                      logger.WithGroup("github.PullRequestEventOpened.Welcome"),
	}

	whHandler.OnPullRequestEventOpened(
		welcomeEvent.Handle,
	)

	e.POST("/github/event", func(c echo.Context) error {
		eventLogger := logger.WithGroup("github").With(
			slog.String("event", c.Request().Header.Get("X-GitHub-Event")),
			slog.String("delivery_id", c.Request().Header.Get("X-GitHub-Delivery")),
		)

		eventLogger.Info("Received GitHub event")

		if err := whHandler.HandleEventRequest(c.Request()); err != nil {
			eventLogger.With(slog.Any("error", err)).Error("Error handling event request")
			return c.String(http.StatusBadRequest, fmt.Sprintf("Error handling event request: %v", err))
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
