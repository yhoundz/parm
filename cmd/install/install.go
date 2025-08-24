/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package install

import (
	"fmt"
	"parm/internal/cmdx"
	"parm/internal/deps"
	gh "parm/internal/github"
	"parm/internal/installer"
	"parm/internal/utils"

	"parm/internal/cmdparser"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var source bool
var (
	branch  string
	commit  string
	release string
)

// installCmd represents the install command
var InstallCmd = &cobra.Command{
	Use:     "install <owner>/<repo>@[release-tag]",
	Aliases: []string{"i"},
	Short:   "Installs a new package",
	Long:    ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if err := cmdx.MarkFlagsRequireFlag(cmd, "release", "source"); err != nil {
			return err
		}

		// TODO: put this check elsewhere?
		if commit != "" {
			// if building from source, git is required
			if err := deps.Require("git"); err != nil {
				return err
			}
		}

		owner, repo, tag, err := cmdparser.ParseRepoReleaseRef(args[0])
		if err == nil {
			if tag != "" {
				if branch != "" || commit != "" || release != "" {
					return fmt.Errorf("cannot mix @tag with --branch/--commit/--release")
				}
				release = tag
				args[0] = owner + "/" + repo
			} else {
				// no tag matched, wait for other args
			}
		} else {
			owner, repo, tag, err := cmdparser.ParseGithubUrlPatternWithRelease(args[0])
			if err != nil {
				if tag != "" {
					// there is a tag,
					if branch != "" || commit != "" || release != "" {
						return fmt.Errorf("cannot mix @tag with --branch/--commit/--release")
					}
					release = tag
					args[0] = owner + "/" + repo
				}
			} else {
				// there is an error
				return fmt.Errorf("Cannot resolve git repository")
			}
		}

		return nil
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := args[0]

		fmt.Println("WARNING: installing with --branch will automatically download from source.")

		ctx := cmd.Context()
		token := viper.GetString("github_api_token")
		client := gh.NewRepoClient(ctx, token)

		inst := installer.New(client)
		owner, repo, _ := cmdparser.ParseRepoRef(pkg)
		opts := installer.InstallOptions{
			Branch:  branch,
			Commit:  commit,
			Release: release,
			Source:  source,
		}

		dest := utils.GetInstallDir(owner, repo)
		fmt.Println(dest)

		err := inst.Install(ctx, dest, owner, repo, opts)
		if err != nil {
			fmt.Print(err)
		}
		return err
	},
}

func init() {
	InstallCmd.PersistentFlags().BoolVarP(&source, "source", "s", false, "Build from source")
	InstallCmd.PersistentFlags().StringVarP(&branch, "branch", "b", "", "Install from this git branch")
	InstallCmd.PersistentFlags().StringVarP(&commit, "commit", "c", "", "Install from this git commit SHA")
	InstallCmd.PersistentFlags().StringVarP(&release, "release", "r", "", "Install binary from this release tag")

	InstallCmd.MarkFlagsMutuallyExclusive("branch", "commit", "release")
	// InstallCmd.MarkFlagsMutuallyExclusive("branch", "commit", "source")
}
