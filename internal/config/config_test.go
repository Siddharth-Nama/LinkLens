package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	unset(t, "PORT")
	t.Setenv("LI_AT", "")
	t.Setenv("LI_JSESSIONID", "")
	t.Setenv("API_KEY", "")
	t.Setenv("HTTP_READ_TIMEOUT", "")
	t.Setenv("HTTP_WRITE_TIMEOUT", "")
	t.Setenv("LINKEDIN_TIMEOUT", "")
	t.Setenv("CACHE_TTL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.HTTPReadTimeout != 15*time.Second {
		t.Errorf("HTTPReadTimeout = %v, want 15s", cfg.HTTPReadTimeout)
	}
	if cfg.HTTPWriteTimeout != 15*time.Second {
		t.Errorf("HTTPWriteTimeout = %v, want 15s", cfg.HTTPWriteTimeout)
	}
	if cfg.LinkedInTimeout != 20*time.Second {
		t.Errorf("LinkedInTimeout = %v, want 20s", cfg.LinkedInTimeout)
	}
	if cfg.CacheTTL != 15*time.Minute {
		t.Errorf("CacheTTL = %v, want 15m", cfg.CacheTTL)
	}
	if cfg.LinkedInConfigured() {
		t.Error("LinkedInConfigured() = true, want false when LI_AT is empty")
	}
	if cfg.APIKeyRequired() {
		t.Error("APIKeyRequired() = true, want false when API_KEY is empty")
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("LI_AT", " cookie-value ")
	t.Setenv("LI_JSESSIONID", " ajax:123 ")
	t.Setenv("API_KEY", " secret ")
	t.Setenv("HTTP_READ_TIMEOUT", "5s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "8s")
	t.Setenv("LINKEDIN_TIMEOUT", "30s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.LinkedInLIAt != "cookie-value" {
		t.Errorf("LinkedInLIAt = %q, want trimmed cookie-value", cfg.LinkedInLIAt)
	}
	if cfg.LinkedInJSESSIONID != "ajax:123" {
		t.Errorf("LinkedInJSESSIONID = %q, want trimmed ajax:123", cfg.LinkedInJSESSIONID)
	}
	if cfg.APIKey != "secret" {
		t.Errorf("APIKey = %q, want trimmed secret", cfg.APIKey)
	}
	if cfg.HTTPReadTimeout != 5*time.Second {
		t.Errorf("HTTPReadTimeout = %v, want 5s", cfg.HTTPReadTimeout)
	}
	if cfg.HTTPWriteTimeout != 8*time.Second {
		t.Errorf("HTTPWriteTimeout = %v, want 8s", cfg.HTTPWriteTimeout)
	}
	if cfg.LinkedInTimeout != 30*time.Second {
		t.Errorf("LinkedInTimeout = %v, want 30s", cfg.LinkedInTimeout)
	}
	if !cfg.LinkedInConfigured() {
		t.Error("LinkedInConfigured() = false, want true when LI_AT is set")
	}
	if !cfg.APIKeyRequired() {
		t.Error("APIKeyRequired() = false, want true when API_KEY is set")
	}
}

func TestLoadRejectsEmptyPort(t *testing.T) {
	t.Setenv("PORT", "   ")
	t.Setenv("HTTP_READ_TIMEOUT", "")
	t.Setenv("HTTP_WRITE_TIMEOUT", "")
	t.Setenv("LINKEDIN_TIMEOUT", "")
	t.Setenv("CACHE_TTL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for empty PORT")
	}
}

func TestLoadRejectsInvalidDuration(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("HTTP_READ_TIMEOUT", "not-a-duration")
	t.Setenv("HTTP_WRITE_TIMEOUT", "")
	t.Setenv("LINKEDIN_TIMEOUT", "")
	t.Setenv("CACHE_TTL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for invalid HTTP_READ_TIMEOUT")
	}
}

func TestLoadRejectsNonPositiveDuration(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("HTTP_READ_TIMEOUT", "0s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "")
	t.Setenv("LINKEDIN_TIMEOUT", "")
	t.Setenv("CACHE_TTL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("Load() error = nil, want error for non-positive HTTP_READ_TIMEOUT")
	}
}

func unset(t *testing.T, key string) {
	t.Helper()
	orig, ok := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
	t.Cleanup(func() {
		if !ok {
			_ = os.Unsetenv(key)
			return
		}
		if err := os.Setenv(key, orig); err != nil {
			t.Errorf("restore %s: %v", key, err)
		}
	})
}
