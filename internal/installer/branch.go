package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"parm/internal/cmdparser"
	gh "parm/internal/github"
)

func (in *Installer) installFromBranch(ctx context.Context, pkgPath, owner, repo, branch string) error {
	valid, _, err := gh.ValidateBranch(ctx, in.client, owner, repo, branch)
	if err != nil {
		return fmt.Errorf("error: cannot resolve branch: %q on %s/%s: %w", branch, owner, repo, err)
	}
	if !valid {
		return fmt.Errorf("error: branch: %s cannot be found", branch)
	}

	cloneLink, _ := cmdparser.BuildGitLink(owner, repo)
	cmd := exec.CommandContext(ctx, "git", "clone",
		"--depth=1", "--recurse-submodules", "--shallow-submodules", "--branch",
		branch, cloneLink, pkgPath)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("git exited with %d: %w", eerr.ExitCode(), eerr)
		}
		return fmt.Errorf("failed to run git clone: %w", err)
	}

	man, err := NewManifest(owner, repo, branch, Branch, true, pkgPath)
	if err != nil {
		return fmt.Errorf("error: failed to create manifest: %w", err)
	}
	return man.WriteManifest(pkgPath)
}
