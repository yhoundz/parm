/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package install

import (
	// "os/exec"
	"fmt"
	"parm/internal/config"
	"parm/internal/deps"

	// gh "parm/internal/github"
	"parm/internal/parser"

	"github.com/spf13/cobra"
)

var source bool
var (
	branch  string
	commit  string
	release string
)

// installCmd represents the install command
var InstallCmd = &cobra.Command{
	Use:   "install <owner>/<repo>@[release-tag]",
	Short: "Installs a new package",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {

		// check if git is installed and in the PATH
		if err := deps.Require("git"); err != nil {
			return err
		}

		owner, repo, tag, err := parser.ParseRepoReleaseRef(args[0])
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
			owner, repo, tag, err := parser.ParseGithubUrlPatternWithRelease(args[0])
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
		// pkg := args[0]
		switch {
		case branch != "":
			fmt.Println("WARNING: installing with --branch will automatically download from source.")
			installPath := config.Cfg.ParmPkgDirPath
		}
		return nil
	},
}

func init() {
	InstallCmd.PersistentFlags().BoolVarP(&source, "source", "s", false, "Build from source")
	InstallCmd.PersistentFlags().StringVarP(&branch, "branch", "b", "", "Install from this git branch")
	InstallCmd.PersistentFlags().StringVarP(&commit, "commit", "c", "", "Install from this git commit SHA")
	InstallCmd.PersistentFlags().StringVarP(&release, "release", "r", "", "Install binary from this release tag")
	InstallCmd.MarkFlagsMutuallyExclusive("branch", "commit", "release")
}
