/*
Copyright © 2025 Alexander Wang
*/
package configure

import (
	"fmt"
	"parm/pkg/cmdparser"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setCmd represents the set command
var SetCmd = &cobra.Command{
	Use:   "set key=value",
	Short: "Sets a key/value pair in the config",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, val := range args {
			k, v, err := cmdparser.StringToString(val)
			if err != nil {
				return err
			}
			viper.Set(k, v)
			fmt.Printf("Set %s = %s\n", k, v)
		}

		if err := viper.WriteConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				if err = viper.SafeWriteConfig(); err != nil {
					return fmt.Errorf("error: failed to create config file: \n%w", err)
				}
			} else {
				return fmt.Errorf("error: failed to write config file: \n%w", err)
			}
		}
		return nil
	},
}

func init() {}
