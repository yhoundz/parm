package installer

import (
	"context"
	"fmt"
	"math"
	"os"
	"parm/internal/core/verify"
	"parm/internal/parmutil"
	"parm/pkg/archive"
	"parm/pkg/progress"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/google/go-github/v74/github"
)

// Does NOT validate the release.
func (in *Installer) installFromRelease(ctx context.Context, pkgPath, owner, repo string, rel *github.RepositoryRelease, opts InstallFlags, hooks *progress.Hooks) (*InstallResult, error) {
	var ass *github.ReleaseAsset
	var err error
	if opts.Asset == nil {
		matches, err := selectReleaseAsset(rel.Assets, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			// TODO: allow users to choose match
			return nil, fmt.Errorf("err: no compatible binary found for release %s", rel.GetTagName())
		}
		// if len(matches) > 1 {
		// 	// TODO: allow users to choose what asset they want installed instead
		// 	return nil
		// }
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
	if err := downloadTo(ctx, archivePath, ass.GetBrowserDownloadURL(), hooks); err != nil {
		return nil, fmt.Errorf("failed to download asset: \n%w", err)
	}

	// TODO: change based on actual verify-level
	if opts.VerifyLevel > 0 {
		if ass.Digest == nil {
			return nil, fmt.Errorf("no upstream digest available for %q; re-run with --no-verify", ass.GetName())
		}
		ok, gen, err := verify.VerifyLevel1(archivePath, *ass.Digest)
		if err != nil {
			return nil, fmt.Errorf("could not verify checksum:\n%q", err)
		}
		if !ok {
			return nil, fmt.Errorf("fatal: checksum invalid:\n\thad %s\n\twanted %s", *gen, *ass.Digest)
		}
	}

	switch {
	case strings.HasSuffix(archivePath, ".tar.gz"), strings.HasSuffix(archivePath, ".tgz"):
		if err := archive.ExtractTarGz(archivePath, tmpDir); err != nil {
			return nil, fmt.Errorf("failed to extract tarball: \n%w", err)
		}
	case strings.HasSuffix(archivePath, ".zip"):
		if err := archive.ExtractZip(archivePath, tmpDir); err != nil {
			return nil, fmt.Errorf("failed to extract zip: \n%w", err)
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
	return nil, fmt.Errorf("no asset by the name of %s was found in release %s", name, rel)
}

// infers the proper release asset based on the name of the asset
func selectReleaseAsset(assets []*github.ReleaseAsset, goos, goarch string) ([]*github.ReleaseAsset, error) {
	type match struct {
		asset *github.ReleaseAsset
		score int
	}

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

	extPref := []string{".tar.gz", ".tgz", ".tar.xz", ".zip", ".bin", ".AppImage"}
	if goos == "windows" {
		extPref = []string{".zip", ".exe", ".msi", ".bin"}
	}

	// other score modifiers
	scoreMods := map[string]int{
		"musl": -1,
	}
	if goos == "windows" {
		scoreMods = map[string]int{}
	}

	// scoring
	scoredMatches := make([]match, len(assets))
	for i, a := range assets {
		scoredMatches[i] = match{asset: a, score: 0}
	}

	const goosMatch = 11
	const goarchMatch = 7
	const prefMatch = 3 // actually a multiplier for preference match
	// INFO: to be used later when adding interactive install.
	// const minScoreMatch = goosMatch + goarchMatch

	for i := range scoredMatches {
		a := &scoredMatches[i]
		name := a.asset.GetName()
		if containsAny(name, gooses[goos]) {
			a.score += goosMatch
		}
		if containsAny(name, goarchs[goarch]) {
			a.score += goarchMatch
		}

		for j, ext := range extPref {
			var mult = float64(prefMatch) * float64((len(extPref) - j))
			var multRounded = int(math.Round(mult))
			if strings.HasSuffix(name, ext) {
				a.score += multRounded
			}
		}

		for j, m := range scoreMods {
			if strings.Contains(name, j) {
				a.score += m
			}
		}
	}

	// sort
	slices.SortStableFunc(scoredMatches, func(a, b match) int {
		if a.score < b.score {
			return 1
		}
		if a.score > b.score {
			return -1
		}
		return 0
	})

	minMatch := scoredMatches[0].score
	// if minMatch < minScoreMatch {
	// 	fmt.Println("warning: selected release asset may not be completely accurate")
	// }

	// find top candidate(s)
	var candidates []*github.ReleaseAsset
	for _, m := range scoredMatches {
		if m.score == minMatch {
			candidates = append(candidates, m.asset)
			continue
		}
		break
	}

	return candidates, nil
}

func containsAny(src string, tokens []string) bool {
	for _, a := range tokens {
		if strings.Contains(src, a) {
			return true
		}
	}
	return false
}
