/*
Copyright © 2025 Alexander Wang
*/
package configure

import (
	"fmt"
	"maps"
	"parm/internal/config"
	"slices"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"configure, cfg"},
	Short:   "Configures parm.",
	Long:    `Prints the current configuration settings to your console.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var settings map[string]any
		if err := mapstructure.Decode(config.Cfg, &settings); err != nil {
			return err
		}
		sorted := slices.Sorted(maps.Keys(settings))
		for _, k := range sorted {
			fmt.Printf("%s: %s\n", k, settings[k])
		}

		return nil
	},
}

func init() {
	ConfigCmd.AddCommand(SetCmd)
	ConfigCmd.AddCommand(ResetCmd)
}
