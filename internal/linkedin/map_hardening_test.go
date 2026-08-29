package linkedin

import (
	"errors"
	"os"
	"testing"
)

func TestMapProfileAmbiguousIncluded(t *testing.T) {
	raw, err := os.ReadFile("testdata/profile_ambiguous.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	got, err := MapProfile(raw, "jane-doe", "in", "canon")
	if err != nil {
		t.Fatalf("MapProfile: %v", err)
	}
	if got.FirstName != "Jane" {
		t.Errorf("FirstName = %q, want Jane", got.FirstName)
	}
	if len(got.Experience) != 1 {
		t.Fatalf("Experience len = %d, want 1", len(got.Experience))
	}
	if got.Experience[0].Title != "Senior Engineer" {
		t.Errorf("Experience = %+v", got.Experience)
	}
}

func TestBelongsToProfile(t *testing.T) {
	urn := "urn:li:fsd_profile:ACoAAAJane"
	tests := []struct {
		name string
		item map[string]any
		want bool
	}{
		{
			name: "no profile ref includes item",
			item: map[string]any{"title": "Engineer"},
			want: true,
		},
		{
			name: "matching profile ref",
			item: map[string]any{"*profile": urn},
			want: true,
		},
		{
			name: "other profile ref excluded",
			item: map[string]any{"*profile": "urn:li:fsd_profile:other"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := belongsToProfile(tt.item, urn); got != tt.want {
				t.Errorf("belongsToProfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapProfileLiveFixtureOptional(t *testing.T) {
	raw, err := os.ReadFile("testdata/profile_live.json")
	if errors.Is(err, os.ErrNotExist) {
		t.Skip("testdata/profile_live.json not found; save a real anonymized LinkedIn response locally to enable")
	}
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	slug := os.Getenv("LIVE_PROFILE_SLUG")
	if slug == "" {
		slug = "jane-doe"
	}

	got, err := MapProfile(raw, slug, "live-test", "live-test")
	if err != nil {
		t.Fatalf("MapProfile: %v", err)
	}
	if got.FullName == "" && got.Headline == "" {
		t.Fatalf("live fixture mapped to empty profile: %+v", got)
	}
}
