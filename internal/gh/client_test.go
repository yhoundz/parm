package gh

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-github/v74/github"
	"github.com/migueleliasweb/go-github-mock/src/mock"
	"github.com/spf13/viper"
)

func TestNew_WithToken(t *testing.T) {
	ctx := context.Background()
	token := "test_token_123"

	provider := New(ctx, token)
	if provider == nil {
		t.Fatal("New() returned nil")
	}

	// Verify we can get services
	repos := provider.Repos()
	if repos == nil {
		t.Error("Repos() returned nil")
	}

	search := provider.Search()
	if search == nil {
		t.Error("Search() returned nil")
	}
}

func TestNew_WithoutToken(t *testing.T) {
	ctx := context.Background()
	token := ""

	provider := New(ctx, token)
	if provider == nil {
		t.Fatal("New() returned nil")
	}

	// Should still work without token (unauthenticated)
	repos := provider.Repos()
	if repos == nil {
		t.Error("Repos() returned nil for unauthenticated client")
	}
}

func TestGetStoredApiKey_FromFallback(t *testing.T) {
	// Clear env vars to ensure fallback logic is tested
	for _, env := range []string{"PARM_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"} {
		t.Setenv(env, "")
	}

	v := viper.New()

	testToken := "fallback_token_123"
	v.Set("github_api_token_fallback", testToken)

	token, err := GetStoredApiKey(v)
	if err != nil {
		t.Fatalf("GetStoredApiKey() error: %v", err)
	}

	if token != testToken {
		t.Errorf("GetStoredApiKey() = %v, want %v", token, testToken)
	}
}

func TestGetStoredApiKey_FromMain(t *testing.T) {
	// Clear env vars
	for _, env := range []string{"PARM_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"} {
		t.Setenv(env, "")
	}

	v := viper.New()

	mainToken := "main_token_123"
	v.Set("github_api_token", mainToken)

	fallbackToken := "fallback_token_456"
	v.Set("github_api_token_fallback", fallbackToken)

	token, err := GetStoredApiKey(v)
	if err != nil {
		t.Fatalf("GetStoredApiKey() error: %v", err)
	}

	// Should prefer main token over fallback
	if token != mainToken {
		t.Errorf("GetStoredApiKey() = %v, want %v (main token)", token, mainToken)
	}
}

func TestGetStoredApiKey_NoToken(t *testing.T) {
	// Clear env vars
	for _, env := range []string{"PARM_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"} {
		t.Setenv(env, "")
	}

	v := viper.New()

	_, err := GetStoredApiKey(v)
	if err == nil {
		t.Error("GetStoredApiKey() should return error when no token is set")
	}
}

func TestGetStoredApiKey_FromEnv(t *testing.T) {
	v := viper.New()

	// Set config values to verify env takes precedence
	v.Set("github_api_token", "config_token_should_be_ignored")
	v.Set("github_api_token_fallback", "fallback_token_should_be_ignored")

	envToken := "env_token_abc"
	t.Setenv("PARM_GITHUB_TOKEN", envToken)

	token, err := GetStoredApiKey(v)
	if err != nil {
		t.Fatalf("GetStoredApiKey() error: %v", err)
	}

	if token != envToken {
		t.Errorf("GetStoredApiKey() = %v, want %v (env token)", token, envToken)
	}
}

func TestGetStoredApiKey_Trimmed(t *testing.T) {
	// Clear env vars
	for _, env := range []string{"PARM_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"} {
		t.Setenv(env, "")
	}

	v := viper.New()

	v.Set("github_api_token", "  token_with_spaces  ")

	token, err := GetStoredApiKey(v)
	if err != nil {
		t.Fatalf("GetStoredApiKey() error: %v", err)
	}

	if token != "token_with_spaces" {
		t.Errorf("GetStoredApiKey() = %v, want token_with_spaces", token)
	}
}

func TestCheckRepositoryVisibility_Public(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposByOwnerByRepo,
			&github.Repository{
				Name:    github.Ptr("repo"),
				Owner:   &github.User{Login: github.Ptr("owner")},
				Private: github.Ptr(false),
			},
		),
	)

	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()

	visibility, err := CheckRepositoryVisibility(ctx, client.Repositories, "owner", "repo")
	if err != nil {
		t.Fatalf("CheckRepositoryVisibility() error: %v", err)
	}

	if visibility != RepoPublic {
		t.Errorf("CheckRepositoryVisibility() = %v, want RepoPublic", visibility)
	}
}

func TestCheckRepositoryVisibility_Private(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatch(
			mock.GetReposByOwnerByRepo,
			&github.Repository{
				Name:    github.Ptr("repo"),
				Owner:   &github.User{Login: github.Ptr("owner")},
				Private: github.Ptr(true),
			},
		),
	)

	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()

	visibility, err := CheckRepositoryVisibility(ctx, client.Repositories, "owner", "repo")
	if err != nil {
		t.Fatalf("CheckRepositoryVisibility() error: %v", err)
	}

	if visibility != RepoPrivate {
		t.Errorf("CheckRepositoryVisibility() = %v, want RepoPrivate", visibility)
	}
}

func TestCheckRepositoryVisibility_NotFound(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"message": "Not Found"}`))
			}),
		),
	)

	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()

	visibility, err := CheckRepositoryVisibility(ctx, client.Repositories, "owner", "nonexistent")
	if err != nil {
		t.Fatalf("CheckRepositoryVisibility() error: %v", err)
	}

	if visibility != RepoNotFound {
		t.Errorf("CheckRepositoryVisibility() = %v, want RepoNotFound", visibility)
	}
}

func TestCheckRepositoryVisibility_NetworkError(t *testing.T) {
	mockedHTTPClient := mock.NewMockedHTTPClient(
		mock.WithRequestMatchHandler(
			mock.GetReposByOwnerByRepo,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"message": "Internal Server Error"}`))
			}),
		),
	)

	client := github.NewClient(mockedHTTPClient)
	ctx := context.Background()

	visibility, err := CheckRepositoryVisibility(ctx, client.Repositories, "owner", "repo")
	if err == nil {
		t.Error("CheckRepositoryVisibility() should return error for non-404 errors")
	}

	if visibility != RepoNotFound {
		t.Errorf("CheckRepositoryVisibility() = %v, want RepoNotFound on error", visibility)
	}
}
