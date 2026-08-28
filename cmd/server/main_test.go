package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Siddharth-Nama/LinkLens/internal/config"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		in      string
		want    slog.Level
		wantErr bool
	}{
		{in: "", want: slog.LevelInfo},
		{in: "info", want: slog.LevelInfo},
		{in: " INFO ", want: slog.LevelInfo},
		{in: "debug", want: slog.LevelDebug},
		{in: "warn", want: slog.LevelWarn},
		{in: "warning", want: slog.LevelWarn},
		{in: "error", want: slog.LevelError},
		{in: "fatal", wantErr: true},
	}
	for _, tt := range tests {
		got, err := parseLogLevel(tt.in)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseLogLevel(%q) error = nil, want error", tt.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseLogLevel(%q) unexpected error: %v", tt.in, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseLogLevel(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestHealth(t *testing.T) {
	tests := []struct {
		name   string
		cfg    config.Config
		wantLI bool
	}{
		{name: "linkedin not configured", cfg: config.Config{}, wantLI: false},
		{name: "linkedin configured", cfg: config.Config{LinkedInLIAt: "cookie"}, wantLI: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rr := httptest.NewRecorder()
			newMux(tt.cfg).ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
			}
			if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}

			var body healthResponse
			if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Status != "ok" {
				t.Errorf("status = %q, want ok", body.Status)
			}
			if body.LinkedInConfigured != tt.wantLI {
				t.Errorf("linkedin_configured = %v, want %v", body.LinkedInConfigured, tt.wantLI)
			}
		})
	}
}

func TestHealthRejectsPOST(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()
	newMux(config.Config{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestRunShutsDownListener(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.Config{
		Port:             "0",
		HTTPReadTimeout:  time.Second,
		HTTPWriteTimeout: time.Second,
	}
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, cfg, log)
	}()

	time.Sleep(150 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("run() unexpected error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("run() did not return after shutdown")
	}
}
