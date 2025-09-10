package uninstaller

import (
	"context"
	"fmt"
	"os"
	"parm/internal/manifest"
	"parm/internal/utils"
	"path/filepath"
)

// TODO: when version management is added?, have an option to remove a specific version
func Uninstall(ctx context.Context, owner, repo string) error {
	dir := utils.GetInstallDir(owner, repo)
	manifest, err := manifest.Read(dir)
	if err != nil {
		return fmt.Errorf("could not read manifest: %w", err)
	}

	var execPaths []string
	for _, path := range manifest.Executables {
		if filepath.IsAbs(path) {
			execPaths = append(execPaths, path)
		} else {
			fullPath := filepath.Join(dir, path)
			execPaths = append(execPaths, fullPath)
		}
	}

	fi, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("ERROR: dir does not exist: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("ERROR: selected item is not a dir: %w", err)
	}

	for _, path := range execPaths {
		isRunning, err := utils.IsProcessRunning(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not check for process: %s: %v\n", path, err)
			continue
		}
		if isRunning {
			return fmt.Errorf("error: cannot uninstall process %s because it is currently running", filepath.Base(path))
		}
	}

	if err = os.RemoveAll(dir); err != nil {
		return fmt.Errorf("ERROR: Cannot remove dir: %s: %w", dir, err)
	}

	return nil
}
