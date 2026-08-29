package linkedin

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestFetchProfileOK(t *testing.T) {
	fixture, err := os.ReadFile("testdata/profile_ok.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var gotPath string
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		if r.Header.Get("Csrf-Token") != "ajax:123" {
			t.Errorf("Csrf-Token = %q", r.Header.Get("Csrf-Token"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(srv.Close)

	c := testClient(srv.URL)
	body, err := c.FetchProfile(context.Background(), "jane-doe")
	if err != nil {
		t.Fatalf("FetchProfile: %v", err)
	}
	if gotPath != profilesPath {
		t.Errorf("path = %q, want %q", gotPath, profilesPath)
	}
	q, err := url.ParseQuery(gotQuery)
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	if q.Get("q") != "memberIdentity" || q.Get("memberIdentity") != "jane-doe" {
		t.Errorf("query = %q", gotQuery)
	}
	if q.Get("decorationId") != DefaultDecorationID {
		t.Errorf("decorationId = %q", q.Get("decorationId"))
	}

	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
}

func TestFetchProfileStatusErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   error
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, want: ErrSessionExpired},
		{name: "forbidden", status: http.StatusForbidden, want: ErrSessionExpired},
		{name: "linkedin 999", status: 999, want: ErrSessionExpired},
		{name: "not found", status: http.StatusNotFound, want: ErrNotFound},
		{name: "rate limited", status: http.StatusTooManyRequests, want: ErrRateLimited},
		{name: "bad gateway", status: http.StatusBadGateway, want: ErrUpstream},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, `{"error":"upstream"}`)
			}))
			t.Cleanup(srv.Close)

			_, err := testClient(srv.URL).FetchProfile(context.Background(), "jane-doe")
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestFetchProfileRejectsEmptyIdentity(t *testing.T) {
	c := New("a", "b", time.Second)
	_, err := c.FetchProfile(context.Background(), "  ")
	if !errors.Is(err, ErrInvalidIdentity) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidIdentity)
	}
}

func testClient(baseURL string) *Client {
	c := New("cookie", `"ajax:123"`, 2*time.Second)
	c.baseURL = baseURL
	return c
}
