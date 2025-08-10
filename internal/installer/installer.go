package installer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	gh "parm/internal/github"
	"parm/internal/parser"
	"parm/internal/utils"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v74/github"
)

type Installer struct {
	client *github.RepositoriesService
}

type InstallOptions struct {
	Branch  string
	Commit  string
	Release string
	Source  bool
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

		cloneLink, _ := parser.BuildGitLink(owner, repo)
		cmd := exec.CommandContext(ctx, "git", "clone",
			"--depth=1", "--recurse-submodules", "--shallow-submodules", "--branch",
			opts.Branch, cloneLink, pkgPath)

		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			if eerr, ok := err.(*exec.ExitError); ok {
				fmt.Printf("git exited with %d\n", eerr.ExitCode())
			} else {
				fmt.Printf("failed to start or was killed: %v\n", err)
			}
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

		cloneLink, _ := parser.BuildGitLink(owner, repo)
		cmd := exec.CommandContext(ctx, "git", "clone",
			"--no-checkout", "--filter=blob:none",
			"--recurse-submodules", "--shallow-submodules",
			cloneLink, pkgPath)
		cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		if err := cmd.Run(); err != nil {
			return err
		}

		// 2) fetch that specific commit (shallow)
		cmd = exec.CommandContext(ctx, "git", "-C", pkgPath, "fetch", "--depth=1", "origin", opts.Commit)
		cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		if err := cmd.Run(); err != nil {
			return err
		}

		// 3) checkout the commit (and submodules at that ref)
		cmd = exec.CommandContext(ctx, "git", "-C", pkgPath, "checkout", "--recurse-submodules", opts.Commit)
		cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		if err := cmd.Run(); err != nil {
			return err
		}

		// 4) ensure submodules materialize shallowly
		cmd = exec.CommandContext(ctx, "git", "-C", pkgPath, "submodule", "update", "--init", "--depth=1", "--recursive")
		cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		return cmd.Run()
	} else if opts.Release != "" {
		valid, rel, err := gh.ValidateRelease(ctx, in.client, owner, repo, opts.Release)
		if err != nil {
			return fmt.Errorf("ERROR: cannot resolve release: %q on %s/%s", opts.Release, owner, repo)
		}
		if !valid {
			return fmt.Errorf("ERROR: release %q is not valid on %s/%s", opts.Release, owner, repo)
		}

		if opts.Source {
			contentOpts := github.RepositoryContentGetOptions{
				Ref: `tags/` + rel.GetTagName(),
			}
			u, _, err := in.client.GetArchiveLink(ctx, owner, repo, github.Tarball, &contentOpts, 0)
			if err != nil {
				return fmt.Errorf("get archive link: %w", err)
			}
			dest := filepath.Join(pkgPath, fmt.Sprintf("%s-%s.tar.gz", repo, rel.GetTagName()))
			// TODO: create version folder
			if err := downloadTo(ctx, u.String(), dest); err != nil {
				return err
			}
			if err := utils.ExtractTarGz(dest, pkgPath); err != nil {
				return err
			}
			return nil
		}

		asset, err := selectReleaseAsset(rel, runtime.GOOS, runtime.GOARCH)
		if err != nil {
			return err
		}
		dest := filepath.Join(pkgPath, asset.GetName())
		// TODO: create version folder
		if err := downloadTo(ctx, asset.GetBrowserDownloadURL(), dest); err != nil {
			return err
		}

		name := strings.ToLower(asset.GetName())
		if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz") {
			if err := utils.ExtractTarGz(dest, pkgPath); err != nil {
				return err
			}
		}
		return nil
	}

	return nil
}

func selectReleaseAsset(rel *github.RepositoryRelease, goos, goarch string) (*github.ReleaseAsset, error) {
	assets := rel.Assets
	if len(assets) == 0 {
		return nil, fmt.Errorf("no assets found on release %q", rel.GetTagName())
	}

	osTokens := map[string][]string{
		"linux":   {"linux"},
		"darwin":  {"darwin", "macos", "mac", "osx"},
		"windows": {"windows", "win64", "win32", "win"},
	}
	archTokens := map[string][]string{
		"amd64": {"amd64", "x86_64", "x64", "64bit", "64-bit"},
		"386":   {"386", "x86", "i386", "32bit", "32-bit"},
		"arm64": {"arm64", "aarch64"},
		"arm":   {"armv7", "armv6", "armhf", "armv7l"},
	}

	// extensions preference by OS
	extPref := []string{".tar.gz", ".tgz", ".tar.xz", ".zip", ".bin", ".exe", ".AppImage"}
	if goos == "windows" {
		extPref = []string{".zip", ".exe"}
	}

	// TODO: rank this elsewhere?
	// rank each asset; pick the highest score
	best := -1
	var winner *github.ReleaseAsset

	for _, a := range assets {
		name := strings.ToLower(a.GetName())
		if skipAsset(name) {
			continue
		}
		score := 0

		// INFO: +6 if it matches goos AND goarch
		// strong matches: os token + arch token present
		if utils.ContainsAny(name, osTokens[goos]) {
			score += 3
		} else {
			// INFO: MUST match goos
			// weak generic archives still acceptable later
			continue
		}
		if utils.ContainsAny(name, archTokens[goarch]) {
			score += 3
		}

		// file extension preference
		for i, ext := range extPref {
			if strings.HasSuffix(name, strings.ToLower(ext)) {
				score += 2 * (len(extPref) - i) // earlier ext = higher score
				break
			}
		}

		// TODO: dunno what to do about this ngl
		// common negatives
		if strings.Contains(name, "musl") && goos == "linux" {
			score -= 1
		}

		if score > best {
			best = score
			winner = a
		}
	}

	if winner != nil {
		return winner, nil
	}

	// fallback if no winner was found
	var osOnly []*github.ReleaseAsset
	for _, a := range assets {
		name := strings.ToLower(a.GetName())
		if skipAsset(name) {
			continue
		}

		fallbackRels := []string{"universal", "noarch", "any", "portable"}

		if utils.ContainsAny(name, osTokens[goos]) ||
			utils.ContainsAny(name, fallbackRels) {
			osOnly = append(osOnly, a)
		}
	}

	if len(osOnly) > 0 {
		// pick by extension preference
		winner := osOnly[0]
		bestIdx := len(extPref)
		for _, a := range osOnly {
			n := strings.ToLower(a.GetName())
			for i, ext := range extPref {
				if strings.HasSuffix(n, ext) && i < bestIdx {
					bestIdx, winner = i, a
					break
				}
			}
		}
		return winner, nil
	}

	return nil, fmt.Errorf("no suitable asset for %s/%s on %q", goos, goarch, rel.GetTagName())
}

func skipAsset(name string) bool {
	// ignore checksum/signature/manifest files
	bad := []string{
		"sha256", "sha512", "checksums", ".sha256", ".sha512", ".sig", ".asc", ".txt", ".sbom",
	}
	for _, b := range bad {
		if strings.Contains(name, b) {
			return true
		}
	}
	return false
}

func downloadTo(ctx context.Context, url, dest string) error {
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

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
