package main

import (
	"log"
	"os"

	"github.com/daily-market-brief/api/internal/db"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://marketbrief:marketbrief_secret@localhost:5432/marketbrief?sslmode=disable"
	}
	d, err := db.New(databaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer d.Close()
	if err := d.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrations ok")
}
