// Package main defines entry point of the application.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	port          = os.Getenv("PORT")
	applicationID = os.Getenv("GITHUB_APP_ID")
)

func main() {
	if port == "" {
		port = "8080"
	}

	privateKey := loadPrivateKey("/etc/secrets/private-key.pem")

	installationIDs, err := getInstallationID(applicationID, privateKey)
	if err != nil {
		log.Fatalf("Unable to get installation IDs: %v", err)
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

	e.GET("/github/installations", func(c echo.Context) error {
		return c.String(http.StatusOK, fmt.Sprintf("%v", installationIDs))
	})

	e.Logger.Fatal(e.Start(":" + port))
}

func getInstallationID(appID string, privateKey []byte) ([]int64, error) {
	jwt, err := generateJWT(appID, privateKey)
	if err != nil {
		return nil, err
	}

	url := "https://api.github.com/app/installations"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwt))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var installations []struct {
		ID int64 `json:"id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&installations)
	if err != nil {
		return nil, err
	}

	var ids []int64
	for _, installation := range installations {
		ids = append(ids, installation.ID)
	}
	return ids, nil
}

func generateJWT(appID string, privateKey []byte) (string, error) {
	now := time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": now,
		"exp": now + 600, // 10 minutes expiration
		"iss": appID,
	})
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func loadPrivateKey(path string) []byte {
	key, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read private key: %v", err)
	}
	return key
}
