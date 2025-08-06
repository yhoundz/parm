package gh

import (
	"context"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

type RepoClient interface {
	ListReleases(ctx context.Context, owner, repo string, opt *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error)
	GetBranch(ctx context.Context, owner, repo, branch string, maxRedirects int) (*github.Branch, *github.Response, error)
	GetCommit(ctx context.Context, owner, repo, sha string, opts *github.ListOptions) (*github.RepositoryCommit, *github.Response, error)
}

func NewRepoClient(ctx context.Context, token string) RepoClient {
	var cli *github.Client
	if token == "" {
		cli = github.NewClient(nil)
	}

	// TODO: change this to handle token expiry??
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	cli = github.NewClient(oauth2.NewClient(ctx, src))
	return cli.Repositories
}
