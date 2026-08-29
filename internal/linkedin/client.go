package linkedin

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36"
	acceptValue      = "application/vnd.linkedin.normalized+json+2.1"
	restliVersion    = "2.0.0"
	referer          = "https://www.linkedin.com/"
	defaultBaseURL   = "https://www.linkedin.com"
)

type Client struct {
	http         *http.Client
	baseURL      string
	decorationID string
	liAt         string
	jsessionID   string
	userAgent    string
}

func New(liAt, jsessionID string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 20 * time.Second
	}
	return &Client{
		http: &http.Client{
			Timeout: timeout,
			// LinkedIn redirects unauthenticated Voyager calls to login; do not follow.
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		baseURL:      defaultBaseURL,
		decorationID: DefaultDecorationID,
		liAt:         sanitizeSessionValue(liAt),
		jsessionID:   sanitizeSessionValue(jsessionID),
		userAgent:    defaultUserAgent,
	}
}

func (c *Client) NewRequest(ctx context.Context, method, rawURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("new linkedin request: %w", err)
	}
	c.applyHeaders(req)
	return req, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.http.Do(req)
}

func (c *Client) applyHeaders(req *http.Request) {
	req.Header.Set("Accept", acceptValue)
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-RestLi-Protocol-Version", restliVersion)
	req.Header.Set("Referer", referer)

	if token := c.jsessionID; token != "" {
		req.Header.Set("Csrf-Token", token)
	}
	if c.liAt != "" {
		req.AddCookie(&http.Cookie{Name: "li_at", Value: c.liAt})
	}
	if c.jsessionID != "" {
		req.AddCookie(&http.Cookie{Name: "JSESSIONID", Value: c.jsessionID})
	}
}
