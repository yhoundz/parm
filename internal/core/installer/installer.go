package installer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"parm/internal/core/uninstaller"
	"parm/internal/manifest"
	"parm/internal/parmutil"
	"parm/pkg/progress"
	"path/filepath"

	"github.com/google/go-github/v74/github"
)

// TODO: symlink binaries to PATH
// TODO: write tests/setup docker
// TODO: update manifest when updating packages

type Installer struct {
	client *github.RepositoriesService
}

type InstallOptions struct {
	Type    manifest.InstallType
	Version string
	Asset   string
}

func New(cli *github.RepositoriesService) *Installer {
	return &Installer{
		client: cli,
	}
}

func (in *Installer) Install(ctx context.Context, owner, repo string, opts InstallOptions, hooks *progress.Hooks) error {
	dest := parmutil.GetInstallDir(owner, repo)
	_, err := os.Stat(dest)

	// if error is something else, ignore it for now and hope it propogates downwards if it's actually serious
	if err == nil {
		if err := uninstaller.Uninstall(ctx, owner, repo); err != nil {
			return err
		}
	}

	dest, err = parmutil.MakeInstallDir(owner, repo, 0o755)
	if err != nil {
		return err
	}
	return in.installFromReleaseByType(ctx, dest, owner, repo, opts, hooks)
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
