package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"parm/internal/cmdparser"
	gh "parm/internal/github"
)

func (in *Installer) installFromCommit(ctx context.Context, pkgPath, owner, repo, commitSHA string) error {
	// TODO: testing
	valid, _, err := gh.ValidateCommit(ctx, in.client, owner, repo, commitSHA)
	if err != nil {
		return fmt.Errorf("ERROR: cannot resolve commit: %q on %s/%s: %w", commitSHA, owner, repo, err)
	}
	if !valid {
		return fmt.Errorf("ERROR: commit %q is not valid on %s/%s: %w", commitSHA, owner, repo, err)
	}

	cloneLink, _ := cmdparser.BuildGitLink(owner, repo)

	var execGitCmd = func(arg ...string) error {
		cmd := exec.CommandContext(ctx, "git", arg...)
		cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git command %v failed: %w", arg, err)
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
	err = execGitCmd("-C", pkgPath, "fetch", "--depth=1", "origin", commitSHA)
	if err != nil {
		return err
	}

	// checkout commit + subms
	err = execGitCmd("-C", pkgPath, "checkout", "--recurse-submodules", commitSHA)
	if err != nil {
		return err
	}

	// ensure submodules materialize shallowly
	err = execGitCmd("-C", pkgPath, "submodule", "update", "--init", "--depth=1", "--recursive")
	if err != nil {
		return err
	}

	man, err := NewManifest(owner, repo, commitSHA, Commit, true, pkgPath)
	if err != nil {
		return fmt.Errorf("error: failed to create manifest: %w", err)
	}
	return man.WriteManifest(pkgPath)
}
