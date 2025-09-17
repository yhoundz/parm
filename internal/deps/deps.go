package deps

import (
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"io"
	"os/exec"
	"runtime"
)

type binFile interface {
	ImportedLibraries() ([]string, error)
	io.Closer
}

func HasDep(dep string) (bool, error) {
	if _, err := exec.LookPath(dep); err != nil {
		return false, fmt.Errorf("fatal: Required dependency '%q' not found in PATH", dep)
	}
	return true, nil
}

func GetBinDeps(path string) ([]string, error) {
	var file binFile
	var err error

	switch runtime.GOOS {
	case "windows":
		file, err = pe.Open(path)
	case "darwin":
		file, err = macho.Open(path)
	case "linux":
		file, err = elf.Open(path)
	default:
		return nil, fmt.Errorf("error: unsupported system")
	}

	if err != nil {
		return nil, err
	}
	libs, err := file.ImportedLibraries()
	if err != nil {
		return nil, err
	}
	if len(libs) == 0 {
		return nil, fmt.Errorf("no libraries were found")
	}

	return libs, nil
}
