package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"parm/internal/config"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
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

func ExtractTarGz(srcPath, destPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// INFO: assumes that the resulting tar.gz will contain a single folder that holds the source code
		// TODO: make this more robust?
		name, ok := stripTopDir(hdr.Name)
		if !ok {
			continue
		}

		target, err := safeJoin(destPath, name)
		if err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, fs.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fs.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			linkTarget := hdr.Linkname
			cleanedTarget := filepath.Clean(filepath.Join(filepath.Dir(name), linkTarget))
			if _, err := safeJoin(destPath, cleanedTarget); err != nil {
				return err
			}
			_ = os.Symlink(linkTarget, target)
		default:
			// nothing?
		}
	}
	return nil
}

func ExtractZip(srcPath, destPath string) error {
	r, err := zip.OpenReader(srcPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		name, ok := stripTopDir(f.Name)
		if !ok {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath, err := safeJoin(destPath, name)
		if err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, f.Mode())
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, f.Mode())
			if err != nil {
				return err
			}
			f, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

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
