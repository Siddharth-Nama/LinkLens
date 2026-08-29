package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Siddharth-Nama/LinkLens/internal/config"
)

type healthResponse struct {
	Status             string `json:"status"`
	LinkedInConfigured bool   `json:"linkedin_configured"`
}

func newHTTPServer(cfg config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  cfg.HTTPReadTimeout,
		WriteTimeout: cfg.HTTPWriteTimeout,
	}
}

func healthHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(healthResponse{
			Status:             "ok",
			LinkedInConfigured: cfg.LinkedInConfigured(),
		})
	}
}

func run(ctx context.Context, cfg config.Config, log *slog.Logger) error {
	log.Info("linklens ready",
		"port", cfg.Port,
		"linkedin_configured", cfg.LinkedInConfigured(),
		"api_key_required", cfg.APIKey != "",
	)
	profiles := newProfileService(cfg)
	return serve(ctx, newHTTPServer(cfg, newMuxWithProfiles(cfg, profiles)), log)
}

func serve(ctx context.Context, srv *http.Server, log *slog.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown: %w", err)
		}
		if err := <-errCh; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		log.Info("shutdown complete")
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
