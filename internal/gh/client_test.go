package gh

import (
	"context"
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
	v := viper.New()

	_, err := GetStoredApiKey(v)
	if err == nil {
		t.Error("GetStoredApiKey() should return error when no token is set")
	}
}

func TestGetStoredApiKey_EmptyTokens(t *testing.T) {
	v := viper.New()

	v.Set("github_api_token", "")
	v.Set("github_api_token_fallback", "")

	_, err := GetStoredApiKey(v)
	if err == nil {
		t.Error("GetStoredApiKey() should return error for empty tokens")
	}
}
