/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package search

import (
	"fmt"

	"github.com/spf13/cobra"
)

var query string

// searchCmd represents the search command
var SearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Searches for repositories",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("search called")
	},
}

func init() {
	SearchCmd.Flags().StringVarP(&query, "query", "q", "", "Searches for the exact query string outlined by the GitHub REST API instead of a general search term.")
}
