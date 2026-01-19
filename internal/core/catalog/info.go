package catalog

import (
	"context"
	"fmt"
	"os"
	"parm/internal/gh"
	"parm/internal/manifest"
	"parm/internal/parmutil"
	"strings"
	"time"

	"github.com/google/go-github/v74/github"
)

type Info struct {
	Owner       string
	Repo        string
	Version     string
	LastUpdated string
	*DownstreamInfo
	*UpstreamInfo
}

type DownstreamInfo struct {
	InstallPath string
}

type UpstreamInfo struct {
	Stars       int
	License     string
	Description string
}

func (info *Info) String() string {
	var out []string
	out = append(out, fmt.Sprintf("Owner: %s", info.Owner))
	out = append(out, fmt.Sprintf("Repo: %s", info.Repo))

	version := info.Version
	if version == "" {
		version = "Not found / No releases"
	}
	out = append(out, fmt.Sprintf("Version: %s", version))

	if info.LastUpdated != "" {
		out = append(out, fmt.Sprintf("LastUpdated: %s", info.LastUpdated))
	}

	if info.DownstreamInfo != nil {
		out = append(out, info.DownstreamInfo.string())
	} else if info.UpstreamInfo != nil {
		out = append(out, info.UpstreamInfo.string())
	} else {
		out = append(out, "Status: Not installed locally")
	}
	return strings.Join(out, "\n")
}

func (info *DownstreamInfo) string() string {
	var out []string
	out = append(out, fmt.Sprintf("InstallPath: %s", info.InstallPath))
	return strings.Join(out, "\n")
}

func (info *UpstreamInfo) string() string {
	var out []string
	out = append(out, fmt.Sprintf("Stars: %d", info.Stars))
	out = append(out, fmt.Sprintf("License: %s", info.License))
	out = append(out, fmt.Sprintf("Description: %s", info.Description))
	return strings.Join(out, "\n")
}

func GetPackageInfo(ctx context.Context, client *github.RepositoriesService, owner, repo string, isUpstream bool) (Info, error) {
	info := Info{
		Owner: owner,
		Repo:  repo,
	}

	if isUpstream {
		// Resolve release first to avoid unnecessary metadata fetch if no releases exist
		rel, err := gh.ResolveReleaseByTag(ctx, client, owner, repo, nil)
		if err != nil {
			return info, err
		}

		info.Version = rel.GetTagName()
		info.LastUpdated = rel.GetPublishedAt().Format(time.DateTime)

		// Now fetch repository metadata safely
		gitRepo, _, err := client.Get(ctx, owner, repo)
		if err != nil {
			// If we found a release but can't get repo info (rare), we still have the version
			return info, nil
		}

		upInfo := UpstreamInfo{
			Stars:       gitRepo.GetStargazersCount(),
			License:     gitRepo.GetLicense().GetName(),
			Description: gitRepo.GetDescription(),
		}
		info.UpstreamInfo = &upInfo
		info.DownstreamInfo = nil
		return info, nil
	}

	pkgPath := parmutil.GetInstallDir(owner, repo)
	_, err := os.Stat(pkgPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Not installed locally, return info with no downstream/upstream
			return info, nil
		}
		return info, fmt.Errorf("error: couldn't access %s: %w", pkgPath, err)
	}
	man, err := manifest.Read(pkgPath)
	if err != nil {
		return info, err
	}
	info.Version = man.Version
	info.LastUpdated = man.LastUpdated

	downInfo := DownstreamInfo{
		InstallPath: pkgPath,
	}
	info.DownstreamInfo = &downInfo
	info.UpstreamInfo = nil

	return info, nil
}
