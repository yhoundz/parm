package sw

import (
	"fmt"
	"parm/internal/cmdutil"
	"parm/internal/core/switcher"
	"parm/internal/manifest"
	"parm/pkg/cmdparser"
	"strings"

	"github.com/spf13/cobra"
)

func NewChannelCmd(f *cmdutil.Factory) *cobra.Command {
	// switchCmd represents the switch command
	var channelCmd = &cobra.Command{
		Use:   "channel",
		Short: "A brief description of your command",
		Args:  cobra.ExactArgs(2),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			pkg := args[0]
			channel := strings.ToLower(args[1])
			// should guarantee to work after PreRunE
			owner, repo, _ := cmdparser.ParseRepoRef(pkg)
			var instType manifest.InstallType
			switch channel {
			case "release":
				instType = manifest.Release
			case "pre-release":
				instType = manifest.PreRelease
			default:
				return fmt.Errorf("error: invalid release channel")
			}
			err := switcher.SwitchChannel(owner, repo, instType)
			if err != nil {
				return err
			}
			return nil
		},
	}

	return channelCmd
}
