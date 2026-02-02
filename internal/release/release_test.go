package release

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-github/v74/github"
)

func TestSelectAsset(t *testing.T) {
	tests := []struct {
		name     string
		assets   []string
		goos     string
		goarch   string
		wantName string
		wantErr  bool
	}{
		{
			name: "Linux AMD64",
			assets: []string{
				"parm-linux-amd64.tar.gz",
				"parm-linux-arm64.tar.gz",
				"parm-darwin-amd64.tar.gz",
			},
			goos:     "linux",
			goarch:   "amd64",
			wantName: "parm-linux-amd64.tar.gz",
		},
		{
			name: "Linux ARM64",
			assets: []string{
				"parm-linux-amd64.tar.gz",
				"parm-linux-arm64.tar.gz",
				"parm-linux-armv7.tar.gz",
			},
			goos:     "linux",
			goarch:   "arm64",
			wantName: "parm-linux-arm64.tar.gz",
		},
		{
			name: "Linux ARMv7",
			assets: []string{
				"parm-linux-amd64.tar.gz",
				"parm-linux-arm64.tar.gz",
				"parm-linux-armv7.tar.gz",
			},
			goos:     "linux",
			goarch:   "arm",
			wantName: "parm-linux-armv7.tar.gz",
		},
		{
			name: "macOS ARM64 (Apple Silicon)",
			assets: []string{
				"parm-darwin-amd64.tar.gz",
				"parm-darwin-arm64.tar.gz",
			},
			goos:     "darwin",
			goarch:   "arm64",
			wantName: "parm-darwin-arm64.tar.gz",
		},
		{
			name: "macOS AMD64 (Intel)",
			assets: []string{
				"parm-darwin-amd64.tar.gz",
				"parm-darwin-arm64.tar.gz",
			},
			goos:     "darwin",
			goarch:   "amd64",
			wantName: "parm-darwin-amd64.tar.gz",
		},
		{
			name: "Windows AMD64",
			assets: []string{
				"parm-windows-amd64.zip",
				"parm-linux-amd64.tar.gz",
			},
			goos:     "windows",
			goarch:   "amd64",
			wantName: "parm-windows-amd64.zip",
		},
		{
			name: "Common Aliases - x86_64",
			assets: []string{
				"tool-linux-x86_64.tar.gz",
			},
			goos:     "linux",
			goarch:   "amd64",
			wantName: "tool-linux-x86_64.tar.gz",
		},
		{
			name: "Common Aliases - aarch64",
			assets: []string{
				"tool-linux-aarch64.tar.gz",
			},
			goos:     "linux",
			goarch:   "arm64",
			wantName: "tool-linux-aarch64.tar.gz",
		},
		{
			name: "ARM variant edge case - armhf",
			assets: []string{
				"tool-linux-armhf.tar.gz",
			},
			goos:     "linux",
			goarch:   "arm",
			wantName: "tool-linux-armhf.tar.gz",
		},
		{
			name: "Ambiguous matches - chooses longest name",
			assets: []string{
				"parm-linux-amd64-debian.tar.gz",
				"parm-linux-amd64.tar.gz",
			},
			goos:     "linux",
			goarch:   "amd64",
			wantName: "parm-linux-amd64-debian.tar.gz",
		},
		{
			name: "No match - OS mismatch",
			assets: []string{
				"parm-windows-amd64.zip",
			},
			goos:    "linux",
			goarch:  "amd64",
			wantErr: true,
		},
		{
			name: "No match - Arch mismatch",
			assets: []string{
				"parm-linux-arm64.tar.gz",
			},
			goos:    "linux",
			goarch:  "amd64",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var assets []*github.ReleaseAsset
			for _, name := range tt.assets {
				n := name
				assets = append(assets, &github.ReleaseAsset{Name: &n})
			}

			got, err := SelectAsset(assets, tt.goos, tt.goarch)
			if (err != nil) != tt.wantErr {
				t.Fatalf("SelectAsset() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) == 0 {
				t.Fatalf("SelectAsset() returned no candidates")
			}
			if got[0].GetName() != tt.wantName {
				t.Fatalf("SelectAsset() got = %v, want %v", got[0].GetName(), tt.wantName)
			}
		})
	}
}

func TestFindBinary(t *testing.T) {
	tmpDir := t.TempDir()

	binaryName := "parm"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	nested := filepath.Join(tmpDir, "nested", "bin")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	path := filepath.Join(nested, binaryName)
	if err := os.WriteFile(path, []byte("binary"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}

	found, err := FindBinary(tmpDir, binaryName)
	if err != nil {
		t.Fatalf("FindBinary() error: %v", err)
	}
	if found != path {
		t.Fatalf("FindBinary() found %s, want %s", found, path)
	}
}

func TestFindBinaryNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "other.txt"), []byte("data"), 0o644); err != nil {
		t.Fatalf("write helper: %v", err)
	}

	binaryName := "parm"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	_, err := FindBinary(tmpDir, binaryName)
	if err == nil {
		t.Fatal("expected error when binary is missing")
	}
}
