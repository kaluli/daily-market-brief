package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

func (db *DB) InsertNewsItem(ctx context.Context, item *NewsItem) error {
	raw := item.Raw
	if raw == nil {
		raw = []byte("{}")
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO news_items (id, published_at, day, source, title, url, tickers, raw, impact_score, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (url) DO NOTHING
	`, item.ID, item.PublishedAt, item.Day, item.Source, item.Title, item.URL, pq.Array(item.Tickers), raw, item.ImpactScore, item.CreatedAt)
	return err
}

func (db *DB) NewsItemsByDay(ctx context.Context, day time.Time) ([]NewsItem, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, published_at, day, source, title, url, tickers, raw, impact_score, created_at
		FROM news_items WHERE day = $1 ORDER BY impact_score DESC
	`, day.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNewsItems(rows)
}

func (db *DB) CountNewsItemsByDay(ctx context.Context, day time.Time) (int, error) {
	var n int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM news_items WHERE day = $1`, day.Format("2006-01-02")).Scan(&n)
	return n, err
}

func (db *DB) URLExists(ctx context.Context, url string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM news_items WHERE url = $1)`, url).Scan(&exists)
	return exists, err
}

func scanNewsItems(rows *sql.Rows) ([]NewsItem, error) {
	var out []NewsItem
	for rows.Next() {
		var item NewsItem
		var tickers pq.StringArray
		if err := rows.Scan(&item.ID, &item.PublishedAt, &item.Day, &item.Source, &item.Title, &item.URL, &tickers, &item.Raw, &item.ImpactScore, &item.CreatedAt); err != nil {
			return nil, err
		}
		item.Tickers = []string(tickers)
		out = append(out, item)
	}
	return out, rows.Err()
}
