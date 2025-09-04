package utils

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
)

func ContainsAny(src string, tokens []string) bool {
	for _, a := range tokens {
		if strings.Contains(src, a) {
			return true
		}
	}
	return false
}

func IsProcessRunning(execPath string) (bool, error) {
	absPath, err := filepath.Abs(execPath)
	if err != nil {
		return false, err
	}
	absPath = filepath.Clean(absPath)

	pses, err := process.Processes()
	if err != nil {
		return false, err
	}

	for _, p := range pses {
		procExe, err := p.Exe()
		if err != nil {
			continue
		}

		resolvedProcExe, err := filepath.EvalSymlinks(procExe)
		if err == nil {
			procExe = resolvedProcExe
		}

		procExe = filepath.Clean(procExe)

		isMatch := false

		if runtime.GOOS == "windows" {
			isMatch = strings.EqualFold(procExe, absPath)
		} else {
			isMatch = procExe == absPath
		}

		if isMatch {
			return true, nil
		}
	}

	return false, nil
}
