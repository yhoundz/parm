/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package configure

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var setValues map[string]string
var resetKey string

var ConfigCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"configure, cfg"},
	Short:   "Configures parm.",
	Long:    `Prints the current configuration settings to your console.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(setValues) > 0 {
			for k, v := range setValues {
				viper.Set(k, v)
				fmt.Printf("Set %s = %s\n", k, v)
			}

			if err := viper.WriteConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); ok {
					if err = viper.SafeWriteConfig(); err != nil {
						return fmt.Errorf("error: failed to create config file: %w", err)
					}
				} else {
					return fmt.Errorf("error: failed to write config file: %w", err)
				}
			}
			return nil
		}

		settings := viper.AllSettings()
		for k, v := range settings {
			fmt.Printf("%s: %s\n", k, v)
		}

		return nil
	},
}

func init() {
	ConfigCmd.PersistentFlags().StringToStringVarP(&setValues, "set", "s", nil, "Set config k/v pairs (e.g. --set key=value)")
	ConfigCmd.PersistentFlags().StringVarP(&resetKey, "reset", "r", "", "Reset config key back to its default value.")
	ConfigCmd.MarkFlagsMutuallyExclusive("reset", "set")
}
