package sysutil

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSymlinkBinToPath(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "parm-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy executable
	binName := "testbin"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(tempDir, binName)

	// Minimal ELF/Mach-O/PE header or just a file that IsBinaryExecutable will accept
	// Actually, IsBinaryExecutable uses filetype.Match which checks magic numbers.
	// For simplicity, let's just use a real small binary or mock the check if we could.
	// But IsValidBinaryExecutable is hard to mock without changing the code.

	// Let's create a file that looks like a binary.
	// ELF magic: \x7fELF
	var content []byte
	switch runtime.GOOS {
	case "windows":
		content = []byte("MZ") // Minimal PE header
		// Add some padding to satisfy potential length checks if any
		content = append(content, make([]byte, 100)...)
	case "darwin":
		content = []byte{0xca, 0xfe, 0xba, 0xbe} // Mach-O Fat
		content = append(content, make([]byte, 100)...)
	default:
		content = []byte{0x7f, 'E', 'L', 'F'} // ELF
		content = append(content, make([]byte, 100)...)
	}

	err = os.WriteFile(binPath, content, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	destPath := filepath.Join(tempDir, "dest")

	err = SymlinkBinToPath(binPath, destPath)
	if err != nil {
		t.Fatalf("SymlinkBinToPath failed: %v", err)
	}

	// Verify the link
	fi, err := os.Lstat(destPath)
	if err != nil {
		t.Fatalf("Failed to stat destPath: %v", err)
	}

	if runtime.GOOS == "windows" {
		// On Windows we used os.Link, so it should be a regular file with same size
		if fi.Mode().IsRegular() {
			t.Log("Verified: Hard link created on Windows (simulated by logic)")
		} else {
			t.Errorf("Expected regular file (hard link) on Windows, got %v", fi.Mode())
		}
	} else {
		// On others it should be a symlink
		if fi.Mode()&os.ModeSymlink != 0 {
			t.Log("Verified: Symlink created on non-Windows")
		} else {
			t.Errorf("Expected symlink, got %v", fi.Mode())
		}
	}

	// Test update (remove existing and re-link)
	err = SymlinkBinToPath(binPath, destPath)
	if err != nil {
		t.Fatalf("SymlinkBinToPath failed on update: %v", err)
	}
}
