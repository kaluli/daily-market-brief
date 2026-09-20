package agents

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/daily-market-brief/api/internal/analyst"
	"github.com/daily-market-brief/api/internal/db"
)

// defaultMaxNewsPerDay caps how many of the day's news items (highest
// impact_score first — see internal/news/impact.go) get sent to the LLM
// analyst per RunDay call. Analyzing every item with a slow local model does
// not scale (a busy day can have 80+ items); the lower-impact ones are
// unlikely to move a trade decision anyway. Override with
// AGENTS_MAX_NEWS_PER_DAY.
const defaultMaxNewsPerDay = 25

func maxNewsPerDay() int {
	if v := os.Getenv("AGENTS_MAX_NEWS_PER_DAY"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultMaxNewsPerDay
}

// TradeView is a JSON-friendly view of one executed trade, returned by RunDay.
type TradeView struct {
	Ticker      string  `json:"ticker"`
	Side        string  `json:"side"`
	Quantity    int64   `json:"quantity"`
	PriceUSD    float64 `json:"price_usd"`
	Reasoning   string  `json:"reasoning"`
	NewsTitle   string  `json:"news_title,omitempty"`
	NewsURL     string  `json:"news_url,omitempty"`
}

// DayResult summarizes what one risk-profile agent did for a single day.
type DayResult struct {
	Profile              string      `json:"risk_profile"`
	NewsAnalyzed         int         `json:"news_analyzed"`
	Trades               []TradeView `json:"trades"`
	CashAfterUSD         float64     `json:"cash_after_usd"`
	Funded               bool        `json:"monthly_allowance_applied"`
	CircuitBreakerActive bool        `json:"circuit_breaker_active"`
	MinSignalStrength    int         `json:"min_signal_strength"`
}

// RunDay funds each portfolio's monthly allowance if due, analyzes the day's
// news with the given analyzer, and lets every risk profile decide and
// execute trades based on that analysis. Skips assets with no price data for
// the day (see docs/AGENTS.md — the price seed step must have run first).
func RunDay(ctx context.Context, database *db.DB, analyzer analyst.Analyzer, day time.Time) ([]DayResult, error) {
	items, err := database.NewsItemsByDay(ctx, day)
	if err != nil {
		return nil, fmt.Errorf("news for day: %w", err)
	}
	// NewsItemsByDay already orders by impact_score DESC, so this keeps the
	// highest-impact items and drops the long tail of low-relevance noise.
	if max := maxNewsPerDay(); len(items) > max {
		items = items[:max]
	}

	type analyzed struct {
		item   db.NewsItem
		result analyst.AnalysisResult
	}
	var analyses []analyzed
	total := len(items)
	for i, item := range items {
		log.Printf("agents run-day %s: analizando noticia %d/%d — %s", day.Format("2006-01-02"), i+1, total, item.Title)
		res, err := analyzer.Analyze(ctx, analyst.NewsInput{Title: item.Title, URL: item.URL, Source: item.Source})
		if err != nil {
			log.Printf("agents run-day %s: noticia %d/%d fallo el analisis: %v", day.Format("2006-01-02"), i+1, total, err)
			continue
		}
		analyses = append(analyses, analyzed{item: item, result: *res})
	}
	log.Printf("agents run-day %s: %d/%d noticias analizadas, decidiendo operaciones...", day.Format("2006-01-02"), len(analyses), total)

	results := make([]DayResult, 0, len(All))
	for _, profile := range All {
		pf, err := database.EnsurePortfolio(ctx, profile.Name, day, profile.MonthlyAllowanceCents, profile.MinSignalStrength)
		if err != nil {
			return nil, fmt.Errorf("ensure portfolio %s: %w", profile.Name, err)
		}
		funded, err := database.EnsureMonthlyFunding(ctx, pf.ID, day)
		if err != nil {
			return nil, fmt.Errorf("fund portfolio %s: %w", profile.Name, err)
		}
		if funded {
			pf, err = database.GetPortfolio(ctx, pf.ID)
			if err != nil {
				return nil, err
			}
		}
		positions, err := database.PositionsByPortfolio(ctx, pf.ID)
		if err != nil {
			return nil, err
		}
		heldQty := map[string]int64{}
		for _, p := range positions {
			heldQty[p.Ticker] = p.Quantity
		}
		openCount := len(positions)
		cashCents := pf.CashCents

		// Risk controls: concentration limit and a drawdown circuit breaker.
		// Both use a same-day snapshot (positions priced once, before the
		// loop) rather than tracking a historical equity peak, which would
		// need a new equity-snapshot table — a known simplification.
		positionsValueCents, err := database.PositionsMarketValueCents(ctx, pf.ID)
		if err != nil {
			return nil, err
		}
		totalEquityCents := cashCents + positionsValueCents
		monthsFunded := int64((day.Year()-SimulationStart.Year())*12+int(day.Month())-int(SimulationStart.Month())) + 1
		if monthsFunded < 1 {
			monthsFunded = 1
		}
		totalFundedCents := monthsFunded * profile.MonthlyAllowanceCents
		drawdownPct := 0.0
		if totalFundedCents > 0 && totalEquityCents < totalFundedCents {
			drawdownPct = 1 - float64(totalEquityCents)/float64(totalFundedCents)
		}
		circuitBreakerActive := drawdownPct > profile.MaxDrawdownPct
		if circuitBreakerActive {
			log.Printf("agents run-day %s: agente %s -> circuit breaker activo (drawdown %.0f%%), solo se permiten ventas", day.Format("2006-01-02"), profile.Name, drawdownPct*100)
		}

		var trades []TradeView
		for _, a := range analyses {
			if a.result.Relevance != analyst.RelevanceMarketMoving {
				continue
			}
			strength, _ := strconv.Atoi(strings.TrimSpace(a.result.SignalStrength))
			if strength < pf.MinSignalStrength {
				continue
			}
			for _, assetName := range a.result.AffectedAssets {
				ticker, ok := MapAssetName(assetName)
				if !ok {
					continue
				}
				biasRaw := strings.TrimSpace(a.result.DirectionalBias[assetName])
				bias := strings.ToLower(biasRaw)
				priceCents, err := database.LatestPriceOnOrBefore(ctx, ticker, day)
				if err != nil || priceCents <= 0 {
					continue // no price data yet for this ticker/day — see docs/AGENTS.md
				}

				reasoning := fmt.Sprintf("%s (strength %d/10) on %q: %s", biasRaw, strength, a.item.Title, a.result.WhyItMatters)

				switch bias {
				case "bullish":
					if circuitBreakerActive {
						continue // drawdown breaker: no new buys, only de-risking sells
					}
					if openCount >= profile.MaxOpenPositions && heldQty[ticker] == 0 {
						continue
					}
					budgetCents := int64(float64(cashCents) * profile.MaxPositionPct)
					qty := budgetCents / priceCents
					if qty < 1 {
						continue
					}
					costCents := qty * priceCents
					if costCents > cashCents {
						continue
					}
					// Concentration limit: cap this ticker's post-trade value
					// at MaxConcentrationPct of total equity.
					maxTickerValueCents := int64(float64(totalEquityCents) * profile.MaxConcentrationPct)
					currentTickerValueCents := heldQty[ticker] * priceCents
					if currentTickerValueCents >= maxTickerValueCents {
						continue // already at/over the concentration cap for this ticker
					}
					if allowed := maxTickerValueCents - currentTickerValueCents; costCents > allowed {
						qty = allowed / priceCents
						if qty < 1 {
							continue
						}
						costCents = qty * priceCents
					}
					t, err := database.ExecuteTrade(ctx, pf.ID, ticker, "buy", qty, priceCents, a.item.ID, reasoning, day)
					if err != nil {
						continue
					}
					cashCents -= costCents
					if heldQty[ticker] == 0 {
						openCount++
					}
					heldQty[ticker] += qty
					trades = append(trades, toView(t, priceCents, reasoning, a.item))

				case "bearish":
					qty := heldQty[ticker]
					if qty <= 0 {
						continue // no naked shorting — only close existing longs
					}
					t, err := database.ExecuteTrade(ctx, pf.ID, ticker, "sell", qty, priceCents, a.item.ID, reasoning, day)
					if err != nil {
						continue
					}
					cashCents += qty * priceCents
					heldQty[ticker] = 0
					openCount--
					trades = append(trades, toView(t, priceCents, reasoning, a.item))
				}
			}
		}

		log.Printf("agents run-day %s: agente %s -> %d operaciones, cash $%.2f", day.Format("2006-01-02"), profile.Name, len(trades), float64(cashCents)/100)
		results = append(results, DayResult{
			Profile:              profile.Name,
			NewsAnalyzed:         len(analyses),
			Trades:               trades,
			CashAfterUSD:         float64(cashCents) / 100,
			Funded:               funded,
			CircuitBreakerActive: circuitBreakerActive,
			MinSignalStrength:    pf.MinSignalStrength,
		})
	}
	return results, nil
}

func toView(t *db.Trade, priceCents int64, reasoning string, item db.NewsItem) TradeView {
	return TradeView{
		Ticker:    t.Ticker,
		Side:      t.Side,
		Quantity:  t.Quantity,
		PriceUSD:  float64(priceCents) / 100,
		Reasoning: reasoning,
		NewsTitle: item.Title,
		NewsURL:   item.URL,
	}
}
