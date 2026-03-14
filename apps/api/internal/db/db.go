package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")

type DB struct {
	*sql.DB
}

func New(databaseURL string) (*DB, error) {
	conn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	return &DB{conn}, nil
}

type NewsItem struct {
	ID          uuid.UUID
	PublishedAt time.Time
	Day         time.Time
	Source      string
	Title       string
	URL         string
	Tickers     []string
	Raw         json.RawMessage
	ImpactScore float64
	CreatedAt   time.Time
}

type DailySummary struct {
	Day           time.Time
	GeneratedAt   time.Time
	Top10         json.RawMessage
	Other90       json.RawMessage
	TextPath      string
	TextSHA256    string
	ItemsAnalyzed int
}
