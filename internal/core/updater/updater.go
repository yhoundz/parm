package updater

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"parm/internal/core/installer"
	"parm/internal/gh"
	"parm/internal/manifest"
	"parm/pkg/progress"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v74/github"
)

// TODO: modify updater to use new symlinking logic
// 12-21-25 ngl idk what ts means lol

type Updater struct {
	client    *github.RepositoriesService
	installer installer.Installer
}

type UpdateStatus string

const (
	StatusUpdated   UpdateStatus = "updated"
	StatusUpToDate  UpdateStatus = "up-to-date"
	StatusNoRelease UpdateStatus = "no-release"
	StatusSkipped   UpdateStatus = "skipped"
)

type UpdateResult struct {
	OldManifest *manifest.Manifest
	*installer.InstallResult
	Status UpdateStatus
}

type UpdateFlags struct {
	Strict bool
}

func New(cli *github.RepositoriesService, rel *installer.Installer) *Updater {
	return &Updater{
		client:    cli,
		installer: *rel,
	}
}

// TODO: update concurrently?
func (up *Updater) Update(ctx context.Context, owner, repo string, installPath string, man *manifest.Manifest, flags *UpdateFlags, hooks *progress.Hooks) (*UpdateResult, error) {
	if man == nil {
		return nil, fmt.Errorf("could not update %s/%s: package manifest not found (is it installed?)", owner, repo)
	}

	var rel *github.RepositoryRelease
	var err error

	switch man.InstallType {
	case manifest.Release:
		rel, _, err = up.client.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			var ghErr *github.ErrorResponse
			if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
				// Distinguish between "repo not found" and "no releases found"
				visibility, vErr := gh.CheckRepositoryVisibility(ctx, up.client, owner, repo)
				if vErr != nil {
					return nil, err // Return original error if we can't check visibility
				}
				if visibility == gh.RepoNotFound {
					return nil, fmt.Errorf("repository %s/%s not found or private", owner, repo)
				}
				return &UpdateResult{Status: StatusNoRelease, OldManifest: man}, nil
			}
			return nil, err
		}
	case manifest.PreRelease:
		rel, err = gh.GetLatestPreRelease(ctx, up.client, owner, repo)
		if err != nil {
			var ghErr *github.ErrorResponse
			if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
				visibility, vErr := gh.CheckRepositoryVisibility(ctx, up.client, owner, repo)
				if vErr == nil && visibility == gh.RepoNotFound {
					return nil, fmt.Errorf("repository %s/%s not found or private", owner, repo)
				}
				return &UpdateResult{Status: StatusNoRelease, OldManifest: man}, nil
			}
			return nil, err
		}
		if rel == nil {
			return &UpdateResult{Status: StatusNoRelease, OldManifest: man}, nil
		}
		// TODO: DRY @installer.go
		if !flags.Strict {
			// expensive!
			relStable, _, err := up.client.GetLatestRelease(ctx, owner, repo)
			if err != nil {
				var ghErr *github.ErrorResponse
				if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
					// No stable release found, stick with pre-release
				} else {
					return nil, err
				}
			}

			if relStable != nil {
				// TODO: abstract elsewhere cuz it's similar to updater.NeedsUpdate?
				currVer, _ := semver.NewVersion(rel.GetTagName())
				stableVer, _ := semver.NewVersion(relStable.GetTagName())
				if stableVer.GreaterThan(currVer) {
					rel = relStable
				}
			}
		}
	default:
		return nil, fmt.Errorf("unsupported install type %q for %s/%s", man.InstallType, owner, repo)
	}

	newVer := rel.GetTagName()

	// only need to check for equality
	if man.Version == newVer {
		return &UpdateResult{Status: StatusUpToDate, OldManifest: man}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("could not fetch latest release for %s/%s: %w", owner, repo, err)
	}

	opts := installer.InstallFlags{
		Type:        man.InstallType,
		Version:     &newVer,
		Asset:       nil,
		Strict:      flags.Strict,
		VerifyLevel: 0,
	}

	res, err := up.installer.Install(ctx, owner, repo, installPath, opts, hooks)
	if err != nil {
		return nil, err
	}
	actual := UpdateResult{
		OldManifest:   man,
		InstallResult: res,
		Status:        StatusUpdated,
	}
	return &actual, nil
}
