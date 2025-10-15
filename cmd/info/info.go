/*
Copyright © 2025 Alexander Wang
*/
package info

import (
	"fmt"
	"parm/internal/cmdutil"
	"parm/internal/core/catalog"
	"parm/internal/gh"
	"parm/pkg/cmdparser"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO: get rid of all this NewProvider bs and replace with a factory pattern
// TODO: don't retrive package info from GitHub if it doesn't have any releases.
// TODO: don't error out if trying to retrieve package info locally if the package doesn't exist.

func NewInfoCmd(f *cmdutil.Factory) *cobra.Command {
	var getUpstream bool
	var infoCmd = &cobra.Command{
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
			client := f.Provider(ctx, token).Repos()
			var owner, repo string

			owner, repo, err = cmdparser.ParseRepoRef(pkg)
			if err != nil {
				owner, repo, err = cmdparser.ParseGithubUrlPattern(pkg)
				if err != nil {
					return err
				}
			}

			info, err := catalog.GetPackageInfo(ctx, client, owner, repo, getUpstream)
			if err != nil {
				return err
			}
			pr := info.String()
			fmt.Println(pr)
			return nil
		},
	}
	infoCmd.Flags().BoolVarP(&getUpstream, "get-upstream", "u", false, "Retrieves the Repository info from the GitHub repository instead of the locally installed package")

	return infoCmd
}
