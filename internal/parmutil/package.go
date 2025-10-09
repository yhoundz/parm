package parmutil

import (
	"fmt"
	"os"
	"parm/internal/config"
	"path/filepath"
)

func MakeInstallDir(owner, repo string, perm os.FileMode) (string, error) {
	path := GetInstallDir(owner, repo)
	err := os.MkdirAll(path, perm)
	if err != nil {
		return "", fmt.Errorf("error: cannot create install dir: \n%w", err)
	}
	return path, nil
}

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
