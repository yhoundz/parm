/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package configure

import (
	"fmt"

	"github.com/spf13/cobra"
	// "github.com/spf13/viper"
)

var ConfigCmd = &cobra.Command{
	Use:     "config",
	Aliases: []string{"configure, cfg"},
	Short:   "Configures parm.",
	Long:    `Prints the current configuration settings to your console.`,
	// INFO: also allow users to set config keys using flags like --set key=value, as well as edit the file directly
	// TODO:
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("config called")
		return nil
	},
}

func init() {

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
