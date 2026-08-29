package linkedin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	DefaultDecorationID = "com.linkedin.voyager.dash.deco.identity.profile.FullProfileWithEntities-93"
	profilesPath        = "/voyager/api/identity/dash/profiles"
	maxResponseBytes    = 8 << 20
)

func (c *Client) FetchProfile(ctx context.Context, memberIdentity string) ([]byte, error) {
	memberIdentity = strings.TrimSpace(memberIdentity)
	if memberIdentity == "" {
		return nil, ErrInvalidIdentity
	}

	rawURL, err := c.profileURL(memberIdentity)
	if err != nil {
		return nil, err
	}

	req, err := c.NewRequest(ctx, http.MethodGet, rawURL)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: read body: %v", ErrUpstream, err)
	}
	if len(body) > maxResponseBytes {
		return nil, fmt.Errorf("%w: response exceeded %d bytes", ErrUpstream, maxResponseBytes)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusUnauthorized, http.StatusForbidden, 999:
		return nil, ErrSessionExpired
	case http.StatusNotFound:
		return nil, ErrNotFound
	case http.StatusTooManyRequests:
		return nil, ErrRateLimited
	default:
		return nil, fmt.Errorf("%w: HTTP %d", ErrUpstream, resp.StatusCode)
	}
}

func (c *Client) profileURL(memberIdentity string) (string, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUpstream, err)
	}
	u.Path = profilesPath
	q := u.Query()
	q.Set("q", "memberIdentity")
	q.Set("memberIdentity", memberIdentity)
	q.Set("decorationId", c.decorationID)
	u.RawQuery = q.Encode()
	return u.String(), nil
}
