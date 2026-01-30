package selfupdate

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"parm/internal/core/verify"

	"github.com/google/go-github/v74/github"
	minioSelfUpdate "github.com/minio/selfupdate"
)

func TestUpdateHandlesArchiveExtraction(t *testing.T) {
	tmpDir := t.TempDir()
	assetName := fmt.Sprintf("parm-%s-%s.zip", runtime.GOOS, runtime.GOARCH)
	binaryName := "parm"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	archivePath := filepath.Join(tmpDir, assetName)
	createZip(t, archivePath, binaryName, []byte("content"))
	digest := fmt.Sprintf("sha256:%s", verifyHash(t, archivePath))

	mux := http.NewServeMux()
	serverURL := ""
	mux.HandleFunc("/repos/yhoundz/parm/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"tag_name":"v1.0.0","assets":[{"name":"%s","browser_download_url":"%s/download/%s","digest":"%s"}]}`,
			assetName, serverURL, assetName, digest)
	})
	mux.HandleFunc("/download/"+assetName, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, archivePath)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	serverURL = server.URL

	client := github.NewClient(server.Client())
	baseURL, err := url.Parse(server.URL + "/")
	if err != nil {
		t.Fatalf("failed to parse base URL: %v", err)
	}
	client.BaseURL = baseURL
	client.UploadURL = baseURL

	origApply := applyFunc
	applied := false
	applyFunc = func(r io.Reader, opts minioSelfUpdate.Options) error {
		applied = true
		return nil
	}
	t.Cleanup(func() {
		applyFunc = origApply
	})

	cfg := Config{
		Owner:          "yhoundz",
		Repo:           "parm",
		Binary:         "parm",
		CurrentVersion: "v0.0.1",
		GitHubClient:   client,
		HTTPClient:     server.Client(),
	}

	if err := Update(context.Background(), cfg, io.Discard, io.Discard); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !applied {
		t.Fatal("applyFunc was not called")
	}
}

func createZip(t *testing.T, path, entry string, contents []byte) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	w, err := zw.Create(entry)
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, err := w.Write(contents); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
}

func verifyHash(t *testing.T, path string) string {
	t.Helper()
	hash, err := verify.GetSha256(path)
	if err != nil {
		t.Fatalf("calculate hash: %v", err)
	}
	return hash
}
