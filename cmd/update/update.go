/*
Copyright © 2025 Alexander Wang
*/
package update

import (
	"context"
	"fmt"
	"parm/internal/cmdutil"
	"parm/internal/core/catalog"
	"parm/internal/core/installer"
	"parm/internal/core/updater"
	"parm/internal/gh"
	"parm/internal/manifest"
	"parm/internal/parmutil"
	"parm/pkg/cmdparser"
	"parm/pkg/sysutil"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	type argsKey struct{}
	var strict bool
	var aKey argsKey

	// updateCmd represents the update command
	var updateCmd = &cobra.Command{
		Use:   "update <owner>/<repo>",
		Short: "Updates a package",
		Long:  `Updates a package to the latest available version.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			var normArgs []string

			if len(args) == 0 {
				mans, err := catalog.GetAllPkgManifest()
				if err != nil {
					fmt.Println("failed to retrieve packages")
					return
				}
				if len(mans) == 0 {
					fmt.Println("no packages to update")
					return
				}

				var newArgs = make([]string, len(mans))
				for i, man := range mans {
					pair := fmt.Sprintf("%s/%s", man.Owner, man.Repo)
					newArgs[i] = pair
				}
				normArgs = newArgs
			} else {
				ignored := make(map[string]bool)

				// remove duplicates and incorrectly formatted or nonexistent packages
				for _, arg := range args {
					owner, repo, err := cmdparser.ParseRepoRef(arg)
					if err != nil {
						owner, repo, err = cmdparser.ParseGithubUrlPattern(arg)
						if err != nil {
							ignored[arg] = true
							fmt.Printf("error: package %s not found, skipping...\n", arg)
							continue
						}
					}
					if _, ok := ignored[arg]; ok {
						// already updated package
						continue
					}
					parsed := fmt.Sprintf("%s/%s", owner, repo)
					normArgs = append(normArgs, parsed)
					ignored[arg] = true
				}
			}
			ctx := context.WithValue(cmd.Context(), aKey, normArgs)
			cmd.SetContext(ctx)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			args, _ := ctx.Value(aKey).([]string)

			token, err := gh.GetStoredApiKey(viper.GetViper())
			if err != nil {
				fmt.Printf("%s\ncontinuing without api key.\n", err)
			}
			client := f.Provider(ctx, token).Repos()
			inst := installer.New(client)
			up := updater.New(client, inst)
			flags := updater.UpdateFlags{
				Strict: strict,
			}

			for _, pkg := range args {
				// guaranteed to work now
				owner, repo, _ := cmdparser.ParseRepoRef(pkg)

				installPath := parmutil.GetInstallDir(owner, repo)
				parentDir, _ := sysutil.GetParentDir(installPath)

				res, err := up.Update(ctx, owner, repo, installPath, &flags, nil)
				if err != nil {
					_ = parmutil.Cleanup(parentDir)
					fmt.Printf("error: failed to update %s/%s:\n\t%q \n", owner, repo, err)
					continue
				}

				// Write new manifest
				old := res.OldManifest
				man, err := manifest.New(owner, repo, res.Version, old.InstallType, res.InstallPath)
				if err != nil {
					return fmt.Errorf("error: failed to create manifest: \n%w", err)
				}
				err = man.Write(res.InstallPath)
				if err != nil {
					return err
				}

				// Symlinked executables to PATH
				binPaths := man.GetFullExecPaths()
				for _, execPath := range binPaths {
					pathToSymLinkTo := parmutil.GetBinDir(man.Repo)

					// TODO: use shims for windows instead?
					err = sysutil.SymlinkBinToPath(execPath, pathToSymLinkTo)
					if err != nil {
						fmt.Println("error: could not symlink binary to PATH")
						continue
					}
				}
			}
			return nil
		},
	}

	updateCmd.Flags().BoolVarP(&strict, "strict", "s", false, "Only available on pre-release channels. Will only install pre-release versions and not stable releases, even if there exists a stable version more up-to-date than a pre-release.")

	return updateCmd
}
