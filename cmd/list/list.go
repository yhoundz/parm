/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package list

import (
	"fmt"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists out currently installed packages",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: list
		return fmt.Errorf("nil")
	},
}

func init() {
}
