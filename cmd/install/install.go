/*
Copyright © 2025 Alexander Wang
*/
package install

import (
	"fmt"
	gh "parm/internal/github"
	"parm/internal/installer"
	"parm/internal/manifest"
	"parm/internal/utils"
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
			release = tag
			args[0] = owner + "/" + repo
		}

		if err := cmdx.MarkFlagsRequireFlag(cmd, "release", "asset"); err != nil {
			return err
		}

		if err := cmdx.MarkFlagsRequireFlag(cmd, "pre-release", "asset"); err != nil {
			return err
		}

		return nil
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := args[0]

		ctx := cmd.Context()
		token := viper.GetString("github_api_token")
		client := gh.NewRepoClient(ctx, token)

		inst := installer.New(client)
		owner, repo, _ := cmdparser.ParseRepoRef(pkg)

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

		pbRender := func(ev progress.Event) {
			pb := progressbar.NewBar()
			pb.Progress(100. * float64(ev.Current) / float64(ev.Total))
		}

		wrapped, done := progress.GetAsyncCallback(pbRender, 128)
		defer done()

		opts := installer.InstallOptions{
			Type:     insType,
			Version:  version,
			Asset:    asset,
			Progress: wrapped,
		}

		dest, err := utils.MakeInstallDir(owner, repo, 0o755)
		if err != nil {
			return err
		}
		fmt.Printf("installing %s/%s to %s\n", owner, repo, dest)

		err = inst.Install(ctx, dest, owner, repo, opts)
		if err != nil {
			fmt.Print(err)
		}

		return err
	},
}

func init() {
	InstallCmd.PersistentFlags().BoolVarP(&pre_release, "pre-release", "p", false, "Installs the latest pre-release binary, if availabale")
	InstallCmd.PersistentFlags().StringVarP(&release, "release", "r", "", "Install binary from this release tag")
	InstallCmd.PersistentFlags().StringVarP(&asset, "asset", "a", "", "Installs a specific asset from a release")

	InstallCmd.MarkFlagsMutuallyExclusive("release", "pre-release")
}
