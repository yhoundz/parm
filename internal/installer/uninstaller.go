package installer

import (
	"context"
	"fmt"
	"os"
	"parm/internal/utils"
)

// TODO: when version management is added?, have an option to remove a specific version
func (in *Installer) Uninstall(ctx context.Context, owner, repo string) error {
	dir := utils.GetInstallDir(owner, repo)
	fi, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("ERROR: dir does not exist: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("ERROR: selected dir is not a file: %w", err)
	}
	if err = os.RemoveAll(dir); err != nil {
		return fmt.Errorf("ERROR: Cannot remove dir: %s: %w", dir, err)
	}

	return nil
}
