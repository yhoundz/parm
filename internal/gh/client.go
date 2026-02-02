package gh

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
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
		if val, ok := os.LookupEnv(env); ok {
			val = strings.TrimSpace(val)
			if val != "" {
				return val, nil
			}
		}
	}

	var tok string
	tok = v.GetString("github_api_token")
	if tok == "" {
		tok = v.GetString("github_api_token_fallback")
	}

	tok = strings.TrimSpace(tok)
	if tok == "" {
		if tok = getTokenFromGitCredential(); tok != "" {
			return tok, nil
		}
		return "", fmt.Errorf("error: api key not found (set PARM_GITHUB_TOKEN/GITHUB_TOKEN/GH_TOKEN, github_api_token_fallback, or a git credential helper)")
	}

	return tok, nil
}

var gitCredentialRunner = runGitCredentialFill

func runGitCredentialFill() (string, error) {
	cmd := exec.Command("git", "credential", "fill")
	cmd.Stdin = strings.NewReader("protocol=https\nhost=github.com\n")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func getTokenFromGitCredential() string {
	out, err := gitCredentialRunner()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "password=") {
			return strings.TrimPrefix(line, "password=")
		}
	}
	return ""
}
