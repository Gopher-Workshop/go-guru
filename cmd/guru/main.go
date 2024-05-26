// Package main defines entry point of the application.
package main

import (
	"fmt"
	"net/http"
	"os"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var port = os.Getenv("PORT")

var (
	githubOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"repo"},
		Endpoint:     github.Endpoint,
	}
)

func main() {
	if port == "" {
		port = "8080"
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from go-guru!")
	})

	e.POST("/github/event", func(c echo.Context) error {
		return c.String(http.StatusOK, "Pong!")
	})

	e.GET("/auth/github/callback", func(c echo.Context) error {
		code := c.QueryParam("code")
		if code == "" {
			return c.String(http.StatusBadRequest, "Missing code query parameter")
		}

		token, err := githubOAuthConfig.Exchange(oauth2.NoContext, code)
		if err != nil {
			return c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to exchange code for token: %s", err.Error()))
		}

		return c.JSON(http.StatusOK, token)
	})

	e.Logger.Fatal(e.Start(":" + port))
}
