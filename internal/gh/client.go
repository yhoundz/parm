package gh

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

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
	for _, env := range []string{"PARM_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"} {
		if val, ok := os.LookupEnv(env); ok && val != "" {
			return strings.TrimSpace(val), nil
		}
	}

	var tok string
	tok = v.GetString("github_api_token")
	if tok == "" {
		tok = v.GetString("github_api_token_fallback")
	}

	tok = strings.TrimSpace(tok)
	if tok == "" {
		return "", fmt.Errorf("error: api key not found")
	}

	return tok, nil
}

// RepositoryVisibility represents the visibility state of a repository
type RepositoryVisibility int

const (
	RepoPublic RepositoryVisibility = iota
	RepoPrivate
	RepoNotFound
)

// CheckRepositoryVisibility checks if a repository is public, private, or doesn't exist
func CheckRepositoryVisibility(ctx context.Context, client *github.RepositoriesService, owner, repo string) (RepositoryVisibility, error) {
	repo_obj, resp, err := client.Get(ctx, owner, repo)
	if err != nil {
		// If we get a 404, the repo doesn't exist or is private (can't tell without auth)
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return RepoNotFound, nil
		}
		return RepoNotFound, err
	}

	if repo_obj.GetPrivate() {
		return RepoPrivate, nil
	}
	return RepoPublic, nil
}

// IsRepositoryPublic checks if a repository is public by attempting to access it without authentication
// Deprecated: Use CheckRepositoryVisibility for better error handling
func IsRepositoryPublic(ctx context.Context, client *github.RepositoriesService, owner, repo string) (bool, error) {
	visibility, err := CheckRepositoryVisibility(ctx, client, owner, repo)
	if err != nil {
		return false, err
	}
	return visibility == RepoPublic, nil
}
