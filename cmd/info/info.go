/*
Copyright © 2025 Alexander Wang
*/
package info

import (
	"fmt"
	gh "parm/internal/github"
	"parm/internal/informer"
	"parm/pkg/cmdparser"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getUpstream bool

var InfoCmd = &cobra.Command{
	Use:   "info <owner>/<repo>",
	Short: "Prints out information about a package",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		pkg := args[0]
		token, err := gh.GetStoredApiKey(viper.GetViper())
		if err != nil {
			fmt.Printf("%s\ncontinuing without api key", err)
		}
		client := gh.NewRepoClient(ctx, token)
		var owner, repo string

		owner, repo, err = cmdparser.ParseRepoRef(pkg)
		if err != nil {
			owner, repo, err = cmdparser.ParseGithubUrlPattern(pkg)
			if err != nil {
				return err
			}
		}

		info, err := informer.GetPackageInfo(ctx, client, owner, repo, getUpstream)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	InfoCmd.PersistentFlags().BoolVarP(&getUpstream, "get-upstream", "u", false, "Retrieves the Repository info from the GitHub repository instead of the locally installed package")
}
