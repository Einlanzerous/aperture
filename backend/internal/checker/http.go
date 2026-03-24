package checker

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultHTTPTimeout = 10 * time.Second
	maxHTTPRedirects   = 5
	userAgent          = "Aperture-HealthCheck/1.0"
)

// HTTPChecker checks a service via an HTTP GET request.
type HTTPChecker struct {
	url    string
	client *http.Client
}

func NewHTTPChecker(url string, skipVerify bool) *HTTPChecker {
	client := &http.Client{
		Timeout: defaultHTTPTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxHTTPRedirects {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
	if skipVerify {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // user-opted skip_verify
		}
	}
	return &HTTPChecker{url: url, client: client}
}

func (c *HTTPChecker) Check(ctx context.Context) (Status, int, int64, string) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return StatusUnknown, 0, 0, fmt.Sprintf("invalid URL: %v", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.client.Do(req)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return StatusUnhealthy, 0, elapsed, err.Error()
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return StatusHealthy, resp.StatusCode, elapsed, ""
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		return StatusDegraded, resp.StatusCode, elapsed, "redirect"
	default:
		return StatusUnhealthy, resp.StatusCode, elapsed, fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
}
