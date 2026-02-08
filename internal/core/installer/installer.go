package installer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"parm/internal/core/uninstaller"
	"parm/internal/gh"
	"parm/internal/manifest"
	"parm/pkg/progress"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v74/github"
)

type Installer struct {
	client *github.RepositoriesService
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

func New(cli *github.RepositoriesService) *Installer {
	return &Installer{
		client: cli,
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
		// e.g. if the user runs "parm install alxrw/parm-e2e --release v1.0.1-beta"
		// correct the release channel to use pre-release instead to match user intent
		if rel.GetPrerelease() {
			opts.Type = manifest.PreRelease
		}
	}

	return in.installFromRelease(ctx, installPath, owner, repo, rel, opts, hooks)
}

func downloadTo(ctx context.Context, destPath, url string, hooks *progress.Hooks) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}

	err = os.MkdirAll(filepath.Dir(destPath), 0o755)
	if err != nil {
		return err
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if hooks == nil {
		_, err = io.Copy(file, resp.Body)
		return err
	}

	var r io.Reader = resp.Body
	var closer io.Closer

	// maybe move hooks out of here?
	if hooks.Decorator != nil {
		wr := hooks.Decorator(progress.StageDownload, resp.Body, resp.ContentLength)
		if rc, ok := wr.(io.ReadCloser); ok {
			r, closer = rc, rc
		} else {
			r = wr
		}
	}

	_, err = io.Copy(file, r)
	if closer != nil {
		_ = closer.Close()
	}

	return err
}
