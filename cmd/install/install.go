/*
Copyright © 2025 Alexander Wang
*/
package install

import (
	"fmt"
	"io"
	"parm/internal/core/installer"
	"parm/internal/gh"
	"parm/internal/manifest"
	"parm/pkg/cmdparser"
	"parm/pkg/cmdx"
	"parm/pkg/progress"

	"fortio.org/progressbar"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pre_release bool
var release string
var asset string

// installCmd represents the install command
var InstallCmd = &cobra.Command{
	Use:   "install <owner>/<repo>@[release-tag]",
	Short: "Installs a new package",
	Long:  ``,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		owner, repo, tag, err := cmdparser.ParseRepoReleaseRef(args[0])
		if err != nil {
			owner, repo, tag, err = cmdparser.ParseGithubUrlPatternWithRelease(args[0])
		}
		if err != nil {
			return fmt.Errorf("cannot resolve git repository from input: %s", args[0])
		}

		// TODO: if tag is a pre-release tag, use the pre-release release channel
		if tag != "" {
			confFlags := []string{"release", "pre-release"}
			for _, flag := range confFlags {
				if cmd.Flags().Changed(flag) {
					return fmt.Errorf("cannot use @version shorthand with the --%s flag", flag)
				}
			}
			cmd.Flags().Set("release", tag)
			args[0] = owner + "/" + repo
		}

		if err := cmdx.MarkFlagsRequireFlag(cmd, "release", "asset"); err != nil {
			if err := cmdx.MarkFlagsRequireFlag(cmd, "pre-release", "asset"); err != nil {
				return err
			}
		}

		return nil
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := args[0]

		ctx := cmd.Context()
		token, _ := gh.GetStoredApiKey(viper.GetViper())
		client := gh.New(ctx, token).Repos()

		inst := installer.New(client)

		var owner, repo string
		var err error

		owner, repo, err = cmdparser.ParseRepoRef(pkg)
		if err != nil {
			owner, repo, err = cmdparser.ParseGithubUrlPattern(pkg)
			if err != nil {
				return err
			}
		}

		var insType manifest.InstallType
		var version string
		if release != "" {
			insType = manifest.Release
			version = release
		} else if pre_release {
			insType = manifest.PreRelease
			// INFO: do nothing, populate version later
			version = ""
		} else {
			insType = manifest.Release
			version = ""
		}

		pb := progressbar.NewBar()

		hooks := &progress.Hooks{
			Decorator: func(stage progress.Stage, r io.Reader, total int64) io.Reader {
				if stage != progress.StageDownload {
					return r
				}
				return progressbar.NewAutoReader(pb, r, total)
			},
			Callback: nil,
		}

		opts := installer.InstallOptions{
			Type:    insType,
			Version: version,
			Asset:   asset,
		}

		fmt.Printf("installing %s/%s\n", owner, repo)

		err = inst.Install(ctx, owner, repo, opts, hooks)
		if err != nil {
			fmt.Printf("%q\n", err)
		}

		fmt.Println()

		return err
	},
}

func init() {
	InstallCmd.PersistentFlags().BoolVarP(&pre_release, "pre-release", "p", false, "Installs the latest pre-release binary, if availabale")
	InstallCmd.PersistentFlags().StringVarP(&release, "release", "r", "", "Install binary from this release tag")
	InstallCmd.PersistentFlags().StringVarP(&asset, "asset", "a", "", "Installs a specific asset from a release")

	InstallCmd.MarkFlagsMutuallyExclusive("release", "pre-release")
}
