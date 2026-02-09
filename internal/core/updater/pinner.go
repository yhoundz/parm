package updater

import (
	"parm/internal/manifest"
	"parm/internal/parmutil"
)

func ChangePinnedStatus(owner, repo string, value bool) (err error, ver string) {
	installDir := parmutil.GetInstallDir(owner, repo)
	man, err := manifest.Read(installDir)
	if err != nil {
		return err, ""
	}

	if man.Pinned == value {
		return nil, man.Version
	}

	man.Pinned = value
	err = man.Write(installDir)
	if err != nil {
		return err, man.Version
	}

	return nil, man.Version
}
