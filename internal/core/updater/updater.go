package updater

import (
	"context"
	"fmt"
	"os"
	"parm/internal/core/installer"
	"parm/internal/gh"
	"parm/internal/manifest"
	"parm/internal/parmutil"
	"parm/pkg/progress"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v74/github"
)

// TODO: modify updater to use new symlinking logic

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
func (up *Updater) Update(ctx context.Context, owner, repo string, installPath string, flags *UpdateFlags, hooks *progress.Hooks) (*UpdateResult, error) {
	installDir := parmutil.GetInstallDir(owner, repo)
	man, err := manifest.Read(installDir)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("package %s/%s does not exist", owner, repo)
		}
		return nil, fmt.Errorf("could not read manifest for %s/%s: \n%w", owner, repo, err)
	}

	var rel *github.RepositoryRelease

	switch man.InstallType {
	case manifest.Release:
		rel, _, err = up.client.GetLatestRelease(ctx, owner, repo)
	case manifest.PreRelease:
		rel, err = gh.GetLatestPreRelease(ctx, up.client, owner, repo)
		// TODO: DRY @installer.go
		if !flags.Strict {
			// expensive!
			relStable, _, err := up.client.GetLatestRelease(ctx, owner, repo)
			if err != nil {
				return nil, err
			}

			// TODO: abstract elsewhere cuz it's similar to updater.NeedsUpdate?
			currV, _ := semver.NewVersion(rel.GetTagName())
			newV, _ := semver.NewVersion(relStable.GetTagName())
			if newV.GreaterThan(currV) {
				rel = relStable
			}
		}
	}

	newVer := rel.GetTagName()
	needsUpdate, err := NeedsUpdate(man.Version, newVer)

	if !needsUpdate {
		return nil, fmt.Errorf("%s/%s is already up to date (ver %s).\n", owner, repo, man.Version)
	}

	if err != nil {
		return nil, fmt.Errorf("could not fetch latest release for %s/%s: \n%w", owner, repo, err)
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

func NeedsUpdate(currVer, latestVer string) (bool, error) {
	currSemVer, err := semver.NewVersion(currVer)
	if err != nil {
		return false, err
	}
	latestSemVer, err := semver.NewVersion(latestVer)
	if err != nil {
		return false, err
	}

	if latestSemVer.GreaterThan(currSemVer) {
		return true, nil
	}

	return false, nil
}
