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

func ValidatePreRelease(
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

func ResolveRelease(ctx context.Context, client *github.RepositoriesService, owner, repo, version string, preRelease bool) (*github.RepositoryRelease, error) {
	if preRelease {
		valid, rel, err := ValidatePreRelease(ctx, client, owner, repo)
		if err != nil {
			return nil, fmt.Errorf("err: cannot resolve pre-release on %s/%s: \n%w", owner, repo, err)
		}
		if !valid {
			return nil, fmt.Errorf("error: no valid pre-release found for %s/%s", owner, repo)
		}

		return rel, nil
	}

	if version != "" {
		valid, rel, err := ValidateRelease(ctx, client, owner, repo, version)
		if err != nil {
			return nil, fmt.Errorf("error: Cannot resolve release %s on %s/%s", version, owner, repo)
		}
		if !valid {
			return nil, fmt.Errorf("error: Release %s not valid, \n%w", version, err)
		}
		return rel, nil
	} else {
		rel, _, err := client.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			var ghErr *github.ErrorResponse
			if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
				return nil, fmt.Errorf("error: no stable release found for %s/%s", owner, repo)
			}
			return nil, fmt.Errorf("error: could not fetch latest release: \n%w", err)
		}
		return rel, nil
	}
}
