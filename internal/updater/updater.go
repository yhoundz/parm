package updater

import (
	"context"
	"fmt"
	"os"
	"parm/internal/installer"
	"parm/internal/manifest"
	"parm/internal/utils"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v74/github"
)

type Updater struct {
	client       *github.RepositoriesService
	relInstaller installer.ReleaseInstaller
}

func New(cli *github.RepositoriesService, rel installer.ReleaseInstaller) *Updater {
	return &Updater{
		client:       cli,
		relInstaller: rel,
	}
}

func (up *Updater) Update(ctx context.Context, owner, repo string) error {
	installDir := utils.GetInstallDir(owner, repo)
	man, err := manifest.Read(installDir)

	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("package %s/%s does not exist", owner, repo)
		}
		return fmt.Errorf("could not read manifest for %s/%s: %w", owner, repo, err)
	}

	switch man.InstallType {
	case manifest.Release, manifest.PreRelease:
		fmt.Printf("Checking for updates for %s/%s...\n", owner, repo)
		rel, _, err := up.client.GetLatestRelease(ctx, owner, repo)

		newVer := rel.GetTagName()

		needsUpdate, err := CheckUpdate(man.Version, newVer)
		if !needsUpdate {
			fmt.Printf("%s/%s is already up to date (ver %s).\n", owner, repo, man.Version)
			return nil
		}

		if err != nil {
			return fmt.Errorf("could not fetch latest release for %s/%s: %w", owner, repo, err)
		}

		if newVer == man.Version {
			fmt.Printf("%s/%s@%s is already up to date.\n", owner, repo, man.Version)
			return nil
		}

		fmt.Printf("Updates found!\n")
		fmt.Printf("Updating %s/%s from %s to %s...\n", owner, repo, man.Version, rel.GetTagName())
		opts := installer.InstallOptions{
			Type:    man.InstallType,
			Version: man.Version,
		}
		return up.updateRelease(ctx, installDir, man, rel, opts)
	}
	return nil
}

func CheckUpdate(currVer, latestVer string) (bool, error) {
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

func (up *Updater) updateRelease(ctx context.Context,
	installDir string,
	man *manifest.Manifest,
	rel *github.RepositoryRelease,
	opts installer.InstallOptions) error {
	owner := man.Owner
	repo := man.Repo

	if man.Version == rel.GetTagName() {
		fmt.Printf("%s/%s@%s is already up to date.\n", owner, repo, man.Version)
		return nil
	}
	fmt.Printf("Updating %s/%s from %s to %s...\n", owner, repo, man.Version, rel.GetTagName())

	return up.relInstaller.InstallFromRelease(ctx, installDir, owner, repo, rel, opts)
}
