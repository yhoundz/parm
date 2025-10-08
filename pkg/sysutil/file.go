package sysutil

import (
	"fmt"
	"io"
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
	pth, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return filepath.Dir(pth), nil
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

	// precheck
	if runtime.GOOS == "windows" {
		// TODO: check for .msi files?
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".exe") {
			return false, nil, nil
		}
	} else { // on unix system
		if info.Mode()&0111 == 0 {
			return false, nil, nil
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return false, nil, err
	}
	defer file.Close()

	const numBytes = 261
	hdr := make([]byte, numBytes)
	n, err := io.ReadAtLeast(file, hdr, 1)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return false, nil, nil
	}
	if n == 0 {
		return false, nil, nil
	}

	kind, _ := filetype.Match(hdr)
	if kind != types.Unknown {
		if kind.Extension == "elf" ||
			kind.Extension == "exe" ||
			kind.Extension == "macho" {
			return true, &kind, nil
		}
	}

	return false, nil, nil
}

func SymlinkBinToPath(binPath, destPath string) (string, error) {
	isBin, err := IsValidBinaryExecutable(binPath)
	if err != nil {
		return "", err
	}
	if !isBin {
		return "", fmt.Errorf("error: provided dir is not a binary")
	}

	absBinDir, err := filepath.Abs(binPath)
	if err != nil {
		return "", err
	}

	newDestPath, err := filepath.Abs(destPath)
	if err != nil {
		return "", err
	}

	if _, err := os.Lstat(newDestPath); err == nil {
		if err := os.Remove(newDestPath); err != nil {
			return "", fmt.Errorf("failed to remove existing symlink at %s\n%w", newDestPath, err)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check destination path: \n%w", err)
	}

	if err := os.Symlink(absBinDir, newDestPath); err != nil {
		return "", err
	}

	return newDestPath, nil
}
