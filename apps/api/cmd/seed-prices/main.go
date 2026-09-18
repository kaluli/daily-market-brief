package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/daily-market-brief/api/internal/db"
)

// seed-prices loads historical daily close prices from CSV files (one per
// ticker, Stooq format: Date,Open,High,Low,Close,Volume) into the
// asset_prices table the investor agents use to price trades.
// See docs/AGENTS.md and scripts/fetch-prices.sh.
func main() {
	dirFlag := flag.String("dir", "", "directory with one CSV per ticker (default: ../../data/prices, or $DATA_PRICES_DIR)")
	flag.Parse()

	priceDir := *dirFlag
	if priceDir == "" {
		priceDir = os.Getenv("DATA_PRICES_DIR")
	}
	if priceDir == "" {
		priceDir = "../../data/prices"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://marketbrief:marketbrief_secret@localhost:5432/marketbrief?sslmode=disable"
	}
	d, err := db.New(databaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer d.Close()

	entries, err := os.ReadDir(priceDir)
	if err != nil {
		log.Fatalf("read dir %s: %v (run scripts/fetch-prices.sh first — see docs/AGENTS.md)", priceDir, err)
	}

	ctx := context.Background()
	totalRows, filesLoaded := 0, 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".csv") {
			continue
		}
		ticker := strings.ToUpper(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
		path := filepath.Join(priceDir, entry.Name())
		n, err := loadCSV(ctx, d, ticker, path)
		if err != nil {
			log.Printf("skip %s: %v", entry.Name(), err)
			continue
		}
		log.Printf("%s: %d rows", ticker, n)
		totalRows += n
		filesLoaded++
	}
	log.Printf("done: %d tickers, %d price rows loaded from %s", filesLoaded, totalRows, priceDir)
}

func loadCSV(ctx context.Context, d *db.DB, ticker, path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		return 0, err
	}
	if len(rows) < 2 {
		return 0, fmt.Errorf("no data rows")
	}

	header := rows[0]
	dateIdx, closeIdx := -1, -1
	for i, col := range header {
		switch strings.ToLower(strings.TrimSpace(col)) {
		case "date":
			dateIdx = i
		case "close":
			closeIdx = i
		}
	}
	if dateIdx == -1 || closeIdx == -1 {
		return 0, fmt.Errorf("missing Date/Close columns (header: %v)", header)
	}

	n := 0
	for _, row := range rows[1:] {
		if len(row) <= dateIdx || len(row) <= closeIdx {
			continue
		}
		day, err := time.Parse("2006-01-02", strings.TrimSpace(row[dateIdx]))
		if err != nil {
			continue
		}
		closeVal, err := strconv.ParseFloat(strings.TrimSpace(row[closeIdx]), 64)
		if err != nil || closeVal <= 0 {
			continue
		}
		closeCents := int64(closeVal*100 + 0.5)
		if err := d.UpsertAssetPrice(ctx, ticker, day, closeCents); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
