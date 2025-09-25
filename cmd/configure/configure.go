/*
Copyright © 2025 Alexander Wang
*/
package configure

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ConfigCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"configure, cfg"},
	Short:   "Configures parm.",
	Long:    `Prints the current configuration settings to your console.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		settings := viper.AllSettings()
		for k, v := range settings {
			fmt.Printf("%s: %s\n", k, v)
		}

		return nil
	},
}

func init() {
	ConfigCmd.AddCommand(SetCmd)
}
