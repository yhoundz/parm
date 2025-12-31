package sysutil

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

	var (
		osPattern      = `(?i)[-_](darwin|macos|linux|windows|win|freebsd|openbsd|netbsd|dragonfly|solaris|sunos|aix)`
		archPattern    = `(?i)[-_](arm64|aarch64|armv7|armv6|amd64|x86_64|x64|386|i386|i686|ppc64|ppc64le|s390x|riscv64)`
		abiPattern     = `(?i)[-_](gnu|musl|mingw|msvc|apple|unknown)`
		versionPattern = `(?i)[-_]v?[0-9]+\.[0-9]+(\.[0-9]+)?`
	)

	result := name
	ext := filepath.Ext(result)
	if strings.ToLower(ext) == ".exe" {
		result = strings.TrimSuffix(result, ext)
	}

	reVersionOSArch := regexp.MustCompile(versionPattern + osPattern + archPattern + `.*$`)
	reOSArch := regexp.MustCompile(osPattern + archPattern + `.*$`)
	reArchABI := regexp.MustCompile(archPattern + "(" + abiPattern + ")?$")
	reOSABI := regexp.MustCompile(osPattern + "(" + abiPattern + ")?$")
	reArch := regexp.MustCompile(archPattern + "$")
	reOS := regexp.MustCompile(osPattern + "$")
	reABI := regexp.MustCompile(abiPattern + "$")

	changed := true
	for changed {
		changed = false
		old := result

		// Strip patterns iteratively
		result = reVersionOSArch.ReplaceAllString(result, "")
		result = reOSArch.ReplaceAllString(result, "")
		result = reArchABI.ReplaceAllString(result, "")
		result = reOSABI.ReplaceAllString(result, "")
		result = reArch.ReplaceAllString(result, "")
		result = reOS.ReplaceAllString(result, "")
		result = reABI.ReplaceAllString(result, "")

		// Strip trailing separators
		result = strings.TrimRight(result, "-_")

		if result != old && result != "" {
			changed = true
		}
	}

	if result == "" {
		// Fallback to original if stripping made it empty (minus extension)
		result = strings.TrimSuffix(name, ext)
	}

	// Restore .exe extension on Windows if it was originally there
	if runtime.GOOS == "windows" && strings.ToLower(ext) == ".exe" {
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

	if runtime.GOOS == "windows" {
		if err := os.Link(binPath, destPath); err != nil {
			return err
		}
	} else {
		if err := os.Symlink(binPath, destPath); err != nil {
			return err
		}
	}

	return nil
}
