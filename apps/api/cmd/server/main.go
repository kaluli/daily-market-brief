package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/daily-market-brief/api/internal/analyst"
	"github.com/daily-market-brief/api/internal/api"
	"github.com/daily-market-brief/api/internal/config"
	"github.com/daily-market-brief/api/internal/db"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://marketbrief:marketbrief_secret@localhost:5432/marketbrief?sslmode=disable"
	}
	summariesPath := os.Getenv("SUMMARIES_PATH")
	if summariesPath == "" {
		summariesPath = "./summaries"
	}
	if abs, err := filepath.Abs(summariesPath); err == nil {
		summariesPath = abs
	}
	d, err := db.New(databaseURL)
	if err != nil {
		log.Fatalf("database: %v (is Postgres running? check DATABASE_URL)", err)
	}
	defer d.Close()

	var a analyst.Analyzer
	if os.Getenv("OPENAI_API_KEY") != "" {
		a = analyst.NewOpenAIAnalyzer("", "")
		log.Print("analyst: using OpenAI (OPENAI_API_KEY set)")
	} else {
		a = analyst.NewStub()
		log.Print("analyst: using stub (set OPENAI_API_KEY for LLM analysis)")
	}
	configDir := config.FindConfigDir()
	srv := api.New(d, summariesPath, a, configDir)
	if configDir != "" {
		log.Print("admin: news sources API enabled (config=" + configDir + ")")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3090"
	}
	log.Printf("listening on http://localhost:%s", port)
	if err := srv.Listen(":" + port); err != nil {
		log.Fatalf("server: %v (try another PORT=3091 if port is in use)", err)
	}
}
