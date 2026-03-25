package semaphore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const maxResponseBody = 1 << 20 // 1 MB

// TaskStatus represents the lifecycle state of a Semaphore task.
type TaskStatus string

const (
	StatusWaiting  TaskStatus = "waiting"
	StatusStarting TaskStatus = "starting"
	StatusRunning  TaskStatus = "running"
	StatusSuccess  TaskStatus = "success"
	StatusError    TaskStatus = "error"
	StatusStopped  TaskStatus = "stopped"
)

// IsTerminal returns true if the task has reached a final state.
func (s TaskStatus) IsTerminal() bool {
	return s == StatusSuccess || s == StatusError || s == StatusStopped
}

// TaskResponse is the subset of Semaphore's task JSON we care about.
type TaskResponse struct {
	ID     int        `json:"id"`
	Status TaskStatus `json:"status"`
}

// Client talks to the Semaphore REST API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a Semaphore API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// do executes an HTTP request against the Semaphore API and decodes the response.
func (c *Client) do(ctx context.Context, method, url string, body io.Reader) (*TaskResponse, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("semaphore returned %d: %s", resp.StatusCode, string(respBody))
	}

	var task TaskResponse
	if err := json.Unmarshal(respBody, &task); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &task, nil
}

// TriggerTask starts a task template and returns the created task.
func (c *Client) TriggerTask(ctx context.Context, projectID, templateID int) (*TaskResponse, error) {
	body, err := json.Marshal(map[string]int{"template_id": templateID})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	url := fmt.Sprintf("%s/api/project/%d/tasks", c.baseURL, projectID)
	return c.do(ctx, http.MethodPost, url, bytes.NewReader(body))
}

// GetTaskStatus fetches the current status of a task.
func (c *Client) GetTaskStatus(ctx context.Context, projectID, taskID int) (*TaskResponse, error) {
	url := fmt.Sprintf("%s/api/project/%d/tasks/%d", c.baseURL, projectID, taskID)
	return c.do(ctx, http.MethodGet, url, nil)
}
