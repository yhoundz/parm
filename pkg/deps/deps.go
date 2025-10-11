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
	"path/filepath"
	"regexp"
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

func GetMissingLibs(ctx context.Context, binPath string) ([]string, error) {
	if _, err := os.Stat(binPath); err != nil {
		return nil, err
	}

	switch runtime.GOOS {
	case "windows":
		return nil, nil
	case "linux":
		return getMissingLibsLinux(ctx, binPath)
	case "darwin":
		return getMissingLibsDarwin(ctx, binPath)
	default:
		return getMissingLibsFallBack(binPath)
	}
}

// uses objdump to find dynamically linked libs
func getMissingLibsLinux(ctx context.Context, binPath string) ([]string, error) {
	objdump := "objdump"
	if err := HasExternalDep(objdump); err != nil {
		return getMissingLibsFallBack(binPath)
	}

	out, err := exec.CommandContext(ctx, objdump, "-p", "--", binPath).Output()
	if err != nil {
		return getMissingLibsFallBack(binPath)
	}

	var deps []string
	reg := regexp.MustCompile(`^\s*NEEDED\s+(.+)$`)
	r := bytes.NewReader(out)
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if match := reg.FindStringSubmatch(line); len(match) == 2 {
			trim := strings.TrimSpace(match[1])
			deps = append(deps, trim)
		}
	}

	return deps, nil
}

// uses otool to find dynamically linked libs
func getMissingLibsDarwin(ctx context.Context, binPath string) ([]string, error) {
	otool := "otool"
	if err := HasExternalDep(otool); err != nil {
		return getMissingLibsFallBack(binPath)
	}

	cmd := exec.CommandContext(ctx, otool, "-L", binPath)
	out, err := cmd.Output()
	if err != nil {
		return getMissingLibsFallBack(binPath)
	}

	var deps []string
	r := bytes.NewReader(out)
	sc := bufio.NewScanner(r)
	first := true
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if first {
			first = false
			continue
		}

		if line == "" {
			continue
		}

		// skip fat headers
		isFat := strings.HasSuffix(line, ":") && strings.Contains(line, "(architecture")
		if isFat {
			continue
		}

		if i := strings.Index(line, " ("); i > 0 {
			name := strings.TrimSpace(line[:i])
			deps = append(deps, name)
		}
	}

	return deps, nil
}

func getMissingLibsFallBack(path string) ([]string, error) {
	var libs []string
	deps, err := GetBinDeps(path)
	if err != nil {
		return nil, err
	}

	for _, dep := range deps {
		hasLib, err := hasSharedLib(dep)
		if err != nil {
			continue
		}
		if !hasLib {
			libs = append(libs, dep)
		}
	}

	return libs, nil
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

	for _, dir := range searchPaths {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return true, nil
		}
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
		return nil, fmt.Errorf("error: failed to open binary: '%s': \n%w", path, err)
	}

	defer file.Close()

	libs, err := file.ImportedLibraries()
	if err != nil {
		return nil, fmt.Errorf("error: failed to get imported libs on %s: \n%w", path, err)
	}

	return libs, nil
}
