package installer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"parm/internal/manifest"
	"parm/pkg/progress"
	"path/filepath"

	"github.com/google/go-github/v74/github"
)

type Installer struct {
	client *github.RepositoriesService
}

type InstallOptions struct {
	Type     manifest.InstallType
	Version  string
	Asset    string
	Progress progress.Callback
}

func New(cli *github.RepositoriesService) *Installer {
	return &Installer{
		client: cli,
	}
}

func (in *Installer) Install(ctx context.Context, pkgPath, owner, repo string, opts InstallOptions) error {
	return in.installFromReleaseByType(ctx, pkgPath, owner, repo, opts)
}

func downloadTo(ctx context.Context, destPath, url string, cb progress.Callback) error {
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

	if cb == nil {
		_, err = io.Copy(file, resp.Body)
		return err
	}

	cb(progress.Event{
		Stage:   progress.StageDownload,
		Current: 0,
		Total:   resp.ContentLength,
	})

	reader := progress.NewReader(resp.Body, resp.ContentLength, progress.StageDownload, cb)
	written, err := io.Copy(file, reader)

	cb(progress.Event{
		Stage:   progress.StageDownload,
		Current: written,
		Total:   resp.ContentLength,
		Done:    true,
	})
	return err
}
