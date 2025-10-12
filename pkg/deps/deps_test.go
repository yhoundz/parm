package deps

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestHasExternalDep(t *testing.T) {
	// Test for a common command that should exist
	commonCmd := "go"
	if runtime.GOOS == "windows" {
		commonCmd = "cmd"
	}

	err := HasExternalDep(commonCmd)
	if err != nil {
		t.Errorf("HasExternalDep(%s) failed: %v", commonCmd, err)
	}

	// Test for non-existent command
	err = HasExternalDep("this-command-definitely-does-not-exist-12345")
	if err == nil {
		t.Error("HasExternalDep() should return error for non-existent command")
	}
}

func TestGetMissingLibs_NonExistentFile(t *testing.T) {
	ctx := context.Background()

	_, err := GetMissingLibs(ctx, "/nonexistent/binary")
	if err == nil {
		t.Error("GetMissingLibs() should return error for non-existent file")
	}
}

func TestGetMissingLibs_TextFile(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a text file
	textFile := filepath.Join(tmpDir, "text.txt")
	os.WriteFile(textFile, []byte("not a binary"), 0644)

	libs, err := GetMissingLibs(ctx, textFile)
	if err != nil {
		t.Logf("Expected error for text file: %v", err)
	}

	// Should return empty or error
	if len(libs) > 0 {
		t.Error("GetMissingLibs() returned libraries for text file")
	}
}

func TestGetBinDeps_InvalidBinary(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid binary file
	invalidBin := filepath.Join(tmpDir, "invalid")
	os.WriteFile(invalidBin, []byte("not a binary"), 0755)

	_, err := GetBinDeps(invalidBin)
	if err == nil {
		t.Error("GetBinDeps() should return error for invalid binary")
	}
}

func TestGetBinDeps_NonExistent(t *testing.T) {
	_, err := GetBinDeps("/nonexistent/binary")
	if err == nil {
		t.Error("GetBinDeps() should return error for non-existent file")
	}
}

func TestHasSharedLib_CommonLibs(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shared library test on Windows")
	}

	// Test for a common library that should exist on most systems
	commonLib := "libc.so.6"
	if runtime.GOOS == "darwin" {
		commonLib = "libSystem.B.dylib"
	}

	hasLib, err := hasSharedLib(commonLib)
	if err != nil {
		t.Logf("Error checking for %s: %v", commonLib, err)
	}

	// Note: This test is environment-dependent
	t.Logf("Has %s: %v", commonLib, hasLib)
}

func TestHasSharedLib_NonExistent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping shared library test on Windows")
	}

	hasLib, err := hasSharedLib("libnonexistent12345.so")
	if err != nil {
		t.Logf("Error (expected): %v", err)
	}

	if hasLib {
		t.Error("hasSharedLib() returned true for non-existent library")
	}
}

func TestGetMissingLibsFallBack(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid binary
	invalidBin := filepath.Join(tmpDir, "invalid")
	os.WriteFile(invalidBin, []byte("not a binary"), 0755)

	_, err := getMissingLibsFallBack(invalidBin)
	if err == nil {
		t.Log("getMissingLibsFallBack() returned no error (acceptable)")
	}
}

func TestGetMissingLibsLinux(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a dummy file
	dummyBin := filepath.Join(tmpDir, "dummy")
	os.WriteFile(dummyBin, []byte("dummy"), 0755)

	// This will likely fail or return empty, but shouldn't crash
	_, err := getMissingLibsLinux(ctx, dummyBin)
	if err != nil {
		t.Logf("Expected error or empty result: %v", err)
	}
}

func TestGetMissingLibsDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping macOS-specific test")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a dummy file
	dummyBin := filepath.Join(tmpDir, "dummy")
	os.WriteFile(dummyBin, []byte("dummy"), 0755)

	// This will likely fail or return empty, but shouldn't crash
	_, err := getMissingLibsDarwin(ctx, dummyBin)
	if err != nil {
		t.Logf("Expected error or empty result: %v", err)
	}
}
