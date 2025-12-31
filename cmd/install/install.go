/*
Copyright © 2025 Alexander Wang
*/
package install

import (
	"fmt"
	"io"
	"parm/internal/cmdutil"
	"parm/internal/core/installer"
	"parm/internal/gh"
	"parm/internal/manifest"
	"parm/internal/parmutil"
	"parm/pkg/cmdparser"
	"parm/pkg/cmdx"
	"parm/pkg/deps"
	"parm/pkg/progress"
	"parm/pkg/sysutil"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
)

func NewInstallCmd(f *cmdutil.Factory) *cobra.Command {
	var pre_release bool
	var release string
	var asset string
	var strict bool
	var no_verify bool

	// installCmd represents the install command
	var installCmd = &cobra.Command{
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

			if tag != "" {
				confFlags := []string{"release", "pre-release"}
				for _, flag := range confFlags {
					if cmd.Flags().Changed(flag) {
						return fmt.Errorf("cannot use @version shorthand with the --%s flag", flag)
					}
				}
				_ = cmd.Flags().Set("release", tag)
				args[0] = owner + "/" + repo
			}

			if !cmd.Flags().Changed("release") && !cmd.Flags().Changed("pre-release") {
				_ = cmd.Flags().Set("release", "")
			}

			if err := cmdx.MarkFlagsRequireFlag(cmd, "release", "asset"); err != nil {
				if err := cmdx.MarkFlagsRequireFlag(cmd, "pre-release", "asset", "strict"); err != nil {
					return err
				}
			}

			return nil
		},
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg := args[0]

			ctx := cmd.Context()

			var owner, repo string
			var err error

			owner, repo, err = cmdparser.ParseRepoRef(pkg)
			if err != nil {
				owner, repo, err = cmdparser.ParseGithubUrlPattern(pkg)
				if err != nil {
					return err
				}
			}

			// Try to get API key, but don't fail if it's not available
			token, _ := gh.GetStoredApiKey(viper.GetViper())

			// Create a temporary client to check if repo is public
			tempClient := f.Provider(ctx, "").Repos()
			isPublic, err := gh.IsRepositoryPublic(ctx, tempClient, owner, repo)
			if err != nil {
				return fmt.Errorf("error: cannot determine repository visibility: %w", err)
			}

			// If repo is private and we don't have a token, fail
			if !isPublic && token == "" {
				return fmt.Errorf("error: api key not found\n\nThis repository is private. Set a GitHub token:\n  export GITHUB_TOKEN=$(gh auth token)")
			}

			client := f.Provider(ctx, token).Repos()

			inst := installer.New(client, token)

			var insType manifest.InstallType
			var version *string
			if release != "" {
				insType = manifest.Release
				version = &release
			} else if pre_release {
				insType = manifest.PreRelease
				// INFO: do nothing, populate version later
				version = nil
			} else {
				// release == ""
				insType = manifest.Release
				version = nil
			}

			pb := mpb.New(mpb.WithWidth(60))

			var bar *mpb.Bar
			hooks := &progress.Hooks{
				Decorator: func(stage progress.Stage, r io.Reader, total int64) io.Reader {
					if stage != progress.StageDownload {
						return r
					}
					if bar != nil {
						return bar.ProxyReader(r)
					}
					bar = pb.AddBar(total,
						mpb.PrependDecorators(
							// decor.Name("downloading"),
							decor.Percentage(decor.WCSyncSpace),
						),
						mpb.AppendDecorators(
							decor.OnComplete(
								decor.EwmaETA(decor.ET_STYLE_GO, 60), "done",
							),
							decor.Name(" "),
							decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 60),
							decor.Name(" "),
							decor.Elapsed(decor.ET_STYLE_GO, decor.WC{W: 5}),
						),
					)
					return bar.ProxyReader(r)
				},
				Callback: nil,
			}

			var ass *string
			if asset == "" {
				ass = nil
			} else {
				ass = &asset
			}

			opts := installer.InstallFlags{
				Type:    insType,
				Version: version,
				Asset:   ass,
				Strict:  strict,
				VerifyLevel: func() uint8 {
					if no_verify {
						return 0
					}
					// TODO: change to actual verify-level once implemented
					return 1
				}(),
			}

			if opts.Version == nil {
				fmt.Printf("Installing %s/%s::latest\n", owner, repo)
			} else {
				fmt.Printf("Installing %s/%s::%s\n", owner, repo, *opts.Version)
			}

			installPath := parmutil.GetInstallDir(owner, repo)
			res, err := inst.Install(ctx, owner, repo, installPath, opts, hooks)
			pb.Wait()
			if err != nil {
				if res == nil {
					return err
				}
				parentDir, cErr := sysutil.GetParentDir(res.InstallPath)
				if cErr == nil {
					err = parmutil.Cleanup(parentDir)
					if err != nil {
						return err
					}
				}
				return err
			}

			man, err := manifest.New(owner, repo, res.Version, opts.Type, res.InstallPath)
			if err != nil {
				return fmt.Errorf("error: failed to create manifest: \n%w", err)
			}
			err = man.Write(res.InstallPath)
			if err != nil {
				return err
			}

			binPaths := man.GetFullExecPaths()

			for _, execPath := range binPaths {
				cleanName := sysutil.CleanBinaryName(filepath.Base(execPath))
				pathToSymLinkTo := parmutil.GetBinDir(cleanName)

				// TODO: use shims for windows instead?
				err = sysutil.SymlinkBinToPath(execPath, pathToSymLinkTo)
				if err != nil {
					return err
				}
				deps, err := deps.GetMissingLibs(ctx, execPath)
				if err != nil {
					return err
				}
				if len(deps) > 0 {
					fmt.Printf("required dependencies found for %s/%s:\n", owner, repo)
					for _, dp := range deps {
						fmt.Println("\t" + dp)
					}
					fmt.Println("Note: this is PURELY informational, and does not necessarily mean that your machine doesn't have these dependencies.")
				}
			}

			fmt.Println()

			return nil
		},
	}

	installCmd.Flags().BoolVarP(&pre_release, "pre-release", "p", false, "Installs the latest pre-release binary, if available")
	installCmd.Flags().BoolVarP(&strict, "strict", "s", false, "Only available with the --pre-release flag. Will only install pre-release versions and not stable releases.")
	installCmd.Flags().BoolVarP(&no_verify, "no-verify", "n", false, "Skips integrity check")
	installCmd.Flags().StringVarP(&release, "release", "r", "", "Install binary from this release tag.")
	installCmd.Flags().StringVarP(&asset, "asset", "a", "", "Installs a specific asset from a release.")

	installCmd.MarkFlagsMutuallyExclusive("release", "pre-release")
	installCmd.MarkFlagsMutuallyExclusive("release", "strict")

	return installCmd
}
