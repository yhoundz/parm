package switcher

import (
	"parm/internal/manifest"
	"parm/internal/parmutil"
)

func SwitchChannel(owner, repo string, channel manifest.InstallType) error {
	installDir := parmutil.GetInstallDir(owner, repo)
	man, err := manifest.Read(installDir)
	if err != nil {
		return err
	}
	if man.InstallType == channel {
		return nil
	}

	man.InstallType = channel
	err = man.Write(installDir)
	if err != nil {
		return err
	}
	return nil
}
