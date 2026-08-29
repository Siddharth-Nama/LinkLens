package service

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestGetByURLUsesCache(t *testing.T) {
	raw, err := os.ReadFile("../linkedin/testdata/profile_ok.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	fetcher := &countingFetcher{body: raw}
	svc := NewProfile(fetcher, time.Minute)

	url := "https://www.linkedin.com/in/jane-doe/"
	if _, err := svc.GetByURL(context.Background(), url); err != nil {
		t.Fatalf("first GetByURL: %v", err)
	}
	if _, err := svc.GetByURL(context.Background(), url); err != nil {
		t.Fatalf("second GetByURL: %v", err)
	}
	if fetcher.calls != 1 {
		t.Fatalf("fetcher calls = %d, want 1 (cache hit)", fetcher.calls)
	}
}

type countingFetcher struct {
	body  []byte
	calls int
}

func (c *countingFetcher) FetchProfile(ctx context.Context, memberIdentity string) ([]byte, error) {
	c.calls++
	return c.body, nil
}
