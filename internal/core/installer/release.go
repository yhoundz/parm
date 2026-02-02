package installer

import (
	"context"
	"fmt"
	"os"
	"parm/internal/core/verify"
	"parm/internal/parmutil"
	"parm/internal/release"
	"parm/pkg/progress"
	"path/filepath"
	"runtime"

	"github.com/google/go-github/v74/github"
)

// Does NOT validate the release.
func (in *Installer) installFromRelease(ctx context.Context, pkgPath, owner, repo string, rel *github.RepositoryRelease, opts InstallFlags, hooks *progress.Hooks) (*InstallResult, error) {
	var ass *github.ReleaseAsset
	var err error
	if opts.Asset == nil {
		matches, err := release.SelectAsset(rel.Assets, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("err: no compatible binary found for release %s", rel.GetTagName())
		}
		ass = matches[0]
	} else {
		ass, err = getAssetByName(rel, *opts.Asset)
		if err != nil {
			return nil, err
		}
	}

	tmpDir, err := parmutil.MakeStagingDir(owner, repo)
	if err != nil {
		return nil, err
	}
	// TODO: Cleanup() instead
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, ass.GetName()) // download destination

	// Default to the browser_download_url, which is intended for direct browser access.
	// When using an auth token (e.g. for private repos), switch to the API asset URL
	// because browser_download_url ignores Authorization headers while the API
	// endpoint (ass.GetURL()) requires and honors them.
	downloadURL := ass.GetBrowserDownloadURL()
	token := in.token
	if token != "" {
		// For private repos, use the API endpoint instead of browser download URL
		downloadURL = ass.GetURL()
	}

	if err := release.DownloadAsset(ctx, nil, downloadURL, archivePath, token, hooks); err != nil {
		return nil, fmt.Errorf("error: failed to download asset: \n%w", err)
	}

	// TODO: change based on actual verify-level
	if opts.VerifyLevel > 0 {
		if ass.Digest == nil {
			return nil, fmt.Errorf("error: no upstream digest available for %q; re-run with --no-verify", ass.GetName())
		}
		ok, gen, err := verify.VerifyLevel1(archivePath, *ass.Digest)
		if err != nil {
			return nil, fmt.Errorf("error: could not verify checksum:\n%q", err)
		}
		if !ok {
			return nil, fmt.Errorf("fatal: checksum invalid:\n\thad %s\n\twanted %s", *gen, *ass.Digest)
		}
	}

	switch {
	case release.IsArchive(archivePath):
		if err := release.ExtractArchive(archivePath, tmpDir); err != nil {
			return nil, fmt.Errorf("error: failed to extract archive: \n%w", err)
		}
	default:
		if runtime.GOOS != "windows" {
			if err := os.Chmod(archivePath, 0o755); err != nil {
				return nil, fmt.Errorf("failed to make binary executable: \n%w", err)
			}
		}
	}

	// TODO: create manifest elsewhere for better separation of concerns?
	// TODO: Return an InstallResult and let the CLI call a manifest writer service.
	// will also help with symlinking

	finalDir, err := parmutil.PromoteStagingDir(pkgPath, tmpDir)
	if err != nil {
		return nil, err
	}

	return &InstallResult{
		InstallPath: finalDir,
		Version:     rel.GetTagName(),
	}, nil
}

// gets release asset by name
func getAssetByName(rel *github.RepositoryRelease, name string) (*github.ReleaseAsset, error) {
	for _, ass := range rel.Assets {
		if *ass.Name == name {
			return ass, nil
		}
	}
	return nil, fmt.Errorf("error: no asset by the name of %s was found in release %s", name, rel)
}
