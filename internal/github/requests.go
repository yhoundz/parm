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
	client *github.RepositoriesService,
	owner, repo string,
	n int) ([]*github.RepositoryRelease, error) {
	rels, _, err := client.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: n})
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
	client *github.RepositoriesService,
	owner, repo, releaseTag string) (bool, *github.RepositoryRelease, error) {

	repository, _, err := client.GetReleaseByTag(ctx, owner, repo, releaseTag)

	if err == nil {
		return true, repository, nil
	}

	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) && ghErr.Response.StatusCode != http.StatusOK {
		// release does not exist
		return false, nil, nil
	}

	// error parsing release
	return false, nil, err
}

func ValidateBranch(
	ctx context.Context,
	client *github.RepositoriesService,
	owner, repo, branch string) (bool, *github.Branch, error) {

	br, _, err := client.GetBranch(ctx, owner, repo, branch, 0)

	if err == nil {
		return true, br, nil
	}

	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) && ghErr.Response.StatusCode != http.StatusOK {
		return false, nil, nil
	}

	return false, nil, err
}

func ValidateCommit(
	ctx context.Context,
	client *github.RepositoriesService,
	owner, repo, commitSha string) (bool, *github.RepositoryCommit, error) {

	commit, _, err := client.GetCommit(ctx, owner, repo, commitSha, nil)

	if err == nil {
		return true, commit, nil
	}

	var ghErr *github.ErrorResponse
	if errors.As(err, &ghErr) && ghErr.Response.StatusCode != http.StatusOK {
		return false, nil, nil
	}

	return false, nil, err
}
