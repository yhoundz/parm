package installer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"parm/internal/cmdparser"
	gh "parm/internal/github"
	"parm/internal/utils"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/google/go-github/v74/github"
)

type Installer struct {
	client *github.RepositoriesService
}

type InstallOptions struct {
	Branch     string
	Commit     string
	Release    string
	Source     bool
	PreRelease bool
}

func New(cli *github.RepositoriesService) *Installer {
	return &Installer{
		client: cli,
	}
}

func (in *Installer) Install(ctx context.Context, pkgPath, owner, repo string, opts InstallOptions) error {
	// WARNING: using --branch or --commit will automatically install from source
	if opts.Branch != "" {
		valid, _, err := gh.ValidateBranch(ctx, in.client, owner, repo, opts.Branch)
		if err != nil {
			fmt.Printf("ERROR: cannot resolve branch: %q on %s/%s", opts.Branch, owner, repo)
			return err
		}
		if !valid {
			return fmt.Errorf("Error: branch: %s cannot be found", opts.Branch)
		}

		cloneLink, _ := cmdparser.BuildGitLink(owner, repo)
		cmd := exec.CommandContext(ctx, "git", "clone",
			"--depth=1", "--recurse-submodules", "--shallow-submodules", "--branch",
			opts.Branch, cloneLink, pkgPath)

		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			if eerr, ok := err.(*exec.ExitError); ok {
				fmt.Printf("git exited with %d\n", eerr.ExitCode())
				return eerr
			} else {
				fmt.Printf("failed to start or was killed: %v\n", err)
			}
			return err
		}

		man, err := NewManifest(owner, repo, opts.Branch, Branch, pkgPath)
		if err != nil {
			return fmt.Errorf("error: failed to create  manifest: %w", err)
		}
		if err := man.WriteManifest(pkgPath); err != nil {
			return fmt.Errorf("error: failed to write manifest: %w", err)
		}
		return nil
	} else if opts.Commit != "" {
		// TODO: testing
		valid, _, err := gh.ValidateCommit(ctx, in.client, owner, repo, opts.Commit)
		if err != nil {
			return fmt.Errorf("ERROR: cannot resolve commit: %q on %s/%s", opts.Commit, owner, repo)
		}
		if !valid {
			return fmt.Errorf("ERROR: commit %q is not valid on %s/%s", opts.Commit, owner, repo)
		}

		cloneLink, _ := cmdparser.BuildGitLink(owner, repo)

		var execGitCmd = func(arg ...string) error {
			cmd := exec.CommandContext(ctx, "git", arg...)
			cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
			if err := cmd.Run(); err != nil {
				return err
			}
			return nil
		}

		// clone
		err = execGitCmd("clone",
			"--no-checkout", "--filter=blob:none",
			"--recurse-submodules", "--shallow-submodules",
			cloneLink, pkgPath)
		if err != nil {
			return err
		}

		// fetch commit
		err = execGitCmd("-C", pkgPath, "fetch", "--depth=1", "origin", opts.Commit)
		if err != nil {
			return err
		}

		// checkout commit + subms
		err = execGitCmd("-C", pkgPath, "checkout", "--recurse-submodules", opts.Commit)
		if err != nil {
			return err
		}

		// ensure submodules materialize shallowly
		err = execGitCmd("-C", pkgPath, "submodule", "update", "--init", "--depth=1", "--recursive")
		if err != nil {
			return err
		}

		man, err := NewManifest(owner, repo, opts.Commit, Commit, pkgPath)
		if err != nil {
			return fmt.Errorf("error: failed to create manifest: %w", err)
		}
		if err = man.WriteManifest(pkgPath); err != nil {
			return fmt.Errorf("error: failed to write manifest: %w", err)
		}

		return nil
	}

	if opts.Release != "" {
		// TODO: redo this part so i actually understand what's going on
		// if source build, download the source code from tarball
		// if not source, find best-matching binary based on GOOS and GOARCH, and then
		// get the download link
		// afterwards, download and extract the tarball to the desired dir.

		// TODO: cleanup if something in the install process goes wrong?
		valid, rel, err := gh.ValidateRelease(ctx, in.client, owner, repo, opts.Release)
		if err != nil {
			return fmt.Errorf("ERROR: Cannot resolve release %s on %s/%s", opts.Release, owner, repo)
		}
		if !valid {
			return fmt.Errorf("ERROR: Release %s not valid, %w", opts.Release, err)
		}

		return in.InstallFromRelease(ctx, pkgPath, owner, repo, rel, opts)

	} else if opts.PreRelease {
		valid, rel, err := gh.ValidatePreRelease(ctx, in.client, owner, repo)
		if err != nil {
			return fmt.Errorf("err: cannot resolve pre-release %s on %s/%s: %w", rel.GetTagName(), owner, repo, err)
		}
		if !valid {
			return fmt.Errorf("error: no valid pre-release found for %s/%s", owner, repo)
		}

		return in.InstallFromRelease(ctx, pkgPath, owner, repo, rel, opts)
	} else {
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

	// none of the options what now?
	// TODO: something happens here? throw an error?
	// return nil
}

func (in *Installer) InstallFromRelease(ctx context.Context, pkgPath, owner, repo string, rel *github.RepositoryRelease, opts InstallOptions) error {
	if opts.Source {
		ref := "tags/" + rel.GetTagName()
		dl, _, err := in.client.GetArchiveLink(
			ctx,
			owner, repo,
			github.Tarball,
			&github.RepositoryContentGetOptions{Ref: ref},
			0,
		)
		if err != nil {
			return fmt.Errorf("ERROR: cannot resolve source for %s/%s, with %w ", owner, repo, err)
		}
		dest := filepath.Join(pkgPath, fmt.Sprintf("%s-%s.tar.gz", repo, rel.GetTagName()))
		if err := downloadTo(ctx, dest, dl.String()); err != nil {
			return err
		}
		if err := utils.ExtractTarGz(dest, pkgPath); err != nil {
			return err
		}
		os.Remove(dest)

		man, err := NewManifest(owner, repo, rel.GetTagName(), Source, pkgPath)
		if err != nil {
			return fmt.Errorf("error: failed to create manifest: %w", err)
		}
		if err := man.WriteManifest(pkgPath); err != nil {
			return fmt.Errorf("error: failed to write manifest: %w", err)
		}
		return nil
	}
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

	ass := matches[0]
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
		// if err := utils.MoveAllFrom(dest, pkgPath); err != nil {
		// 	return err
		// }
	}

	man, err := NewManifest(owner, repo, rel.GetTagName(), Release, pkgPath)
	if err != nil {
		return fmt.Errorf("error: failed to create manifest: %w", err)
	}
	if err := man.WriteManifest(pkgPath); err != nil {
		return fmt.Errorf("error: failed to write manifest: %w", err)
	}

	return nil
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
