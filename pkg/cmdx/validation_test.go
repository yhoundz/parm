package cmdx

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestMarkFlagsRequireFlag(t *testing.T) {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			name:      "dependent flag without required flag should error",
			args:      []string{"--child"},
			expectErr: true,
		},
		{
			name:      "dependent flag with required flag should not error",
			args:      []string{"--parent", "--child"},
			expectErr: false,
		},
		{
			name:      "no flags should not error",
			args:      []string{},
			expectErr: false,
		},
		{
			name:      "only required flag should not error",
			args:      []string{"--parent"},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var parentFlag bool
			var childFlag bool

			cmd := &cobra.Command{
				Use: "test",
				RunE: func(cmd *cobra.Command, args []string) error {
					return MarkFlagsRequireFlag(cmd, "parent", "child")
				},
			}

			cmd.Flags().BoolVar(&parentFlag, "parent", false, "parent flag")
			cmd.Flags().BoolVar(&childFlag, "child", false, "child flag")

			cmd.SetArgs(tc.args)
			err := cmd.Execute()

			if tc.expectErr && err == nil {
				t.Errorf("Expected an error, but got nil")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("Expected no error, but got: %v", err)
			}
		})
	}
}
