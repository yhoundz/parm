package gh

import (
	"context"
	"os"
	"testing"

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
		old := os.Getenv(env)
		os.Unsetenv(env)
		defer func(k, v string) {
			if v != "" {
				os.Setenv(k, v)
			}
		}(env, old)
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
		old := os.Getenv(env)
		os.Unsetenv(env)
		defer func(k, v string) {
			if v != "" {
				os.Setenv(k, v)
			}
		}(env, old)
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
		old := os.Getenv(env)
		os.Unsetenv(env)
		defer func(k, v string) {
			if v != "" {
				os.Setenv(k, v)
			}
		}(env, old)
	}

	v := viper.New()

	_, err := GetStoredApiKey(v)
	if err == nil {
		t.Error("GetStoredApiKey() should return error when no token is set")
	}
}

func TestGetStoredApiKey_FromEnv(t *testing.T) {
	v := viper.New()

	envToken := "env_token_abc"
	os.Setenv("PARM_GITHUB_TOKEN", envToken)
	defer os.Unsetenv("PARM_GITHUB_TOKEN")

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
		old := os.Getenv(env)
		os.Unsetenv(env)
		defer func(k, v string) {
			if v != "" {
				os.Setenv(k, v)
			}
		}(env, old)
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
