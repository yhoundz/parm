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

// uses magic numbers to determine if a file is a binary executable
func IsBinaryExecutable(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".exe") {
			return false, nil
		}
	} else { // on unix system
		if info.Mode()&0111 == 0 {
			return false, nil
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	const numBytes = 261
	hdr := make([]byte, numBytes)
	n, err := io.ReadAtLeast(file, hdr, 1)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return false, nil
	}
	if n == 0 {
		return false, nil
	}

	kind, _ := filetype.Match(hdr)
	if kind != types.Unknown {
		if kind.Extension == "elf" ||
			kind.Extension == "exe" ||
			kind.Extension == "macho" {
			return true, nil
		}
	}

	return false, nil
}
