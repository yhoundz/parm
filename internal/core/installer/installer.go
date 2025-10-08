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
	"parm/internal/parmutil"
	"parm/pkg/progress"
	"path/filepath"

	"github.com/google/go-github/v74/github"
)

// TODO: when installing, check if dir exists before overwriting it
// TODO: if download fails for some reason at any point, remove all traces of partially installed dirs
// TODO: Check for dependencies after installation and bubble them up to the user
// TODO: write tests/setup docker
// TODO: update manifest when updating packages
// TODO: create install scripts: .sh, .ps1, .fish
// TODO: create section on how to add packages to parm in README.md
// TODO: make readme prettier w/ html/css + CI/CD badges
// TODO: restructure README
// TODO: add migrate command if user changes bin or install dir in config

type Installer struct {
	client *github.RepositoriesService
}

type InstallFlags struct {
	Type    manifest.InstallType
	Version string
	Asset   string
}

func New(cli *github.RepositoriesService) *Installer {
	return &Installer{
		client: cli,
	}
}

func (in *Installer) Install(ctx context.Context, owner, repo string, opts InstallFlags, hooks *progress.Hooks) error {
	dest := parmutil.GetInstallDir(owner, repo)
	f, _ := os.Stat(dest)

	// if error is something else, ignore it for now and hope it propogates downwards if it's actually serious
	if f != nil {
		if err := uninstaller.Uninstall(ctx, owner, repo); err != nil {
			return err
		}
	}

	dest, err := parmutil.MakeInstallDir(owner, repo, 0o755)
	if err != nil {
		return err
	}

	// validate opts and correct them
	isPre := opts.Type == manifest.PreRelease
	rel, err := gh.ResolveRelease(ctx, in.client, owner, repo, opts.Version, isPre)
	if err != nil {
		return err
	}
	if rel.GetPrerelease() && !isPre {
		opts.Type = manifest.PreRelease
	}

	return in.InstallFromRelease(ctx, dest, owner, repo, rel, opts, hooks)
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
