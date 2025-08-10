package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestContainsAny(t *testing.T) {
	tests := []struct {
		s      string
		tokens []string
		want   bool
	}{
		{"tool_linux_amd64.tar.gz", []string{"linux"}, true},
		{"TOOL_DARWIN_X64.ZIP", []string{"darwin"}, true}, // case-insensitive
		{"win.exe", []string{"linux", "darwin"}, false},
		{"", []string{"any"}, false},
	}
	for _, tc := range tests {
		if got := ContainsAny(tc.s, tc.tokens); got != tc.want {
			t.Fatalf("ContainsAny(%q,%v)=%v want %v", tc.s, tc.tokens, got, tc.want)
		}
	}
}

func TestExtractTarGz(t *testing.T) {
	td := t.TempDir()
	data := makeTarGz(t, map[string]string{
		"dir/file.txt": "hello",
	})
	arc := filepath.Join(td, "a.tar.gz")
	if err := os.WriteFile(arc, data, 0o644); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(td, "out")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := ExtractTarGz(arc, dest); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join(dest, "dir", "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "hello" {
		t.Fatalf("got %q", b)
	}
}

func makeTarGz(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{
			Name: name, Mode: 0o644, Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()
	gz.Close()
	return buf.Bytes()
}
