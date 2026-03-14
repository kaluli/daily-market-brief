package db

import (
	"context"
	"database/sql"
	"time"
)

func (db *DB) InsertDailySummary(ctx context.Context, s *DailySummary) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO daily_summaries (day, generated_at, top10, other90, text_path, text_sha256, items_analyzed)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (day) DO UPDATE SET
			generated_at = EXCLUDED.generated_at,
			top10 = EXCLUDED.top10,
			other90 = EXCLUDED.other90,
			text_path = EXCLUDED.text_path,
			text_sha256 = EXCLUDED.text_sha256,
			items_analyzed = EXCLUDED.items_analyzed
	`, s.Day.Format("2006-01-02"), s.GeneratedAt, s.Top10, s.Other90, s.TextPath, s.TextSHA256, s.ItemsAnalyzed)
	return err
}

func (db *DB) GetDailySummary(ctx context.Context, day time.Time) (*DailySummary, error) {
	var s DailySummary
	var dayVal interface{}
	err := db.QueryRowContext(ctx, `
		SELECT day::text, generated_at, top10, other90, text_path, text_sha256, items_analyzed
		FROM daily_summaries WHERE day = $1
	`, day.Format("2006-01-02")).Scan(&dayVal, &s.GeneratedAt, &s.Top10, &s.Other90, &s.TextPath, &s.TextSHA256, &s.ItemsAnalyzed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if str, ok := dayVal.(string); ok {
		s.Day, _ = time.Parse("2006-01-02", str)
	}
	return &s, nil
}

func (db *DB) GetDailySummariesRange(ctx context.Context, from, to time.Time) ([]DailySummary, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT day::text, generated_at, top10, other90, text_path, text_sha256, items_analyzed
		FROM daily_summaries WHERE day >= $1 AND day <= $2 ORDER BY day DESC
	`, from.Format("2006-01-02"), to.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DailySummary
	for rows.Next() {
		var s DailySummary
		var dayVal interface{}
		if err := rows.Scan(&dayVal, &s.GeneratedAt, &s.Top10, &s.Other90, &s.TextPath, &s.TextSHA256, &s.ItemsAnalyzed); err != nil {
			return nil, err
		}
		if str, ok := dayVal.(string); ok {
			s.Day, _ = time.Parse("2006-01-02", str)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (db *DB) ListDaysWithSummaries(ctx context.Context) ([]time.Time, error) {
	rows, err := db.QueryContext(ctx, `SELECT day::text FROM daily_summaries ORDER BY day DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []time.Time
	for rows.Next() {
		var dayStr string
		if err := rows.Scan(&dayStr); err != nil {
			return nil, err
		}
		t, _ := time.Parse("2006-01-02", dayStr)
		out = append(out, t)
	}
	return out, rows.Err()
}
