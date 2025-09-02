package installer

import (
	"context"
	"fmt"
	"os"
	gh "parm/internal/github"
	"parm/internal/utils"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v74/github"
)

func (in *Installer) installFromReleaseByType(ctx context.Context, pkgPath, owner, repo string, opts InstallOptions) error {
	var rel *github.RepositoryRelease

	switch opts.Type {
	case Release:
		// TODO: cleanup if something in the install process goes wrong?
		valid, rel, err := gh.ValidateRelease(ctx, in.client, owner, repo, opts.Version)
		if err != nil {
			return fmt.Errorf("ERROR: Cannot resolve release %s on %s/%s", opts.Version, owner, repo)
		}
		if !valid {
			return fmt.Errorf("ERROR: Release %s not valid, %w", opts.Version, err)
		}

		return in.InstallFromRelease(ctx, pkgPath, owner, repo, rel, opts)
	case PreRelease:
		valid, rel, err := gh.ValidatePreRelease(ctx, in.client, owner, repo)
		if err != nil {
			return fmt.Errorf("err: cannot resolve pre-release %s on %s/%s: %w", rel.GetTagName(), owner, repo, err)
		}
		if !valid {
			return fmt.Errorf("error: no valid pre-release found for %s/%s", owner, repo)
		}

		return in.InstallFromRelease(ctx, pkgPath, owner, repo, rel, opts)
	default:
		return fmt.Errorf("invalid install type for release-based installation: %s", opts.Type)
	}

	return in.InstallFromRelease(ctx, pkgPath, owner, repo, rel, opts)
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

		man, err := NewManifest(owner, repo, rel.GetTagName(), Release, true, pkgPath)
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

	installT := Release
	if opts.Type == PreRelease {
		installT = PreRelease
	}

	man, err := NewManifest(owner, repo, rel.GetTagName(), installT, false, pkgPath)
	if err != nil {
		return fmt.Errorf("error: failed to create manifest: %w", err)
	}
	if err := man.WriteManifest(pkgPath); err != nil {
		return fmt.Errorf("error: failed to write manifest: %w", err)
	}

	return nil
}
