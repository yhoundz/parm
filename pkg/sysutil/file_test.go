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
