/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package verify

import (
	"fmt"

	"github.com/spf13/cobra"
)

// verifyCmd represents the verify command
var VerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("verify called")
	},
}

func init() {}
