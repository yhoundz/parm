package catalog

import (
	"os"
	"path/filepath"
	"testing"

	"parm/internal/manifest"

	"github.com/spf13/viper"
)

func TestGetAllPkgManifest_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	v := viper.New()
	v.Set("parm_pkg_path", tmpDir)
	viper.Set("parm_pkg_path", tmpDir)
	defer viper.Reset()

	manifests, err := GetAllPkgManifest()
	if err != nil {
		t.Fatalf("GetAllPkgManifest() error: %v", err)
	}

	if len(manifests) != 0 {
		t.Errorf("GetAllPkgManifest() returned %d manifests, want 0", len(manifests))
	}
}

func TestGetAllPkgManifest_SinglePackage(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package structure
	pkgDir := filepath.Join(tmpDir, "owner", "repo")
	os.MkdirAll(pkgDir, 0755)

	// Create manifest
	m := &manifest.Manifest{
		Owner:       "owner",
		Repo:        "repo",
		Version:     "v1.0.0",
		InstallType: manifest.Release,
		Executables: []string{"bin/app"},
		LastUpdated: "2025-01-01 12:00:00",
	}
	m.Write(pkgDir)

	viper.Set("parm_pkg_path", tmpDir)
	defer viper.Reset()

	manifests, err := GetAllPkgManifest()
	if err != nil {
		t.Fatalf("GetAllPkgManifest() error: %v", err)
	}

	if len(manifests) != 1 {
		t.Fatalf("GetAllPkgManifest() returned %d manifests, want 1", len(manifests))
	}

	if manifests[0].Owner != "owner" {
		t.Errorf("Manifest owner = %v, want owner", manifests[0].Owner)
	}

	if manifests[0].Repo != "repo" {
		t.Errorf("Manifest repo = %v, want repo", manifests[0].Repo)
	}
}

func TestGetAllPkgManifest_MultiplePackages(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple packages
	packages := []struct {
		owner string
		repo  string
	}{
		{"owner1", "repo1"},
		{"owner1", "repo2"},
		{"owner2", "repo1"},
	}

	for _, pkg := range packages {
		pkgDir := filepath.Join(tmpDir, pkg.owner, pkg.repo)
		os.MkdirAll(pkgDir, 0755)

		m := &manifest.Manifest{
			Owner:       pkg.owner,
			Repo:        pkg.repo,
			Version:     "v1.0.0",
			InstallType: manifest.Release,
			Executables: []string{"bin/app"},
			LastUpdated: "2025-01-01 12:00:00",
		}
		m.Write(pkgDir)
	}

	viper.Set("parm_pkg_path", tmpDir)
	defer viper.Reset()

	manifests, err := GetAllPkgManifest()
	if err != nil {
		t.Fatalf("GetAllPkgManifest() error: %v", err)
	}

	if len(manifests) != 3 {
		t.Errorf("GetAllPkgManifest() returned %d manifests, want 3", len(manifests))
	}
}

func TestGetAllPkgManifest_SkipInvalid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid package
	validDir := filepath.Join(tmpDir, "owner1", "repo1")
	os.MkdirAll(validDir, 0755)

	m1 := &manifest.Manifest{
		Owner:       "owner1",
		Repo:        "repo1",
		Version:     "v1.0.0",
		InstallType: manifest.Release,
		Executables: []string{"bin/app"},
		LastUpdated: "2025-01-01 12:00:00",
	}
	m1.Write(validDir)

	// Create directory without manifest
	invalidDir := filepath.Join(tmpDir, "owner2", "repo2")
	os.MkdirAll(invalidDir, 0755)

	// Create directory with corrupted manifest
	corruptDir := filepath.Join(tmpDir, "owner3", "repo3")
	os.MkdirAll(corruptDir, 0755)
	manifestPath := filepath.Join(corruptDir, manifest.ManifestFileName)
	os.WriteFile(manifestPath, []byte("corrupted json"), 0644)

	viper.Set("parm_pkg_path", tmpDir)
	defer viper.Reset()

	manifests, err := GetAllPkgManifest()
	if err != nil {
		t.Fatalf("GetAllPkgManifest() error: %v", err)
	}

	// Should only return the valid manifest
	if len(manifests) != 1 {
		t.Errorf("GetAllPkgManifest() returned %d manifests, want 1", len(manifests))
	}

	if manifests[0].Owner != "owner1" {
		t.Errorf("Manifest owner = %v, want owner1", manifests[0].Owner)
	}
}

func TestGetInstalledPkgInfo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package
	pkgDir := filepath.Join(tmpDir, "owner", "repo")
	os.MkdirAll(pkgDir, 0755)

	m := &manifest.Manifest{
		Owner:       "owner",
		Repo:        "repo",
		Version:     "v1.0.0",
		InstallType: manifest.Release,
		Executables: []string{"bin/app"},
		LastUpdated: "2025-01-01 12:00:00",
	}
	m.Write(pkgDir)

	viper.Set("parm_pkg_path", tmpDir)
	defer viper.Reset()

	infos, data, err := GetInstalledPkgInfo()
	if err != nil {
		t.Fatalf("GetInstalledPkgInfo() error: %v", err)
	}

	if len(infos) != 1 {
		t.Fatalf("GetInstalledPkgInfo() returned %d infos, want 1", len(infos))
	}

	if data.NumPkgs != 1 {
		t.Errorf("NumPkgs = %v, want 1", data.NumPkgs)
	}

	// Check info string format
	expectedSubstr := "owner/repo"
	if !contains(infos[0], expectedSubstr) {
		t.Errorf("Info string %q should contain %q", infos[0], expectedSubstr)
	}
}

func TestGetInstalledPkgInfo_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	viper.Set("parm_pkg_path", tmpDir)
	defer viper.Reset()

	infos, data, err := GetInstalledPkgInfo()
	if err != nil {
		t.Fatalf("GetInstalledPkgInfo() error: %v", err)
	}

	if len(infos) != 0 {
		t.Errorf("GetInstalledPkgInfo() returned %d infos, want 0", len(infos))
	}

	if data.NumPkgs != 0 {
		t.Errorf("NumPkgs = %v, want 0", data.NumPkgs)
	}
}

func TestGetAllPkgManifest_NoConfigPath(t *testing.T) {
	viper.Reset()
	viper.Set("parm_pkg_path", "")

	_, err := GetAllPkgManifest()
	if err == nil {
		t.Error("GetAllPkgManifest() should return error when parm_pkg_path is empty")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && anySubstring(s, substr)))
}

func anySubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
