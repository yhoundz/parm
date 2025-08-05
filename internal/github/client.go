package gh

import (
	"context"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

func NewClient(ctx context.Context, token string) *github.Client {
	if token == "" {
		return github.NewClient(nil)
	}

	// TODO: change this to handle token expiry??
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return github.NewClient(oauth2.NewClient(ctx, src))
}
