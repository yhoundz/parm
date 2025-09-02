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
		return fmt.Errorf("error: cannot resolve branch: %q on %s/%s", branch, owner, repo)
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
			fmt.Printf("git exited with %d\n", eerr.ExitCode())
			return eerr
		} else {
			fmt.Printf("failed to start or was killed: %v\n", err)
		}
		return err
	}

	man, err := NewManifest(owner, repo, branch, Branch, true, pkgPath)
	if err != nil {
		return fmt.Errorf("error: failed to create  manifest: %w", err)
	}
	if err := man.WriteManifest(pkgPath); err != nil {
		return fmt.Errorf("error: failed to write manifest: %w", err)
	}
	return nil
}
