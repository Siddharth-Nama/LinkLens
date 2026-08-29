package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Siddharth-Nama/LinkLens/internal/config"
	"github.com/Siddharth-Nama/LinkLens/internal/linkedin"
	"github.com/Siddharth-Nama/LinkLens/internal/profile"
	"github.com/Siddharth-Nama/LinkLens/internal/service"
)

type handlerStubFetcher struct {
	body []byte
	err  error
}

func (s handlerStubFetcher) FetchProfile(_ context.Context, _ string) ([]byte, error) {
	return s.body, s.err
}

func TestProfileHandlerMissingURL(t *testing.T) {
	cfg := config.Config{LinkedInLIAt: "cookie"}
	svc := service.NewProfile(handlerStubFetcher{})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/profiles", nil)
	newMuxWithProfiles(cfg, svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
	assertErrorCode(t, rr, profile.CodeInvalidURL)
}

func TestProfileHandlerLinkedInNotConfigured(t *testing.T) {
	cfg := config.Config{}
	svc := service.NewProfile(nil)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/profiles?url=https://www.linkedin.com/in/jane-doe/", nil)
	newMuxWithProfiles(cfg, svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rr.Code)
	}
	assertErrorCode(t, rr, profile.CodeSessionExpired)
}

func TestProfileHandlerSuccess(t *testing.T) {
	raw, err := os.ReadFile("../../internal/linkedin/testdata/profile_ok.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	cfg := config.Config{LinkedInLIAt: "cookie"}
	svc := service.NewProfile(handlerStubFetcher{body: raw})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/profiles?url=https://www.linkedin.com/in/jane-doe/", nil)
	newMuxWithProfiles(cfg, svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}

	var out profile.Profile
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.FirstName != "Jane" {
		t.Errorf("FirstName = %q", out.FirstName)
	}
}

func TestProfileHandlerSessionExpired(t *testing.T) {
	cfg := config.Config{LinkedInLIAt: "cookie"}
	svc := service.NewProfile(handlerStubFetcher{err: linkedin.ErrSessionExpired})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/profiles?url=https://www.linkedin.com/in/jane-doe/", nil)
	newMuxWithProfiles(cfg, svc).ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rr.Code)
	}
	assertErrorCode(t, rr, profile.CodeSessionExpired)
}

func assertErrorCode(t *testing.T, rr *httptest.ResponseRecorder, code string) {
	t.Helper()
	var body profile.ErrorBody
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if body.Error.Code != code {
		t.Errorf("error code = %q, want %q", body.Error.Code, code)
	}
}
