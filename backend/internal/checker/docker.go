package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

const (
	defaultDockerSocket      = "/var/run/docker.sock"
	defaultDockerDialTimeout = 5 * time.Second
	defaultDockerTimeout     = 10 * time.Second
	maxDockerResponseBytes   = 10 << 20 // 10 MB

	dockerHealthHealthy   = "healthy"
	dockerHealthUnhealthy = "unhealthy"
	dockerHealthStarting  = "starting"
)

// DockerChecker checks a container's state via the Docker Engine API over a Unix socket.
type DockerChecker struct {
	container string
	client    *http.Client
}

func NewDockerChecker(container string, socketPath string) *DockerChecker {
	if socketPath == "" {
		socketPath = defaultDockerSocket
	}
	return &DockerChecker{
		container: container,
		client: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return (&net.Dialer{Timeout: defaultDockerDialTimeout}).DialContext(ctx, "unix", socketPath)
				},
			},
			Timeout: defaultDockerTimeout,
		},
	}
}

// Check returns (status, statusCode=0, responseTime=0, message) since Docker checks have no HTTP status code or response time.
func (c *DockerChecker) Check(ctx context.Context) (Status, int, int64, string) {
	status, msg := c.inspect(ctx)
	return status, 0, 0, msg
}

type dockerInspect struct {
	State struct {
		Status  string `json:"Status"`
		Running bool   `json:"Running"`
		Health  *struct {
			Status string `json:"Status"`
		} `json:"Health"`
	} `json:"State"`
}

func (c *DockerChecker) inspect(ctx context.Context) (Status, string) {
	url := fmt.Sprintf("http://localhost/containers/%s/json", c.container)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return StatusUnknown, err.Error()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return StatusUnhealthy, fmt.Sprintf("docker socket: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return StatusUnknown, "container not found"
	case http.StatusOK:
		// handled below
	default:
		return StatusUnhealthy, fmt.Sprintf("docker API: HTTP %d", resp.StatusCode)
	}

	var inspect dockerInspect
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxDockerResponseBytes)).Decode(&inspect); err != nil {
		return StatusUnknown, fmt.Sprintf("parse error: %v", err)
	}

	if !inspect.State.Running {
		return StatusUnhealthy, fmt.Sprintf("container is %s", inspect.State.Status)
	}

	if h := inspect.State.Health; h != nil {
		switch h.Status {
		case dockerHealthHealthy:
			return StatusHealthy, ""
		case dockerHealthUnhealthy:
			return StatusUnhealthy, "healthcheck failing"
		case dockerHealthStarting:
			return StatusDegraded, "healthcheck starting"
		}
	}

	return StatusHealthy, "running (no healthcheck)"
}
