package installer

import (
	"context"
	"os"
	"parm/internal/core/uninstaller"
	"parm/internal/gh"
	"parm/internal/manifest"
	"parm/pkg/progress"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v74/github"
)

type Installer struct {
	client *github.RepositoriesService
	token  string // GitHub token for private repo access
}

type InstallFlags struct {
	Type        manifest.InstallType
	Version     *string
	Asset       *string
	Strict      bool
	VerifyLevel uint8
}

type InstallResult struct {
	InstallPath string
	Version     string
}

func New(cli *github.RepositoriesService, token string) *Installer {
	return &Installer{
		client: cli,
		token:  token,
	}
}

func (in *Installer) Install(ctx context.Context, owner, repo string, installPath string, opts InstallFlags, hooks *progress.Hooks) (*InstallResult, error) {
	f, _ := os.Stat(installPath)
	var err error

	// if error is something else, ignore it for now and hope it propogates downwards if it's actually serious
	if f != nil {
		if err := uninstaller.Uninstall(ctx, owner, repo); err != nil {
			return nil, err
		}
	}

	var rel *github.RepositoryRelease
	if opts.Type == manifest.PreRelease {
		rel, _ = gh.ResolvePreRelease(ctx, in.client, owner, repo)
		if !opts.Strict {
			// expensive!
			relStable, err := gh.ResolveReleaseByTag(ctx, in.client, owner, repo, nil)
			if err != nil {
				return nil, err
			}

			// TODO: abstract elsewhere cuz it's similar to updater.NeedsUpdate
			currVer, _ := semver.NewVersion(rel.GetTagName())
			newVer, _ := semver.NewVersion(relStable.GetTagName())
			if newVer.GreaterThan(currVer) {
				rel = relStable
			}
		}
	} else {
		rel, err = gh.ResolveReleaseByTag(ctx, in.client, owner, repo, opts.Version)
		if err != nil {
			return nil, err
		}

		// we get to this point if the user installs a pre-release using the --release flag
		// e.g. if the user runs "parm install yhoundz/parm-e2e --release v1.0.1-beta"
		// correct the release channel to use pre-release instead to match user intent
		if rel.GetPrerelease() {
			opts.Type = manifest.PreRelease
		}
	}

	return in.installFromRelease(ctx, installPath, owner, repo, rel, opts, hooks)
}
