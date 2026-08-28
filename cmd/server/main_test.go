package main

import (
	"context"
	"io"
	"log/slog"
	"testing"

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

func TestRunReturnsWhenContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := run(ctx, config.Config{Port: "8080"}, log); err != nil {
		t.Fatalf("run() unexpected error: %v", err)
	}
}
