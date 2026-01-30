package selfupdate

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFindBinaryInArchive(t *testing.T) {
	tmpDir := t.TempDir()

	binaryName := "parm"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	// Create nested directory structure with binary
	binaryPath := filepath.Join(tmpDir, "nested", "dir", binaryName)
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(binaryPath, []byte("binary content"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create another file that's not the binary
	otherPath := filepath.Join(tmpDir, "other.txt")
	if err := os.WriteFile(otherPath, []byte("other"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test binary finding logic (simulating the filepath.Walk from Update function)
	foundPath := ""
	err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == binaryName {
			foundPath = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if foundPath == "" {
		t.Error("Binary not found")
	}

	if foundPath != binaryPath {
		t.Errorf("Found path = %v, want %v", foundPath, binaryPath)
	}
}

func TestFindBinaryInArchive_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory with no binary
	otherPath := filepath.Join(tmpDir, "other.txt")
	if err := os.WriteFile(otherPath, []byte("other"), 0644); err != nil {
		t.Fatal(err)
	}

	binaryName := "parm"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	// Test binary finding logic
	foundPath := ""
	err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == binaryName {
			foundPath = path
			return filepath.SkipAll
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	if foundPath != "" {
		t.Error("Should not have found binary")
	}
}
