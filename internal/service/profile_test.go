package service

import (
	"context"
	"os"
	"testing"

	"github.com/Siddharth-Nama/LinkLens/internal/linkedin"
)

type stubFetcher struct {
	body []byte
	err  error
}

func (s stubFetcher) FetchProfile(ctx context.Context, memberIdentity string) ([]byte, error) {
	return s.body, s.err
}

func TestGetByURLSuccess(t *testing.T) {
	raw, err := os.ReadFile("../linkedin/testdata/profile_ok.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	svc := NewProfile(stubFetcher{body: raw})
	got, err := svc.GetByURL(context.Background(), "https://www.linkedin.com/in/jane-doe/")
	if err != nil {
		t.Fatalf("GetByURL: %v", err)
	}
	if got.FirstName != "Jane" || got.PublicIdentifier != "jane-doe" {
		t.Errorf("profile = %+v", got)
	}
}

func TestGetByURLInvalidInput(t *testing.T) {
	svc := NewProfile(stubFetcher{})
	_, err := svc.GetByURL(context.Background(), "https://www.linkedin.com/company/acme")
	if err == nil {
		t.Fatal("expected error for company url")
	}
}

func TestGetByURLPropagatesLinkedInError(t *testing.T) {
	svc := NewProfile(stubFetcher{err: linkedin.ErrSessionExpired})
	_, err := svc.GetByURL(context.Background(), "https://www.linkedin.com/in/jane-doe/")
	if err != linkedin.ErrSessionExpired {
		t.Fatalf("error = %v", err)
	}
}
