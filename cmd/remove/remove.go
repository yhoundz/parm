/*
Copyright © 2025 Alexander Wang
*/
package remove

import (
	"fmt"
	"parm/internal/cmdutil"
	"parm/internal/core/uninstaller"
	"parm/pkg/cmdparser"

	"github.com/spf13/cobra"
)

func NewRemoveCmd(f *cmdutil.Factory) *cobra.Command {
	// uninstallCmd represents the uninstall command
	var RemoveCmd = &cobra.Command{
		Use:     "remove <owner>/<repo>...",
		Aliases: []string{"uninstall", "rm"},
		Short:   "Uninstalls a parm package",
		Long:    `Uninstalls a parm package. Does not remove the configuration files`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()
			removed := make(map[string]bool)

			for _, pkg := range args {
				if _, ok := removed[pkg]; ok {
					// already processed the package uninstallation
					continue
				}
				removed[pkg] = true
				owner, repo, err := cmdparser.ParseRepoRef(pkg)

				if err != nil {
					fmt.Printf("invalid package ref: %q: %s\n", pkg, err)
					continue
				}

				err = uninstaller.RemoveSymlink(ctx, owner, repo)
				if err != nil {
					fmt.Printf("error: cannot remove symlink for %s/%s:\n%q", owner, repo, err)
				}

				err = uninstaller.Uninstall(ctx, owner, repo)
				if err != nil {
					fmt.Printf("error: cannot uninstall %s: %s\n", pkg, err)
				}
			}
		},
	}

	return RemoveCmd
}
