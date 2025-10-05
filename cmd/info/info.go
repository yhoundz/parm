/*
Copyright © 2025 Alexander Wang
*/
package info

import (
	"fmt"

	"github.com/spf13/cobra"
)

var setUpstream *bool

var InfoCmd = &cobra.Command{
	Use:   "info <owner>/<repo>",
	Short: "Prints out information about a package",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("info called")
		return nil
	},
}

func init() {
	InfoCmd.PersistentFlags().BoolVarP()
}
