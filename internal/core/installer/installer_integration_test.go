package installer

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"parm/internal/config"
	"parm/internal/manifest"

	"github.com/google/go-github/v74/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestInstall_FullWorkflow_Release(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir
	config.Cfg.ParmBinPath = filepath.Join(tmpDir, "bin")

	// Create test archive with a binary
	archivePath := createTestTarGzWithBinary(t, tmpDir)

	// Create HTTP server to serve the archive
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, archivePath)
	}))
	defer server.Close()

	// Create mock GitHub client
	assetName := fmt.Sprintf("test-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesLatestByOwnerByRepo,
			&github.RepositoryRelease{
				TagName: github.Ptr("v1.0.0"),
				Assets: []*github.ReleaseAsset{
					{
						Name:               github.Ptr(assetName),
						BrowserDownloadURL: github.Ptr(server.URL + "/asset"),
					},
				},
			},
		),
	)

	client := github.NewClient(mockedHTTPClient)
	installer := New(client.Repositories, "")

	ctx := context.Background()
	installPath := filepath.Join(tmpDir, "owner", "repo")

	opts := InstallFlags{
		Type:        manifest.Release,
		Version:     nil,
		Asset:       nil,
		Strict:      false,
		VerifyLevel: 0,
	}

	result, err := installer.Install(ctx, "owner", "repo", installPath, opts, nil)
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	if result == nil {
		t.Fatal("Install() returned nil result")
	}

	if result.Version != "v1.0.0" {
		t.Errorf("Version = %v, want v1.0.0", result.Version)
	}

	// Verify installation directory exists
	if _, err := os.Stat(result.InstallPath); os.IsNotExist(err) {
		t.Error("Installation directory was not created")
	}
}

func TestInstall_PreRelease(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	archivePath := createTestTarGzWithBinary(t, tmpDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, archivePath)
	}))
	defer server.Close()

	assetName := fmt.Sprintf("test-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesByOwnerByRepo,
			[]*github.RepositoryRelease{
				{
					TagName:    github.Ptr("v1.0.0-beta"),
					Prerelease: github.Ptr(true),
					Assets: []*github.ReleaseAsset{
						{
							Name:               github.Ptr(assetName),
							BrowserDownloadURL: github.Ptr(server.URL + "/asset"),
						},
					},
				},
			},
		),
		mock.WithRequestMatch(
			mock.GetReposReleasesLatestByOwnerByRepo,
			&github.RepositoryRelease{
				TagName: github.Ptr("v0.9.0"),
				Assets: []*github.ReleaseAsset{
					{
						Name:               github.Ptr(assetName),
						BrowserDownloadURL: github.Ptr(server.URL + "/asset"),
					},
				},
			},
		),
	)

	client := github.NewClient(mockedHTTPClient)
	installer := New(client.Repositories, "")

	ctx := context.Background()
	installPath := filepath.Join(tmpDir, "owner", "repo")

	opts := InstallFlags{
		Type:        manifest.PreRelease,
		Version:     nil,
		Asset:       nil,
		Strict:      false,
		VerifyLevel: 0,
	}

	result, err := installer.Install(ctx, "owner", "repo", installPath, opts, nil)
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	if result == nil {
		t.Fatal("Install() returned nil result")
	}

	// With strict=false, should install stable if newer
	t.Logf("Installed version: %s", result.Version)
}

func TestInstall_SpecificAsset(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	archivePath := createTestZipWithBinary(t, tmpDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, archivePath)
	}))
	defer server.Close()

	specificAsset := "custom-asset.zip"
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesLatestByOwnerByRepo,
			&github.RepositoryRelease{
				TagName: github.Ptr("v1.0.0"),
				Assets: []*github.ReleaseAsset{
					{
						Name:               github.Ptr(specificAsset),
						BrowserDownloadURL: github.Ptr(server.URL + "/asset"),
					},
				},
			},
		),
	)

	client := github.NewClient(mockedHTTPClient)
	installer := New(client.Repositories, "")

	ctx := context.Background()
	installPath := filepath.Join(tmpDir, "owner", "repo")

	opts := InstallFlags{
		Type:        manifest.Release,
		Version:     nil,
		Asset:       &specificAsset,
		Strict:      false,
		VerifyLevel: 0,
	}

	result, err := installer.Install(ctx, "owner", "repo", installPath, opts, nil)
	if err != nil {
		t.Fatalf("Install() error: %v", err)
	}

	if result == nil {
		t.Fatal("Install() returned nil result")
	}
}

func TestDownloadTo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test content
	testContent := []byte("test file content")

	// Create HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer server.Close()

	ctx := context.Background()
	destPath := filepath.Join(tmpDir, "downloaded.txt")

	err := downloadTo(ctx, destPath, server.URL, "", nil)
	if err != nil {
		t.Fatalf("downloadTo() error: %v", err)
	}

	// Verify file was created
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != string(testContent) {
		t.Errorf("Downloaded content = %q, want %q", content, testContent)
	}
}

func TestDownloadTo_404(t *testing.T) {
	tmpDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	ctx := context.Background()
	destPath := filepath.Join(tmpDir, "downloaded.txt")

	err := downloadTo(ctx, destPath, server.URL, "", nil)
	if err == nil {
		t.Error("downloadTo() should return error for 404")
	}
}

func TestInstallFromRelease_TarGz(t *testing.T) {
	tmpDir := t.TempDir()

	archivePath := createTestTarGzWithBinary(t, tmpDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, archivePath)
	}))
	defer server.Close()

	assetName := "test.tar.gz"
	release := &github.RepositoryRelease{
		TagName: github.Ptr("v1.0.0"),
		Assets: []*github.ReleaseAsset{
			{
				Name:               github.Ptr(assetName),
				BrowserDownloadURL: github.Ptr(server.URL + "/asset"),
			},
		},
	}

	installer := &Installer{}
	ctx := context.Background()
	pkgPath := filepath.Join(tmpDir, "install")

	opts := InstallFlags{
		Type:        manifest.Release,
		Version:     nil,
		Asset:       nil,
		Strict:      false,
		VerifyLevel: 0,
	}

	result, err := installer.installFromRelease(ctx, pkgPath, "owner", "repo", release, opts, nil)
	if err != nil {
		t.Fatalf("installFromRelease() error: %v", err)
	}

	if result == nil {
		t.Fatal("installFromRelease() returned nil")
	}

	if result.Version != "v1.0.0" {
		t.Errorf("Version = %v, want v1.0.0", result.Version)
	}
}

func TestInstallFromRelease_Zip(t *testing.T) {
	tmpDir := t.TempDir()

	archivePath := createTestZipWithBinary(t, tmpDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, archivePath)
	}))
	defer server.Close()

	assetName := "test.zip"
	release := &github.RepositoryRelease{
		TagName: github.Ptr("v1.0.0"),
		Assets: []*github.ReleaseAsset{
			{
				Name:               github.Ptr(assetName),
				BrowserDownloadURL: github.Ptr(server.URL + "/asset"),
			},
		},
	}

	installer := &Installer{}
	ctx := context.Background()
	pkgPath := filepath.Join(tmpDir, "install")

	opts := InstallFlags{
		Type:        manifest.Release,
		Version:     nil,
		Asset:       &assetName,
		Strict:      false,
		VerifyLevel: 0,
	}

	result, err := installer.installFromRelease(ctx, pkgPath, "owner", "repo", release, opts, nil)
	if err != nil {
		t.Fatalf("installFromRelease() error: %v", err)
	}

	if result == nil {
		t.Fatal("installFromRelease() returned nil")
	}
}

// Helper functions
func createTestTarGzWithBinary(t *testing.T, baseDir string) string {
	archivePath := filepath.Join(baseDir, "test.tar.gz")

	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Create a fake binary content
	binContent := []byte("fake binary content")

	hdr := &tar.Header{
		Name: "testbin",
		Mode: 0755,
		Size: int64(len(binContent)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}

	if _, err := tw.Write(binContent); err != nil {
		t.Fatal(err)
	}

	return archivePath
}

func createTestZipWithBinary(t *testing.T, baseDir string) string {
	archivePath := filepath.Join(baseDir, "test.zip")

	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	binContent := []byte("fake binary content")

	fw, err := zw.Create("testbin")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := io.Writer.Write(fw, binContent); err != nil {
		t.Fatal(err)
	}

	return archivePath
}

func TestDownloadTo_WithAuth(t *testing.T) {
	tmpDir := t.TempDir()

	testContent := []byte("private repo content")
	expectedToken := "test-token-123"

	// Create HTTP server that requires authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+expectedToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		accept := r.Header.Get("Accept")
		if accept != "application/octet-stream" {
			t.Errorf("Expected Accept header 'application/octet-stream', got %q", accept)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(testContent)
	}))
	defer server.Close()

	ctx := context.Background()
	destPath := filepath.Join(tmpDir, "downloaded.txt")

	err := downloadTo(ctx, destPath, server.URL, expectedToken, nil)
	if err != nil {
		t.Fatalf("downloadTo() with auth error: %v", err)
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != string(testContent) {
		t.Errorf("Downloaded content = %q, want %q", content, testContent)
	}
}

func TestDownloadTo_WithAuth_Unauthorized(t *testing.T) {
	tmpDir := t.TempDir()

	// Create HTTP server that requires authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer correct-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx := context.Background()
	destPath := filepath.Join(tmpDir, "downloaded.txt")

	err := downloadTo(ctx, destPath, server.URL, "wrong-token", nil)
	if err == nil {
		t.Error("downloadTo() should return error for unauthorized request")
	}
}
