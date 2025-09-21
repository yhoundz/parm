/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package list

import (
	"fmt"
	"parm/internal/lister"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists out currently installed packages",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		list, data, err := lister.GetInstalledPkgInfo()
		if err != nil {
			return err
		}
		for _, pkg := range list {
			fmt.Println(pkg)
		}
		fmt.Printf("Total: %d packages installed.", data.NumPkgs)
		return nil
	},
}

func init() {}
