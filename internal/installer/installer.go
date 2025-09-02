package installer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"parm/internal/utils"
	"path/filepath"
	"slices"
	"strings"

	"github.com/google/go-github/v74/github"
)

type Installer struct {
	client *github.RepositoriesService
}

type InstallOptions struct {
	Type    InstallType
	Version string
	Source  bool
}

func New(cli *github.RepositoriesService) *Installer {
	return &Installer{
		client: cli,
	}
}

func (in *Installer) Install(ctx context.Context, pkgPath, owner, repo string, opts InstallOptions) error {
	switch opts.Type {
	case Branch:
		return in.installFromBranch(ctx, pkgPath, owner, repo, opts.Version)
	case Commit:
		return in.installFromCommit(ctx, pkgPath, owner, repo, opts.Version)
	case Release, PreRelease:
		return in.installFromReleaseByType(ctx, pkgPath, owner, repo, opts)
	default:
		rel, _, err := in.client.GetLatestRelease(ctx, owner, repo)
		if err != nil {
			var ghErr *github.ErrorResponse
			if errors.As(err, &ghErr) && ghErr.Response.StatusCode == http.StatusNotFound {
				return fmt.Errorf("no stable release found for %s/%s", owner, repo)
			}
			return fmt.Errorf("could not fetch latest release: %w", err)
		}

		return in.InstallFromRelease(ctx, pkgPath, owner, repo, rel, opts)
	}
}

func downloadTo(ctx context.Context, destPath, url string) error {
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

	_, err = io.Copy(file, resp.Body)
	return err
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

	// fmt.Print(candidates)
	return candidates, nil
}
