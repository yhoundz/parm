package utils

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode"
)

func ContainsAny(s string, tokens []string) bool {
	for _, t := range tokens {
		if t != "" && strings.Contains(s, strings.ToLower(t)) {
			return true
		}
	}
	return false
}

func ExtractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
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

		target, err := safeJoin(destDir, hdr.Name)
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
			// optional: handle symlinks; skip if not needed
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return err
			}
		default:
			// ignore other types
		}
	}
	return nil
}

func safeJoin(dir, name string) (string, error) {
	p := filepath.Join(dir, name)
	base := filepath.Clean(dir) + string(os.PathSeparator)
	if !strings.HasPrefix(filepath.Clean(p)+string(os.PathSeparator), base) {
		return "", fmt.Errorf("illegal path in archive: %q", name)
	}
	return p, nil
}

// TODO: put this elsewhere?
func SanitizePath(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '/' || r == '\\' || r == 0 {
			b.WriteRune('-')
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}

	name := b.String()

	if runtime.GOOS == "windows" {
		name = strings.TrimRight(name, ". ")
		if isWinReserved(name) {
			name = name + "-"
		}
	}

	if name == "" {
		name = hashPrefix(s, 12)
	}

	const maxLen = 38
	if len(name) > maxLen {
		name = name[:maxLen-13] + "-" + hashPrefix(s, 12)
	}

	name = filepath.Base(name)
	if name == "." || name == string(filepath.Separator) {
		name = hashPrefix(s, 12)
	}

	return name
}

func hashPrefix(s string, n int) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:])[:n]
}

func isWinReserved(name string) bool {
	n := strings.ToLower(name)
	switch n {
	case "con", "prn", "aux", "nul":
		return true
	}
	if strings.HasPrefix(n, "com") && len(n) == 4 && n[3] >= '1' && n[3] <= '9' {
		return true
	}
	if strings.HasPrefix(n, "lpt") && len(n) == 4 && n[3] >= '1' && n[3] <= '9' {
		return true
	}
	return false
}
