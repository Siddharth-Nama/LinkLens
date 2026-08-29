package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Siddharth-Nama/LinkLens/internal/cache"
	"github.com/Siddharth-Nama/LinkLens/internal/linkedin"
	"github.com/Siddharth-Nama/LinkLens/internal/profile"
	"github.com/Siddharth-Nama/LinkLens/internal/profileurl"
)

var ErrLinkedInNotConfigured = errors.New("linkedin session is not configured")

type Fetcher interface {
	FetchProfile(ctx context.Context, memberIdentity string) ([]byte, error)
}

type ProfileService struct {
	fetcher Fetcher
	cache   *cache.TTL[profile.Profile]
}

func NewProfile(fetcher Fetcher, cacheTTL time.Duration) *ProfileService {
	svc := &ProfileService{fetcher: fetcher}
	if cacheTTL > 0 {
		svc.cache = cache.NewTTL[profile.Profile](cacheTTL)
	}
	return svc
}

func (s *ProfileService) GetByURL(ctx context.Context, rawURL string) (profile.Profile, error) {
	parsed, err := profileurl.Parse(rawURL)
	if err != nil {
		return profile.Profile{}, err
	}
	if s.fetcher == nil {
		return profile.Profile{}, ErrLinkedInNotConfigured
	}

	if s.cache != nil {
		if cached, ok := s.cache.Get(parsed.PublicIdentifier); ok {
			cached.InputURL = parsed.Input
			cached.CanonicalURL = parsed.CanonicalURL
			return cached, nil
		}
	}

	body, err := s.fetcher.FetchProfile(ctx, parsed.PublicIdentifier)
	if err != nil {
		return profile.Profile{}, err
	}

	out, err := linkedin.MapProfile(body, parsed.PublicIdentifier, parsed.Input, parsed.CanonicalURL)
	if err != nil {
		return profile.Profile{}, err
	}
	if s.cache != nil {
		s.cache.Set(parsed.PublicIdentifier, out)
	}
	return out, nil
}

func InvalidURLMessage(err error) string {
	switch {
	case errors.Is(err, profileurl.ErrEmpty):
		return "profile url query param is required"
	case errors.Is(err, profileurl.ErrTooLong):
		return "profile url is too long"
	case errors.Is(err, profileurl.ErrNotLinkedIn):
		return "url must be a linkedin.com profile link"
	case errors.Is(err, profileurl.ErrNotProfile):
		return "url must point to a linkedin /in/ profile"
	case errors.Is(err, profileurl.ErrBadSlug):
		return "linkedin profile identifier is invalid"
	default:
		return fmt.Sprintf("invalid profile url: %v", err)
	}
}