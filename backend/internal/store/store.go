package store

import (
	"context"
	"time"
)

// DateFormat is the canonical date layout used for daily summary keys.
const DateFormat = "2006-01-02"

// CheckRecord is a single raw health-check result destined for persistent storage.
type CheckRecord struct {
	ServiceName  string    `json:"serviceName"`
	ServiceType  string    `json:"serviceType"`
	Status       string    `json:"status"`
	StatusCode   int       `json:"statusCode,omitempty"`
	ResponseTime int64     `json:"responseTime,omitempty"` // milliseconds
	Message      string    `json:"message,omitempty"`
	CheckedAt    time.Time `json:"checkedAt"`
}

// DailySummary is an aggregated view of one service's checks for a single day.
type DailySummary struct {
	ServiceName     string    `json:"serviceName"`
	Date            time.Time `json:"date"`
	TotalChecks     int       `json:"totalChecks"`
	HealthyChecks   int       `json:"healthyChecks"`
	UnhealthyChecks int       `json:"unhealthyChecks"`
	DegradedChecks  int       `json:"degradedChecks"`
	AvgResponseMs   int64     `json:"avgResponseMs"`
	MinResponseMs   int64     `json:"minResponseMs"`
	MaxResponseMs   int64     `json:"maxResponseMs"`
	UptimePct       float64   `json:"uptimePct"`
}

// Store is the persistence interface for check history.
// Implementations must be safe for concurrent use.
type Store interface {
	// RecordCheck persists a single raw check result.
	RecordCheck(ctx context.Context, record CheckRecord) error

	// GetRecentHistory returns raw check records for a service within the given duration.
	GetRecentHistory(ctx context.Context, serviceName string, since time.Duration) ([]CheckRecord, error)

	// GetDailySummaries returns aggregated daily summaries for a service in [from, to].
	GetDailySummaries(ctx context.Context, serviceName string, from, to time.Time) ([]DailySummary, error)

	// Compact rolls raw records older than rawMaxAge into daily summaries, then deletes them.
	Compact(ctx context.Context, rawMaxAge time.Duration) error

	// Prune deletes daily summaries older than summaryMaxAge.
	Prune(ctx context.Context, summaryMaxAge time.Duration) error

	// Close releases any underlying resources (DB connections, etc.).
	Close() error
}
