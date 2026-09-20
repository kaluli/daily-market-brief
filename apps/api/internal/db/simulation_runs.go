package db

import (
	"context"
	"database/sql"
	"time"
)

// SimulationRun records the outcome of processing one calendar day through
// cmd/simulate: how many news items were analyzed, how many trades each
// agent made (0 is a valid, real result — not a failure), whether it
// errored, and how long it took. One row per day (upserted), so "was this
// day processed?" is a direct query instead of inferring it from the
// presence or absence of trades in the trades table.
type SimulationRun struct {
	Day                time.Time
	NewsAnalyzed       int
	RiskyTrades        int
	ConservativeTrades int
	Status             string // "ok" | "error"
	ErrorMessage       string
	Model              string
	DurationMS         int64
	StartedAt          time.Time
	FinishedAt         time.Time
}

// UpsertSimulationRun records (or overwrites) one day's run outcome.
func (db *DB) UpsertSimulationRun(ctx context.Context, r SimulationRun) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO simulation_runs (day, news_analyzed, risky_trades, conservative_trades, status, error_message, model, duration_ms, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (day) DO UPDATE SET
			news_analyzed = EXCLUDED.news_analyzed,
			risky_trades = EXCLUDED.risky_trades,
			conservative_trades = EXCLUDED.conservative_trades,
			status = EXCLUDED.status,
			error_message = EXCLUDED.error_message,
			model = EXCLUDED.model,
			duration_ms = EXCLUDED.duration_ms,
			started_at = EXCLUDED.started_at,
			finished_at = EXCLUDED.finished_at
	`, r.Day.Format("2006-01-02"), r.NewsAnalyzed, r.RiskyTrades, r.ConservativeTrades, r.Status, r.ErrorMessage, r.Model, r.DurationMS, r.StartedAt, r.FinishedAt)
	return err
}

// HasSuccessfulRun reports whether a day already has a status='ok' row in
// simulation_runs — used by cmd/simulate to skip days it already processed
// unless -force is passed.
func (db *DB) HasSuccessfulRun(ctx context.Context, day time.Time) (bool, error) {
	var status string
	err := db.QueryRowContext(ctx, `SELECT status FROM simulation_runs WHERE day = $1`, day.Format("2006-01-02")).Scan(&status)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return status == "ok", nil
}
