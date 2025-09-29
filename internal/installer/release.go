package installer

import (
	"context"
	"fmt"
	"math"
	"os"
	gh "parm/internal/github"
	"parm/internal/manifest"
	"parm/internal/utils"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/google/go-github/v74/github"
)

type ReleaseInstaller interface {
	InstallFromRelease(ctx context.Context, pkgPath, owner, repo string, rel *github.RepositoryRelease, opts InstallOptions) error
}

func (in *Installer) installFromReleaseByType(ctx context.Context, pkgPath, owner, repo string, opts InstallOptions) error {
	isPre := opts.Type == manifest.PreRelease
	rel, err := gh.ResolveRelease(ctx, in.client, owner, repo, opts.Version, isPre)
	if err != nil {
		return err
	}

	return in.InstallFromRelease(ctx, pkgPath, owner, repo, rel, opts)
}

// Does NOT validate the release.
func (in *Installer) InstallFromRelease(ctx context.Context, pkgPath, owner, repo string, rel *github.RepositoryRelease, opts InstallOptions) error {
	var ass *github.ReleaseAsset
	var err error
	if opts.Asset == "" {
		matches, err := selectReleaseAsset(rel.Assets, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			return err
		}
		if matches == nil {
			// TODO: allow users to choose what asset they want installed instead
			return fmt.Errorf("err: No install matches found")
		}
		if len(matches) == 0 {
			// TODO: allow users to choose match
			return fmt.Errorf("err: no compatible binary found for release %s", rel.GetTagName())
		}
		// if len(matches) > 1 {
		// 	// TODO: allow users to choose what asset they want installed instead
		// 	return nil
		// }
		ass = matches[0]

	} else {
		ass, err = getAssetByName(rel, opts.Asset)
		if err != nil {
			return err
		}
	}

	dest := filepath.Join(pkgPath, ass.GetName()) // download destination
	if err := downloadTo(ctx, dest, ass.GetBrowserDownloadURL()); err != nil {
		return fmt.Errorf("ERROR: failed to download asset: %w", err)
	}

	switch {
	case strings.HasSuffix(dest, ".tar.gz"), strings.HasSuffix(dest, ".tgz"):
		if err := utils.ExtractTarGz(dest, pkgPath); err != nil {
			return fmt.Errorf("ERROR: failed to extract tarball: %w", err)
		}
		os.Remove(dest)
	case strings.HasSuffix(dest, ".zip"):
		if err := utils.ExtractZip(dest, pkgPath); err != nil {
			return fmt.Errorf("ERROR: failed to extract zip: %w", err)
		}
		os.Remove(dest)
	default:
		if runtime.GOOS != "windows" {
			if err := os.Chmod(dest, 0o755); err != nil {
				return fmt.Errorf("failed to make binary executable: %w", err)
			}
		}
	}

	// TODO: create manifest elsewhere?
	man, err := manifest.New(owner, repo, rel.GetTagName(), opts.Type, pkgPath)
	if err != nil {
		return fmt.Errorf("error: failed to create manifest: %w", err)
	}
	return man.Write(pkgPath)
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
	const minScoreMatch = goosMatch + goarchMatch

	for i := range scoredMatches {
		a := &scoredMatches[i]
		name := a.asset.GetName()
		if utils.ContainsAny(name, gooses[goos]) {
			a.score += goosMatch
		}
		if utils.ContainsAny(name, goarchs[goarch]) {
			a.score += goarchMatch
		}

		for j, ext := range extPref {
			var mult float64 = float64(prefMatch) * float64((len(extPref) - j))
			var multRounded int = int(math.Round(mult))
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
	if minMatch < minScoreMatch {
		return nil, fmt.Errorf("ERROR: Cannot find sufficient matches")
	}

	// find top candidate(s)
	var candidates []*github.ReleaseAsset
	for _, m := range scoredMatches {
		if m.score == minMatch {
			candidates = append(candidates, m.asset)
			continue
		}
		break
	}

	fmt.Print(candidates[0])
	return candidates, nil
}
