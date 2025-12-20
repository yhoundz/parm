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

// CleanBinaryName strips common platform/architecture suffixes from binary names
// e.g., "nuclei-linux-amd64" -> "nuclei", "tool_darwin_arm64" -> "tool"
func CleanBinaryName(name string) string {
	if name == "" {
		return name
	}

	// Common platform identifiers to strip (order matters - check longer patterns first)
	platforms := []string{
		// Combined long patterns first
		"x86_64-unknown-linux-gnu", "x86_64-unknown-linux-musl",
		"x86_64-apple-darwin", "aarch64-apple-darwin",
		"x86_64-pc-windows-msvc", "x86_64-pc-windows-gnu",
		"unknown-linux-gnu", "unknown-linux-musl",
		"apple-darwin", "pc-windows",
		// Architecture names
		"amd64", "x86_64", "x64", "arm64", "aarch64", "386", "i386", "x86", "armv7", "armv6", "armhf",
		// OS names
		"linux", "darwin", "macos", "mac", "osx", "windows", "win", "win64", "win32",
	}

	result := name

	// Remove file extension first if present (but not for extensionless binaries)
	ext := filepath.Ext(result)
	if ext == ".exe" {
		result = strings.TrimSuffix(result, ext)
	}

	// Keep stripping platform suffixes until no more can be removed
	changed := true
	for changed {
		changed = false
		for _, p := range platforms {
			// Match patterns like -linux-amd64, _linux_amd64, -linux, _amd64, etc.
			suffixes := []string{
				"-" + p,
				"_" + p,
				"." + p,
			}
			for _, suffix := range suffixes {
				if strings.HasSuffix(strings.ToLower(result), strings.ToLower(suffix)) {
					newResult := result[:len(result)-len(suffix)]
					if newResult != "" { // Don't strip if it would result in empty string
						result = newResult
						changed = true
					}
				}
			}
		}
	}

	// If we stripped everything or result is empty, return original
	if result == "" {
		return name
	}

	// Restore .exe extension on Windows
	if runtime.GOOS == "windows" && ext == ".exe" {
		result += ext
	}

	return result
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

	if err := os.Symlink(binPath, destPath); err != nil {
		return err
	}

	return nil
}
