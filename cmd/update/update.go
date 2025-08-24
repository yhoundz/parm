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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("update called")
	},
}

func init() {}
