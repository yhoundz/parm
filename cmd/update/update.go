/*
Copyright © 2025 Alexander Wang
*/
package update

import (
	"fmt"
	gh "parm/internal/github"
	"parm/internal/installer"
	"parm/internal/updater"
	"parm/pkg/cmdparser"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// updateCmd represents the update command
var UpdateCmd = &cobra.Command{
	Use:   "update <owner>/<repo>",
	Short: "Updates a package",
	Long:  `Updates a package to the latest available version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		token, err := gh.GetStoredApiKey(viper.GetViper())
		if err != nil {
			fmt.Printf("%s\ncontinuing without api key.\n", err)
		}
		client := gh.NewRepoClient(ctx, token)
		inst := installer.New(client)
		up := updater.New(client, inst)
		updated := make(map[string]bool)

		for _, pkg := range args {
			if _, ok := updated[pkg]; ok {
				// already updated package
				continue
			}
			updated[pkg] = true

			var owner, repo string
			var err error

			owner, repo, err = cmdparser.ParseRepoRef(pkg)
			if err != nil {
				owner, repo, err = cmdparser.ParseGithubUrlPattern(pkg)
				if err != nil {
					fmt.Printf("%s\n", err)
					continue
				}
			}

			// TODO: change this later
			err = up.Update(ctx, owner, repo, nil)
			if err != nil {
				fmt.Printf("error: failed to update %s/%s\n", owner, repo)
			}
		}
		return nil
	},
}

func init() {}
