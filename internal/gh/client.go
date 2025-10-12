package gh

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v74/github"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

type Provider interface {
	Repos() *github.RepositoriesService
	Search() *github.SearchService
}

type client struct {
	c *github.Client
}

func (cli *client) Repos() *github.RepositoriesService { return cli.c.Repositories }
func (cli *client) Search() *github.SearchService      { return cli.c.Search }

type Option func(*clientOptions)

type clientOptions struct {
	hc *http.Client
}

func WithHTTPClient(hc *http.Client) Option {
	return func(c *clientOptions) {
		c.hc = hc
	}
}

func New(ctx context.Context, token string, opts ...Option) Provider {
	var cliOpts clientOptions
	for _, opt := range opts {
		opt(&cliOpts)
	}

	hc := cliOpts.hc
	if hc == nil && token != "" {
		src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		hc = oauth2.NewClient(ctx, src)
	}

	cli := github.NewClient(hc)
	return &client{
		c: cli,
	}
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
