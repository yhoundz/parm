/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package sw

import (
	"parm/internal/cmdutil"

	"github.com/spf13/cobra"
)

func NewSwitchCmd(f *cmdutil.Factory) *cobra.Command {
	// switchCmd represents the switch command
	var switchCmd = &cobra.Command{
		Use:   "switch",
		Short: "A brief description of your command",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	switchCmd.AddCommand(
		NewChannelCmd(f),
	)

	return switchCmd
}
