package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

// dockerClient talks to the Docker daemon via the Unix socket.
// No third-party SDK required — the Docker Engine API is plain HTTP.
var dockerTransport = &http.Client{
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "unix", "/var/run/docker.sock")
		},
	},
	Timeout: 10 * time.Second,
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

// checkDocker inspects a container by name or ID and returns its health status.
func checkDocker(ctx context.Context, containerName string) (Status, string) {
	url := fmt.Sprintf("http://localhost/containers/%s/json", containerName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return StatusUnknown, err.Error()
	}

	resp, err := dockerTransport.Do(req)
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
	if err := json.NewDecoder(resp.Body).Decode(&inspect); err != nil {
		return StatusUnknown, fmt.Sprintf("parse error: %v", err)
	}

	if !inspect.State.Running {
		return StatusUnhealthy, fmt.Sprintf("container is %s", inspect.State.Status)
	}

	if h := inspect.State.Health; h != nil {
		switch h.Status {
		case "healthy":
			return StatusHealthy, ""
		case "unhealthy":
			return StatusUnhealthy, "healthcheck failing"
		case "starting":
			return StatusDegraded, "healthcheck starting"
		}
	}

	// Container is running but has no HEALTHCHECK directive.
	return StatusHealthy, "running (no healthcheck)"
}
