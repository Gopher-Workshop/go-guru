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

	logger := slog.Default()

	handle := githubevents.New(webhookSecretKey)

	e.POST("/github/event", func(c echo.Context) error {
		if err := handle.HandleEventRequest(c.Request()); err != nil {
			logger.Error("Error handling event request: %v", err)
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
