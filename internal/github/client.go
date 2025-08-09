package gh

import (
	"context"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

func NewRepoClient(ctx context.Context, token string) *github.RepositoriesService {
	var cli *github.Client
	if token == "" {
		cli = github.NewClient(nil)
	}

	// TODO: change this to handle token expiry??
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	cli = github.NewClient(oauth2.NewClient(ctx, src))
	return cli.Repositories
}
