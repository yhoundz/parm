package installer

import (
	"runtime"
	"testing"

	"github.com/google/go-github/v74/github"
)

func TestSelectReleaseAsset_LinuxAmd64(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping Linux-specific test")
	}

	assets := []*github.ReleaseAsset{
		{Name: github.Ptr("app-linux-amd64.tar.gz")},
		{Name: github.Ptr("app-darwin-amd64.tar.gz")},
		{Name: github.Ptr("app-windows-amd64.zip")},
	}

	matches, err := selectReleaseAsset(assets, "linux", "amd64")
	if err != nil {
		t.Fatalf("selectReleaseAsset() error: %v", err)
	}

	if len(matches) == 0 {
		t.Fatal("selectReleaseAsset() returned no matches")
	}

	// Should select Linux asset
	if matches[0].GetName() != "app-linux-amd64.tar.gz" {
		t.Errorf("selectReleaseAsset() = %v, want app-linux-amd64.tar.gz", matches[0].GetName())
	}
}

func TestSelectReleaseAsset_DarwinArm64(t *testing.T) {
	assets := []*github.ReleaseAsset{
		{Name: github.Ptr("app-linux-amd64.tar.gz")},
		{Name: github.Ptr("app-darwin-arm64.tar.gz")},
		{Name: github.Ptr("app-darwin-amd64.tar.gz")},
	}

	matches, err := selectReleaseAsset(assets, "darwin", "arm64")
	if err != nil {
		t.Fatalf("selectReleaseAsset() error: %v", err)
	}

	if len(matches) == 0 {
		t.Fatal("selectReleaseAsset() returned no matches")
	}

	// Should select macOS ARM64 asset
	if matches[0].GetName() != "app-darwin-arm64.tar.gz" {
		t.Errorf("selectReleaseAsset() = %v, want app-darwin-arm64.tar.gz", matches[0].GetName())
	}
}

func TestSelectReleaseAsset_WindowsAmd64(t *testing.T) {
	assets := []*github.ReleaseAsset{
		{Name: github.Ptr("app-linux-amd64.tar.gz")},
		{Name: github.Ptr("app-windows-amd64.zip")},
		{Name: github.Ptr("app-darwin-amd64.tar.gz")},
	}

	matches, err := selectReleaseAsset(assets, "windows", "amd64")
	if err != nil {
		t.Fatalf("selectReleaseAsset() error: %v", err)
	}

	if len(matches) == 0 {
		t.Fatal("selectReleaseAsset() returned no matches")
	}

	// Should select Windows asset (zip preferred)
	if matches[0].GetName() != "app-windows-amd64.zip" {
		t.Errorf("selectReleaseAsset() = %v, want app-windows-amd64.zip", matches[0].GetName())
	}
}

func TestSelectReleaseAsset_PreferTarGz(t *testing.T) {
	assets := []*github.ReleaseAsset{
		{Name: github.Ptr("app-linux-amd64.zip")},
		{Name: github.Ptr("app-linux-amd64.tar.gz")},
	}

	matches, err := selectReleaseAsset(assets, "linux", "amd64")
	if err != nil {
		t.Fatalf("selectReleaseAsset() error: %v", err)
	}

	if len(matches) == 0 {
		t.Fatal("selectReleaseAsset() returned no matches")
	}

	// Should prefer .tar.gz over .zip on Linux
	if matches[0].GetName() != "app-linux-amd64.tar.gz" {
		t.Errorf("selectReleaseAsset() = %v, want app-linux-amd64.tar.gz", matches[0].GetName())
	}
}

func TestSelectReleaseAsset_NoMatch(t *testing.T) {
	assets := []*github.ReleaseAsset{
		{Name: github.Ptr("app-linux-amd64.tar.gz")},
		{Name: github.Ptr("app-darwin-amd64.tar.gz")},
	}

	// Request Windows asset when only Linux/Darwin available
	matches, err := selectReleaseAsset(assets, "windows", "amd64")
	if err != nil {
		t.Fatalf("selectReleaseAsset() error: %v", err)
	}

	// Should return empty or low-score matches
	if len(matches) > 0 {
		t.Logf("selectReleaseAsset() returned %d matches (may have low scores)", len(matches))
	}
}

func TestSelectReleaseAsset_AlternativeNames(t *testing.T) {
	assets := []*github.ReleaseAsset{
		{Name: github.Ptr("app-MacOS-X86_64.tar.gz")},
		{Name: github.Ptr("app-Linux-X86_64.tar.gz")},
	}

	// Test alternative OS/arch names with capitalization
	matches, err := selectReleaseAsset(assets, "darwin", "amd64")
	if err != nil {
		t.Fatalf("selectReleaseAsset() error: %v", err)
	}

	if len(matches) == 0 {
		t.Fatal("selectReleaseAsset() returned no matches")
	}

	// Should match "MacOS" (case-insensitive) and "X86_64"
	if matches[0].GetName() != "app-MacOS-X86_64.tar.gz" {
		t.Errorf("selectReleaseAsset() = %v, want app-MacOS-X86_64.tar.gz", matches[0].GetName())
	}
}

func TestSelectReleaseAsset_MuslPenalty(t *testing.T) {
	assets := []*github.ReleaseAsset{
		{Name: github.Ptr("app-linux-amd64-musl.tar.gz")},
		{Name: github.Ptr("app-linux-amd64.tar.gz")},
	}

	matches, err := selectReleaseAsset(assets, "linux", "amd64")
	if err != nil {
		t.Fatalf("selectReleaseAsset() error: %v", err)
	}

	if len(matches) == 0 {
		t.Fatal("selectReleaseAsset() returned no matches")
	}

	// Should prefer non-musl version
	if matches[0].GetName() != "app-linux-amd64.tar.gz" {
		t.Errorf("selectReleaseAsset() = %v, want app-linux-amd64.tar.gz (non-musl)", matches[0].GetName())
	}
}

func TestGetAssetByName(t *testing.T) {
	rel := &github.RepositoryRelease{
		Assets: []*github.ReleaseAsset{
			{Name: github.Ptr("asset1.tar.gz")},
			{Name: github.Ptr("asset2.zip")},
		},
	}

	asset, err := getAssetByName(rel, "asset1.tar.gz")
	if err != nil {
		t.Fatalf("getAssetByName() error: %v", err)
	}

	if asset == nil {
		t.Fatal("getAssetByName() returned nil")
	}

	if asset.GetName() != "asset1.tar.gz" {
		t.Errorf("getAssetByName() = %v, want asset1.tar.gz", asset.GetName())
	}
}

func TestGetAssetByName_NotFound(t *testing.T) {
	rel := &github.RepositoryRelease{
		Assets: []*github.ReleaseAsset{
			{Name: github.Ptr("asset1.tar.gz")},
		},
	}

	_, err := getAssetByName(rel, "nonexistent.zip")
	if err == nil {
		t.Error("getAssetByName() should return error for non-existent asset")
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		tokens []string
		want   bool
	}{
		{"match first", "app-linux-amd64", []string{"linux", "darwin"}, true},
		{"match second", "app-darwin-arm64", []string{"linux", "darwin"}, true},
		{"no match", "app-windows-amd64", []string{"linux", "darwin"}, false},
		{"empty tokens", "app-linux-amd64", []string{}, false},
		{"empty src", "", []string{"linux"}, false},
		{"partial match", "app-x86_64", []string{"x86_64", "amd64"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsAny(tt.src, tt.tokens)
			if got != tt.want {
				t.Errorf("containsAny(%q, %v) = %v, want %v", tt.src, tt.tokens, got, tt.want)
			}
		})
	}
}

func TestSelectReleaseAsset_Arm32(t *testing.T) {
	assets := []*github.ReleaseAsset{
		{Name: github.Ptr("app-linux-armv7.tar.gz")},
		{Name: github.Ptr("app-linux-arm64.tar.gz")},
		{Name: github.Ptr("app-linux-amd64.tar.gz")},
	}

	matches, err := selectReleaseAsset(assets, "linux", "arm")
	if err != nil {
		t.Fatalf("selectReleaseAsset() error: %v", err)
	}

	if len(matches) == 0 {
		t.Fatal("selectReleaseAsset() returned no matches")
	}

	// Should select ARM32 variant
	if matches[0].GetName() != "app-linux-armv7.tar.gz" {
		t.Errorf("selectReleaseAsset() = %v, want app-linux-armv7.tar.gz", matches[0].GetName())
	}
}

func TestSelectReleaseAsset_MultipleMatches(t *testing.T) {
	assets := []*github.ReleaseAsset{
		{Name: github.Ptr("app-linux-amd64.tar.gz")},
		{Name: github.Ptr("app-linux-x86_64.tar.gz")},
	}

	matches, err := selectReleaseAsset(assets, "linux", "amd64")
	if err != nil {
		t.Fatalf("selectReleaseAsset() error: %v", err)
	}

	if len(matches) == 0 {
		t.Fatal("selectReleaseAsset() returned no matches")
	}

	// Both should match, function returns top candidates
	t.Logf("Found %d matches", len(matches))
}
