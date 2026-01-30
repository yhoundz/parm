package gh

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/go-github/v74/github"
)

func GetLatestPreRelease(
	ctx context.Context,
	client *github.RepositoriesService,
	owner, repo string,
) (*github.RepositoryRelease, error) {
	// WARNING: this doesn't always work, especially if the latest pre-release is not within the past 30 (?) releases, or if maintainer releases versions out of order
	rels, _, err := client.ListReleases(ctx, owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("could not list releases for %s/%s: \n%w", owner, repo, err)
	}

	for _, rel := range rels {
		if rel.GetPrerelease() {
			return rel, nil
		}
	}

	return nil, nil
}

func validatePreRelease(
	ctx context.Context,
	client *github.RepositoriesService,
	owner, repo string,
) (bool, *github.RepositoryRelease, error) {
	rel, err := GetLatestPreRelease(ctx, client, owner, repo)

	if err != nil {
		return false, nil, err
	}

	if rel != nil {
		return true, rel, nil
	}

	return false, nil, nil
}

func validateRelease(
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

func ResolvePreRelease(ctx context.Context, client *github.RepositoriesService, owner, repo string) (*github.RepositoryRelease, error) {
	valid, rel, err := validatePreRelease(ctx, client, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("err: cannot resolve pre-release on %s/%s: \n%w", owner, repo, err)
	}
	if !valid {
		return nil, fmt.Errorf("error: no valid pre-release found for %s/%s", owner, repo)
	}

	return rel, nil
}

// Retrieves a GitHub RepositoryRelease. If provided version string is nil, then return the latest stable release
func ResolveReleaseByTag(ctx context.Context, client *github.RepositoriesService, owner, repo string, version *string) (*github.RepositoryRelease, error) {
	if version != nil {
		valid, rel, err := validateRelease(ctx, client, owner, repo, *version)
		if err != nil {
			return nil, fmt.Errorf("error: Cannot resolve release %s on %s/%s", *version, owner, repo)
		}
		if !valid {
			return nil, fmt.Errorf("error: Release %s not valid", *version)
		}
		return rel, nil
	} else {
		rel, _, err := client.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			var ghErr *github.ErrorResponse
			if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
				return nil, fmt.Errorf("error: no stable release found for %s/%s (repo may be missing or token lacks permission to view releases)", owner, repo)
			}
			return nil, fmt.Errorf("error: could not fetch latest release: \n%w", err)
		}
		return rel, nil
	}
}
