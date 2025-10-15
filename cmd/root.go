/*
Copyright © 2025 Alexander Wang
*/
package cmd

import (
	"os"
	"parm/cmd/configure"
	"parm/cmd/info"
	"parm/cmd/install"
	"parm/cmd/list"
	"parm/cmd/remove"
	"parm/cmd/update"
	"parm/internal/cmdutil"
	"parm/internal/config"
	"parm/internal/gh"
	"parm/parmver"

	"github.com/spf13/cobra"
)

func NewRootCmd(f *cmdutil.Factory) *cobra.Command {
	// rootCmd represents the base command when called without any subcommands
	var rootCmd = &cobra.Command{
		Use:   "parm",
		Short: "A zero-root, GitHub-native CLI package manager for installing and managing any GitHub-hosted tool.",
		Long: `Parm is a thin CLI tool that downloads and installs prebuilt
your programs. It has zero dependencies, zero root access, and is truly
cross-platform on Windows, Linux, and MacOS.`,
		Version: parmver.AppVersion.String(),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			err := config.Init()
			if err != nil {
				return err
			}
			return nil
		},
	}

	rootCmd.AddCommand(
		configure.NewConfigureCmd(f),
		install.NewInstallCmd(f),
		remove.NewRemoveCmd(f),
		update.NewUpdateCmd(f),
		list.NewListCmd(f),
		info.NewInfoCmd(f),
		// search.NewSearchCmd(f),
	)

	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	factory := &cmdutil.Factory{
		Provider: gh.New,
	}
	err := NewRootCmd(factory).Execute()
	if err != nil {
		os.Exit(1)
	}
}
