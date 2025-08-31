package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"parm/internal/cmdparser"
	gh "parm/internal/github"
)

func (in *Installer) installFromCommit(ctx context.Context, pkgPath, owner, repo string, opts InstallOptions) error {
	// TODO: testing
	valid, _, err := gh.ValidateCommit(ctx, in.client, owner, repo, opts.Version)
	if err != nil {
		return fmt.Errorf("ERROR: cannot resolve commit: %q on %s/%s", opts.Version, owner, repo)
	}
	if !valid {
		return fmt.Errorf("ERROR: commit %q is not valid on %s/%s", opts.Version, owner, repo)
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
	err = execGitCmd("-C", pkgPath, "fetch", "--depth=1", "origin", opts.Version)
	if err != nil {
		return err
	}

	// checkout commit + subms
	err = execGitCmd("-C", pkgPath, "checkout", "--recurse-submodules", opts.Version)
	if err != nil {
		return err
	}

	// ensure submodules materialize shallowly
	err = execGitCmd("-C", pkgPath, "submodule", "update", "--init", "--depth=1", "--recursive")
	if err != nil {
		return err
	}

	man, err := NewManifest(owner, repo, opts.Version, Commit, true, pkgPath)
	if err != nil {
		return fmt.Errorf("error: failed to create manifest: %w", err)
	}
	if err = man.WriteManifest(pkgPath); err != nil {
		return fmt.Errorf("error: failed to write manifest: %w", err)
	}

	return nil
}
