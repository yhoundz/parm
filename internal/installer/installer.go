package installer

import "parm/internal/deps"

type InstallOptions struct {
	Branch  string
	Commit  string
	Release string
	Source  bool
}

func Install(pkgPath, owner, repo, release string) error {
	if err := deps.Require("tar"); err != nil {
		return err
	}
	return nil
}
