// Package main defines entry point of the application.
package main

import (
	"net/http"
	"os"

	echo "github.com/labstack/echo/v4"
)

var port = os.Getenv("PORT")

func main() {
	if port == "" {
		port = "8080"
	}

	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello from go-guru!")
	})

	e.POST("/github/events", func(c echo.Context) error {
		return c.String(http.StatusOK, "Pong!")
	})

	e.Logger.Fatal(e.Start(":" + port))
}
