package updater

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"parm/internal/config"
	"parm/internal/core/installer"
	"parm/internal/manifest"

	"github.com/google/go-github/v74/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestUpdate_Success(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	// Create installed package with old version
	pkgDir := filepath.Join(tmpDir, "owner", "repo")
	os.MkdirAll(pkgDir, 0755)

	m := &manifest.Manifest{
		Owner:       "owner",
		Repo:        "repo",
		Version:     "v1.0.0",
		InstallType: manifest.Release,
		Executables: []string{},
		LastUpdated: "2025-01-01 12:00:00",
	}
	m.Write(pkgDir)

	// Create test archive
	archivePath := createTestArchive(t, tmpDir)

	// Create HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, archivePath)
	}))
	defer server.Close()

	// Create mock GitHub client with newer version
	assetName := fmt.Sprintf("test-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	releaseResponse := &github.RepositoryRelease{
		TagName: github.Ptr("v2.0.0"),
		Assets: []*github.ReleaseAsset{
			{
				Name:               github.Ptr(assetName),
				BrowserDownloadURL: github.Ptr(server.URL + "/asset"),
			},
		},
	}

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesLatestByOwnerByRepo,
			releaseResponse,
			releaseResponse, // Provide twice in case called multiple times
		),
		mock.WithRequestMatch(
			mock.GetReposReleasesTagsByOwnerByRepoByTag,
			releaseResponse,
		),
		mock.WithRequestMatchHandler(
			mock.GetReposReleasesAssetsByOwnerByRepoByAssetId,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, server.URL+"/asset", http.StatusFound)
			}),
		),
	)

	client := github.NewClient(mockedHTTPClient)
	inst := installer.New(client.Repositories)
	updater := New(client.Repositories, inst)

	ctx := context.Background()
	installPath := pkgDir

	flags := &UpdateFlags{
		Strict: false,
	}

	man, err := manifest.Read(installPath)
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	result, err := updater.Update(ctx, "owner", "repo", installPath, man, flags, nil)

	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	if result == nil {
		t.Fatal("Update() returned nil result")
	}

	if result.OldManifest.Version != "v1.0.0" {
		t.Errorf("Old version = %v, want v1.0.0", result.OldManifest.Version)
	}

	if result.Version != "v2.0.0" {
		t.Errorf("New version = %v, want v2.0.0", result.Version)
	}
}

func TestUpdate_AlreadyUpToDate(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	// Create installed package with current version
	pkgDir := filepath.Join(tmpDir, "owner", "repo")
	os.MkdirAll(pkgDir, 0755)

	m := &manifest.Manifest{
		Owner:       "owner",
		Repo:        "repo",
		Version:     "v1.0.0",
		InstallType: manifest.Release,
		Executables: []string{},
		LastUpdated: "2025-01-01 12:00:00",
	}
	m.Write(pkgDir)

	// Mock GitHub client with same version
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesLatestByOwnerByRepo,
			&github.RepositoryRelease{
				TagName: github.Ptr("v1.0.0"),
			},
		),
	)

	client := github.NewClient(mockedHTTPClient)
	inst := installer.New(client.Repositories)
	updater := New(client.Repositories, inst)

	ctx := context.Background()
	installPath := pkgDir

	flags := &UpdateFlags{
		Strict: false,
	}

	man, err := manifest.Read(installPath)
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	_, err = updater.Update(ctx, "owner", "repo", installPath, man, flags, nil)

	if err == nil {
		t.Error("Update() should return error when already up to date")
	}
}

func TestUpdate_PackageNotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	mockedHTTPClient := mock.NewMockedHTTPClient()
	client := github.NewClient(mockedHTTPClient)
	inst := installer.New(client.Repositories)
	updater := New(client.Repositories, inst)

	ctx := context.Background()
	installPath := filepath.Join(tmpDir, "owner", "nonexistent")

	flags := &UpdateFlags{
		Strict: false,
	}

	man, err := manifest.Read(installPath)
	if err == nil {
		t.Error("Update() should return an error since the manifest doesn't exist.")
	}
	_, err = updater.Update(ctx, "owner", "nonexistent", installPath, man, flags, nil)
	if err == nil {
		t.Error("Update() should return error for non-installed package")
	}
}

func TestUpdate_PreReleaseChannel(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	// Create installed pre-release package
	pkgDir := filepath.Join(tmpDir, "owner", "repo")
	os.MkdirAll(pkgDir, 0755)

	m := &manifest.Manifest{
		Owner:       "owner",
		Repo:        "repo",
		Version:     "v1.0.0-beta",
		InstallType: manifest.PreRelease,
		Executables: []string{},
		LastUpdated: "2025-01-01 12:00:00",
	}
	m.Write(pkgDir)

	archivePath := createTestArchive(t, tmpDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, archivePath)
	}))
	defer server.Close()

	assetName := fmt.Sprintf("test-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)

	preReleaseResponse := []*github.RepositoryRelease{
		{
			TagName:    github.Ptr("v2.0.0-beta"),
			Prerelease: github.Ptr(true),
			Assets: []*github.ReleaseAsset{
				{
					Name:               github.Ptr(assetName),
					BrowserDownloadURL: github.Ptr(server.URL + "/asset"),
				},
			},
		},
	}

	stableReleaseResponse := &github.RepositoryRelease{
		TagName: github.Ptr("v1.5.0"),
		Assets: []*github.ReleaseAsset{
			{
				Name:               github.Ptr(assetName),
				BrowserDownloadURL: github.Ptr(server.URL + "/asset"),
			},
		},
	}

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesByOwnerByRepo,
			preReleaseResponse,
			preReleaseResponse, // Provide twice
		),
		mock.WithRequestMatch(
			mock.GetReposReleasesLatestByOwnerByRepo,
			stableReleaseResponse,
			stableReleaseResponse, // Provide twice
		),
		mock.WithRequestMatchHandler(
			mock.GetReposReleasesAssetsByOwnerByRepoByAssetId,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, server.URL+"/asset", http.StatusFound)
			}),
		),
	)

	client := github.NewClient(mockedHTTPClient)
	inst := installer.New(client.Repositories)
	updater := New(client.Repositories, inst)

	ctx := context.Background()
	installPath := pkgDir

	flags := &UpdateFlags{
		Strict: false,
	}

	man, err := manifest.Read(installPath)
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	result, err := updater.Update(ctx, "owner", "repo", installPath, man, flags, nil)
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	// With strict=false, should update to stable if newer
	t.Logf("Updated to version: %s", result.Version)
}

func TestUpdate_StrictPreRelease(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	// Create installed pre-release package
	pkgDir := filepath.Join(tmpDir, "owner", "repo")
	os.MkdirAll(pkgDir, 0755)

	m := &manifest.Manifest{
		Owner:       "owner",
		Repo:        "repo",
		Version:     "v1.0.0-alpha",
		InstallType: manifest.PreRelease,
		Executables: []string{},
		LastUpdated: "2025-01-01 12:00:00",
	}
	m.Write(pkgDir)

	archivePath := createTestArchive(t, tmpDir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, archivePath)
	}))
	defer server.Close()

	assetName := fmt.Sprintf("test-%s-%s.tar.gz", runtime.GOOS, runtime.GOARCH)

	preReleaseResponse := []*github.RepositoryRelease{
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
	}

	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesByOwnerByRepo,
			preReleaseResponse,
			preReleaseResponse, // Provide twice in case called multiple times
		),
		mock.WithRequestMatchHandler(
			mock.GetReposReleasesAssetsByOwnerByRepoByAssetId,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, server.URL+"/asset", http.StatusFound)
			}),
		),
	)

	client := github.NewClient(mockedHTTPClient)
	inst := installer.New(client.Repositories)
	updater := New(client.Repositories, inst)

	ctx := context.Background()
	installPath := pkgDir

	flags := &UpdateFlags{
		Strict: true,
	}

	man, err := manifest.Read(installPath)
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}
	result, err := updater.Update(ctx, "owner", "repo", installPath, man, flags, nil)
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	// Should stay on pre-release channel
	if result.Version != "v1.0.0-beta" {
		t.Errorf("Version = %v, want v1.0.0-beta", result.Version)
	}
}

// Helper function
func createTestArchive(t *testing.T, baseDir string) string {
	archivePath := filepath.Join(baseDir, "update-test.tar.gz")

	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	content := []byte("updated binary content")

	hdr := &tar.Header{
		Name: "testbin",
		Mode: 0755,
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}

	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}

	return archivePath
}
