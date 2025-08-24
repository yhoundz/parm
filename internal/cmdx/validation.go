package cmdx

import (
	"fmt"

	"github.com/spf13/cobra"
)

func MarkFlagsRequireFlag(cmd *cobra.Command, reqFlag string, depFlags ...string) error {
	reqFlagSet := cmd.Flags().Changed(reqFlag)
	if !reqFlagSet {
		for _, depFlag := range depFlags {
			if cmd.Flags().Changed(depFlag) {
				return fmt.Errorf("flag --%s is only valid when used with --%s", depFlag, reqFlag)
			}
		}
	}
	return nil
}
