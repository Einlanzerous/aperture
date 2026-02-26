package checker

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many redirects")
		}
		return nil
	},
}

// checkHTTP performs a GET request against url and returns the health status,
// HTTP status code, response time in milliseconds, and an optional error message.
func checkHTTP(ctx context.Context, url string) (Status, int, int64, string) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return StatusUnknown, 0, 0, fmt.Sprintf("invalid URL: %v", err)
	}
	req.Header.Set("User-Agent", "Aperture-HealthCheck/1.0")

	resp, err := httpClient.Do(req)
	elapsed := time.Since(start).Milliseconds()
	if err != nil {
		return StatusUnhealthy, 0, elapsed, err.Error()
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return StatusHealthy, resp.StatusCode, elapsed, ""
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		// Treat unresolved redirects as degraded.
		return StatusDegraded, resp.StatusCode, elapsed, "redirect"
	default:
		return StatusUnhealthy, resp.StatusCode, elapsed, fmt.Sprintf("HTTP %d", resp.StatusCode)
	}
}
