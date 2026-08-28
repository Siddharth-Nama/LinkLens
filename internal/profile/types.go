package profile

import "time"

type Profile struct {
	InputURL         string          `json:"inputUrl"`
	CanonicalURL     string          `json:"canonicalUrl"`
	PublicIdentifier string          `json:"publicIdentifier"`
	ProfileID        string          `json:"profileId,omitempty"`
	FirstName        string          `json:"firstName"`
	LastName         string          `json:"lastName"`
	FullName         string          `json:"fullName"`
	Headline         string          `json:"headline"`
	About            string          `json:"about"`
	Location         Location        `json:"location"`
	Industry         string          `json:"industry"`
	ProfilePictures  []Image         `json:"profilePictures"`
	BackgroundImage  *Image          `json:"backgroundImage"`
	ConnectionsCount *int            `json:"connectionsCount"`
	FollowersCount   *int            `json:"followersCount"`
	Experience       []Experience    `json:"experience"`
	Education        []Education     `json:"education"`
	Skills           []Skill         `json:"skills"`
	Certifications   []Certification `json:"certifications"`
	Languages        []Language      `json:"languages"`
	Volunteer        []Volunteer     `json:"volunteer"`
	Projects         []Project       `json:"projects"`
	Honors           []Honor         `json:"honors"`
	Publications     []Publication   `json:"publications"`
	FetchedAt        time.Time       `json:"fetchedAt"`
	Partial          bool            `json:"partial"`
	MissingSections  []string        `json:"missingSections"`
}

type Location struct {
	Country string `json:"country"`
	City    string `json:"city"`
	Raw     string `json:"raw"`
}

type Image struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type Date struct {
	Month *int `json:"month,omitempty"`
	Year  *int `json:"year,omitempty"`
}

type Experience struct {
	Title       string `json:"title"`
	CompanyName string `json:"companyName"`
	CompanyURL  string `json:"companyUrl,omitempty"`
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
	StartDate   *Date  `json:"startDate"`
	EndDate     *Date  `json:"endDate"`
	IsCurrent   bool   `json:"isCurrent"`
}

type Education struct {
	SchoolName  string `json:"schoolName"`
	Degree      string `json:"degree,omitempty"`
	Field       string `json:"field,omitempty"`
	StartDate   *Date  `json:"startDate"`
	EndDate     *Date  `json:"endDate"`
	Description string `json:"description,omitempty"`
}

type Skill struct {
	Name             string `json:"name"`
	EndorsementCount *int   `json:"endorsementCount"`
}

type Certification struct {
	Name      string `json:"name"`
	Authority string `json:"authority,omitempty"`
	URL       string `json:"url,omitempty"`
	IssuedOn  *Date  `json:"issuedOn"`
	ExpiresOn *Date  `json:"expiresOn"`
}

type Language struct {
	Name        string `json:"name"`
	Proficiency string `json:"proficiency,omitempty"`
}

type Volunteer struct {
	Role         string `json:"role"`
	Organization string `json:"organization"`
	Cause        string `json:"cause,omitempty"`
	Description  string `json:"description,omitempty"`
	StartDate    *Date  `json:"startDate"`
	EndDate      *Date  `json:"endDate"`
}

type Project struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	StartDate   *Date  `json:"startDate"`
	EndDate     *Date  `json:"endDate"`
}

type Honor struct {
	Title       string `json:"title"`
	Issuer      string `json:"issuer,omitempty"`
	IssuedOn    *Date  `json:"issuedOn"`
	Description string `json:"description,omitempty"`
}

type Publication struct {
	Title       string `json:"title"`
	Publisher   string `json:"publisher,omitempty"`
	URL         string `json:"url,omitempty"`
	PublishedOn *Date  `json:"publishedOn"`
	Description string `json:"description,omitempty"`
}

func New() Profile {
	return Profile{
		ProfilePictures: []Image{},
		Experience:      []Experience{},
		Education:       []Education{},
		Skills:          []Skill{},
		Certifications:  []Certification{},
		Languages:       []Language{},
		Volunteer:       []Volunteer{},
		Projects:        []Project{},
		Honors:          []Honor{},
		Publications:    []Publication{},
		MissingSections: []string{},
	}
}
