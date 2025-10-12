package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetDefaultPrefixDir(t *testing.T) {
	// Save original env vars
	origXDG := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", origXDG)

	tests := []struct {
		name          string
		goos          string
		xdgDataHome   string
		shouldContain string
	}{
		{
			name:          "Linux with XDG_DATA_HOME",
			goos:          "linux",
			xdgDataHome:   "/custom/data",
			shouldContain: "parm",
		},
		{
			name:          "Linux without XDG_DATA_HOME",
			goos:          "linux",
			xdgDataHome:   "",
			shouldContain: "parm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.goos != runtime.GOOS {
				t.Skipf("Skipping test for %s on %s", tt.goos, runtime.GOOS)
			}

			if tt.xdgDataHome != "" {
				os.Setenv("XDG_DATA_HOME", tt.xdgDataHome)
			} else {
				os.Unsetenv("XDG_DATA_HOME")
			}

			path, err := GetDefaultPrefixDir()
			if err != nil {
				t.Fatalf("GetDefaultPrefixDir() error: %v", err)
			}

			if path == "" {
				t.Error("GetDefaultPrefixDir() returned empty string")
			}

			// Check if path contains expected substring
			if tt.shouldContain != "" && filepath.Base(path) != tt.shouldContain {
				t.Errorf("GetDefaultPrefixDir() = %v, should contain %v", path, tt.shouldContain)
			}
		})
	}
}

func TestGetDefaultPrefixDir_AllPlatforms(t *testing.T) {
	path, err := GetDefaultPrefixDir()
	if err != nil {
		t.Fatalf("GetDefaultPrefixDir() error: %v", err)
	}

	if path == "" {
		t.Error("GetDefaultPrefixDir() returned empty string")
	}

	// Verify it ends with "parm"
	if filepath.Base(path) != "parm" {
		t.Errorf("GetDefaultPrefixDir() = %v, expected to end with 'parm'", path)
	}

	// Platform-specific checks
	switch runtime.GOOS {
	case "linux":
		// Should be in .local/share or XDG_DATA_HOME
		if !filepath.IsAbs(path) {
			t.Error("Expected absolute path on Linux")
		}
	case "darwin":
		// Should be in Library/Application Support
		if !filepath.IsAbs(path) {
			t.Error("Expected absolute path on macOS")
		}
	case "windows":
		// Should be in home directory
		if !filepath.IsAbs(path) {
			t.Error("Expected absolute path on Windows")
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	if DefaultCfg == nil {
		t.Fatal("DefaultCfg is nil")
	}

	if DefaultCfg.ParmPkgPath == "" {
		t.Error("DefaultCfg.ParmPkgPath is empty")
	}

	if DefaultCfg.ParmBinPath == "" {
		t.Error("DefaultCfg.ParmBinPath is empty")
	}

	// GitHubApiTokenFallback can be empty
}
