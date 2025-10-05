package utils

import (
	"fmt"
	"os"
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
	installPath := config.Cfg.ParmPkgDirPath
	dest := filepath.Join(installPath, owner, repo)
	return dest
}
