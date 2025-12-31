package sysutil

import (
	"testing"
)

func TestSafeJoin_ValidPaths(t *testing.T) {
	tests := []struct {
		name    string
		root    string
		path    string
		wantErr bool
	}{
		{"simple file", "/tmp/root", "file.txt", false},
		{"nested file", "/tmp/root", "dir/file.txt", false},
		{"dot path", "/tmp/root", "./file.txt", false},
		{"traversal attempt", "/tmp/root", "../etc/passwd", true},
		{"complex traversal", "/tmp/root", "dir/../../etc/passwd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SafeJoin(tt.root, tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SafeJoin() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("SafeJoin() unexpected error: %v", err)
				}
				if result == "" {
					t.Error("SafeJoin() returned empty string")
				}
			}
		})
	}
}

func TestCleanBinaryName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"tool-v1.0.0-darwin-arm64", "tool"},
		{"tool_linux_amd64", "tool"},
		{"tool-1.0.0-linux-x86_64", "tool"},
		{"tool-windows-amd64.exe", "tool"},
		{"tool-arm64", "tool"},
		{"tool-macos", "tool"},
		{"tool-gnu", "tool"},
		{"tool-musl", "tool"},
		{"tool-x86_64-unknown-linux-gnu", "tool"},
		{"tool_darwin_arm64", "tool"},
		{"tool-v2.3.4-win-x64", "tool"},
		{"simple-tool", "simple-tool"},
		{"tool-1.2.3", "tool-1.2.3"}, // Shouldn't strip version if not followed by OS/Arch (per original script logic)
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := CleanBinaryName(tt.input)
			if got != tt.expected {
				t.Errorf("CleanBinaryName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

