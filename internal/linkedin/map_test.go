package linkedin

import (
	"os"
	"testing"
)

func TestMapProfileRichFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/profile_rich.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	got, err := MapProfile(raw, "jane-doe", "https://www.linkedin.com/in/jane-doe/", "https://www.linkedin.com/in/jane-doe/")
	if err != nil {
		t.Fatalf("MapProfile: %v", err)
	}

	if got.FullName != "Jane Doe" {
		t.Errorf("FullName = %q", got.FullName)
	}
	if got.Headline != "Software Engineer at Acme" {
		t.Errorf("Headline = %q", got.Headline)
	}
	if got.About != "Building reliable systems." {
		t.Errorf("About = %q", got.About)
	}
	if got.Location.Raw != "San Francisco Bay Area" {
		t.Errorf("Location.Raw = %q", got.Location.Raw)
	}
	if got.ConnectionsCount == nil || *got.ConnectionsCount != 500 {
		t.Errorf("ConnectionsCount = %v", got.ConnectionsCount)
	}
	if len(got.ProfilePictures) != 1 || got.ProfilePictures[0].URL == "" {
		t.Errorf("ProfilePictures = %+v", got.ProfilePictures)
	}
	if got.BackgroundImage == nil || got.BackgroundImage.URL == "" {
		t.Errorf("BackgroundImage = %+v", got.BackgroundImage)
	}
	if len(got.Experience) != 1 || got.Experience[0].Title != "Senior Engineer" {
		t.Errorf("Experience = %+v", got.Experience)
	}
	if !got.Experience[0].IsCurrent {
		t.Error("expected current experience")
	}
	if len(got.Education) != 1 || got.Education[0].SchoolName != "State University" {
		t.Errorf("Education = %+v", got.Education)
	}
	if len(got.Skills) != 2 || got.Skills[0].Name != "Go" {
		t.Errorf("Skills = %+v", got.Skills)
	}
	if len(got.Certifications) != 1 {
		t.Errorf("Certifications = %+v", got.Certifications)
	}
	if len(got.Languages) != 1 {
		t.Errorf("Languages = %+v", got.Languages)
	}
	if got.Partial {
		t.Errorf("Partial = true, want false for rich fixture")
	}
}

func TestMapProfileMinimalFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/profile_ok.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	got, err := MapProfile(raw, "jane-doe", "in", "canon")
	if err != nil {
		t.Fatalf("MapProfile: %v", err)
	}
	if got.FirstName != "Jane" || got.LastName != "Doe" {
		t.Errorf("name = %s %s", got.FirstName, got.LastName)
	}
	if !got.Partial {
		t.Error("expected partial profile for minimal fixture")
	}
	for _, section := range []string{"about", "experience", "education", "skills"} {
		found := false
		for _, m := range got.MissingSections {
			if m == section {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missingSections should include %q, got %v", section, got.MissingSections)
		}
	}
}

func TestMapProfileNotFound(t *testing.T) {
	raw := []byte(`{"data":{"*elements":[]},"included":[]}`)
	_, err := MapProfile(raw, "missing", "", "")
	if err != ErrNotFound {
		t.Fatalf("error = %v, want %v", err, ErrNotFound)
	}
}

func TestMapProfileInvalidJSON(t *testing.T) {
	_, err := MapProfile([]byte("{"), "x", "", "")
	if err == nil {
		t.Fatal("expected parse error")
	}
}
