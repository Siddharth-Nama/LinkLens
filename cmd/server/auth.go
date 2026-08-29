package main

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/Siddharth-Nama/LinkLens/internal/config"
	"github.com/Siddharth-Nama/LinkLens/internal/profile"
)

const apiKeyHeader = "X-API-Key"

func apiKeyAuthorized(cfg config.Config, r *http.Request) bool {
	if !cfg.APIKeyRequired() {
		return true
	}
	got := strings.TrimSpace(r.Header.Get(apiKeyHeader))
	if got == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(cfg.APIKey)) == 1
}

func writeUnauthorized(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, profile.NewError(
		profile.CodeUnauthorized,
		"missing or invalid api key",
	))
}
