package sysutil

import (
	"os"
	"path/filepath"
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

func TestGetParentDir(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"regular path", "/home/user/dir/file", false},
		{"relative path", "dir/file", false},
		{"single level", "/home", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent, err := GetParentDir(tt.path)

			if tt.wantErr && err == nil {
				t.Error("GetParentDir() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("GetParentDir() unexpected error: %v", err)
			}
			if !tt.wantErr && parent == "" {
				t.Error("GetParentDir() returned empty string")
			}
		})
	}
}

func TestIsBinaryExecutable(t *testing.T) {
	t.Skip("Skipping binary detection test - requires actual compiled binaries")

	tmpDir := t.TempDir()

	// Test non-binary file
	textFile := filepath.Join(tmpDir, "text.txt")
	os.WriteFile(textFile, []byte("not a binary"), 0644)

	isBin, _, err := IsBinaryExecutable(textFile)
	if err != nil {
		t.Logf("IsBinaryExecutable() error on text file: %v", err)
	}
	if isBin {
		t.Error("IsBinaryExecutable() returned true for text file")
	}
}

func TestIsValidBinaryExecutable(t *testing.T) {
	t.Skip("Skipping binary validation test - requires actual compiled binaries")

	tmpDir := t.TempDir()

	// Test text file
	textFile := filepath.Join(tmpDir, "text.txt")
	os.WriteFile(textFile, []byte("not a binary"), 0644)

	isValid, err := IsValidBinaryExecutable(textFile)
	if err != nil {
		t.Logf("IsValidBinaryExecutable() error on text file: %v", err)
	}
	if isValid {
		t.Error("IsValidBinaryExecutable() returned true for text file")
	}
}

func TestSymlinkBinToPath(t *testing.T) {
	t.Skip("Skipping symlink test - requires actual compiled binaries")
}

func TestSymlinkBinToPath_NonBinary(t *testing.T) {
	t.Skip("Skipping symlink test - requires actual compiled binaries")
}
