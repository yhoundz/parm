/*
Copyright © 2025 Alexander Wang
*/
package selfupdate

import (
	"context"
	"fmt"
	"os"

	"parm/internal/cmdutil"
	"parm/internal/selfupdate"
	"parm/parmver"

	"github.com/spf13/cobra"
)

func NewSelfUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "self-update",
		Short: "Update parm to the latest version",
		Long:  `Update the parm binary to the latest stable release.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			owner := parmver.Owner
			if owner == "" {
				owner = "yhoundz"
			}
			repo := parmver.Repo
			if repo == "" {
				repo = "parm"
			}

			if err := selfupdate.Update(context.Background(), selfupdate.Config{
				Owner:          owner,
				Repo:           repo,
				Binary:         "parm",
				CurrentVersion: parmver.StringVersion,
			}, os.Stdout, os.Stderr); err != nil {
				return fmt.Errorf("self-update failed: %w", err)
			}
			return nil
		},
	}
	return cmd
}
