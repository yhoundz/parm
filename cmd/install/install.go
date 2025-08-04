/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package install

import (
	"parm/internal/deps"

	"github.com/spf13/cobra"
)

var source bool
var channel string

// installCmd represents the install command
var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs a new package",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return deps.Require("git")
	},
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	InstallCmd.Flags().BoolVarP(&source, "source", "s", false, "Build from source")
	InstallCmd.Flags().StringVarP(&channel, "channel", "ch", "stable", "Release channel: nightly, stable, etc.")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
