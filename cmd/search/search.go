/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package search

import (
	"fmt"

	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var SearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Searches for repositories",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("search called")
	},
}

func init() {}
