package informer

import (
	"context"
	"fmt"
	"os"
	"parm/internal/utils"

	"github.com/google/go-github/v74/github"
)

type Info struct {
	Owner       string
	Repo        string
	Release     string
	LastUpdated string
	downstream  DownstreamInfo
	upstream    UpstreamInfo
}

type DownstreamInfo struct {
	InstallPath string
}

type UpstreamInfo struct {
	Stars       int
	Description string
}

func GetPackageInfo(ctx context.Context, client *github.RepositoriesService, owner, repo string, isUpstream bool) (Info, error) {
	var info Info
	pkgPath := utils.GetInstallDir(owner, repo)
	_, err := os.Stat(pkgPath)
	if err != nil {
		return info, fmt.Errorf("error: there was an error accessing %s:\n%w", pkgPath, err)
	}

	return info, nil
}
