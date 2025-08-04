package deps

import (
	"fmt"
	"os/exec"
)

func Require(dep string) error {
	if _, err := exec.LookPath(dep); err != nil {
		return fmt.Errorf("fatal: Required dependency '%q' not found in PATH", dep)
	}
	return nil
}
