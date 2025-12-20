package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"parm/pkg/sysutil"
	"path/filepath"
	"strings"
)

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

		name := hdr.Name
		target, err := sysutil.SafeJoin(destPath, name)
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
			// Extract only permission bits from tar header mode
			mode := fs.FileMode(hdr.Mode) & 0o777
			// Ensure executables remain executable (at least 0o755 for owner)
			if mode&0o100 != 0 {
				mode |= 0o755
			} else if mode == 0 {
				mode = 0o644
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
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
			if _, err := sysutil.SafeJoin(destPath, cleanedTarget); err != nil {
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
		name := f.Name

		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		fpath, err := sysutil.SafeJoin(destPath, name)
		if err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0o755)
		} else {
			var fdir string
			if lastIndex := strings.LastIndex(fpath, string(os.PathSeparator)); lastIndex > -1 {
				fdir = fpath[:lastIndex]
			}

			err = os.MkdirAll(fdir, 0o755)
			if err != nil {
				return err
			}
			// Extract only permission bits and ensure executables remain executable
			mode := f.Mode() & 0o777
			if mode&0o100 != 0 {
				mode |= 0o755
			} else if mode == 0 {
				mode = 0o644
			}
			outFile, err := os.OpenFile(
				fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
