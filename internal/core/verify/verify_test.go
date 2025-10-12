package verify

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestGetSha256(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file with known content
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	os.WriteFile(testFile, content, 0644)

	// Calculate expected hash
	h := sha256.New()
	h.Write(content)
	expected := fmt.Sprintf("%x", h.Sum(nil))

	// Get hash using function
	hash, err := GetSha256(testFile)
	if err != nil {
		t.Fatalf("GetSha256() error: %v", err)
	}

	if hash != expected {
		t.Errorf("GetSha256() = %v, want %v", hash, expected)
	}
}

func TestGetSha256_NonExistent(t *testing.T) {
	_, err := GetSha256("/nonexistent/file")
	if err == nil {
		t.Error("GetSha256() should return error for non-existent file")
	}
}

func TestGetSha256_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	emptyFile := filepath.Join(tmpDir, "empty.txt")
	os.WriteFile(emptyFile, []byte{}, 0644)

	hash, err := GetSha256(emptyFile)
	if err != nil {
		t.Fatalf("GetSha256() error: %v", err)
	}

	// SHA256 of empty file
	expected := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	if hash != expected {
		t.Errorf("GetSha256() = %v, want %v", hash, expected)
	}
}

func TestVerifyLevel1_MatchingChecksum(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	os.WriteFile(testFile, content, 0644)

	// Calculate hash
	h := sha256.New()
	h.Write(content)
	expectedHash := fmt.Sprintf("sha256:%x", h.Sum(nil))

	// Verify
	ok, generatedHash, err := VerifyLevel1(testFile, expectedHash)
	if err != nil {
		t.Fatalf("VerifyLevel1() error: %v", err)
	}

	if !ok {
		t.Error("VerifyLevel1() returned false for matching checksum")
	}

	if generatedHash == nil {
		t.Error("VerifyLevel1() returned nil generated hash")
	} else if *generatedHash != expectedHash {
		t.Errorf("Generated hash = %v, want %v", *generatedHash, expectedHash)
	}
}

func TestVerifyLevel1_MismatchedChecksum(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test content")
	os.WriteFile(testFile, content, 0644)

	// Use wrong hash
	wrongHash := "sha256:0000000000000000000000000000000000000000000000000000000000000000"

	// Verify
	ok, _, err := VerifyLevel1(testFile, wrongHash)
	if err != nil {
		t.Fatalf("VerifyLevel1() error: %v", err)
	}

	if ok {
		t.Error("VerifyLevel1() returned true for mismatched checksum")
	}
}

func TestVerifyLevel1_NonExistentFile(t *testing.T) {
	_, _, err := VerifyLevel1("/nonexistent/file", "sha256:abc123")
	if err == nil {
		t.Error("VerifyLevel1() should return error for non-existent file")
	}
}

func TestVerifyLevel1_KnownHash(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file with known content
	testFile := filepath.Join(tmpDir, "hello.txt")
	os.WriteFile(testFile, []byte("hello world"), 0644)

	// Known SHA256 hash of "hello world"
	knownHash := "sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9"

	ok, _, err := VerifyLevel1(testFile, knownHash)
	if err != nil {
		t.Fatalf("VerifyLevel1() error: %v", err)
	}

	if !ok {
		t.Error("VerifyLevel1() returned false for known correct hash")
	}
}

func TestVerifyLevel1_DifferentContent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two files with different content
	file1 := filepath.Join(tmpDir, "file1.txt")
	file2 := filepath.Join(tmpDir, "file2.txt")
	os.WriteFile(file1, []byte("content1"), 0644)
	os.WriteFile(file2, []byte("content2"), 0644)

	// Get hash of file1
	hash1, _ := GetSha256(file1)
	upstreamHash := fmt.Sprintf("sha256:%s", hash1)

	// Verify file2 with file1's hash (should fail)
	ok, _, err := VerifyLevel1(file2, upstreamHash)
	if err != nil {
		t.Fatalf("VerifyLevel1() error: %v", err)
	}

	if ok {
		t.Error("VerifyLevel1() returned true for different content")
	}
}
