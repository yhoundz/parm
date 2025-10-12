/*
Copyright © 2025 Alexander Wang
*/
package update

import (
	"fmt"
	"parm/internal/core/installer"
	"parm/internal/core/updater"
	"parm/internal/gh"
	"parm/internal/manifest"
	"parm/internal/parmutil"
	"parm/pkg/cmdparser"
	"parm/pkg/sysutil"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var strict bool

// updateCmd represents the update command
var UpdateCmd = &cobra.Command{
	Use:   "update <owner>/<repo>",
	Short: "Updates a package",
	Long:  `Updates a package to the latest available version.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		for i, arg := range args {
			owner, repo, err := cmdparser.ParseRepoRef(arg)
			if err != nil {
				owner, repo, err = cmdparser.ParseGithubUrlPattern(arg)
				if err != nil {
					args = slices.Delete(args, i, i+1)
					fmt.Printf("error: package %s not found, skipping...\n", arg)
					continue
				}
				args[i] = fmt.Sprintf("%s/%s", owner, repo)
			}
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		token, err := gh.GetStoredApiKey(viper.GetViper())
		if err != nil {
			fmt.Printf("%s\ncontinuing without api key.\n", err)
		}
		client := gh.New(ctx, token).Repos()
		inst := installer.New(client)
		up := updater.New(client, inst)
		updated := make(map[string]bool)
		flags := updater.UpdateFlags{
			Strict: strict,
		}

		for _, pkg := range args {
			if _, ok := updated[pkg]; ok {
				// already updated package
				continue
			}
			updated[pkg] = true

			// guaranteed to work now
			owner, repo, _ := cmdparser.ParseRepoRef(pkg)

			installPath := parmutil.GetInstallDir(owner, repo)
			parentDir, _ := sysutil.GetParentDir(installPath)

			res, err := up.Update(ctx, owner, repo, installPath, &flags, nil)
			if err != nil {
				_ = parmutil.Cleanup(parentDir)
				fmt.Printf("error: failed to update %s/%s\n", owner, repo)
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

func init() {
	UpdateCmd.Flags().BoolVarP(&strict, "strict", "s", false, "Only available on pre-release channels. Will only install pre-release versions and not stable releases, even if there exists a stable version more up-to-date than a pre-release.")
}
