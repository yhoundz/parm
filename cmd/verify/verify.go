/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package verify

import (
	"github.com/spf13/cobra"
)

var sha256 string

// verifyCmd represents the verify command
var VerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	VerifyCmd.Flags().StringVarP(&sha256, "sha256", "s", "", "Sha256 flag")
	VerifyCmd.MarkFlagRequired("sha256")
}
