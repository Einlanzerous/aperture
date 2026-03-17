package postgres

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/aperture-dashboard/aperture/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/001_init.sql
var initSQL string

// PGStore implements store.Store backed by PostgreSQL.
type PGStore struct {
	pool *pgxpool.Pool
}

// Open creates a new Postgres-backed store, running migrations on startup.
func Open(ctx context.Context, dsn string) (*PGStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}

	if _, err := pool.Exec(ctx, initSQL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres migrate: %w", err)
	}

	return &PGStore{pool: pool}, nil
}

func (s *PGStore) RecordCheck(ctx context.Context, record store.CheckRecord) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO check_history (service, status, status_code, response_ms, message, checked_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		record.ServiceName, record.Status, record.StatusCode,
		record.ResponseTime, record.Message, record.CheckedAt.UTC(),
	)
	return err
}

func (s *PGStore) GetRecentHistory(ctx context.Context, serviceName string, since time.Duration) ([]store.CheckRecord, error) {
	cutoff := time.Now().UTC().Add(-since)
	rows, err := s.pool.Query(ctx,
		`SELECT service, status, status_code, response_ms, message, checked_at
		 FROM check_history
		 WHERE service = $1 AND checked_at >= $2
		 ORDER BY checked_at ASC`,
		serviceName, cutoff,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanCheckRecords(rows)
}

func (s *PGStore) GetDailySummaries(ctx context.Context, serviceName string, from, to time.Time) ([]store.DailySummary, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT service, date, total_checks, healthy_checks, unhealthy_checks, degraded_checks,
		        avg_response_ms, min_response_ms, max_response_ms, uptime_pct
		 FROM check_daily_summary
		 WHERE service = $1 AND date >= $2 AND date <= $3
		 ORDER BY date ASC`,
		serviceName, from.UTC(), to.UTC(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanDailySummaries(rows)
}

func (s *PGStore) Compact(ctx context.Context, rawMaxAge time.Duration) error {
	cutoff := time.Now().UTC().Add(-rawMaxAge)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("compact begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Upsert daily summaries from raw records older than the cutoff.
	_, err = tx.Exec(ctx, `
		INSERT INTO check_daily_summary
			(service, date, total_checks, healthy_checks, unhealthy_checks, degraded_checks,
			 avg_response_ms, min_response_ms, max_response_ms, uptime_pct)
		SELECT
			service,
			DATE(checked_at) AS date,
			COUNT(*)                                                          AS total_checks,
			SUM(CASE WHEN status = 'healthy'   THEN 1 ELSE 0 END)            AS healthy_checks,
			SUM(CASE WHEN status = 'unhealthy'  THEN 1 ELSE 0 END)           AS unhealthy_checks,
			SUM(CASE WHEN status = 'degraded'   THEN 1 ELSE 0 END)           AS degraded_checks,
			COALESCE(AVG(CASE WHEN response_ms > 0 THEN response_ms END), 0) AS avg_response_ms,
			COALESCE(MIN(CASE WHEN response_ms > 0 THEN response_ms END), 0) AS min_response_ms,
			COALESCE(MAX(CASE WHEN response_ms > 0 THEN response_ms END), 0) AS max_response_ms,
			ROUND(
				CAST(SUM(CASE WHEN status = 'healthy' THEN 1 ELSE 0 END) AS DOUBLE PRECISION) /
				NULLIF(COUNT(*), 0) * 100,
				2
			) AS uptime_pct
		FROM check_history
		WHERE checked_at < $1
		GROUP BY service, DATE(checked_at)
		ON CONFLICT(service, date) DO UPDATE SET
			total_checks     = EXCLUDED.total_checks,
			healthy_checks   = EXCLUDED.healthy_checks,
			unhealthy_checks = EXCLUDED.unhealthy_checks,
			degraded_checks  = EXCLUDED.degraded_checks,
			avg_response_ms  = EXCLUDED.avg_response_ms,
			min_response_ms  = EXCLUDED.min_response_ms,
			max_response_ms  = EXCLUDED.max_response_ms,
			uptime_pct       = EXCLUDED.uptime_pct
	`, cutoff)
	if err != nil {
		return fmt.Errorf("compact upsert: %w", err)
	}

	_, err = tx.Exec(ctx, `DELETE FROM check_history WHERE checked_at < $1`, cutoff)
	if err != nil {
		return fmt.Errorf("compact delete: %w", err)
	}

	return tx.Commit(ctx)
}

func (s *PGStore) Prune(ctx context.Context, summaryMaxAge time.Duration) error {
	cutoff := time.Now().UTC().Add(-summaryMaxAge)
	_, err := s.pool.Exec(ctx, `DELETE FROM check_daily_summary WHERE date < $1`, cutoff)
	return err
}

func (s *PGStore) Close() error {
	s.pool.Close()
	return nil
}

// scanCheckRecords reads pgx rows into CheckRecord slices.
func scanCheckRecords(rows pgx.Rows) ([]store.CheckRecord, error) {
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

// scanDailySummaries reads pgx rows into DailySummary slices.
func scanDailySummaries(rows pgx.Rows) ([]store.DailySummary, error) {
	var out []store.DailySummary
	for rows.Next() {
		var d store.DailySummary
		if err := rows.Scan(
			&d.ServiceName, &d.Date, &d.TotalChecks, &d.HealthyChecks,
			&d.UnhealthyChecks, &d.DegradedChecks, &d.AvgResponseMs,
			&d.MinResponseMs, &d.MaxResponseMs, &d.UptimePct,
		); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
