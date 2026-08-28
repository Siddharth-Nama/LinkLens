package profile

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestNewMarshalsEmptySlicesNotNull(t *testing.T) {
	p := New()
	p.FetchedAt = time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)

	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	s := string(raw)
	for _, key := range []string{
		`"profilePictures":[]`,
		`"experience":[]`,
		`"education":[]`,
		`"skills":[]`,
		`"missingSections":[]`,
	} {
		if !strings.Contains(s, key) {
			t.Errorf("JSON missing %s\n%s", key, s)
		}
	}
	if !strings.Contains(s, `"connectionsCount":null`) {
		t.Errorf("JSON missing connectionsCount null\n%s", s)
	}
	if !strings.Contains(s, `"backgroundImage":null`) {
		t.Errorf("JSON missing backgroundImage null\n%s", s)
	}
}

func TestProfileRoundTrip(t *testing.T) {
	n := 12
	month := 3
	year := 2020
	p := New()
	p.InputURL = "https://www.linkedin.com/in/jane-doe/"
	p.CanonicalURL = "https://www.linkedin.com/in/jane-doe/"
	p.PublicIdentifier = "jane-doe"
	p.FirstName = "Jane"
	p.LastName = "Doe"
	p.FullName = "Jane Doe"
	p.Headline = "Engineer"
	p.About = "About text"
	p.Location = Location{Country: "US", City: "SF", Raw: "San Francisco Bay Area"}
	p.ConnectionsCount = &n
	p.Experience = []Experience{{
		Title:       "SWE",
		CompanyName: "Acme",
		StartDate:   &Date{Month: &month, Year: &year},
		IsCurrent:   true,
	}}
	p.FetchedAt = time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)

	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Profile
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.FullName != "Jane Doe" || got.PublicIdentifier != "jane-doe" {
		t.Errorf("round trip identity = %+v", got)
	}
	if got.ConnectionsCount == nil || *got.ConnectionsCount != 12 {
		t.Errorf("connectionsCount = %v", got.ConnectionsCount)
	}
	if len(got.Experience) != 1 || got.Experience[0].Title != "SWE" {
		t.Errorf("experience = %+v", got.Experience)
	}
	if got.Experience[0].StartDate == nil || got.Experience[0].StartDate.Year == nil || *got.Experience[0].StartDate.Year != 2020 {
		t.Errorf("startDate = %+v", got.Experience[0].StartDate)
	}
}

func TestErrorBodyJSON(t *testing.T) {
	raw, err := json.Marshal(NewError(CodeInvalidURL, "not a LinkedIn profile URL"))
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := `{"error":{"code":"INVALID_URL","message":"not a LinkedIn profile URL"}}`
	if string(raw) != want {
		t.Errorf("JSON = %s, want %s", raw, want)
	}
}
