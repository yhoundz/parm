package updater

import (
	"parm/internal/manifest"
	"parm/internal/parmutil"
)

func ChangePinnedStatus(owner, repo string, value bool) (ver string, err error) {
	installDir := parmutil.GetInstallDir(owner, repo)
	man, err := manifest.Read(installDir)
	if err != nil {
		return "", err
	}

	if man.Pinned == value {
		return man.Version, nil
	}

	man.Pinned = value
	err = man.Write(installDir)
	if err != nil {
		return man.Version, err
	}

	return man.Version, nil
}
