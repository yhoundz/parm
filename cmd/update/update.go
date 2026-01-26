/*
Copyright © 2025 Alexander Wang
*/
package update

import (
	"context"
	"errors"
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
	"strings"
	"time"

	"github.com/google/go-github/v74/github"
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

			token, tokenErr := gh.GetStoredApiKey(viper.GetViper())
			client := f.Provider(ctx, token).Repos()
			inst := installer.New(client, token)
			up := updater.New(client, inst)
			flags := updater.UpdateFlags{
				Strict: strict,
			}

			stats := struct {
				total       int
				updated     int
				upToDate    int
				skipped     int
				pinned      int
				noRelease   int
				failed      int
				rateLimited int
			}{}

			type failure struct {
				pkg string
				err error
			}
			var failures []failure
			var rateLimitErr *github.RateLimitError

			for _, pkg := range args {
				stats.total++
				// guaranteed to work now
				owner, repo, _ := cmdparser.ParseRepoRef(pkg)

				installPath := parmutil.GetInstallDir(owner, repo)
				parentDir, _ := sysutil.GetParentDir(installPath)

				man, err := manifest.Read(installPath)
				if err != nil {
					fmt.Printf("! skipped %s/%s: not installed correctly\n", owner, repo)
					stats.skipped++
					continue
				}

				if man.Pinned {
					fmt.Printf("- %s/%s is pinned to %s\n", owner, repo, man.Version)
					stats.pinned++
					continue
				}

				res, err := up.Update(ctx, owner, repo, installPath, man, &flags, nil)
				if err != nil {
					_ = parmutil.Cleanup(parentDir)
					stats.failed++

					if errors.As(err, &rateLimitErr) {
						stats.rateLimited++
						fmt.Printf("x failed %s/%s: rate limited\n", owner, repo)
					} else {
						fmt.Printf("x failed %s/%s\n", owner, repo)
					}
					failures = append(failures, failure{pkg: pkg, err: err})
					continue
				}

				switch res.Status {
				case updater.StatusUpToDate:
					fmt.Printf("✓ %s/%s is already up to date (%s)\n", owner, repo, man.Version)
					stats.upToDate++
					continue
				case updater.StatusNoRelease:
					fmt.Printf("? %s/%s: no releases found\n", owner, repo)
					stats.noRelease++
					continue
				case updater.StatusSkipped:
					fmt.Printf("! skipped %s/%s\n", owner, repo)
					stats.skipped++
					continue
				}

				// Write new manifest
				old := res.OldManifest
				man, err = manifest.New(owner, repo, res.Version, old.InstallType, res.InstallPath)
				// TODO: maybe set this pinned thing somewhere else
				man.Pinned = old.Pinned

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
				fmt.Printf("* updated %s/%s: %s -> %s\n", owner, repo, old.Version, res.Version)
				stats.updated++
			}

			if len(failures) > 0 {
				fmt.Println("\nFailures:")
				if stats.rateLimited > 0 {
					fmt.Printf("  - %d packages failed due to GitHub rate limiting", stats.rateLimited)
					if rateLimitErr != nil {
						fmt.Printf(" (reset in %v)", time.Until(rateLimitErr.Rate.Reset.Time))
					}
					fmt.Println()
					if tokenErr != nil {
						fmt.Println("    Tip: Set GITHUB_TOKEN to increase your rate limit.")
					} else {
						fmt.Println("    Tip: Your current token might be hitting limits or lacks necessary scopes.")
					}
				}

				for _, f := range failures {
					if !errors.As(f.err, &rateLimitErr) {
						if strings.Contains(f.err.Error(), "not found or private") {
							if tokenErr != nil {
								fmt.Printf("  - %s: %v (Tip: Try setting GITHUB_TOKEN for private repo access)\n", f.pkg, f.err)
							} else {
								fmt.Printf("  - %s: %v (Tip: Ensure your token is granted access to this repo; fine-grained tokens must explicitly include it)\n", f.pkg, f.err)
							}
						} else {
							fmt.Printf("  - %s: %v\n", f.pkg, f.err)
						}
					}
				}
			}

			if stats.total > 0 {
				fmt.Printf("\nSummary: %d total, %d updated, %d up-to-date", stats.total, stats.updated, stats.upToDate)
				if stats.pinned > 0 {
					fmt.Printf(", %d pinned", stats.pinned)
				}
				if stats.noRelease > 0 {
					fmt.Printf(", %d no release", stats.noRelease)
				}
				if stats.failed > 0 {
					fmt.Printf(", %d failed", stats.failed)
				}
				if stats.skipped > 0 {
					fmt.Printf(", %d skipped", stats.skipped)
				}
				fmt.Println()
			}
			return nil
		},
	}

	updateCmd.Flags().BoolVarP(&strict, "strict", "s", false, "Only available on pre-release channels. Will only install pre-release versions and not stable releases, even if there exists a stable version more up-to-date than a pre-release.")

	return updateCmd
}
