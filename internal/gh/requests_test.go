package gh

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-github/v74/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
)

func TestGetLatestPreRelease(t *testing.T) {
	// Create mock client with pre-release data
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesByOwnerByRepo,
			[]*github.RepositoryRelease{
				{
					TagName:    github.Ptr("v1.0.1"),
					Prerelease: github.Ptr(false),
				},
				{
					TagName:    github.Ptr("v1.0.0-beta"),
					Prerelease: github.Ptr(true),
				},
			},
		),
	)
	
	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	
	rel, err := GetLatestPreRelease(ctx, client.Repositories, "owner", "repo")
	if err != nil {
		t.Fatalf("GetLatestPreRelease() error: %v", err)
	}
	
	if rel == nil {
		t.Fatal("GetLatestPreRelease() returned nil")
	}
	
	if !rel.GetPrerelease() {
		t.Error("GetLatestPreRelease() returned non-prerelease")
	}
	
	if rel.GetTagName() != "v1.0.0-beta" {
		t.Errorf("GetLatestPreRelease() tag = %v, want v1.0.0-beta", rel.GetTagName())
	}
}

func TestGetLatestPreRelease_NoPreRelease(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesByOwnerByRepo,
			[]*github.RepositoryRelease{
				{
					TagName:    github.Ptr("v1.0.0"),
					Prerelease: github.Ptr(false),
				},
			},
		),
	)
	
	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	
	rel, err := GetLatestPreRelease(ctx, client.Repositories, "owner", "repo")
	if err != nil {
		t.Fatalf("GetLatestPreRelease() error: %v", err)
	}
	
	if rel != nil {
		t.Error("GetLatestPreRelease() should return nil when no pre-release exists")
	}
}

func TestResolvePreRelease(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesByOwnerByRepo,
			[]*github.RepositoryRelease{
				{
					TagName:    github.Ptr("v2.0.0-alpha"),
					Prerelease: github.Ptr(true),
				},
			},
		),
	)
	
	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	
	rel, err := ResolvePreRelease(ctx, client.Repositories, "owner", "repo")
	if err != nil {
		t.Fatalf("ResolvePreRelease() error: %v", err)
	}
	
	if rel == nil {
		t.Fatal("ResolvePreRelease() returned nil")
	}
	
	if rel.GetTagName() != "v2.0.0-alpha" {
		t.Errorf("ResolvePreRelease() tag = %v, want v2.0.0-alpha", rel.GetTagName())
	}
}

func TestResolvePreRelease_NoPreRelease(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesByOwnerByRepo,
			[]*github.RepositoryRelease{},
		),
	)
	
	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	
	_, err := ResolvePreRelease(ctx, client.Repositories, "owner", "repo")
	if err == nil {
		t.Error("ResolvePreRelease() should return error when no pre-release found")
	}
}

func TestResolveReleaseByTag_SpecificTag(t *testing.T) {
	tag := "v1.0.0"
	
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesTagsByOwnerByRepoByTag,
			&github.RepositoryRelease{
				TagName:    github.Ptr(tag),
				Prerelease: github.Ptr(false),
			},
		),
	)
	
	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	
	rel, err := ResolveReleaseByTag(ctx, client.Repositories, "owner", "repo", &tag)
	if err != nil {
		t.Fatalf("ResolveReleaseByTag() error: %v", err)
	}
	
	if rel == nil {
		t.Fatal("ResolveReleaseByTag() returned nil")
	}
	
	if rel.GetTagName() != tag {
		t.Errorf("ResolveReleaseByTag() tag = %v, want %v", rel.GetTagName(), tag)
	}
}

func TestResolveReleaseByTag_LatestRelease(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesLatestByOwnerByRepo,
			&github.RepositoryRelease{
				TagName:    github.Ptr("v2.0.0"),
				Prerelease: github.Ptr(false),
			},
		),
	)
	
	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	
	rel, err := ResolveReleaseByTag(ctx, client.Repositories, "owner", "repo", nil)
	if err != nil {
		t.Fatalf("ResolveReleaseByTag() error: %v", err)
	}
	
	if rel == nil {
		t.Fatal("ResolveReleaseByTag() returned nil")
	}
	
	if rel.GetTagName() != "v2.0.0" {
		t.Errorf("ResolveReleaseByTag() tag = %v, want v2.0.0", rel.GetTagName())
	}
}

func TestResolveReleaseByTag_NonExistentTag(t *testing.T) {
	tag := "v99.99.99"
	
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposReleasesTagsByOwnerByRepoByTag,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message": "Not Found"}`))
			}),
		),
	)
	
	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	
	_, err := ResolveReleaseByTag(ctx, client.Repositories, "owner", "repo", &tag)
	if err == nil {
		t.Error("ResolveReleaseByTag() should return error for non-existent tag")
	}
}

func TestValidatePreRelease(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesByOwnerByRepo,
			[]*github.RepositoryRelease{
				{
					TagName:    github.Ptr("v1.0.0-rc1"),
					Prerelease: github.Ptr(true),
				},
			},
		),
	)
	
	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	
	valid, rel, err := validatePreRelease(ctx, client.Repositories, "owner", "repo")
	if err != nil {
		t.Fatalf("validatePreRelease() error: %v", err)
	}
	
	if !valid {
		t.Error("validatePreRelease() returned false for valid pre-release")
	}
	
	if rel == nil {
		t.Error("validatePreRelease() returned nil release")
	}
}

func TestValidateRelease(t *testing.T) {
	tag := "v1.0.0"
	
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposReleasesTagsByOwnerByRepoByTag,
			&github.RepositoryRelease{
				TagName: github.Ptr(tag),
			},
		),
	)
	
	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()
	
	valid, rel, err := validateRelease(ctx, client.Repositories, "owner", "repo", tag)
	if err != nil {
		t.Fatalf("validateRelease() error: %v", err)
	}
	
	if !valid {
		t.Error("validateRelease() returned false for valid release")
	}
	
	if rel == nil {
		t.Error("validateRelease() returned nil release")
	}
}

