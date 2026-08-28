package profileurl

import (
	"errors"
	"strings"
	"testing"
)

func TestParseSuccess(t *testing.T) {
	tests := []struct {
		in   string
		slug string
	}{
		{in: "https://www.linkedin.com/in/jane-doe/", slug: "jane-doe"},
		{in: "https://linkedin.com/in/jane-doe", slug: "jane-doe"},
		{in: "http://www.linkedin.com/in/jane-doe", slug: "jane-doe"},
		{in: "www.linkedin.com/in/jane-doe", slug: "jane-doe"},
		{in: "https://in.linkedin.com/in/jane-doe", slug: "jane-doe"},
		{in: "https://uk.linkedin.com/in/jane-doe/", slug: "jane-doe"},
		{in: "https://www.linkedin.com/in/jane-doe?trk=public", slug: "jane-doe"},
		{in: "  https://www.linkedin.com/in/jane-doe/  ", slug: "jane-doe"},
		{in: "https://www.linkedin.com/in/jane-doe/overlay/photo/", slug: "jane-doe"},
		{in: "https://www.linkedin.com/mwlite/in/jane-doe", slug: "jane-doe"},
		{in: "https://www.linkedin.com/in/BillGates", slug: "BillGates"},
	}

	for _, tt := range tests {
		got, err := Parse(tt.in)
		if err != nil {
			t.Errorf("Parse(%q) unexpected error: %v", tt.in, err)
			continue
		}
		if got.PublicIdentifier != tt.slug {
			t.Errorf("Parse(%q).PublicIdentifier = %q, want %q", tt.in, got.PublicIdentifier, tt.slug)
		}
		wantCanonical := "https://www.linkedin.com/in/" + tt.slug + "/"
		if got.CanonicalURL != wantCanonical {
			t.Errorf("Parse(%q).CanonicalURL = %q, want %q", tt.in, got.CanonicalURL, wantCanonical)
		}
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		in   string
		want error
	}{
		{in: "", want: ErrEmpty},
		{in: "   ", want: ErrEmpty},
		{in: strings.Repeat("a", maxURLLength+1), want: ErrTooLong},
		{in: "https://example.com/in/jane-doe", want: ErrNotLinkedIn},
		{in: "https://linkedin.com.evil.example/in/jane-doe", want: ErrNotLinkedIn},
		{in: "https://notlinkedin.com/in/jane-doe", want: ErrNotLinkedIn},
		{in: "https://www.linkedin.com/company/tross", want: ErrNotProfile},
		{in: "https://www.linkedin.com/school/stanford-university", want: ErrNotProfile},
		{in: "https://www.linkedin.com/feed/", want: ErrNotProfile},
		{in: "https://www.linkedin.com/in/", want: ErrNotProfile},
		{in: "https://www.linkedin.com/", want: ErrNotProfile},
		{in: "https://www.linkedin.com/in/ab", want: ErrBadSlug},
		{in: "https://www.linkedin.com/in/-jane-doe", want: ErrBadSlug},
		{in: "https://www.linkedin.com/in/jane-doe-", want: ErrBadSlug},
		{in: "https://www.linkedin.com/in/jane_doe", want: ErrBadSlug},
		{in: "://broken", want: ErrInvalid},
	}

	for _, tt := range tests {
		_, err := Parse(tt.in)
		if !errors.Is(err, tt.want) {
			t.Errorf("Parse(%q) error = %v, want %v", tt.in, err, tt.want)
		}
	}
}
