package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/aperture-dashboard/aperture/internal/store"

	_ "modernc.org/sqlite"
)

//go:embed migrations/001_init.sql
var initSQL string

// SQLiteStore implements store.Store backed by a SQLite database.
type SQLiteStore struct {
	db *sql.DB
}

// Open creates a new SQLite-backed store, running migrations on startup.
func Open(ctx context.Context, dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}

	// Enable WAL mode for better concurrent read/write performance.
	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite WAL: %w", err)
	}

	if _, err := db.ExecContext(ctx, initSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite migrate: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) RecordCheck(ctx context.Context, record store.CheckRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO check_history (service, status, status_code, response_ms, message, checked_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		record.ServiceName, record.Status, record.StatusCode,
		record.ResponseTime, record.Message, record.CheckedAt.UTC(),
	)
	return err
}

func (s *SQLiteStore) GetRecentHistory(ctx context.Context, serviceName string, since time.Duration) ([]store.CheckRecord, error) {
	cutoff := time.Now().UTC().Add(-since)
	rows, err := s.db.QueryContext(ctx,
		`SELECT service, status, status_code, response_ms, message, checked_at
		 FROM check_history
		 WHERE service = ? AND checked_at >= ?
		 ORDER BY checked_at ASC`,
		serviceName, cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCheckRecords(rows)
}

func (s *SQLiteStore) GetDailySummaries(ctx context.Context, serviceName string, from, to time.Time) ([]store.DailySummary, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT service, date, total_checks, healthy_checks, unhealthy_checks, degraded_checks,
		        avg_response_ms, min_response_ms, max_response_ms, uptime_pct
		 FROM check_daily_summary
		 WHERE service = ? AND date >= ? AND date <= ?
		 ORDER BY date ASC`,
		serviceName, from.UTC().Format(store.DateFormat), to.UTC().Format(store.DateFormat),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDailySummaries(rows)
}

func (s *SQLiteStore) Compact(ctx context.Context, rawMaxAge time.Duration) error {
	cutoff := time.Now().UTC().Add(-rawMaxAge)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("compact begin tx: %w", err)
	}
	defer tx.Rollback()

	// Upsert daily summaries from raw records older than the cutoff.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO check_daily_summary
			(service, date, total_checks, healthy_checks, unhealthy_checks, degraded_checks,
			 avg_response_ms, min_response_ms, max_response_ms, uptime_pct)
		SELECT
			service,
			DATE(checked_at) AS date,
			COUNT(*)                                                       AS total_checks,
			SUM(CASE WHEN status = 'healthy'   THEN 1 ELSE 0 END)         AS healthy_checks,
			SUM(CASE WHEN status = 'unhealthy'  THEN 1 ELSE 0 END)        AS unhealthy_checks,
			SUM(CASE WHEN status = 'degraded'   THEN 1 ELSE 0 END)        AS degraded_checks,
			COALESCE(AVG(CASE WHEN response_ms > 0 THEN response_ms END), 0) AS avg_response_ms,
			COALESCE(MIN(CASE WHEN response_ms > 0 THEN response_ms END), 0) AS min_response_ms,
			COALESCE(MAX(CASE WHEN response_ms > 0 THEN response_ms END), 0) AS max_response_ms,
			ROUND(
				CAST(SUM(CASE WHEN status = 'healthy' THEN 1 ELSE 0 END) AS REAL) /
				NULLIF(COUNT(*), 0) * 100,
				2
			) AS uptime_pct
		FROM check_history
		WHERE checked_at < ?
		GROUP BY service, DATE(checked_at)
		ON CONFLICT(service, date) DO UPDATE SET
			total_checks     = excluded.total_checks,
			healthy_checks   = excluded.healthy_checks,
			unhealthy_checks = excluded.unhealthy_checks,
			degraded_checks  = excluded.degraded_checks,
			avg_response_ms  = excluded.avg_response_ms,
			min_response_ms  = excluded.min_response_ms,
			max_response_ms  = excluded.max_response_ms,
			uptime_pct       = excluded.uptime_pct
	`, cutoff)
	if err != nil {
		return fmt.Errorf("compact upsert: %w", err)
	}

	// Delete the compacted raw records.
	_, err = tx.ExecContext(ctx, `DELETE FROM check_history WHERE checked_at < ?`, cutoff)
	if err != nil {
		return fmt.Errorf("compact delete: %w", err)
	}

	return tx.Commit()
}

func (s *SQLiteStore) Prune(ctx context.Context, summaryMaxAge time.Duration) error {
	cutoff := time.Now().UTC().Add(-summaryMaxAge).Format(store.DateFormat)
	_, err := s.db.ExecContext(ctx, `DELETE FROM check_daily_summary WHERE date < ?`, cutoff)
	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// scanCheckRecords reads rows into CheckRecord slices.
func scanCheckRecords(rows *sql.Rows) ([]store.CheckRecord, error) {
	var out []store.CheckRecord
	for rows.Next() {
		var r store.CheckRecord
		if err := rows.Scan(&r.ServiceName, &r.Status, &r.StatusCode, &r.ResponseTime, &r.Message, &r.CheckedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// scanDailySummaries reads rows into DailySummary slices.
func scanDailySummaries(rows *sql.Rows) ([]store.DailySummary, error) {
	var out []store.DailySummary
	for rows.Next() {
		var d store.DailySummary
		var dateStr string
		if err := rows.Scan(
			&d.ServiceName, &dateStr, &d.TotalChecks, &d.HealthyChecks,
			&d.UnhealthyChecks, &d.DegradedChecks, &d.AvgResponseMs,
			&d.MinResponseMs, &d.MaxResponseMs, &d.UptimePct,
		); err != nil {
			return nil, err
		}
		d.Date, _ = time.Parse(store.DateFormat, dateStr)
		out = append(out, d)
	}
	return out, rows.Err()
}
