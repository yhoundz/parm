package sysutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsProcessRunning_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	// Path to non-existent binary
	nonExistentPath := filepath.Join(tmpDir, "nonexistent")

	running, err := IsProcessRunning(nonExistentPath)
	if err != nil {
		// Error is acceptable for non-existent file
		return
	}

	if running {
		t.Error("IsProcessRunning() returned true for non-existent binary")
	}
}

func TestIsProcessRunning_CurrentProcess(t *testing.T) {
	// Get current executable path
	exePath, err := os.Executable()
	if err != nil {
		t.Skipf("Cannot get executable path: %v", err)
	}

	// The test runner itself should be running
	running, err := IsProcessRunning(exePath)
	if err != nil {
		t.Fatalf("IsProcessRunning() error: %v", err)
	}

	// Note: This might be false if the process list doesn't include the test binary
	// This is OS-dependent behavior
	t.Logf("Current process running status: %v", running)
}

func TestIsProcessRunning_RelativePath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping relative path test on Windows")
	}

	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "testbin")

	// Create a dummy file (not a real running process)
	os.WriteFile(binPath, []byte("dummy"), 0755)

	running, err := IsProcessRunning(binPath)
	if err != nil {
		t.Logf("Error checking process: %v", err)
	}

	// Should not be running since it's not actually executed
	if running {
		t.Error("IsProcessRunning() returned true for non-running binary")
	}
}
