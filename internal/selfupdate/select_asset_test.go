package selfupdate

import (
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
			name: "Ambiguous matches - chooses highest score",
			assets: []string{
				"parm-linux-amd64-debian.tar.gz",
				"parm-linux-amd64.tar.gz",
			},
			goos:     "linux",
			goarch:   "amd64",
			wantName: "parm-linux-amd64-debian.tar.gz", // Both match equally in score (10+5), first one wins
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
			for _, n := range tt.assets {
				name := n
				assets = append(assets, &github.ReleaseAsset{
					Name: &name,
				})
			}

			got, err := selectAsset(assets, tt.goos, tt.goarch)
			if (err != nil) != tt.wantErr {
				t.Fatalf("selectAsset() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if got.GetName() != tt.wantName {
					t.Errorf("selectAsset() got = %v, want %v", got.GetName(), tt.wantName)
				}
			}
		})
	}
}
