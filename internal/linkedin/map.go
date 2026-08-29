package linkedin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Siddharth-Nama/LinkLens/internal/profile"
)

func MapProfile(raw []byte, publicIdentifier, inputURL, canonicalURL string) (profile.Profile, error) {
	var payload voyagerPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return profile.Profile{}, fmt.Errorf("parse linkedin response: %w", err)
	}

	included := indexIncluded(payload.Included)
	root, ok := findProfileEntity(included, publicIdentifier, payload.rootURNs())
	if !ok {
		return profile.Profile{}, ErrNotFound
	}

	out := profile.New()
	out.InputURL = inputURL
	out.CanonicalURL = canonicalURL
	out.PublicIdentifier = publicIdentifier
	out.FetchedAt = time.Now().UTC()

	out.FirstName = str(root["firstName"])
	out.LastName = str(root["lastName"])
	out.FullName = strings.TrimSpace(out.FirstName + " " + out.LastName)
	out.Headline = str(root["headline"])
	out.About = firstNonEmpty(str(root["summary"]), str(root["about"]))
	out.Industry = firstNonEmpty(str(root["industryName"]), str(root["industry"]))
	out.ProfileID = urnID(str(root["entityUrn"]))
	profileURN := str(root["entityUrn"])

	out.Location = mapLocation(root, included)
	out.ProfilePictures = mapProfilePictures(root)
	out.BackgroundImage = mapBackgroundImage(root)
	out.ConnectionsCount = intPtr(root["connections"])
	out.FollowersCount = intPtr(root["followersCount"])

	out.Experience = mapExperiences(included, profileURN)
	out.Education = mapEducation(included, profileURN)
	out.Skills = mapSkills(included, profileURN)
	out.Certifications = mapCertifications(included, profileURN)
	out.Languages = mapLanguages(included, profileURN)
	out.Volunteer = mapVolunteer(included, profileURN)
	out.Projects = mapProjects(included, profileURN)
	out.Honors = mapHonors(included, profileURN)
	out.Publications = mapPublications(included, profileURN)

	out.MissingSections = missingSections(out)
	out.Partial = len(out.MissingSections) > 0
	return out, nil
}

type voyagerPayload struct {
	Data struct {
		Elements     []json.RawMessage `json:"elements"`
		StarElements []string          `json:"*elements"`
	} `json:"data"`
	Included []map[string]any `json:"included"`
}

func (p voyagerPayload) rootURNs() []string {
	if len(p.Data.StarElements) > 0 {
		return p.Data.StarElements
	}
	urns := make([]string, 0, len(p.Data.Elements))
	for _, el := range p.Data.Elements {
		var obj map[string]any
		if err := json.Unmarshal(el, &obj); err != nil {
			continue
		}
		if urn := str(obj["entityUrn"]); urn != "" {
			urns = append(urns, urn)
		}
	}
	return urns
}

type includedIndex struct {
	byURN  map[string]map[string]any
	byType map[string][]map[string]any
	all    []map[string]any
}

func indexIncluded(items []map[string]any) includedIndex {
	idx := includedIndex{
		byURN:  make(map[string]map[string]any, len(items)),
		byType: make(map[string][]map[string]any),
		all:    items,
	}
	for _, item := range items {
		if urn := str(item["entityUrn"]); urn != "" {
			idx.byURN[urn] = item
		}
		if typ := str(item["$type"]); typ != "" {
			idx.byType[typ] = append(idx.byType[typ], item)
		}
	}
	return idx
}

func findProfileEntity(idx includedIndex, publicIdentifier string, rootURNs []string) (map[string]any, bool) {
	for _, urn := range rootURNs {
		if ent, ok := idx.byURN[urn]; ok {
			return ent, true
		}
	}
	for _, ent := range idx.all {
		if str(ent["publicIdentifier"]) == publicIdentifier {
			return ent, true
		}
	}
	for _, ent := range idx.byType["com.linkedin.voyager.dash.identity.profile.Profile"] {
		if str(ent["publicIdentifier"]) == publicIdentifier {
			return ent, true
		}
	}
	return nil, false
}

func belongsToProfile(item map[string]any, profileURN string) bool {
	if profileURN == "" {
		return true
	}
	refs := profileRefs(item)
	if len(refs) == 0 {
		return true
	}
	for _, ref := range refs {
		if ref == profileURN {
			return true
		}
	}
	return false
}

func profileRefs(item map[string]any) []string {
	var refs []string
	for _, key := range []string{"*profile", "profileUrn", "*profileUrn"} {
		if v := str(item[key]); v != "" {
			refs = append(refs, v)
		}
	}
	if p, ok := item["profile"].(map[string]any); ok {
		if v := str(p["entityUrn"]); v != "" {
			refs = append(refs, v)
		}
	}
	return refs
}

func mapLocation(root map[string]any, idx includedIndex) profile.Location {
	loc := profile.Location{}
	if geo := str(root["geoLocationName"]); geo != "" {
		loc.Raw = geo
		return loc
	}
	if geo := str(root["locationName"]); geo != "" {
		loc.Raw = geo
		return loc
	}
	if ref := str(root["*geo"]); ref != "" {
		if ent, ok := idx.byURN[ref]; ok {
			loc.Country = str(ent["country"])
			loc.City = firstNonEmpty(str(ent["city"]), str(ent["defaultLocalizedName"]))
			loc.Raw = firstNonEmpty(str(ent["defaultLocalizedName"]), loc.City, loc.Country)
		}
	}
	return loc
}

func mapProfilePictures(root map[string]any) []profile.Image {
	if pics := imagesFromNode(root["profilePicture"]); len(pics) > 0 {
		return pics
	}
	return imagesFromNode(root["displayPictureUrl"])
}

func mapBackgroundImage(root map[string]any) *profile.Image {
	pics := imagesFromNode(root["backgroundPicture"])
	if len(pics) == 0 {
		return nil
	}
	return &pics[0]
}

func imagesFromNode(node any) []profile.Image {
	switch v := node.(type) {
	case string:
		if v == "" {
			return nil
		}
		return []profile.Image{{URL: v}}
	case map[string]any:
		if ref, ok := v["displayImageReference"].(map[string]any); ok {
			return vectorImages(ref["vectorImage"])
		}
		if ref, ok := v["vectorImage"].(map[string]any); ok {
			return vectorImages(ref)
		}
		if url := str(v["rootUrl"]); url != "" {
			return []profile.Image{{URL: url}}
		}
	}
	return nil
}

func vectorImages(node any) []profile.Image {
	vm, ok := node.(map[string]any)
	if !ok {
		return nil
	}
	root := str(vm["rootUrl"])
	artifacts, _ := vm["artifacts"].([]any)
	out := make([]profile.Image, 0, len(artifacts))
	for _, a := range artifacts {
		am, ok := a.(map[string]any)
		if !ok {
			continue
		}
		seg := str(am["fileIdentifyingUrlPathSegment"])
		if seg == "" {
			continue
		}
		img := profile.Image{URL: root + seg}
		if w, ok := am["width"].(float64); ok {
			img.Width = int(w)
		}
		if h, ok := am["height"].(float64); ok {
			img.Height = int(h)
		}
		out = append(out, img)
	}
	return out
}

func mapExperiences(idx includedIndex, profileURN string) []profile.Experience {
	var out []profile.Experience
	for typ, items := range idx.byType {
		if !strings.HasSuffix(typ, ".Position") || strings.Contains(typ, "PositionGroup") {
			continue
		}
		for _, item := range items {
			if !belongsToProfile(item, profileURN) {
				continue
			}
			exp := profile.Experience{
				Title:       firstNonEmpty(str(item["title"]), str(item["multiLocaleTitle"])),
				CompanyName: companyName(item, idx),
				Location:    firstNonEmpty(str(item["geoLocationName"]), str(item["locationName"])),
				Description: firstNonEmpty(str(item["description"]), str(item["multiLocaleDescription"])),
				StartDate:   mapDateRange(item, "start"),
				EndDate:     mapDateRange(item, "end"),
			}
			exp.IsCurrent = boolVal(item["current"]) || exp.EndDate == nil
			if exp.CompanyName == "" && exp.Title == "" {
				continue
			}
			out = append(out, exp)
		}
	}
	return out
}

func companyName(item map[string]any, idx includedIndex) string {
	if name := str(item["companyName"]); name != "" {
		return name
	}
	if ref := str(item["*company"]); ref != "" {
		if ent, ok := idx.byURN[ref]; ok {
			return firstNonEmpty(str(ent["name"]), str(ent["universalName"]))
		}
	}
	if company, ok := item["company"].(map[string]any); ok {
		return firstNonEmpty(str(company["name"]), str(company["universalName"]))
	}
	return ""
}

func mapEducation(idx includedIndex, profileURN string) []profile.Education {
	var out []profile.Education
	for typ, items := range idx.byType {
		if !strings.HasSuffix(typ, ".Education") {
			continue
		}
		for _, item := range items {
			if !belongsToProfile(item, profileURN) {
				continue
			}
			ed := profile.Education{
				SchoolName:  schoolName(item, idx),
				Degree:      firstNonEmpty(str(item["degreeName"]), str(item["degree"])),
				Field:       firstNonEmpty(str(item["fieldOfStudy"]), str(item["fieldOfStudyName"])),
				Description: str(item["description"]),
				StartDate:   mapDateRange(item, "start"),
				EndDate:     mapDateRange(item, "end"),
			}
			if ed.SchoolName == "" {
				continue
			}
			out = append(out, ed)
		}
	}
	return out
}

func schoolName(item map[string]any, idx includedIndex) string {
	if name := str(item["schoolName"]); name != "" {
		return name
	}
	if ref := str(item["*school"]); ref != "" {
		if ent, ok := idx.byURN[ref]; ok {
			return str(ent["name"])
		}
	}
	if school, ok := item["school"].(map[string]any); ok {
		return str(school["name"])
	}
	return ""
}

func mapSkills(idx includedIndex, profileURN string) []profile.Skill {
	var out []profile.Skill
	for typ, items := range idx.byType {
		if !strings.HasSuffix(typ, ".Skill") {
			continue
		}
		for _, item := range items {
			if !belongsToProfile(item, profileURN) {
				continue
			}
			name := firstNonEmpty(str(item["name"]), str(item["skillName"]))
			if name == "" {
				continue
			}
			out = append(out, profile.Skill{Name: name, EndorsementCount: intPtr(item["endorsementCount"])})
		}
	}
	return out
}

func mapCertifications(idx includedIndex, profileURN string) []profile.Certification {
	var out []profile.Certification
	for typ, items := range idx.byType {
		if !strings.HasSuffix(typ, ".Certification") {
			continue
		}
		for _, item := range items {
			if !belongsToProfile(item, profileURN) {
				continue
			}
			name := str(item["name"])
			if name == "" {
				continue
			}
			out = append(out, profile.Certification{
				Name:      name,
				Authority: str(item["authority"]),
				URL:       str(item["url"]),
				IssuedOn:  mapDateRange(item, "start"),
				ExpiresOn: mapDateRange(item, "end"),
			})
		}
	}
	return out
}

func mapLanguages(idx includedIndex, profileURN string) []profile.Language {
	var out []profile.Language
	for typ, items := range idx.byType {
		if !strings.HasSuffix(typ, ".Language") {
			continue
		}
		for _, item := range items {
			if !belongsToProfile(item, profileURN) {
				continue
			}
			name := firstNonEmpty(str(item["name"]), str(item["language"]))
			if name == "" {
				continue
			}
			out = append(out, profile.Language{Name: name, Proficiency: str(item["proficiency"])})
		}
	}
	return out
}

func mapVolunteer(idx includedIndex, profileURN string) []profile.Volunteer {
	var out []profile.Volunteer
	for typ, items := range idx.byType {
		if !strings.HasSuffix(typ, ".VolunteerExperience") {
			continue
		}
		for _, item := range items {
			if !belongsToProfile(item, profileURN) {
				continue
			}
			role := firstNonEmpty(str(item["role"]), str(item["title"]))
			org := firstNonEmpty(str(item["companyName"]), str(item["organization"]))
			if role == "" && org == "" {
				continue
			}
			out = append(out, profile.Volunteer{
				Role:         role,
				Organization: org,
				Cause:        str(item["cause"]),
				Description:  str(item["description"]),
				StartDate:    mapDateRange(item, "start"),
				EndDate:      mapDateRange(item, "end"),
			})
		}
	}
	return out
}

func mapProjects(idx includedIndex, profileURN string) []profile.Project {
	var out []profile.Project
	for typ, items := range idx.byType {
		if !strings.HasSuffix(typ, ".Project") {
			continue
		}
		for _, item := range items {
			if !belongsToProfile(item, profileURN) {
				continue
			}
			name := str(item["title"])
			if name == "" {
				name = str(item["name"])
			}
			if name == "" {
				continue
			}
			out = append(out, profile.Project{
				Name:        name,
				Description: str(item["description"]),
				URL:         str(item["url"]),
				StartDate:   mapDateRange(item, "start"),
				EndDate:     mapDateRange(item, "end"),
			})
		}
	}
	return out
}

func mapHonors(idx includedIndex, profileURN string) []profile.Honor {
	var out []profile.Honor
	for typ, items := range idx.byType {
		if !strings.HasSuffix(typ, ".Honor") {
			continue
		}
		for _, item := range items {
			if !belongsToProfile(item, profileURN) {
				continue
			}
			title := firstNonEmpty(str(item["title"]), str(item["name"]))
			if title == "" {
				continue
			}
			out = append(out, profile.Honor{
				Title:       title,
				Issuer:      str(item["issuer"]),
				IssuedOn:    mapDateRange(item, "start"),
				Description: str(item["description"]),
			})
		}
	}
	return out
}

func mapPublications(idx includedIndex, profileURN string) []profile.Publication {
	var out []profile.Publication
	for typ, items := range idx.byType {
		if !strings.HasSuffix(typ, ".Publication") {
			continue
		}
		for _, item := range items {
			if !belongsToProfile(item, profileURN) {
				continue
			}
			title := str(item["name"])
			if title == "" {
				title = str(item["title"])
			}
			if title == "" {
				continue
			}
			out = append(out, profile.Publication{
				Title:       title,
				Publisher:   str(item["publisher"]),
				URL:         str(item["url"]),
				PublishedOn: mapDateRange(item, "start"),
				Description: str(item["description"]),
			})
		}
	}
	return out
}

func mapDateRange(item map[string]any, which string) *profile.Date {
	if period, ok := item["timePeriod"].(map[string]any); ok {
		if which == "start" {
			return mapLiDate(period["startDate"])
		}
		return mapLiDate(period["endDate"])
	}
	if dr, ok := item["dateRange"].(map[string]any); ok {
		if which == "start" {
			return mapLiDate(dr["start"])
		}
		return mapLiDate(dr["end"])
	}
	return nil
}

func mapLiDate(node any) *profile.Date {
	m, ok := node.(map[string]any)
	if !ok || m == nil {
		return nil
	}
	d := &profile.Date{}
	if month, ok := m["month"].(float64); ok {
		mi := int(month)
		d.Month = &mi
	}
	if year, ok := m["year"].(float64); ok {
		yi := int(year)
		d.Year = &yi
	}
	if d.Month == nil && d.Year == nil {
		return nil
	}
	return d
}

func missingSections(p profile.Profile) []string {
	checks := []struct {
		name  string
		empty bool
	}{
		{"about", p.About == ""},
		{"experience", len(p.Experience) == 0},
		{"education", len(p.Education) == 0},
		{"skills", len(p.Skills) == 0},
		{"certifications", len(p.Certifications) == 0},
		{"languages", len(p.Languages) == 0},
		{"profilePictures", len(p.ProfilePictures) == 0},
	}
	var missing []string
	for _, c := range checks {
		if c.empty {
			missing = append(missing, c.name)
		}
	}
	return missing
}

func str(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return fmt.Sprintf("%.0f", t)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func boolVal(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

func intPtr(v any) *int {
	switch t := v.(type) {
	case float64:
		i := int(t)
		return &i
	case int:
		return &t
	default:
		return nil
	}
}

func urnID(urn string) string {
	if urn == "" {
		return ""
	}
	parts := strings.Split(urn, ":")
	return parts[len(parts)-1]
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
