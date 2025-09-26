package gh

import (
	"context"
	"fmt"

	"github.com/google/go-github/v74/github"
	"github.com/spf13/viper"
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

// returns the current API key, or nil if there is none
func GetStoredApiKey(v *viper.Viper) (string, error) {
	var tok string
	tok = v.GetString("github_api_token")
	if tok == "" {
		tok = v.GetString("github_api_token_fallback")
		if tok == "" {
			return "", fmt.Errorf("error: api key not found")
		}
	}

	return tok, nil
}
