package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestExtractTarGz_Valid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test tar.gz file
	tarPath := filepath.Join(tmpDir, "test.tar.gz")
	createTestTarGz(t, tarPath, map[string]string{
		"file1.txt":     "content1",
		"dir/file2.txt": "content2",
	})

	// Extract
	extractDir := filepath.Join(tmpDir, "extracted")
	err := ExtractTarGz(tarPath, extractDir)
	if err != nil {
		t.Fatalf("ExtractTarGz failed: %v", err)
	}

	// Verify extracted files
	content1, err := os.ReadFile(filepath.Join(extractDir, "file1.txt"))
	if err != nil || string(content1) != "content1" {
		t.Errorf("file1.txt not extracted correctly")
	}

	content2, err := os.ReadFile(filepath.Join(extractDir, "dir", "file2.txt"))
	if err != nil || string(content2) != "content2" {
		t.Errorf("dir/file2.txt not extracted correctly")
	}
}

func TestExtractTarGz_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a malicious tar.gz with path traversal
	tarPath := filepath.Join(tmpDir, "evil.tar.gz")
	f, err := os.Create(tarPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	// Try to write outside extraction directory
	hdr := &tar.Header{
		Name: "../../../etc/evil.txt",
		Mode: 0644,
		Size: 4,
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte("evil"))
	tw.Close()
	gw.Close()

	// Attempt extraction
	extractDir := filepath.Join(tmpDir, "extracted")
	err = ExtractTarGz(tarPath, extractDir)

	// Should fail due to path traversal protection
	if err == nil {
		t.Error("Expected error for path traversal, got nil")
	}
}

func TestExtractTarGz_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid tar.gz
	invalidPath := filepath.Join(tmpDir, "invalid.tar.gz")
	os.WriteFile(invalidPath, []byte("not a tar.gz file"), 0644)

	extractDir := filepath.Join(tmpDir, "extracted")
	err := ExtractTarGz(invalidPath, extractDir)

	if err == nil {
		t.Error("Expected error for invalid tar.gz, got nil")
	}
}

func TestExtractZip_Valid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test zip file
	zipPath := filepath.Join(tmpDir, "test.zip")
	createTestZip(t, zipPath, map[string]string{
		"file1.txt":     "content1",
		"dir/file2.txt": "content2",
	})

	// Extract
	extractDir := filepath.Join(tmpDir, "extracted")
	err := ExtractZip(zipPath, extractDir)
	if err != nil {
		t.Fatalf("ExtractZip failed: %v", err)
	}

	// Verify extracted files
	content1, err := os.ReadFile(filepath.Join(extractDir, "file1.txt"))
	if err != nil || string(content1) != "content1" {
		t.Errorf("file1.txt not extracted correctly")
	}

	content2, err := os.ReadFile(filepath.Join(extractDir, "dir", "file2.txt"))
	if err != nil || string(content2) != "content2" {
		t.Errorf("dir/file2.txt not extracted correctly")
	}
}

func TestExtractZip_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a malicious zip with path traversal
	zipPath := filepath.Join(tmpDir, "evil.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)

	// Try to write outside extraction directory
	fw, _ := zw.Create("../../../etc/evil.txt")
	fw.Write([]byte("evil"))
	zw.Close()

	// Attempt extraction
	extractDir := filepath.Join(tmpDir, "extracted")
	err = ExtractZip(zipPath, extractDir)

	// Should fail due to path traversal protection
	if err == nil {
		t.Error("Expected error for path traversal, got nil")
	}
}

func TestExtractZip_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create zip with nested directories
	zipPath := filepath.Join(tmpDir, "nested.zip")
	createTestZip(t, zipPath, map[string]string{
		"a/b/c/file.txt": "nested",
	})

	extractDir := filepath.Join(tmpDir, "extracted")
	err := ExtractZip(zipPath, extractDir)
	if err != nil {
		t.Fatalf("ExtractZip failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(extractDir, "a", "b", "c", "file.txt"))
	if err != nil || string(content) != "nested" {
		t.Errorf("nested file not extracted correctly")
	}
}

func TestExtractZip_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid zip
	invalidPath := filepath.Join(tmpDir, "invalid.zip")
	os.WriteFile(invalidPath, []byte("not a zip file"), 0644)

	extractDir := filepath.Join(tmpDir, "extracted")
	err := ExtractZip(invalidPath, extractDir)

	if err == nil {
		t.Error("Expected error for invalid zip, got nil")
	}
}

func TestExtractTarGz_Permissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir := t.TempDir()

	// Create tar.gz with specific permissions
	tarPath := filepath.Join(tmpDir, "perms.tar.gz")
	f, _ := os.Create(tarPath)
	defer f.Close()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: "executable.sh",
		Mode: 0755,
		Size: 4,
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte("test"))
	tw.Close()
	gw.Close()

	extractDir := filepath.Join(tmpDir, "extracted")
	err := ExtractTarGz(tarPath, extractDir)
	if err != nil {
		t.Fatalf("ExtractTarGz failed: %v", err)
	}

	info, err := os.Stat(filepath.Join(extractDir, "executable.sh"))
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode()&0111 == 0 {
		t.Error("Executable permissions not preserved")
	}
}

// Helper functions
func createTestTarGz(t *testing.T, path string, files map[string]string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
}

func createTestZip(t *testing.T, path string, files map[string]string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	
	zw := zip.NewWriter(f)
	defer zw.Close()
	
	for name, content := range files {
		// Create file header with proper permissions
		fh := &zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		}
		fh.SetMode(0644) // Set readable permissions
		
		fw, err := zw.CreateHeader(fh)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
}
