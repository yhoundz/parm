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
var pre_release bool
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

		owner, repo, tag, err := cmdparser.ParseRepoReleaseRef(args[0])
		if err != nil {
			owner, repo, tag, err = cmdparser.ParseGithubUrlPatternWithRelease(args[0])
		}
		if err != nil {
			return fmt.Errorf("cannot resolve git repository from input: %s", args[0])
		}

		if tag != "" {
			confFlags := []string{"branch", "commit", "release", "pre-release"}
			for _, flag := range confFlags {
				if cmd.Flags().Changed(flag) {
					return fmt.Errorf("cannot use @version shorthand with the --%s flag", flag)
				}
			}
			release = tag
			args[0] = owner + "/" + repo
		}

		// TODO: put this check elsewhere?
		if commit != "" {
			// if building from source, git is required
			if err := deps.Require("git"); err != nil {
				return err
			}
		}

		if err := cmdx.MarkFlagsRequireFlag(cmd, "release", "source"); err != nil {
			return err
		}

		if err := cmdx.MarkFlagsRequireFlag(cmd, "pre-release", "source"); err != nil {
			return err
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
			Branch:     branch,
			Commit:     commit,
			Release:    release,
			Source:     source,
			PreRelease: pre_release,
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
	InstallCmd.PersistentFlags().BoolVarP(&pre_release, "pre-release", "p", false, "Installs the latest pre-release binary, if availabale")
	InstallCmd.PersistentFlags().StringVarP(&branch, "branch", "b", "", "Install from this git branch")
	InstallCmd.PersistentFlags().StringVarP(&commit, "commit", "c", "", "Install from this git commit SHA")
	InstallCmd.PersistentFlags().StringVarP(&release, "release", "r", "", "Install binary from this release tag")

	InstallCmd.MarkFlagsMutuallyExclusive("branch", "commit", "release", "pre-release")
}
