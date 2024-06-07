package github

import (
	"context"
	"sync"

	"github.com/google/go-github/v62/github"
	"github.com/jferrl/go-githubauth"
	"golang.org/x/oauth2"
)

// Installations represents a cache repository for GitHub App installations.
// wihin the context of the application.
// The token source is used to create a new GitHub client.
// It provides the necessary application authentication for the GitHub API.
type Installations struct {
	src           oauth2.TokenSource
	installations map[int64]*github.Client

	mu sync.Mutex
}

// NewInstallations creates a new cache repository for GitHub App installations.
func NewInstallations(src oauth2.TokenSource) *Installations {
	return &Installations{
		src:           src,
		installations: make(map[int64]*github.Client),
	}
}

// Client returns a GitHub client for the given installation ID.
// If the client is not already cached, it will be created and stored.
// The client is created using the provided token source.
func (i *Installations) Client(id int64) *github.Client {
	i.mu.Lock()
	defer i.mu.Unlock()

	if client, ok := i.installations[id]; ok {
		return client
	}

	cli := github.NewClient(
		oauth2.NewClient(
			context.Background(),
			githubauth.NewInstallationTokenSource(id, i.src),
		),
	)

	i.installations[id] = cli

	return cli
}
