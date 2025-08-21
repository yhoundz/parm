/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package uninstall

import (
	"fmt"
	"parm/internal/cmdparser"
	"parm/internal/installer"

	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var UninstallCmd = &cobra.Command{
	Use:   "uninstall <owner>/<repo>",
	Short: "Uninstalls a parm package",
	Long:  `Uninstalls a parm package. Does not remove the configuration files`,
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := args[0]
		ctx := cmd.Context()
		owner, repo, err := cmdparser.ParseRepoRef(pkg)

		if err != nil {
			return fmt.Errorf("invalid package ref: %q: %w", pkg, err)
		}

		inst := installer.New(nil)
		return inst.Uninstall(ctx, owner, repo)
	},
}

func init() {}
