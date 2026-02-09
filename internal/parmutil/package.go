package parmutil

import (
	"fmt"
	"os"
	"parm/internal/config"
	"path/filepath"
	"strings"
)

const STAGING_DIR_PREFIX string = ".staging-"

func MakeInstallDir(owner, repo string, perm os.FileMode) (string, error) {
	path := GetInstallDir(owner, repo)
	err := os.MkdirAll(path, perm)
	if err != nil {
		return "", fmt.Errorf("cannot create install dir: \n%w", err)
	}
	return path, nil
}

// Generates install directory for a package. Does not guarantee that the directory actually exists.
func GetInstallDir(owner, repo string) string {
	installPath := config.Cfg.ParmPkgPath
	dest := filepath.Join(installPath, owner, repo)
	return dest
}

func GetBinDir(repoName string) string {
	binPath := config.Cfg.ParmBinPath
	dest := filepath.Join(binPath, repoName)
	return dest
}

func MakeStagingDir(owner, repo string) (string, error) {
	parentDir := filepath.Join(config.Cfg.ParmPkgPath, owner)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return "", err
	}
	var tmpDir string
	var err error
	if tmpDir, err = os.MkdirTemp(parentDir, STAGING_DIR_PREFIX+repo+"-"); err != nil {
		return "", err
	}
	return tmpDir, nil
}

func PromoteStagingDir(final, staging string) (string, error) {
	if fi, err := os.Stat(final); err == nil && fi.IsDir() {
		if err := os.RemoveAll(final); err != nil {
			return "", fmt.Errorf("failed removing existing install: %w", err)
		}
	}
	if err := os.Rename(staging, final); err != nil {
		return "", fmt.Errorf("staging promotion failed: %w", err)
	}
	return final, nil
}

// dir should be the owner dir, at $PKG_ROOT/owner/
// TODO: fix this mess
func Cleanup(dir string) error {
	if dir == "" {
		// fail silently
		return nil
	}

	var fi os.FileInfo
	var err error
	if fi, err = os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a dir", dir)
	}
	dr, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, item := range dr {
		if strings.HasPrefix(item.Name(), STAGING_DIR_PREFIX) {
			path := filepath.Join(dir, item.Name())
			_ = os.RemoveAll(path)
		}
	}

	dr, _ = os.ReadDir(dir)
	if len(dr) == 0 {
		_ = os.Remove(dir)
	}

	return nil
}
