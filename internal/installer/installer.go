package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	gh "parm/internal/github"
	"parm/internal/parser"
)

type Installer struct {
	client gh.RepoClient
}

type InstallOptions struct {
	Branch  string
	Commit  string
	Release string
	Source  bool
}

func New(cli gh.RepoClient) *Installer {
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

		cloneLink, _ := parser.BuildGitLink(owner, repo)
		if valid {
			cmd := exec.CommandContext(ctx, "git", "clone",
				"--depth=1", "--recurse-submodules", "--shallow-submodules", "--branch",
				opts.Branch, cloneLink, pkgPath)

			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin

			// TODO: figure out what this dodes
			if err := cmd.Run(); err != nil {
				if eerr, ok := err.(*exec.ExitError); ok {
					fmt.Printf("git exited with %d\n", eerr.ExitCode())
				} else {
					fmt.Printf("failed to start or was killed: %v\n", err)
				}
			}
			return nil
		}
	} else if opts.Commit != "" {
		// TODO:
		valid, _, err := gh.ValidateCommit(ctx, in.client, owner, repo, opts.Commit)
		if err != nil {
			fmt.Printf("ERROR: cannot resolve commit: %q on %s/%s", opts.Commit, owner, repo)
			return err
		}

		cloneLink, _ := parser.BuildGitLink(owner, repo)
		if valid {
			cmd := exec.CommandContext(ctx, "git", "clone",
				"--depth=1", "--recurse-submodules", "--shallow-submodules",
				opts.Branch, cloneLink, pkgPath)

			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin

			// TODO: figure out what this dodes
			if err := cmd.Run(); err != nil {
				if eerr, ok := err.(*exec.ExitError); ok {
					fmt.Printf("git exited with %d\n", eerr.ExitCode())
				} else {
					fmt.Printf("failed to start or was killed: %v\n", err)
				}
			}
			return err
		}
	}
	if opts.Source {

	}

	return nil
}
