package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	gh "parm/internal/github"
	"parm/internal/parser"

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
		// TODO: redo this part so i actually understand what's going on
		// if source build, download the source code from tarball
		// if not source, find best-matching binary based on GOOS and GOARCH, and then
		// get the download link
		// afterwards, download and extract the tarball to the desired dir.
		if opts.Source {
		}
	}

	return nil
}
