/*
Copyright © 2025 Alexander Wang
*/
package remove

import (
	"fmt"
	"parm/internal/uninstaller"
	"parm/pkg/cmdparser"

	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var RemoveCmd = &cobra.Command{
	Use:     "remove <owner>/<repo>...",
	Aliases: []string{"uninstall"},
	Short:   "Uninstalls a parm package",
	Long:    `Uninstalls a parm package. Does not remove the configuration files`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, pkg := range args {
			ctx := cmd.Context()
			owner, repo, err := cmdparser.ParseRepoRef(pkg)

			if err != nil {
				fmt.Printf("invalid package ref: %q: %s\n", pkg, err)
				continue
			}

			err = uninstaller.Uninstall(ctx, owner, repo)
			if err != nil {
				fmt.Printf("error: cannot uninstall %s: %s\n", pkg, err)
			}
		}
	},
}

func init() {}
