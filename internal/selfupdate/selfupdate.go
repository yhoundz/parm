package selfupdate

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"parm/pkg/sysutil"

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

	asset, err := selectAsset(rel.Assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

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
			if err := extractZip(archivePath, tmpDir); err != nil {
				return err
			}
		} else {
			if err := extractTarGz(archivePath, tmpDir); err != nil {
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

func selectAsset(assets []*github.ReleaseAsset, goos, goarch string) (*github.ReleaseAsset, error) {
	gooses := map[string][]string{
		"windows": {"windows", "win64", "win32", "win"},
		"darwin":  {"macos", "darwin", "mac", "osx"},
		"linux":   {"linux"},
	}
	goarchs := map[string][]string{
		"amd64": {"amd64", "x86_64", "x64", "64bit", "64-bit"},
		"386":   {"386", "x86", "i386", "32bit", "32-bit"},
		"arm64": {"arm64", "aarch64"},
		"arm":   {"armv7", "armv6", "armhf", "armv7l"},
	}

	var bestMatch *github.ReleaseAsset
	maxScore := -1

	for _, a := range assets {
		name := strings.ToLower(a.GetName())
		score := 0

		// Check OS
		osMatch := false
		for _, token := range gooses[goos] {
			if strings.Contains(name, token) {
				osMatch = true
				break
			}
		}
		if !osMatch {
			continue
		}
		score += 10

		// Check Arch
		archMatch := false
		for _, token := range goarchs[goarch] {
			if strings.Contains(name, token) {
				archMatch = true
				break
			}
		}
		if !archMatch {
			continue
		}
		score += 5

		if score > maxScore {
			maxScore = score
			bestMatch = a
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no compatible asset found for %s/%s", goos, goarch)
	}

	return bestMatch, nil
}

func extractTarGz(srcPath, destPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target, err := sysutil.SafeJoin(destPath, hdr.Name)
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			mode := os.FileMode(hdr.Mode) & 0o777
			if mode == 0 {
				mode = 0o644
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractZip(srcPath, destPath string) error {
	r, err := zip.OpenReader(srcPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if err := func() error {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			fpath, err := sysutil.SafeJoin(destPath, f.Name)
			if err != nil {
				return err
			}

			if f.FileInfo().IsDir() {
				return os.MkdirAll(fpath, 0o755)
			}

			if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
				return err
			}
			mode := f.Mode() & 0o777
			if mode == 0 {
				mode = 0o644
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, rc)
			return err
		}(); err != nil {
			return err
		}
	}
	return nil
}
