package semaphore

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aperture-dashboard/aperture/internal/config"
)

// ActionState is the public view of an action and its current task state.
type ActionState struct {
	Name        string     `json:"name"`
	ProjectID   int        `json:"projectId"`
	TemplateID  int        `json:"templateId"`
	Category    string     `json:"category,omitempty"`
	Icon        string     `json:"icon,omitempty"`
	Size        string     `json:"size,omitempty"`
	TaskID      int        `json:"taskId,omitempty"`
	TaskStatus  TaskStatus `json:"taskStatus"`
	TriggeredAt *time.Time `json:"triggeredAt,omitempty"`
}

// Manager tracks action state and delegates to the Semaphore client.
type Manager struct {
	client *Client
	mu     sync.RWMutex
	states map[string]*ActionState
}

// NewManager creates a Manager from config and a Semaphore client.
func NewManager(client *Client, actions []config.ActionConfig) *Manager {
	states := make(map[string]*ActionState, len(actions))
	for _, a := range actions {
		states[a.Name] = &ActionState{
			Name:       a.Name,
			ProjectID:  a.ProjectID,
			TemplateID: a.TemplateID,
			Category:   a.Category,
			Icon:       a.Icon,
			Size:       a.Size,
			TaskStatus: "idle",
		}
	}
	return &Manager{client: client, states: states}
}

// ListActions returns a snapshot of all action states.
func (m *Manager) ListActions() []ActionState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]ActionState, 0, len(m.states))
	for _, s := range m.states {
		out = append(out, *s)
	}
	return out
}

// Trigger fires a Semaphore task for the named action.
func (m *Manager) Trigger(ctx context.Context, name string) (*ActionState, error) {
	m.mu.Lock()
	state, ok := m.states[name]
	if !ok {
		m.mu.Unlock()
		return nil, fmt.Errorf("unknown action %q", name)
	}

	// Don't allow re-triggering while a task is in-flight.
	if state.TaskID != 0 && !state.TaskStatus.IsTerminal() && state.TaskStatus != "idle" {
		m.mu.Unlock()
		return nil, fmt.Errorf("action %q already has a task in progress", name)
	}
	m.mu.Unlock()

	task, err := m.client.TriggerTask(ctx, state.ProjectID, state.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("trigger task: %w", err)
	}

	now := time.Now()

	m.mu.Lock()
	state.TaskID = task.ID
	state.TaskStatus = task.Status
	state.TriggeredAt = &now
	snapshot := *state
	m.mu.Unlock()

	return &snapshot, nil
}

// GetStatus refreshes and returns the status of the named action's current task.
func (m *Manager) GetStatus(ctx context.Context, name string) (*ActionState, error) {
	m.mu.RLock()
	state, ok := m.states[name]
	if !ok {
		m.mu.RUnlock()
		return nil, fmt.Errorf("unknown action %q", name)
	}

	// No task to poll — return current state.
	if state.TaskID == 0 || state.TaskStatus.IsTerminal() || state.TaskStatus == "idle" {
		snapshot := *state
		m.mu.RUnlock()
		return &snapshot, nil
	}

	projectID := state.ProjectID
	taskID := state.TaskID
	m.mu.RUnlock()

	task, err := m.client.GetTaskStatus(ctx, projectID, taskID)
	if err != nil {
		return nil, fmt.Errorf("get task status: %w", err)
	}

	m.mu.Lock()
	state.TaskStatus = task.Status
	snapshot := *state
	m.mu.Unlock()

	return &snapshot, nil
}
