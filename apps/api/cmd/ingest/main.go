package main

import (
	"context"
	"log"
	"os"

	"github.com/daily-market-brief/api/internal/config"
	"github.com/daily-market-brief/api/internal/db"
	"github.com/daily-market-brief/api/internal/news"
)

func main() {
	configDir := config.FindConfigDir()
	if configDir == "" {
		configDir = os.Getenv("CONFIG_DIR")
	}
	if configDir == "" {
		configDir = "config"
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
	ctx := context.Background()
	n, err := news.Harvest(ctx, d, configDir)
	if err != nil {
		log.Fatalf("harvest: %v", err)
	}
	log.Printf("ingest ok: %d new items", n)
}
