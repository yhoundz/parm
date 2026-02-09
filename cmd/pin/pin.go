/*
Copyright © 2025 Alexander Wang
*/
package pin

import (
	"fmt"
	"parm/internal/cmdutil"
	"parm/internal/core/updater"
	"parm/pkg/cmdparser"

	"github.com/spf13/cobra"
)

func NewPinCmd(f *cmdutil.Factory) *cobra.Command {
	var pinCmd = &cobra.Command{
		Use:   "pin <owner>/<repo>",
		Short: "Pins a package to prevent updates.",
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, pkg := range args {
				owner, repo, err := cmdparser.ParseRepoRef(pkg)
				if err != nil {
					owner, repo, err = cmdparser.ParseGithubUrlPattern(pkg)
					if err != nil {
						return err
					}
				}

				if err != nil {
					fmt.Printf("cannot read manifest for %s/%s", owner, repo)
					continue
				}

				err, ver := updater.ChangePinnedStatus(owner, repo, true)
				if err != nil {
					fmt.Printf("unable to update pinned status for %s/%s", owner, repo)
					continue
				}

				fmt.Printf("Successfully pinned %s/%s to version %s", owner, repo, ver)
			}
			return nil
		},
	}

	return pinCmd
}
