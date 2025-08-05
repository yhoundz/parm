package gh

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/v74/github"
)

func GetNLatestReleases(
	ctx context.Context,
	client *github.Client,
	owner, repo string,
	n int) ([]*github.RepositoryRelease, error) {
	rels, _, err := client.Repositories.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: n})
	if err != nil {
		return nil, fmt.Errorf("list releases: %w", err)
	}
	if len(rels) == 0 {
		return nil, fmt.Errorf("no releases for %s/%s", owner, repo)
	}

	return rels, nil
}

func ValidateRelease(
	ctx context.Context,
	client *github.Client,
	owner, repo, releaseTag string) (bool, *github.RepositoryRelease, error) {

	repository, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, releaseTag)

	if err == nil {
		return true, repository, nil
	}

	var ghErr *github.ErrorResponse

	if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
		// release does not exist
		return false, nil, nil
	}

	// error parsing release
	return false, nil, err
}
