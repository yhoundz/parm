/*
Copyright © 2025 yhoundz
*/
package list

import (
	"fmt"
	"parm/internal/cmdutil"
	"parm/internal/core/catalog"

	"github.com/spf13/cobra"
)

func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists out currently installed packages",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			list, data, err := catalog.GetInstalledPkgInfo()
			if err != nil {
				return err
			}
			for _, pkg := range list {
				fmt.Println(pkg)
			}
			fmt.Printf("Total: %d packages installed.\n", data.NumPkgs)
			return nil
		},
	}

	return listCmd
}
