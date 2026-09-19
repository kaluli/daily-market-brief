package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/daily-market-brief/api/internal/agents"
	"github.com/daily-market-brief/api/internal/db"
)

// rewind-agents borra los trades de ambos agentes ejecutados a partir de
// -from (inclusive) y recalcula cash/posiciones reproduciendo el historial
// de trades anterior a esa fecha mas el fondeo mensual correspondiente.
// Es la forma segura de "pisar" un tramo de la simulacion para volver a
// correrlo con cmd/simulate, en vez de re-ejecutar cmd/simulate sobre dias
// ya procesados (que no tiene proteccion de idempotencia y duplicaria los
// trades).
//
// Por defecto corre en modo -dry-run=true: solo muestra que haria, no toca
// la base. Revisa la salida y despues corre con -dry-run=false para aplicar.
//
func main() {
	fromStr := flag.String("from", "", "primer dia a borrar y re-simular, YYYY-MM-DD (requerido)")
	dryRun := flag.Bool("dry-run", true, "true: solo muestra que haria. false: aplica los cambios")
	flag.Parse()

	if *fromStr == "" {
		log.Fatal("uso: go run cmd/rewind-agents/main.go -from=YYYY-MM-DD [-dry-run=false]")
	}
	from, err := time.Parse("2006-01-02", *fromStr)
	if err != nil {
		log.Fatalf("invalid -from: %v", err)
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

	if *dryRun {
		fmt.Println("=== DRY RUN: no se va a modificar nada. Corré con -dry-run=false para aplicar. ===")
	}

	for _, profile := range agents.All {
		pf, err := d.GetPortfolioByRiskProfile(ctx, profile.Name)
		if err != nil {
			log.Fatalf("%s: %v", profile.Name, err)
		}

		allTrades, err := d.TradesByPortfolioRange(ctx, pf.ID, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), time.Now().UTC())
		if err != nil {
			log.Fatalf("%s: %v", profile.Name, err)
		}

		cashCents := int64(0)
		monthCursor := time.Date(agents.SimulationStart.Year(), agents.SimulationStart.Month(), 1, 0, 0, 0, 0, time.UTC)
		cutoffMonth := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
		for !monthCursor.After(cutoffMonth) {
			cashCents += pf.MonthlyAllowanceCents
			monthCursor = monthCursor.AddDate(0, 1, 0)
		}

		type posState struct {
			qty int64
			avg int64
		}
		positions := map[string]*posState{}

		var kept, removed, warnings int
		for _, t := range allTrades {
			if !t.ExecutedAt.Before(from) {
				removed++
				continue
			}
			kept++
			amount := t.Quantity * t.PriceCents
			p, ok := positions[t.Ticker]
			if !ok {
				p = &posState{}
				positions[t.Ticker] = p
			}
			if t.Side == "buy" {
				cashCents -= amount
				newQty := p.qty + t.Quantity
				if newQty > 0 {
					p.avg = (p.qty*p.avg + t.Quantity*t.PriceCents) / newQty
				}
				p.qty = newQty
			} else {
				cashCents += amount
				if t.Quantity > p.qty {
					// El motor real solo vende lo que efectivamente tiene en
					// ese momento (nunca en corto), así que esto significa
					// que el orden reconstruido de trades del mismo día no
					// coincide con el orden real — el resultado puede no ser
					// exacto para este ticker.
					warnings++
					fmt.Printf("  ADVERTENCIA: venta de %d %s el %s pero solo se tenian %d reconstruidos — orden ambiguo, revisar a mano\n",
						t.Quantity, t.Ticker, t.ExecutedAt.Format("2006-01-02"), p.qty)
				}
				p.qty -= t.Quantity
				if p.qty < 0 {
					p.qty = 0
				}
			}
		}

		fmt.Printf("\n=== %s (%s) ===\n", profile.Label, profile.Name)
		fmt.Printf("  trades: %d se mantienen (antes de %s), %d se borran (desde %s en adelante)\n", kept, from.Format("2006-01-02"), removed, from.Format("2006-01-02"))
		fmt.Printf("  cash: $%.2f -> $%.2f\n", float64(pf.CashCents)/100, float64(cashCents)/100)
		if warnings > 0 {
			fmt.Printf("  *** %d advertencia(s) de orden ambiguo — revisar antes de aplicar con -dry-run=false ***\n", warnings)
		}
		if cashCents < 0 {
			fmt.Printf("  *** CASH NEGATIVO — esto no deberia pasar nunca. NO apliques (-dry-run=false) hasta entender por que. ***\n")
		}
		hasPositions := false
		for ticker, p := range positions {
			if p.qty > 0 {
				hasPositions = true
				fmt.Printf("  posicion resultante: %d %s @ avg $%.2f\n", p.qty, ticker, float64(p.avg)/100)
			}
		}
		if !hasPositions {
			fmt.Println("  sin posiciones abiertas resultantes")
		}

		if *dryRun {
			continue
		}
		if warnings > 0 || cashCents < 0 {
			log.Fatalf("%s: no se aplica — hay advertencias o cash negativo, revisa el dry-run primero", profile.Name)
		}

		var rows []db.Position
		for ticker, p := range positions {
			if p.qty > 0 {
				rows = append(rows, db.Position{Ticker: ticker, Quantity: p.qty, AvgCostCents: p.avg})
			}
		}
		if err := d.RewindPortfolio(ctx, pf.ID, from, cashCents, rows); err != nil {
			log.Fatalf("%s: rewind fallo: %v", profile.Name, err)
		}
		fmt.Println("  OK: aplicado.")
	}
}
