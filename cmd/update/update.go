/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package update

import (
	"fmt"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates a package",
	Long:  `Updates a package to the latest available version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("update called")
		return nil
	},
}

func init() {}
