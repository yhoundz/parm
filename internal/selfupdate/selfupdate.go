package selfupdate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"parm/internal/core/verify"
	"parm/internal/release"
	"parm/parmver"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v74/github"
	minioSelfUpdate "github.com/minio/selfupdate"
)

var (
	applyFunc = minioSelfUpdate.Apply
)

type Config struct {
	Owner          string
	Repo           string
	Binary         string
	CurrentVersion string
	GitHubClient   *github.Client
	HTTPClient     *http.Client
}

func Run(stdout, stderr io.Writer) error {
	owner := parmver.Owner
	if owner == "" {
		owner = "yhoundz"
	}
	repo := parmver.Repo
	if repo == "" {
		repo = "parm"
	}
	return Update(context.Background(), Config{
		Owner:          owner,
		Repo:           repo,
		Binary:         "parm",
		CurrentVersion: parmver.StringVersion,
	}, stdout, stderr)
}

func Update(ctx context.Context, cfg Config, stdout, stderr io.Writer) error {
	if cfg.Owner == "" || cfg.Repo == "" {
		return fmt.Errorf("owner and repo must be configured")
	}

	out := stdout
	if out == nil {
		out = io.Discard
	}
	errOut := stderr
	if errOut == nil {
		errOut = io.Discard
	}

	ghClient := cfg.GitHubClient
	if ghClient == nil {
		ghClient = github.NewClient(nil)
	}

	repos := ghClient.Repositories
	rel, _, err := repos.GetLatestRelease(ctx, cfg.Owner, cfg.Repo)
	if err != nil {
		return fmt.Errorf("could not fetch latest release: %w", err)
	}

	latestTag := rel.GetTagName()
	if latestTag == "" {
		return errors.New("latest release missing tag name")
	}

	if upToDate(cfg.CurrentVersion, latestTag) {
		fmt.Fprintf(out, "parm is already up to date (%s)\n", cfg.CurrentVersion)
		return nil
	}

	fmt.Fprintf(out, "Updating parm %s -> %s...\n", cfg.CurrentVersion, latestTag)

	matches, err := release.SelectAsset(rel.Assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("no compatible release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}
	asset := matches[0]

	tmpDir, err := os.MkdirTemp("", "parm-selfupdate-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	assetPath := filepath.Join(tmpDir, asset.GetName())
	if err := release.DownloadAsset(ctx, cfg.HTTPClient, asset.GetBrowserDownloadURL(), assetPath, "", nil); err != nil {
		return err
	}

	if err := verifyAssetChecksum(out, errOut, assetPath, asset); err != nil {
		return err
	}

	binary := cfg.Binary
	if binary == "" {
		binary = "parm"
	}
	if runtime.GOOS == "windows" && !strings.HasSuffix(binary, ".exe") {
		binary += ".exe"
	}

	targetPath := assetPath
	if release.IsArchive(asset.GetName()) {
		if err := release.ExtractArchive(assetPath, tmpDir); err != nil {
			return err
		}
		targetPath, err = release.FindBinary(tmpDir, binary)
		if err != nil {
			return err
		}
	}

	f, err := os.Open(targetPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := applyFunc(f, minioSelfUpdate.Options{}); err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	fmt.Fprintf(out, "Successfully updated to %s\n", latestTag)
	return nil
}

func upToDate(current, latest string) bool {
	if current == "" {
		return false
	}

	currentVer, errC := semver.NewVersion(current)
	latestVer, errL := semver.NewVersion(latest)
	if errC == nil && errL == nil {
		return !latestVer.GreaterThan(currentVer)
	}

	return current == latest
}

func verifyAssetChecksum(stdout, stderr io.Writer, path string, asset *github.ReleaseAsset) error {
	if asset == nil || asset.Digest == nil {
		fmt.Fprintln(stderr, "selfupdate: no checksum provided for asset; skipping verification")
		return nil
	}

	ok, generated, err := verify.VerifyLevel1(path, *asset.Digest)
	if err != nil {
		return fmt.Errorf("failed to verify checksum for %q: %w", asset.GetName(), err)
	}
	if !ok {
		return fmt.Errorf("checksum mismatch for %q: had %s, wanted %s", asset.GetName(), *generated, *asset.Digest)
	}

	fmt.Fprintf(stdout, "Verified checksum %s\n", *generated)
	return nil
}
