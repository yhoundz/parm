/*
Copyright © 2025 A.W. <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"parm/cmd/configure"
	"parm/cmd/install"
	"parm/cmd/remove"
	"parm/cmd/update"
	"parm/internal/config"
	"parm/internal/parmver"

	"github.com/spf13/cobra"
)

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
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(configure.ConfigCmd)
	rootCmd.AddCommand(install.InstallCmd)
	rootCmd.AddCommand(remove.RemoveCmd)
	rootCmd.AddCommand(update.UpdateCmd)
}
