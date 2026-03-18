package checker

import (
	"time"

	"github.com/aperture-dashboard/aperture/internal/config"
)

// Status represents the health state of a monitored service.
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
)

const defaultCheckTimeout = 15 * time.Second

// ServiceStatus holds the latest check result for a single service.
type ServiceStatus struct {
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	URL           string    `json:"url,omitempty"`
	Container     string    `json:"container,omitempty"`
	Status        Status    `json:"status"`
	StatusCode    int       `json:"statusCode,omitempty"`
	ResponseTime  int64     `json:"responseTime,omitempty"` // milliseconds
	Message       string    `json:"message,omitempty"`
	CheckedAt     time.Time `json:"checkedAt"`
	Icon          string    `json:"icon,omitempty"`
	Category      string    `json:"category,omitempty"`
	Href          string    `json:"href,omitempty"`
	Size          string    `json:"size,omitempty"`
	DetailDefault bool      `json:"detailDefault,omitempty"`
}

// newServiceStatus creates a ServiceStatus pre-populated from config fields.
func newServiceStatus(svc config.ServiceConfig) *ServiceStatus {
	return &ServiceStatus{
		Name:          svc.Name,
		Type:          string(svc.Type),
		URL:           svc.URL,
		Container:     svc.Container,
		Icon:          svc.Icon,
		Category:      svc.Category,
		Href:          svc.Href,
		Size:          svc.Size,
		DetailDefault: svc.DetailDefault,
		CheckedAt:     time.Now(),
	}
}
