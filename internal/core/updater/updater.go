package updater

import (
	"context"
	"fmt"
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

type UpdateResult struct {
	OldManifest *manifest.Manifest
	*installer.InstallResult
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
	var rel *github.RepositoryRelease
	var err error

	switch man.InstallType {
	case manifest.Release:
		rel, _, err = up.client.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			return nil, err
		}
	case manifest.PreRelease:
		rel, err = gh.GetLatestPreRelease(ctx, up.client, owner, repo)
		if err != nil {
			return nil, err
		}
		// TODO: DRY @installer.go
		if !flags.Strict {
			// expensive!
			relStable, _, err := up.client.GetLatestRelease(ctx, owner, repo)
			if err != nil {
				return nil, err
			}

			// TODO: abstract elsewhere cuz it's similar to updater.NeedsUpdate?
			currVer, _ := semver.NewVersion(rel.GetTagName())
			stableVer, _ := semver.NewVersion(relStable.GetTagName())
			if stableVer.GreaterThan(currVer) {
				rel = relStable
			}
		}
	}

	newVer := rel.GetTagName()

	// only need to check for equality
	if man.Version == newVer {
		return nil, fmt.Errorf("%s/%s is already up to date (ver %s)", owner, repo, man.Version)
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
	}
	return &actual, nil
}
