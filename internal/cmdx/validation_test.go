package cmdx

import (
	"testing"

	"github.com/spf13/cobra"
)

var parentFlag bool
var childFlag bool

func TestMarkFlagsRequireFlag(t *testing.T) {
	tc := []struct {
		name   string
		args   []string
		expErr bool
	}{}
}
