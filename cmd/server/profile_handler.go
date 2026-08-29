package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Siddharth-Nama/LinkLens/internal/config"
	"github.com/Siddharth-Nama/LinkLens/internal/linkedin"
	"github.com/Siddharth-Nama/LinkLens/internal/profile"
	"github.com/Siddharth-Nama/LinkLens/internal/profileurl"
	"github.com/Siddharth-Nama/LinkLens/internal/service"
)

func newMux(cfg config.Config) *http.ServeMux {
	return newMuxWithProfiles(cfg, newProfileService(cfg))
}

func newMuxWithProfiles(cfg config.Config, profiles *service.ProfileService) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler(cfg))
	mux.HandleFunc("GET /v1/profiles", profileHandler(cfg, profiles))
	return mux
}

func profileHandler(cfg config.Config, profiles *service.ProfileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !apiKeyAuthorized(cfg, r) {
			writeUnauthorized(w)
			return
		}

		if !cfg.LinkedInConfigured() {
			writeJSON(w, http.StatusBadGateway, profile.NewError(
				profile.CodeSessionExpired,
				"linkedin session is not configured on the server",
			))
			return
		}

		rawURL := r.URL.Query().Get("url")
		out, err := profiles.GetByURL(r.Context(), rawURL)
		if err != nil {
			writeProfileError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, out)
	}
}

func writeProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, profileurl.ErrEmpty),
		errors.Is(err, profileurl.ErrTooLong),
		errors.Is(err, profileurl.ErrInvalid),
		errors.Is(err, profileurl.ErrNotLinkedIn),
		errors.Is(err, profileurl.ErrNotProfile),
		errors.Is(err, profileurl.ErrBadSlug):
		writeJSON(w, http.StatusBadRequest, profile.NewError(profile.CodeInvalidURL, service.InvalidURLMessage(err)))
	case errors.Is(err, linkedin.ErrNotFound):
		writeJSON(w, http.StatusNotFound, profile.NewError(profile.CodeNotFound, "linkedin profile not found or not visible"))
	case errors.Is(err, linkedin.ErrSessionExpired), errors.Is(err, service.ErrLinkedInNotConfigured):
		writeJSON(w, http.StatusBadGateway, profile.NewError(profile.CodeSessionExpired, "linkedin session expired or not authorized"))
	case errors.Is(err, linkedin.ErrRateLimited):
		writeJSON(w, http.StatusTooManyRequests, profile.NewError(profile.CodeRateLimited, "linkedin rate limit reached, try again later"))
	case errors.Is(err, linkedin.ErrUpstream), errors.Is(err, linkedin.ErrInvalidIdentity):
		writeJSON(w, http.StatusBadGateway, profile.NewError(profile.CodeUpstreamError, "failed to fetch profile from linkedin"))
	default:
		writeJSON(w, http.StatusInternalServerError, profile.NewError(profile.CodeInternal, "unexpected server error"))
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func newProfileService(cfg config.Config) *service.ProfileService {
	if !cfg.LinkedInConfigured() {
		return service.NewProfile(nil)
	}
	li := linkedin.New(cfg.LinkedInLIAt, cfg.LinkedInJSESSIONID, cfg.LinkedInTimeout)
	return service.NewProfile(li)
}
