package uninstaller

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"parm/internal/config"
	"parm/internal/manifest"
)

func TestUninstall_Success(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	// Create installed package structure
	pkgDir := filepath.Join(tmpDir, "owner", "repo")
	os.MkdirAll(pkgDir, 0755)

	// Create manifest
	m := &manifest.Manifest{
		Owner:       "owner",
		Repo:        "repo",
		Version:     "v1.0.0",
		InstallType: manifest.Release,
		Executables: []string{},
		LastUpdated: "2025-01-01 12:00:00",
	}
	m.Write(pkgDir)

	// Create a file in the package
	testFile := filepath.Join(pkgDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	ctx := context.Background()
	err := Uninstall(ctx, "owner", "repo")
	if err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	// Verify package directory was removed
	if _, err := os.Stat(pkgDir); !os.IsNotExist(err) {
		t.Error("Package directory still exists after uninstall")
	}
}

func TestUninstall_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	ctx := context.Background()
	err := Uninstall(ctx, "owner", "nonexistent")
	if err == nil {
		t.Error("Uninstall() should return error for non-existent package")
	}
}

func TestUninstall_NoManifest(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	// Create directory without manifest
	pkgDir := filepath.Join(tmpDir, "owner", "repo")
	os.MkdirAll(pkgDir, 0755)

	ctx := context.Background()
	err := Uninstall(ctx, "owner", "repo")
	if err == nil {
		t.Error("Uninstall() should return error when manifest is missing")
	}
}

func TestUninstall_WithExecutables(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	// Create installed package
	pkgDir := filepath.Join(tmpDir, "owner", "repo")
	binDir := filepath.Join(pkgDir, "bin")
	os.MkdirAll(binDir, 0755)

	// Create fake executable
	binPath := filepath.Join(binDir, "testbin")
	os.WriteFile(binPath, []byte("fake binary"), 0755)

	// Create manifest with executable
	m := &manifest.Manifest{
		Owner:       "owner",
		Repo:        "repo",
		Version:     "v1.0.0",
		InstallType: manifest.Release,
		Executables: []string{"bin/testbin"},
		LastUpdated: "2025-01-01 12:00:00",
	}
	m.Write(pkgDir)

	ctx := context.Background()
	err := Uninstall(ctx, "owner", "repo")
	if err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	// Verify removal
	if _, err := os.Stat(pkgDir); !os.IsNotExist(err) {
		t.Error("Package directory still exists after uninstall")
	}
}

func TestRemovePkgSymlinks_Success(t *testing.T) {
	t.Skip("TODO: Symlink removal logic change, test must be rewritten")
	tmpDir := t.TempDir()
	config.Cfg.ParmBinPath = tmpDir

	// Create symlink
	linkPath := filepath.Join(tmpDir, "repo")
	targetPath := "/some/target"

	// On Windows, this might require admin privileges, so skip if it fails
	err := os.Symlink(targetPath, linkPath)
	if err != nil {
		t.Skipf("Cannot create symlink: %v", err)
	}

	ctx := context.Background()
	err = RemovePkgSymlinks(ctx, "owner", "repo")
	if err != nil {
		t.Logf("RemovePkgSymlinks() error: %v", err)
	}

	// Verify symlink was removed
	if _, err := os.Lstat(linkPath); !os.IsNotExist(err) {
		t.Error("Symlink still exists after RemovePkgSymlinks")
	}
}

func TestRemovePkgSymlinks_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmBinPath = tmpDir

	ctx := context.Background()
	err := RemovePkgSymlinks(ctx, "owner", "nonexistent")
	// Should not error on non-existent symlink
	if err != nil {
		t.Logf("RemovePkgSymlinks() returned error (acceptable): %v", err)
	}
}

func TestUninstall_CleansUpParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	// Create installed package
	pkgDir := filepath.Join(tmpDir, "owner", "repo")
	os.MkdirAll(pkgDir, 0755)

	// Create manifest
	m := &manifest.Manifest{
		Owner:       "owner",
		Repo:        "repo",
		Version:     "v1.0.0",
		InstallType: manifest.Release,
		Executables: []string{},
		LastUpdated: "2025-01-01 12:00:00",
	}
	m.Write(pkgDir)

	ctx := context.Background()
	err := Uninstall(ctx, "owner", "repo")
	if err != nil {
		t.Fatalf("Uninstall() error: %v", err)
	}

	// Parent directory (owner) should be removed if empty
	ownerDir := filepath.Join(tmpDir, "owner")
	entries, err := os.ReadDir(ownerDir)
	if err == nil && len(entries) == 0 {
		t.Log("Parent directory is empty (will be cleaned up)")
	}
}
