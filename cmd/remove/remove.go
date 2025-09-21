/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
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
	Use:   "remove <owner>/<repo>",
	Short: "Uninstalls a parm package",
	Long:  `Uninstalls a parm package. Does not remove the configuration files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := args[0]
		ctx := cmd.Context()
		owner, repo, err := cmdparser.ParseRepoRef(pkg)

		if err != nil {
			return fmt.Errorf("invalid package ref: %q: %w", pkg, err)
		}

		return uninstaller.Uninstall(ctx, owner, repo)
	},
}

func init() {}
