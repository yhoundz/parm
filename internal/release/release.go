package release

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"parm/pkg/archive"
	"parm/pkg/progress"

	"github.com/google/go-github/v74/github"
	"golang.org/x/exp/slices"
)

var errBinaryFound = errors.New("binary found")

// SelectAsset chooses the best matching release asset for the given OS/ARCH.
// It returns all assets that tie for the highest score so the caller can decide.
func SelectAsset(assets []*github.ReleaseAsset, goos, goarch string) ([]*github.ReleaseAsset, error) {
	if len(assets) == 0 {
		return nil, fmt.Errorf("no assets provided for %s/%s", goos, goarch)
	}

	type match struct {
		asset *github.ReleaseAsset
		score int
	}

	gooses := map[string][]string{
		"windows": {"windows", "win64", "win32", "win"},
		"darwin":  {"macos", "darwin", "mac", "osx"},
		"linux":   {"linux"},
	}
	goarchs := map[string][]string{
		"amd64": {"amd64", "x86_64", "x64", "64bit", "64-bit"},
		"386":   {"386", "x86", "i386", "32bit", "32-bit"},
		"arm64": {"arm64", "aarch64"},
		"arm":   {"armv7", "armv6", "armhf", "armv7l"},
	}
	goosTokens, ok := gooses[goos]
	if !ok {
		return nil, fmt.Errorf("unsupported os %q", goos)
	}
	goarchTokens, ok := goarchs[goarch]
	if !ok {
		return nil, fmt.Errorf("unsupported arch %q", goarch)
	}

	extPref := []string{".tar.gz", ".tgz", ".tar.xz", ".zip", ".bin", ".appimage"}
	if goos == "windows" {
		extPref = []string{".zip", ".exe", ".msi", ".bin"}
	}

	scoreMods := map[string]int{
		"musl": -1,
	}
	if goos == "windows" {
		scoreMods = map[string]int{}
	}

	scoredMatches := make([]match, 0, len(assets))
	for _, a := range assets {
		name := strings.ToLower(a.GetName())
		if !containsAny(name, goosTokens) {
			continue
		}
		if !containsAny(name, goarchTokens) {
			continue
		}
		scoredMatches = append(scoredMatches, match{asset: a})
	}
	if len(scoredMatches) == 0 {
		return nil, fmt.Errorf("no compatible assets found for %s/%s", goos, goarch)
	}

	const goosMatch = 11
	const goarchMatch = 7
	const prefMatch = 3

	for i := range scoredMatches {
		entry := &scoredMatches[i]
		name := strings.ToLower(entry.asset.GetName())

		if containsAny(name, goosTokens) {
			entry.score += goosMatch
		}
		if containsAny(name, goarchTokens) {
			entry.score += goarchMatch
		}

		for j, ext := range extPref {
			mult := float64(prefMatch) * float64(len(extPref)-j)
			if strings.HasSuffix(name, ext) {
				entry.score += int(math.Round(mult))
			}
		}

		for substr, mod := range scoreMods {
			if strings.Contains(name, substr) {
				entry.score += mod
			}
		}
	}

	slices.SortStableFunc(scoredMatches, func(a, b match) int {
		if a.score < b.score {
			return 1
		}
		if a.score > b.score {
			return -1
		}
		return 0
	})

	bestScore := scoredMatches[0].score
	candidates := []*github.ReleaseAsset{}
	for _, m := range scoredMatches {
		if m.score == bestScore {
			candidates = append(candidates, m.asset)
			continue
		}
		break
	}

	return candidates, nil
}

// DownloadAsset downloads a release asset to a local path, optionally using a GitHub token and progress hooks.
func DownloadAsset(ctx context.Context, client *http.Client, url, destPath, token string, hooks *progress.Hooks) error {
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/octet-stream")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}

	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := io.Reader(resp.Body)
	var closer io.Closer
	if hooks != nil && hooks.Decorator != nil {
		wrapped := hooks.Decorator(progress.StageDownload, resp.Body, resp.ContentLength)
		if rc, ok := wrapped.(io.ReadCloser); ok {
			reader = rc
			closer = rc
		} else {
			reader = wrapped
		}
	}

	if _, err := io.Copy(file, reader); err != nil {
		if closer != nil {
			_ = closer.Close()
		}
		return err
	}

	if closer != nil {
		_ = closer.Close()
	}

	return nil
}

// IsArchive reports whether a filename looks like a supported archive.
func IsArchive(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, ".tar.gz") ||
		strings.HasSuffix(lower, ".tgz") ||
		strings.HasSuffix(lower, ".zip")
}

// ExtractArchive extracts the contents of .tar.gz, .tgz, or .zip files into destDir.
func ExtractArchive(archivePath, destDir string) error {
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return archive.ExtractZip(archivePath, destDir)
	default:
		return archive.ExtractTarGz(archivePath, destDir)
	}
}

// FindBinary walks destDir looking for the named binary and returns the path once found.
func FindBinary(root, binaryName string) (string, error) {
	if binaryName == "" {
		return "", errors.New("binary name is empty")
	}

	var binaryPath string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == binaryName {
			binaryPath = path
			return errBinaryFound
		}
		return nil
	})

	if errors.Is(err, errBinaryFound) {
		err = nil
	}
	if err != nil {
		return "", err
	}
	if binaryPath == "" {
		return "", fmt.Errorf("binary %s not found in %s", binaryName, root)
	}
	return binaryPath, nil
}

func containsAny(src string, tokens []string) bool {
	for _, token := range tokens {
		if strings.Contains(src, token) {
			return true
		}
	}
	return false
}
