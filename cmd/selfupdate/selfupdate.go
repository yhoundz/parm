/*
Copyright © 2025 Alexander Wang
*/
package selfupdate

import (
	"fmt"
	"os"

	"parm/internal/cmdutil"
	selfupdatepkg "parm/internal/selfupdate"

	"github.com/spf13/cobra"
)

func NewSelfUpdateCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "self-update",
		Short: "Update parm to the latest version",
		Long:  `Update the parm binary to the latest stable release.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := selfupdatepkg.Run(os.Stdout, os.Stderr); err != nil {
				return fmt.Errorf("self-update failed: %w", err)
			}
			return nil
		},
	}
	return cmd
}
