package catalog

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"parm/internal/config"
	"parm/internal/manifest"

	"github.com/google/go-github/v74/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestGetPackageInfo_Downstream(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	// Create installed package
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

	ctx := context.Background()
	info, err := GetPackageInfo(ctx, nil, "owner", "repo", false)
	if err != nil {
		t.Fatalf("GetPackageInfo() error: %v", err)
	}

	if info.Owner != "owner" {
		t.Errorf("Owner = %v, want owner", info.Owner)
	}

	if info.Repo != "repo" {
		t.Errorf("Repo = %v, want repo", info.Repo)
	}

	if info.Version != "v1.0.0" {
		t.Errorf("Version = %v, want v1.0.0", info.Version)
	}

	if info.DownstreamInfo == nil {
		t.Error("DownstreamInfo should not be nil")
	}

	if info.UpstreamInfo != nil {
		t.Error("UpstreamInfo should be nil for downstream")
	}
}

func TestGetPackageInfo_Upstream(t *testing.T) {
	// Create mock GitHub client
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposByOwnerByRepo,
			&github.Repository{
				Name:            github.Ptr("repo"),
				Owner:           &github.User{Login: github.Ptr("owner")},
				StargazersCount: github.Ptr(100),
				License:         &github.License{Name: github.Ptr("MIT")},
				Description:     github.Ptr("Test repository"),
			},
		),
		mock.WithRequestMatch(
			mock.GetReposReleasesLatestByOwnerByRepo,
			&github.RepositoryRelease{
				TagName:     github.Ptr("v1.0.0"),
				PublishedAt: &github.Timestamp{},
			},
		),
	)

	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()

	info, err := GetPackageInfo(ctx, client.Repositories, "owner", "repo", true)
	if err != nil {
		t.Fatalf("GetPackageInfo() error: %v", err)
	}

	if info.Owner != "owner" {
		t.Errorf("Owner = %v, want owner", info.Owner)
	}

	if info.Repo != "repo" {
		t.Errorf("Repo = %v, want repo", info.Repo)
	}

	if info.UpstreamInfo == nil {
		t.Fatal("UpstreamInfo should not be nil")
	}

	if info.UpstreamInfo.Stars != 100 {
		t.Errorf("Stars = %v, want 100", info.UpstreamInfo.Stars)
	}

	if info.UpstreamInfo.License != "MIT" {
		t.Errorf("License = %v, want MIT", info.UpstreamInfo.License)
	}

	if info.DownstreamInfo != nil {
		t.Error("DownstreamInfo should be nil for upstream")
	}
}

func TestGetPackageInfo_DownstreamNotInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	config.Cfg.ParmPkgPath = tmpDir

	ctx := context.Background()
	_, err := GetPackageInfo(ctx, nil, "owner", "nonexistent", false)
	if err == nil {
		t.Error("GetPackageInfo() should return error for non-existent package")
	}
}

func TestInfo_String(t *testing.T) {
	info := Info{
		Owner:       "owner",
		Repo:        "repo",
		Version:     "v1.0.0",
		LastUpdated: "2025-01-01 12:00:00",
		DownstreamInfo: &DownstreamInfo{
			InstallPath: "/path/to/install",
		},
	}

	str := info.String()

	if str == "" {
		t.Error("String() returned empty string")
	}

	// Should contain key information
	if !anySubstring(str, "owner") {
		t.Error("String() should contain owner")
	}

	if !anySubstring(str, "repo") {
		t.Error("String() should contain repo")
	}

	if !anySubstring(str, "v1.0.0") {
		t.Error("String() should contain version")
	}
}

func TestDownstreamInfo_String(t *testing.T) {
	info := &DownstreamInfo{
		InstallPath: "/path/to/install",
	}

	str := info.string()

	if str == "" {
		t.Error("string() returned empty string")
	}

	if !anySubstring(str, "/path/to/install") {
		t.Error("string() should contain install path")
	}
}

func TestUpstreamInfo_String(t *testing.T) {
	info := &UpstreamInfo{
		Stars:       100,
		License:     "MIT",
		Description: "Test description",
	}

	str := info.string()

	if str == "" {
		t.Error("string() returned empty string")
	}

	if !anySubstring(str, "100") {
		t.Error("string() should contain stars count")
	}

	if !anySubstring(str, "MIT") {
		t.Error("string() should contain license")
	}

	if !anySubstring(str, "Test description") {
		t.Error("string() should contain description")
	}
}
