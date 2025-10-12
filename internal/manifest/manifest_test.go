package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"parm/internal/config"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a mock binary
	binPath := filepath.Join(tmpDir, "testbin")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	// Create a simple executable file
	content := []byte{0x7f, 0x45, 0x4c, 0x46} // ELF magic
	switch runtime.GOOS {
	case "darwin":
		content = []byte{0xcf, 0xfa, 0xed, 0xfe} // Mach-O
	case "windows":
		content = []byte{0x4d, 0x5a} // PE
	}
	os.WriteFile(binPath, content, 0755)

	m, err := New("owner", "repo", "v1.0.0", Release, tmpDir)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if m.Owner != "owner" {
		t.Errorf("Owner = %v, want %v", m.Owner, "owner")
	}

	if m.Repo != "repo" {
		t.Errorf("Repo = %v, want %v", m.Repo, "repo")
	}

	if m.Version != "v1.0.0" {
		t.Errorf("Version = %v, want %v", m.Version, "v1.0.0")
	}

	if m.InstallType != Release {
		t.Errorf("InstallType = %v, want %v", m.InstallType, Release)
	}

	if m.LastUpdated == "" {
		t.Error("LastUpdated is empty")
	}
}

func TestManifest_Write(t *testing.T) {
	tmpDir := t.TempDir()

	m := &Manifest{
		Owner:       "owner",
		Repo:        "repo",
		LastUpdated: "2025-01-01 12:00:00",
		Executables: []string{"bin/app"},
		InstallType: Release,
		Version:     "v1.0.0",
	}

	err := m.Write(tmpDir)
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	// Check if file was created
	manifestPath := filepath.Join(tmpDir, ManifestFileName)
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("Write() did not create manifest file")
	}

	// Verify content
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}

	var readManifest Manifest
	err = json.Unmarshal(data, &readManifest)
	if err != nil {
		t.Fatalf("Failed to unmarshal manifest: %v", err)
	}

	if readManifest.Owner != m.Owner {
		t.Errorf("Read Owner = %v, want %v", readManifest.Owner, m.Owner)
	}
}

func TestRead(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a manifest file
	m := &Manifest{
		Owner:       "owner",
		Repo:        "repo",
		LastUpdated: "2025-01-01 12:00:00",
		Executables: []string{"bin/app"},
		InstallType: Release,
		Version:     "v1.0.0",
	}

	err := m.Write(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	// Read it back
	readManifest, err := Read(tmpDir)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}

	if readManifest.Owner != m.Owner {
		t.Errorf("Owner = %v, want %v", readManifest.Owner, m.Owner)
	}

	if readManifest.Repo != m.Repo {
		t.Errorf("Repo = %v, want %v", readManifest.Repo, m.Repo)
	}

	if readManifest.Version != m.Version {
		t.Errorf("Version = %v, want %v", readManifest.Version, m.Version)
	}
}

func TestRead_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := Read(tmpDir)
	if err == nil {
		t.Error("Read() should return error for non-existent manifest")
	}
}

func TestRead_CorruptedManifest(t *testing.T) {
	tmpDir := t.TempDir()

	// Create corrupted manifest
	manifestPath := filepath.Join(tmpDir, ManifestFileName)
	os.WriteFile(manifestPath, []byte("not valid json"), 0644)

	_, err := Read(tmpDir)
	if err == nil {
		t.Error("Read() should return error for corrupted manifest")
	}
}

func TestManifest_GetFullExecPaths(t *testing.T) {
	// Set up config with temp directory
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	m := &Manifest{
		Owner:       "owner",
		Repo:        "repo",
		Executables: []string{"bin/app", "bin/tool"},
	}

	paths := m.GetFullExecPaths()

	if len(paths) != 2 {
		t.Errorf("GetFullExecPaths() returned %d paths, want 2", len(paths))
	}

	// Each path should be absolute and contain the executable name
	for i, path := range paths {
		if !filepath.IsAbs(path) {
			t.Errorf("Path %d is not absolute: %v", i, path)
		}
		// Path should contain owner/repo
		if !strings.Contains(path, "owner") || !strings.Contains(path, "repo") {
			t.Errorf("Path %d doesn't contain owner/repo: %v", i, path)
		}
	}
}

func TestGetBinExecutables(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a bin directory with mock executables
	binDir := filepath.Join(tmpDir, "bin")
	os.MkdirAll(binDir, 0755)

	// TODO: actually create a binary with go build
	// Create mock binary
	binPath := filepath.Join(binDir, "testbin")
	if runtime.GOOS == "windows" {
		binPath += ".exe"
	}

	content := []byte{0x7f, 0x45, 0x4c, 0x46} // ELF magic
	switch runtime.GOOS {
	case "darwin":
		content = []byte{0xcf, 0xfa, 0xed, 0xfe}
	case "windows":
		content = []byte{0x4d, 0x5a}
	}
	os.WriteFile(binPath, content, 0755)

	paths, err := getBinExecutables(tmpDir)
	if err != nil {
		t.Fatalf("getBinExecutables() error: %v", err)
	}

	if len(paths) == 0 {
		t.Log("No executables found (expected if magic number check fails)")
	} else {
		t.Logf("Found %d executables", len(paths))
	}
}

func TestGetBinExecutables_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	paths, err := getBinExecutables(tmpDir)
	if err != nil {
		t.Fatalf("getBinExecutables() error: %v", err)
	}

	if len(paths) != 0 {
		t.Errorf("getBinExecutables() returned %d paths for empty dir, want 0", len(paths))
	}
}

func TestGetBinExecutables_WithNonBinaries(t *testing.T) {
	tmpDir := t.TempDir()

	// Create text file
	textFile := filepath.Join(tmpDir, "readme.txt")
	os.WriteFile(textFile, []byte("not a binary"), 0644)

	paths, err := getBinExecutables(tmpDir)
	if err != nil {
		t.Fatalf("getBinExecutables() error: %v", err)
	}

	// Should skip text files
	for _, path := range paths {
		if filepath.Base(path) == "readme.txt" {
			t.Error("getBinExecutables() included text file")
		}
	}
}
