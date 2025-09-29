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
		for _, pkg := range args {
			var owner, repo string
			var err error

			owner, repo, err = cmdparser.ParseRepoRef(pkg)
			if err != nil {
				owner, repo, err = cmdparser.ParseGithubUrlPattern(pkg)
				if err != nil {
					return err
				}
			}

			ctx := cmd.Context()
			token := viper.GetString("github_api_token")
			client := gh.NewRepoClient(ctx, token)
			inst := installer.New(client)
			up := updater.New(client, inst)

			err = up.Update(ctx, owner, repo)
			if err != nil {
				fmt.Printf("error: failed to update %s/%s", owner, repo)
			}
		}
		return nil
	},
}

func init() {}
