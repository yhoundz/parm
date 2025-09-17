package utils

import (
	"fmt"
	"io"
	"os"
	"parm/internal/config"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

func MoveAllFrom(src, dest string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, e := range entries {
		oldPath := filepath.Join(src, e.Name())
		newPath := filepath.Join(dest, e.Name())

		if err := os.Rename(oldPath, newPath); err != nil {
			return err
		}
	}
	return nil
}

func safeJoin(root, name string) (string, error) {
	cleaned := filepath.Clean(name)
	target := filepath.Join(root, cleaned)

	root = filepath.Clean(root) + string(os.PathSeparator)
	if !strings.HasPrefix(target, root) {
		return "", fmt.Errorf("tar entry %q escapes extraction dir", name)
	}
	return target, nil
}

func stripTopDir(path string) (string, bool) {
	if i := strings.IndexByte(path, '/'); i >= 0 {
		out := path[i+1:]
		if out == "" {
			return "", false
		}
		return out, true
	}
	return "", false
}

func GetInstallDir(owner, repo string) string {
	installPath := config.Cfg.ParmPkgDirPath
	dir := owner + "-" + repo
	dest := filepath.Join(installPath, dir)
	return dest
}

// checks if a file is a binary executable, and then checks if it's even able to be run by the user.
func IsValidBinaryExecutable(path string) (bool, error) {
	kind, err := IsBinaryExecutable(path)
	if err != nil {
		return false, err
	}
	if kind == nil {
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
func IsBinaryExecutable(path string) (*types.Type, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// precheck
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".exe") {
			return nil, nil
		}
	} else { // on unix system
		if info.Mode()&0111 == 0 {
			return nil, nil
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	const numBytes = 261
	hdr := make([]byte, numBytes)
	n, err := io.ReadAtLeast(file, hdr, 1)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, nil
	}
	if n == 0 {
		return nil, nil
	}

	kind, _ := filetype.Match(hdr)
	if kind != types.Unknown {
		if kind.Extension == "elf" ||
			kind.Extension == "exe" ||
			kind.Extension == "macho" {
			return &kind, nil
		}
	}

	return nil, nil
}

func SymlinkBinToPath(binPath, destPath string) (string, error) {
	isBin, err := IsValidBinaryExecutable(binPath)
	if err != nil {
		return "", err
	}
	if isBin {
		return "", fmt.Errorf("error: provided dir is not a binary")
	}

	absBinDir, err := filepath.Abs(binPath)
	if err != nil {
		return "", err
	}

	newDestPath := filepath.Join(destPath, filepath.Base(absBinDir))

	if _, err := os.Lstat(newDestPath); err != nil {
		if err := os.Remove(newDestPath); err != nil {
			return "", fmt.Errorf("failed to remove existing symlink at %w", err)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to check destination path: %w", err)
	}

	if err := os.Symlink(absBinDir, newDestPath); err != nil {
		return "", err
	}

	return "", nil
}
