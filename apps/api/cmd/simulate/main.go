package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/daily-market-brief/api/internal/agents"
	"github.com/daily-market-brief/api/internal/analyst"
	"github.com/daily-market-brief/api/internal/db"
)

// simulate walks a date range day by day, letting the risky/conservative
// agents analyze that day's news and trade (agents.RunDay — same logic the
// API's /api/agents/run-day uses), crediting each portfolio's monthly
// fictitious allowance automatically. Every 7 days it asks the feedback
// agent for a short weekly review and saves it (also printed live).
//
// Safe to run in slices (e.g. -from=2026-03-01 -to=2026-03-31, then later
// -from=2026-04-01 -to=2026-04-30): portfolio state lives in the database,
// not in this process, so a later run continues from where the last one left
// off. Each day's outcome is recorded in simulation_runs, so by default a
// day that already has a successful run is skipped (idempotent re-runs of
// the same range are safe); pass -force to reprocess anyway (still not
// duplicate-safe within that forced re-run — it will execute new trades on
// top of the old ones).
func main() {
	fromStr := flag.String("from", "2026-03-01", "primer dia YYYY-MM-DD (incluido)")
	toStr := flag.String("to", "", "ultimo dia YYYY-MM-DD (incluido; default: hoy UTC)")
	force := flag.Bool("force", false, "reprocesar dias que ya tienen una corrida exitosa registrada")
	flag.Parse()

	from, err := time.Parse("2006-01-02", *fromStr)
	if err != nil {
		log.Fatalf("invalid -from: %v", err)
	}
	to := time.Now().UTC()
	to = time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	if *toStr != "" {
		to, err = time.Parse("2006-01-02", *toStr)
		if err != nil {
			log.Fatalf("invalid -to: %v", err)
		}
	}
	if to.Before(from) {
		log.Fatalf("-to (%s) es anterior a -from (%s)", to.Format("2006-01-02"), from.Format("2006-01-02"))
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

	analyzer, provider := analyst.NewAnalyzerFromEnv()
	completeFeedback, feedbackProvider := analyst.NewChatCompleterFromEnv()
	log.Printf("simulate: %s -> %s (analyst: %s, feedback: %s)", from.Format("2006-01-02"), to.Format("2006-01-02"), provider, feedbackProvider)

	ctx := context.Background()
	weekStart := from
	totalDays := 0

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = provider
	}

	for day := from; !day.After(to); day = day.AddDate(0, 0, 1) {
		totalDays++
		log.Printf("=== dia %s (%d/%d) ===", day.Format("2006-01-02"), totalDays, int(to.Sub(from).Hours()/24)+1)

		if !*force {
			if ok, err := d.HasSuccessfulRun(ctx, day); err == nil && ok {
				log.Printf("dia %s: ya tiene una corrida exitosa registrada, se salta (usa -force para reprocesar)", day.Format("2006-01-02"))
				daysInWeek := int(day.Sub(weekStart).Hours()/24) + 1
				if daysInWeek >= 7 || day.Equal(to) {
					weekStart = day.AddDate(0, 0, 1)
				}
				continue
			}
		}

		started := time.Now()
		results, err := agents.RunDay(ctx, d, analyzer, day)
		run := db.SimulationRun{
			Day: day, Model: model, StartedAt: started, FinishedAt: time.Now(),
			DurationMS: time.Since(started).Milliseconds(),
		}
		if err != nil {
			log.Printf("dia %s: error: %v (sigo con el proximo dia)", day.Format("2006-01-02"), err)
			run.Status = "error"
			run.ErrorMessage = err.Error()
		} else {
			run.Status = "ok"
			for _, r := range results {
				log.Printf("dia %s: agente %s -> %d operaciones, cash $%.2f", day.Format("2006-01-02"), r.Profile, len(r.Trades), r.CashAfterUSD)
				run.NewsAnalyzed = r.NewsAnalyzed
				switch r.Profile {
				case agents.Risky.Name:
					run.RiskyTrades = len(r.Trades)
				case agents.Conservative.Name:
					run.ConservativeTrades = len(r.Trades)
				}
			}
		}
		if err := d.UpsertSimulationRun(ctx, run); err != nil {
			log.Printf("dia %s: no se pudo registrar la corrida: %v", day.Format("2006-01-02"), err)
		}

		daysInWeek := int(day.Sub(weekStart).Hours()/24) + 1
		if daysInWeek >= 7 || day.Equal(to) {
			weeklyFeedback(ctx, d, completeFeedback, feedbackProvider, weekStart, day)
			weekStart = day.AddDate(0, 0, 1)
		}
	}

	log.Printf("=== simulacion terminada: %d dias procesados (%s a %s) ===", totalDays, from.Format("2006-01-02"), to.Format("2006-01-02"))
	printFinalSummary(ctx, d)
}

func weeklyFeedback(ctx context.Context, d *db.DB, complete analyst.ChatCompleteFunc, provider string, weekStart, weekEnd time.Time) {
	log.Printf("--- feedback semanal: %s a %s ---", weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"))
	// Scoped to [weekStart, weekEnd] — NOT the unscoped BuildPortfolioView,
	// which pulls the last 10 trades ever and previously caused the coach to
	// misattribute real trades from other weeks to this one.
	risky, err1 := agents.BuildPortfolioViewRange(ctx, d, agents.Risky, &weekStart, &weekEnd)
	conservative, err2 := agents.BuildPortfolioViewRange(ctx, d, agents.Conservative, &weekStart, &weekEnd)
	if err1 != nil || err2 != nil {
		log.Printf("feedback semanal: no se pudo armar el resumen de carteras (%v / %v)", err1, err2)
		return
	}

	question := fmt.Sprintf(
		"Revision semanal: semana del %s al %s. Dame un resumen breve de como le fue a cada cartera esta semana y que deberia tener en cuenta la que viene.",
		weekStart.Format("2006-01-02"), weekEnd.Format("2006-01-02"),
	)
	answer, err := complete(ctx, agents.FeedbackSystemPrompt, agents.BuildFeedbackUserPrompt(risky, conservative, question))
	if err != nil {
		log.Printf("feedback semanal: error: %v", err)
		return
	}
	log.Printf("feedback semanal:\n%s", answer)

	var recs []agents.Recommendation
	for _, profile := range agents.All {
		pf, err := d.GetPortfolioByRiskProfile(ctx, profile.Name)
		if err != nil {
			continue
		}
		rec, err := agents.ComputeRecommendation(ctx, d, profile, pf, weekEnd)
		if err == nil && rec != nil {
			recs = append(recs, *rec)
			log.Printf("feedback semanal: recomendacion para %s: %s", profile.Name, rec.Reason)
		}
	}

	if _, err := d.InsertFeedbackWithRecommendations(ctx, question, answer, provider, weekEnd, recs); err != nil {
		log.Printf("feedback semanal: no se pudo guardar: %v", err)
	}
}

func printFinalSummary(ctx context.Context, d *db.DB) {
	for _, profile := range agents.All {
		v, err := agents.BuildPortfolioView(ctx, d, profile)
		if err != nil {
			log.Printf("%s: error armando resumen final: %v", profile.Name, err)
			continue
		}
		log.Printf("%s: cash $%.2f, posiciones $%.2f, equity total $%.2f", v.Label, v.CashUSD, v.PositionsValueUSD, v.TotalEquityUSD)
	}
}
