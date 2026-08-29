package cache

import (
	"testing"
	"time"
)

func TestTTLHitMissAndExpiry(t *testing.T) {
	c := NewTTL[string](50 * time.Millisecond)
	c.Set("jane-doe", "cached")

	got, ok := c.Get("jane-doe")
	if !ok || got != "cached" {
		t.Fatalf("Get() = %q %v, want cached true", got, ok)
	}

	_, ok = c.Get("missing")
	if ok {
		t.Fatal("expected cache miss")
	}

	time.Sleep(60 * time.Millisecond)
	_, ok = c.Get("jane-doe")
	if ok {
		t.Fatal("expected expired entry to miss")
	}
}
