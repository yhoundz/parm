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
	if opts.Source {
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
					opts.Branch, cloneLink)

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
		} else if opts.Commit != "" {
			// TODO:
		}
	}

	return nil
}
