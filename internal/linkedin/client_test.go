package linkedin

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNewRequestSetsVoyagerHeaders(t *testing.T) {
	c := New("cookie-value", `"ajax:123"`, 5*time.Second)
	req, err := c.NewRequest(context.Background(), http.MethodGet, "https://www.linkedin.com/voyager/api/identity/dash/profiles")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}

	if got := req.Header.Get("Accept"); got != acceptValue {
		t.Errorf("Accept = %q, want %q", got, acceptValue)
	}
	if got := req.Header.Get("X-RestLi-Protocol-Version"); got != restliVersion {
		t.Errorf("X-RestLi-Protocol-Version = %q, want %q", got, restliVersion)
	}
	if got := req.Header.Get("User-Agent"); got != defaultUserAgent {
		t.Errorf("User-Agent = %q", got)
	}
	if got := req.Header.Get("Referer"); got != referer {
		t.Errorf("Referer = %q, want %q", got, referer)
	}
	if got := req.Header.Get("Csrf-Token"); got != "ajax:123" {
		t.Errorf("Csrf-Token = %q, want ajax:123 without quotes", got)
	}

	cookies := map[string]string{}
	for _, cookie := range req.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}
	if cookies["li_at"] != "cookie-value" {
		t.Errorf("li_at cookie = %q, want cookie-value", cookies["li_at"])
	}
	if cookies["JSESSIONID"] != "ajax:123" {
		t.Errorf("JSESSIONID cookie = %q, want ajax:123", cookies["JSESSIONID"])
	}
}

func TestCsrfTokenWithoutQuotes(t *testing.T) {
	c := New("x", "ajax:456", time.Second)
	req, err := c.NewRequest(context.Background(), http.MethodGet, "https://www.linkedin.com/")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if got := req.Header.Get("Csrf-Token"); got != "ajax:456" {
		t.Errorf("Csrf-Token = %q, want ajax:456", got)
	}
}

func TestNewRequestOmitsAuthWhenEmpty(t *testing.T) {
	c := New("", "", time.Second)
	req, err := c.NewRequest(context.Background(), http.MethodGet, "https://www.linkedin.com/")
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	if req.Header.Get("Csrf-Token") != "" {
		t.Errorf("Csrf-Token = %q, want empty", req.Header.Get("Csrf-Token"))
	}
	if len(req.Cookies()) != 0 {
		t.Errorf("cookies = %v, want none", req.Cookies())
	}
}

func TestNewUsesDefaultTimeout(t *testing.T) {
	c := New("a", "b", 0)
	if c.http.Timeout != 20*time.Second {
		t.Errorf("Timeout = %v, want 20s", c.http.Timeout)
	}
}

func TestNewRequestInvalidURL(t *testing.T) {
	c := New("a", "b", time.Second)
	_, err := c.NewRequest(context.Background(), http.MethodGet, "://bad")
	if err == nil {
		t.Fatal("NewRequest error = nil, want error")
	}
}
