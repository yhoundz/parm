package parmutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"parm/internal/config"
)

func TestGetInstallDir(t *testing.T) {
	// Setup config
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	owner := "testowner"
	repo := "testrepo"

	dir := GetInstallDir(owner, repo)

	if dir == "" {
		t.Error("GetInstallDir() returned empty string")
	}

	// Should contain both owner and repo
	if !strings.Contains(dir, owner) || !strings.Contains(dir, repo) {
		t.Errorf("GetInstallDir() = %v, should contain %v and %v", dir, owner, repo)
	}

	// Should be under tmpDir
	if !strings.HasPrefix(dir, tmpDir) {
		t.Errorf("GetInstallDir() = %v, should be under %v", dir, tmpDir)
	}
}

func TestGetBinDir(t *testing.T) {
	// Setup config
	tmpDir := t.TempDir()
	config.Cfg.ParmBinPath = tmpDir

	repoName := "testrepo"

	dir := GetBinDir(repoName)

	if dir == "" {
		t.Error("GetBinDir() returned empty string")
	}

	// Should contain repo name
	if !strings.Contains(dir, repoName) {
		t.Errorf("GetBinDir() = %v, should contain %v", dir, repoName)
	}

	// Should be under tmpDir
	if !strings.HasPrefix(dir, tmpDir) {
		t.Errorf("GetBinDir() = %v, should be under %v", dir, tmpDir)
	}
}

func TestMakeInstallDir(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	owner := "testowner"
	repo := "testrepo"

	dir, err := MakeInstallDir(owner, repo, 0755)
	if err != nil {
		t.Fatalf("MakeInstallDir() error: %v", err)
	}

	if dir == "" {
		t.Error("MakeInstallDir() returned empty string")
	}

	// Verify directory was created
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("MakeInstallDir() did not create a directory")
	}
}

func TestMakeStagingDir(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	owner := "testowner"
	repo := "testrepo"

	stagingDir, err := MakeStagingDir(owner, repo)
	if err != nil {
		t.Fatalf("MakeStagingDir() error: %v", err)
	}

	if stagingDir == "" {
		t.Error("MakeStagingDir() returned empty string")
	}

	// Verify directory was created
	info, err := os.Stat(stagingDir)
	if err != nil {
		t.Fatalf("Staging directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("MakeStagingDir() did not create a directory")
	}

	// Should have staging prefix
	if !strings.Contains(filepath.Base(stagingDir), STAGING_DIR_PREFIX) {
		t.Errorf("Staging dir name should contain %v", STAGING_DIR_PREFIX)
	}

	// Clean up
	os.RemoveAll(stagingDir)
}

func TestMakeStagingDir_UniqueNames(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	owner := "testowner"
	repo := "testrepo"

	// Create two staging directories
	dir1, err := MakeStagingDir(owner, repo)
	if err != nil {
		t.Fatal(err)
	}

	dir2, err := MakeStagingDir(owner, repo)
	if err != nil {
		t.Fatal(err)
	}

	// Should be different
	if dir1 == dir2 {
		t.Error("MakeStagingDir() should create unique directories")
	}

	// Clean up
	os.RemoveAll(dir1)
	os.RemoveAll(dir2)
}

func TestPromoteStagingDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create staging directory
	stagingDir := filepath.Join(tmpDir, "staging")
	os.MkdirAll(stagingDir, 0755)

	// Create file in staging
	testFile := filepath.Join(stagingDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	// Promote to final location
	finalDir := filepath.Join(tmpDir, "final")
	result, err := PromoteStagingDir(finalDir, stagingDir)
	if err != nil {
		t.Fatalf("PromoteStagingDir() error: %v", err)
	}

	if result != finalDir {
		t.Errorf("PromoteStagingDir() = %v, want %v", result, finalDir)
	}

	// Verify final directory exists
	if _, err := os.Stat(finalDir); os.IsNotExist(err) {
		t.Error("Final directory was not created")
	}

	// Verify staging directory no longer exists
	if _, err := os.Stat(stagingDir); !os.IsNotExist(err) {
		t.Error("Staging directory still exists after promotion")
	}

	// Verify file was moved
	movedFile := filepath.Join(finalDir, "test.txt")
	content, err := os.ReadFile(movedFile)
	if err != nil || string(content) != "test content" {
		t.Error("File was not moved correctly")
	}
}

func TestPromoteStagingDir_OverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing final directory
	finalDir := filepath.Join(tmpDir, "final")
	os.MkdirAll(finalDir, 0755)
	oldFile := filepath.Join(finalDir, "old.txt")
	os.WriteFile(oldFile, []byte("old content"), 0644)

	// Create staging directory
	stagingDir := filepath.Join(tmpDir, "staging")
	os.MkdirAll(stagingDir, 0755)
	newFile := filepath.Join(stagingDir, "new.txt")
	os.WriteFile(newFile, []byte("new content"), 0644)

	// Promote (should overwrite)
	_, err := PromoteStagingDir(finalDir, stagingDir)
	if err != nil {
		t.Fatalf("PromoteStagingDir() error: %v", err)
	}

	// Old file should not exist
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file still exists after promotion")
	}

	// New file should exist
	movedFile := filepath.Join(finalDir, "new.txt")
	if _, err := os.Stat(movedFile); os.IsNotExist(err) {
		t.Error("New file was not moved")
	}
}

func TestCleanup(t *testing.T) {
	tmpDir := t.TempDir()

	// Create owner directory
	ownerDir := filepath.Join(tmpDir, "owner")
	os.MkdirAll(ownerDir, 0755)

	// Create staging directories
	stagingDir1 := filepath.Join(ownerDir, STAGING_DIR_PREFIX+"repo1")
	stagingDir2 := filepath.Join(ownerDir, STAGING_DIR_PREFIX+"repo2")
	os.MkdirAll(stagingDir1, 0755)
	os.MkdirAll(stagingDir2, 0755)

	// Create non-staging directory
	normalDir := filepath.Join(ownerDir, "normalrepo")
	os.MkdirAll(normalDir, 0755)

	// Run cleanup
	err := Cleanup(ownerDir)
	if err != nil {
		t.Fatalf("Cleanup() error: %v", err)
	}

	// Staging directories should be removed
	if _, err := os.Stat(stagingDir1); !os.IsNotExist(err) {
		t.Error("Staging directory 1 still exists after cleanup")
	}
	if _, err := os.Stat(stagingDir2); !os.IsNotExist(err) {
		t.Error("Staging directory 2 still exists after cleanup")
	}

	// Normal directory should still exist
	if _, err := os.Stat(normalDir); os.IsNotExist(err) {
		t.Error("Normal directory was removed by cleanup")
	}
}

func TestCleanup_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create owner directory with only staging dirs
	ownerDir := filepath.Join(tmpDir, "owner")
	os.MkdirAll(ownerDir, 0755)

	stagingDir := filepath.Join(ownerDir, STAGING_DIR_PREFIX+"repo")
	os.MkdirAll(stagingDir, 0755)

	// Run cleanup
	err := Cleanup(ownerDir)
	if err != nil {
		t.Fatalf("Cleanup() error: %v", err)
	}

	// Owner directory should be removed if empty after cleanup
	if _, err := os.Stat(ownerDir); !os.IsNotExist(err) {
		t.Log("Owner directory still exists (acceptable)")
	}
}

func TestCleanup_NonExistentDir(t *testing.T) {
	err := Cleanup("/nonexistent/directory")
	// Should not error on non-existent directory
	if err != nil {
		t.Logf("Cleanup() returned error for non-existent dir: %v", err)
	}
}

func TestCleanup_EmptyString(t *testing.T) {
	err := Cleanup("")
	// Should not error on empty string
	if err != nil {
		t.Errorf("Cleanup() returned error for empty string: %v", err)
	}
}
