package installer

import (
	"context"
	"fmt"
	"os"
	"parm/internal/utils"
)

func (in *Installer) Update(ctx context.Context, owner, repo string) error {
	installDir := utils.GetInstallDir(owner, repo)
	man, err := ReadManifest(installDir)

	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("package %s/%s does not exist", owner, repo)
		}
		return fmt.Errorf("could not read manifest for %s/%s: %w", owner, repo, err)
	}

	switch man.InstallType {
	case Release:
		fmt.Printf("Checking for updates for %s/%s...\n", owner, repo)
		rel, _, err := in.client.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			return fmt.Errorf("could not fetch latest release for %s/%s: %w", owner, repo, err)
		}

		if man.Version == rel.GetTagName() {
			fmt.Printf("%s/%s@%s is already up to date.", owner, repo, man.Version)
			return nil
		}

		fmt.Printf("Updates found!")
		fmt.Printf("Updating %s/%s from %s to %s...\n", owner, repo, man.Version, rel.GetTagName())
		opts := InstallOptions{}
		return in.Install(ctx, installDir, owner, repo, opts)
	case PreRelease:

	default:
	}
	return nil
}
