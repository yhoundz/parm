/*
Copyright © 2025 A.W. <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"parm/cmd/configure"
	"parm/cmd/install"
	"parm/internal/version"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "parm",
	Short: "A zero-root, GitHub-native CLI package manager for installing and managing any GitHub-hosted tool.",
	Long: `Parm (Probably a Repo Manager) is a thin CLI tool that downloads
and installs prebuilt binaries or builds from source for any
GitHub repository, keeping everything neatly isolated and in a single
binary. It requires only Git and your shell, avoids system-wide changes
or root access, and gives you full control over your programs.`,
	Version: version.AppVersion.String(),
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
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.parm.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
