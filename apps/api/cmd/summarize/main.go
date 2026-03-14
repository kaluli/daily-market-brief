package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/daily-market-brief/api/internal/db"
	"github.com/daily-market-brief/api/internal/summarizer"
)

func main() {
	dayStr := flag.String("day", "", "day YYYY-MM-DD (default: today UTC)")
	flag.Parse()
	var day time.Time
	if *dayStr != "" {
		var err error
		day, err = time.Parse("2006-01-02", *dayStr)
		if err != nil {
			log.Fatalf("invalid day: %v", err)
		}
	} else {
		day = time.Now().UTC()
		day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	}
	summariesPath := os.Getenv("SUMMARIES_PATH")
	if summariesPath == "" {
		summariesPath = "./summaries"
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
	items, err := d.NewsItemsByDay(ctx, day)
	if err != nil {
		log.Fatalf("news by day: %v", err)
	}
	sm := summarizer.NewExtractive()
	res, err := sm.Summarize(ctx, day, items)
	if err != nil {
		log.Fatalf("summarize: %v", err)
	}
	textPath, sha256Hex, err := summarizer.WriteResult(summariesPath, day, res)
	if err != nil {
		log.Fatalf("write result: %v", err)
	}
	ds := summarizer.ToDailySummary(day, res, textPath, sha256Hex)
	if err := d.InsertDailySummary(ctx, ds); err != nil {
		log.Fatalf("insert summary: %v", err)
	}
	log.Printf("summarize ok: %s (%d lines, %d items)", day.Format("2006-01-02"), len(res.Lines), res.ItemsAnalyzed)
}
