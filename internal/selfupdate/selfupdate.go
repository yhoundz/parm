package selfupdate

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"parm/internal/core/installer"
	"parm/pkg/archive"

	"github.com/google/go-github/v74/github"
	"github.com/minio/selfupdate"
)

type Config struct {
	Owner          string
	Repo           string
	Binary         string
	CurrentVersion string
}

func Update(ctx context.Context, cfg Config, stdout, stderr io.Writer) error {
	client := github.NewClient(nil)
	repos := client.Repositories

	rel, _, err := repos.GetLatestRelease(ctx, cfg.Owner, cfg.Repo)
	if err != nil {
		return fmt.Errorf("could not fetch latest release: %w", err)
	}

	if rel.GetTagName() == cfg.CurrentVersion {
		fmt.Fprintf(stdout, "parm is already up to date (%s)\n", cfg.CurrentVersion)
		return nil
	}

	fmt.Fprintf(stdout, "Updating parm %s -> %s...\n", cfg.CurrentVersion, rel.GetTagName())

	matches, err := installer.SelectReleaseAsset(rel.Assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("no compatible asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	asset := matches[0]

	downloadURL := asset.GetBrowserDownloadURL()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request for asset download: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download asset: status %d", resp.StatusCode)
	}

	var reader io.Reader = resp.Body
	name := asset.GetName()

	// If it's an archive, we need to extract the binary first
	if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz") || strings.HasSuffix(name, ".zip") {
		tmpDir, err := os.MkdirTemp("", "parm-selfupdate-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDir)

		archivePath := filepath.Join(tmpDir, name)
		f, err := os.Create(archivePath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, resp.Body); err != nil {
			f.Close()
			return err
		}
		f.Close()

		if strings.HasSuffix(name, ".zip") {
			if err := archive.ExtractZip(archivePath, tmpDir); err != nil {
				return err
			}
		} else {
			if err := archive.ExtractTarGz(archivePath, tmpDir); err != nil {
				return err
			}
		}

		// Look for the binary in the extracted files
		binaryName := cfg.Binary
		if runtime.GOOS == "windows" {
			binaryName += ".exe"
		}

		binaryPath := ""
		err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.Name() == binaryName {
				binaryPath = path
				return filepath.SkipDir
			}
			return nil
		})
		if err != nil {
			return err
		}

		if binaryPath == "" {
			return fmt.Errorf("binary %s not found in archive", binaryName)
		}

		binFile, err := os.Open(binaryPath)
		if err != nil {
			return err
		}
		defer binFile.Close()
		reader = binFile

		if err := selfupdate.Apply(reader, selfupdate.Options{}); err != nil {
			return fmt.Errorf("failed to apply update: %w", err)
		}
	} else {
		if err := selfupdate.Apply(reader, selfupdate.Options{}); err != nil {
			return fmt.Errorf("failed to apply update: %w", err)
		}
	}

	fmt.Fprintf(stdout, "Successfully updated to %s\n", rel.GetTagName())
	return nil
}
