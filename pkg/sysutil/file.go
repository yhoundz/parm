package sysutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

func SafeJoin(root, name string) (string, error) {
	cleaned := filepath.Clean(name)
	target := filepath.Join(root, cleaned)

	root = filepath.Clean(root) + string(os.PathSeparator)
	if !strings.HasPrefix(target, root) {
		return "", fmt.Errorf("tar entry %q escapes extraction dir", name)
	}
	return target, nil
}

func GetParentDir(path string) (string, error) {
	abs := filepath.IsAbs(path)
	if !abs {
		return "", fmt.Errorf("filepath must be absolute")
	}

	return filepath.Dir(path), nil
}

// checks if a file is a binary executable, and then checks if it's even able to be run by the user.
func IsValidBinaryExecutable(path string) (bool, error) {
	isBin, kind, err := IsBinaryExecutable(path)
	if err != nil {
		return false, err
	}
	if kind == nil || !isBin {
		return false, err
	}

	switch runtime.GOOS {
	case "windows":
		return kind.Extension == "exe", nil
	case "darwin":
		return kind.Extension == "macho", nil
	case "linux":
		return kind.Extension == "elf", nil
	}

	return false, nil
}

// uses magic numbers to determine if a file is a binary executable
func IsBinaryExecutable(path string) (bool, *types.Type, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, nil, err
	}

	if info.IsDir() {
		return false, nil, nil
	}

	// precheck for Windows only - requires .exe extension
	if runtime.GOOS == "windows" {
		// TODO: check for .msi files?
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".exe") {
			return false, nil, nil
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return false, nil, err
	}
	defer file.Close()

	buf, err := os.ReadFile(path)
	if err != nil {
		return false, nil, err
	}

	kind, _ := filetype.Match(buf)
	if kind != types.Unknown {
		if kind.Extension == "elf" ||
			kind.Extension == "exe" ||
			kind.Extension == "macho" {
			return true, &kind, nil
		}
	}

	return false, nil, nil
}

func SymlinkBinToPath(binPath, destPath string) error {
	isBin, err := IsValidBinaryExecutable(binPath)
	if err != nil {
		return err
	}
	if !isBin {
		return fmt.Errorf("error: provided dir is not a binary")
	}

	// Ensure the binary has execute permissions
	if runtime.GOOS != "windows" {
		if err := os.Chmod(binPath, 0o755); err != nil {
			return fmt.Errorf("failed to set execute permission on %s\n%w", binPath, err)
		}
	}

	if _, err := os.Lstat(destPath); err == nil {
		if err := os.Remove(destPath); err != nil {
			return fmt.Errorf("failed to remove existing symlink at %s\n%w", destPath, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check destination path: \n%w", err)
	}

	if runtime.GOOS == "windows" {
		shimContent := fmt.Sprintf("@echo off\n\"%s\" %%*\n", binPath)
		if err := os.WriteFile(destPath, []byte(shimContent), 0o755); err != nil {
			return fmt.Errorf("failed to create windows shim at %s\n%w", destPath, err)
		}
	} else {
		if err := os.Symlink(binPath, destPath); err != nil {
			return err
		}
	}

	return nil
}
