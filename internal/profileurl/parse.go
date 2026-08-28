package profileurl

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

const maxURLLength = 2048

var (
	ErrEmpty       = errors.New("profile URL is empty")
	ErrTooLong     = errors.New("profile URL exceeds maximum length")
	ErrInvalid     = errors.New("profile URL is not a valid URL")
	ErrNotLinkedIn = errors.New("URL is not a LinkedIn host")
	ErrNotProfile  = errors.New("URL is not a LinkedIn profile")
	ErrBadSlug     = errors.New("profile identifier is invalid")
)

type Parsed struct {
	Input            string
	CanonicalURL     string
	PublicIdentifier string
}

func Parse(raw string) (Parsed, error) {
	input := strings.TrimSpace(raw)
	if input == "" {
		return Parsed{}, ErrEmpty
	}
	if len(input) > maxURLLength {
		return Parsed{}, ErrTooLong
	}

	toParse := input
	if !strings.Contains(toParse, "://") {
		toParse = "https://" + toParse
	}

	u, err := url.Parse(toParse)
	if err != nil {
		return Parsed{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return Parsed{}, ErrInvalid
	}

	host := strings.ToLower(u.Hostname())
	if !isLinkedInHost(host) {
		return Parsed{}, ErrNotLinkedIn
	}

	parts := splitPath(u.Path)
	slug, err := profileSlug(parts)
	if err != nil {
		return Parsed{}, err
	}

	return Parsed{
		Input:            input,
		CanonicalURL:     "https://www.linkedin.com/in/" + slug + "/",
		PublicIdentifier: slug,
	}, nil
}

func isLinkedInHost(host string) bool {
	return host == "linkedin.com" || strings.HasSuffix(host, ".linkedin.com")
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	raw := strings.Split(path, "/")
	parts := make([]string, 0, len(raw))
	for _, p := range raw {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func profileSlug(parts []string) (string, error) {
	for i, part := range parts {
		if !strings.EqualFold(part, "in") {
			continue
		}
		if i+1 >= len(parts) {
			return "", ErrNotProfile
		}
		slug, err := url.PathUnescape(parts[i+1])
		if err != nil {
			return "", ErrBadSlug
		}
		slug = strings.TrimSpace(slug)
		if !validPublicIdentifier(slug) {
			return "", ErrBadSlug
		}
		return slug, nil
	}

	if len(parts) > 0 {
		switch strings.ToLower(parts[0]) {
		case "company", "school", "showcase", "jobs", "feed", "pub":
			return "", ErrNotProfile
		}
	}
	return "", ErrNotProfile
}

func validPublicIdentifier(slug string) bool {
	if n := len(slug); n < 3 || n > 100 {
		return false
	}
	for i, r := range slug {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
		case r == '-' && i > 0 && i < len(slug)-1:
		default:
			return false
		}
	}
	return true
}
