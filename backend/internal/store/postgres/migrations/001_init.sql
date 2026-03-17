CREATE TABLE IF NOT EXISTS check_history (
    id            BIGSERIAL PRIMARY KEY,
    service       TEXT      NOT NULL,
    status        TEXT      NOT NULL,
    status_code   INTEGER   NOT NULL DEFAULT 0,
    response_ms   BIGINT    NOT NULL DEFAULT 0,
    message       TEXT      NOT NULL DEFAULT '',
    checked_at    TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_check_history_lookup
    ON check_history (service, checked_at);

CREATE TABLE IF NOT EXISTS check_daily_summary (
    id               BIGSERIAL PRIMARY KEY,
    service          TEXT    NOT NULL,
    date             DATE    NOT NULL,
    total_checks     INTEGER NOT NULL,
    healthy_checks   INTEGER NOT NULL,
    unhealthy_checks INTEGER NOT NULL,
    degraded_checks  INTEGER NOT NULL,
    avg_response_ms  BIGINT  NOT NULL DEFAULT 0,
    min_response_ms  BIGINT  NOT NULL DEFAULT 0,
    max_response_ms  BIGINT  NOT NULL DEFAULT 0,
    uptime_pct       DOUBLE PRECISION NOT NULL DEFAULT 0,
    UNIQUE(service, date)
);
