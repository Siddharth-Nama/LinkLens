package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port               string
	LinkedInLIAt       string
	LinkedInJSESSIONID string
	APIKey             string
	HTTPReadTimeout    time.Duration
	HTTPWriteTimeout   time.Duration
	LinkedInTimeout    time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Port:               port(),
		LinkedInLIAt:       strings.TrimSpace(os.Getenv("LI_AT")),
		LinkedInJSESSIONID: strings.TrimSpace(os.Getenv("LI_JSESSIONID")),
		APIKey:             strings.TrimSpace(os.Getenv("API_KEY")),
	}

	var err error
	cfg.HTTPReadTimeout, err = envDuration("HTTP_READ_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.HTTPWriteTimeout, err = envDuration("HTTP_WRITE_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	cfg.LinkedInTimeout, err = envDuration("LINKEDIN_TIMEOUT", 20*time.Second)
	if err != nil {
		return Config{}, err
	}

	if cfg.Port == "" {
		return Config{}, fmt.Errorf("PORT must not be empty")
	}

	return cfg, nil
}

func (c Config) LinkedInConfigured() bool {
	return c.LinkedInLIAt != ""
}

func (c Config) APIKeyRequired() bool {
	return c.APIKey != ""
}

func port() string {
	raw, ok := os.LookupEnv("PORT")
	if !ok {
		return "8080"
	}
	return strings.TrimSpace(raw)
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: %w", key, raw, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s must be greater than 0", key)
	}
	return d, nil
}
