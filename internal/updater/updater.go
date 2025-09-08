package updater

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"parm/internal/deps"
	"parm/internal/installer"
	"parm/internal/manifest"
	"parm/internal/utils"
	"strings"
	"time"

	"github.com/google/go-github/v74/github"
)

type Updater struct {
	client       *github.RepositoriesService
	relInstaller installer.ReleaseInstaller
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
		if err != nil {
			return fmt.Errorf("could not fetch latest release for %s/%s: %w", owner, repo, err)
		}

		if man.Version == rel.GetTagName() {
			fmt.Printf("%s/%s@%s is already up to date.", owner, repo, man.Version)
			return nil
		}

		fmt.Printf("Updates found!")
		fmt.Printf("Updating %s/%s from %s to %s...\n", owner, repo, man.Version, rel.GetTagName())
		opts := installer.InstallOptions{
			Type:    man.InstallType,
			Version: man.Version,
			Source:  man.IsSource,
		}
		return up.updateRelease(ctx, installDir, man, rel, opts)
	case manifest.Branch:
		if err := deps.Require("git"); err != nil {
			return err
		}

		fmt.Printf("Checking for updates for %s/%s on %s", owner, repo, man.Version)
		oldSHA, err := getGitCommit(ctx, installDir)
		if err != nil {
			return fmt.Errorf("could not retrieve current commit for %s/%s: %w", owner, repo, err)
		}

		fmt.Println("Retrieving latest changes...")
		if err := gitPull(ctx, installDir); err != nil {
			return fmt.Errorf("failed to update for %s/%s: %w", owner, repo, err)
		}

		newSHA, err := getGitCommit(ctx, installDir)
		if err != nil {
			return fmt.Errorf("could not retrieve new commit for %s/%s: %w", owner, repo, err)
		}

		if oldSHA == newSHA {
			fmt.Printf("%s/%s is already up to date.\n", owner, repo)
			return nil
		}

		fmt.Printf("Successfully updated %s/%s", owner, repo)
		man.LastUpdated = time.Now().UTC().Format(time.RFC3339)
		man.Version = newSHA

		if err := man.Write(installDir); err != nil {
			return fmt.Errorf("failed to update manifest for %s/%s: %w", owner, repo, err)
		}

		fmt.Printf("WARNING: You will need to rebuild the application from source.")
		return nil
	case manifest.Commit:
		fmt.Printf("Update failed: commit installs cannot be updated.")
		return nil
	}
	return nil
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

func getGitCommit(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitPull(ctx context.Context, repoPath string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", repoPath, "pull")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pull failed:\n%s", string(out))
	}
	return nil
}
