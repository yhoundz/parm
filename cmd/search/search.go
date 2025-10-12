/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package search

import (
	"fmt"
	"parm/internal/core/catalog"
	"parm/internal/gh"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var query string

var NewProvider = gh.New

// searchCmd represents the search command
var SearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Searches for repositories",
	Args: func(cmd *cobra.Command, args []string) error {
		if query != "" && len(args) > 0 {
			return fmt.Errorf("error: cannot have any args with the --query flag.")
		} else {
			exp := cobra.ExactArgs(1)
			return exp(cmd, args)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		token, err := gh.GetStoredApiKey(viper.GetViper())
		if err != nil {
			return err
		}
		client := NewProvider(ctx, token)
		var opts = catalog.RepoSearchOptions{
			Key:   nil,
			Query: nil,
		}
		if query != "" {
			opts.Query = &query
		} else {
			opts.Key = &args[0]
		}
		// TODO: finish this up
		_, err = catalog.SearchRepo(ctx, client.Search(), opts)
		return nil
	},
}

func init() {
	SearchCmd.Flags().StringVarP(&query, "query", "q", "", "Searches for the exact query string outlined by the GitHub REST API instead of a general search term.")
}
