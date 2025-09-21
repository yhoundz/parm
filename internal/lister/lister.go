package lister

import (
	"fmt"
	"os"
	"parm/internal/manifest"
	"path/filepath"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type PkgListData struct {
	NumPkgs int
}

func GetInstalledPkgInfo() ([]string, PkgListData, error) {
	mans, err := GetAllPkgManifest()
	var data PkgListData
	if err != nil {
		return nil, data, err
	}
	var infos []string
	for _, man := range mans {
		str := fmt.Sprintf("%s/%s || ver. %s", man.Owner, man.Repo, man.Version)
		infos = append(infos, str)
	}

	data.NumPkgs = len(infos)
	return infos, data, nil
}

func GetAllPkgManifest() ([]*manifest.Manifest, error) {
	pkgDir := viper.GetViper().Get("parm_pkg_dir_path")
	if pkgDir == nil {
		return nil, fmt.Errorf("error: parm_pkg_dir_path could not be found")
	}
	pkgDirPath := cast.ToString(pkgDir)

	var mans []*manifest.Manifest
	entries, err := os.ReadDir(pkgDirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range entries {
		if !file.IsDir() {
			continue
		}
		fullFilePath := filepath.Join(pkgDirPath, file.Name())
		man, err := manifest.Read(fullFilePath)
		if err != nil {
			// cannot find manifest, assume it's not an installation folder and continue
			continue
		}
		mans = append(mans, man)
	}

	return mans, nil
}
