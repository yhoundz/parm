/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package configure

import (
	"fmt"
	"parm/internal/config"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var resetAll bool

// resetCmd represents the reset command
var ResetCmd = &cobra.Command{
	Use:   "reset <key>",
	Short: "Resets key/value pairs to their default value.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		var cfgMap map[string]any
		if err := mapstructure.Decode(config.DefaultCfg, &cfgMap); err != nil {
			return err
		}
		for _, arg := range args {
			def, ok := cfgMap[arg]
			if !ok {
				return fmt.Errorf("error: %s is not a valid configuration key", arg)
			}

			viper.Set(arg, def)
			fmt.Printf("Reset %s to default value: %v\n", arg, def)
			if err := viper.WriteConfig(); err != nil {
				return fmt.Errorf("error: failed to write config file: %w", err)
			}
		}
		return nil
	},
	Args: func(cmd *cobra.Command, args []string) error {
		resetAllFlag, err := cmd.Flags().GetBool("all")
		if err != nil {
			return err
		}

		if resetAllFlag && len(args) > 0 {
			return fmt.Errorf("error: no arguments accepted when using the --all flag.")
		}
		return nil
	},
}

func init() {
	ResetCmd.PersistentFlags().BoolVarP(&resetAll, "all", "a", false, "Resets all config values to their defaults.")
}
