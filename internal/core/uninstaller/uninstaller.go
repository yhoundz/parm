package uninstaller

import (
	"context"
	"fmt"
	"os"
	"parm/internal/manifest"
	"parm/internal/parmutil"
	"parm/pkg/sysutil"
	"path/filepath"
)

// TODO: when version management is added?, have an option to remove a specific version
func Uninstall(ctx context.Context, owner, repo string) error {
	dir := parmutil.GetInstallDir(owner, repo)
	fi, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("error: dir does not exist: \n%w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("error: selected item is not a dir: \n%w", err)
	}

	manifest, err := manifest.Read(dir)
	if err != nil {
		return fmt.Errorf("error: could not read manifest: \n%w", err)
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

	for _, path := range execPaths {
		isRunning, err := sysutil.IsProcessRunning(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not check for process: %s: %v\n", path, err)
			continue
		}
		if isRunning {
			return fmt.Errorf("error: cannot uninstall process %s because it is currently running", filepath.Base(path))
		}
	}

	if err = os.RemoveAll(dir); err != nil {
		return fmt.Errorf("error: Cannot remove dir: %s: \n%w", dir, err)
	}

	parentDir, err := sysutil.GetParentDir(dir)
	// NOTE: don't want to error out here if it fails
	if err != nil {
		return nil
	}

	entries, err := os.ReadDir(parentDir)
	if err == nil && len(entries) == 0 {
		_ = os.Remove(parentDir)
		return nil
	}

	return nil
}
