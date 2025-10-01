package deps

import (
	"bufio"
	"bytes"
	"context"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type binFile interface {
	ImportedLibraries() ([]string, error)
	io.Closer
}

func HasExternalDep(dep string) error {
	if _, err := exec.LookPath(dep); err != nil {
		return fmt.Errorf("error: dependency '%q' not found in PATH", dep)
	}
	return nil
}

/*
naive implementation of "ldd" on unix systems. may not be
completely accurate.
*/
func getMissingLibsLinux(ctx context.Context, path string) ([]string, error) {
	ldd := "ldd"
	if err := HasExternalDep(ldd); err != nil {
		return getMissingLibsFallBack(path)
	}

	cmd := exec.CommandContext(ctx, ldd, path)
	out, err := cmd.Output()
	if err != nil {
		return getMissingLibsFallBack(path)
	}

	var missingDeps []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.Contains(line, "=> not found") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				missingDeps = append(missingDeps, parts[0])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return getMissingLibsFallBack(path)
	}

	return missingDeps, nil
}

func getMissingLibsDarwin(ctx context.Context, path string) ([]string, error) {
	otool := "otool"
	if err := HasExternalDep(otool); err != nil {
		return getMissingLibsFallBack(path)
	}

	// cmd := exec.CommandContext(ctx, otool, "-L", path)
	// _, err := cmd.Output()
	// if err != nil {
	// 	return getMissingLibsFallBack(path)
	// }

	// WARNING: this functionality is not implemented yet for macOS
	return getMissingLibsFallBack(path)
}

func getMissingLibsFallBack(path string) ([]string, error) {
	return nil, nil
}

func hasSharedLib(name string) (bool, error) {
	var searchPaths []string
	switch runtime.GOOS {
	case "linux":
		// WARNING: only for 64-bit OSes for linux
		searchPaths = []string{
			"/usr/local/lib/x86_64-linux-gnu",
			"/lib/x86_64-linux-gnu",
			"/usr/lib/x86_64-linux-gnu",
			"/usr/lib/x86_64-linux-gnu64",
			"/usr/local/lib64",
			"/lib64",
			"/usr/lib64",
			"/usr/local/lib",
			"/lib",
			"/usr/lib",
			"/usr/x86_64-linux-gnu/lib64",
			"/usr/x86_64-linux-gnu/lib",
		}
		if env := os.Getenv("LD_LIBRARY_PATH"); env != "" {
			searchPaths = append(strings.Split(env, ":"), searchPaths...)
		}
	case "darwin":
		searchPaths = []string{
			"/usr/lib/",
			"/System/Library/Frameworks/",
			"/System/Library/PrivateFrameworks/",
			"/Library/Frameworks/",
			"/usr/local/lib/",
		}
		if env := os.Getenv("DYLD_LIBRARY_PATH"); env != "" {
			searchPaths = append(strings.Split(env, ":"), searchPaths...)
		}
	case "windows":
		return false, fmt.Errorf("warning: cannot check dependencies at this time")
	}
	return false, nil
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
		return nil, fmt.Errorf("error: failed to open binary: '%s': %w", path, err)
	}

	defer file.Close()

	libs, err := file.ImportedLibraries()
	if err != nil {
		return nil, fmt.Errorf("error: failed to get imported libs on %s: %w", path, err)
	}

	return libs, nil
}
