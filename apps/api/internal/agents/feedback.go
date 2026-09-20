package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/daily-market-brief/api/internal/db"
)

// PositionView is a human/JSON-friendly view of one open position.
type PositionView struct {
	Ticker     string  `json:"ticker"`
	Quantity   int64   `json:"quantity"`
	AvgCostUSD float64 `json:"avg_cost_usd"`
}

// TradeSummaryView is a human/JSON-friendly view of one past trade.
type TradeSummaryView struct {
	Ticker       string  `json:"ticker"`
	Side         string  `json:"side"`
	Quantity     int64   `json:"quantity"`
	PriceUSD     float64 `json:"price_usd"`
	ExecutedAt   string  `json:"executed_at"`
	Reasoning    string  `json:"reasoning"`
	CashAfterUSD float64 `json:"cash_after_usd"` // portfolio cash right after this trade executed
}

// PortfolioView is a full snapshot of one agent's simulated portfolio.
type PortfolioView struct {
	RiskProfile       string             `json:"risk_profile"`
	Label             string             `json:"label"`
	CashUSD           float64            `json:"cash_usd"`
	PositionsValueUSD float64            `json:"positions_value_usd"`
	TotalEquityUSD    float64            `json:"total_equity_usd"`
	MinSignalStrength int                `json:"min_signal_strength"`
	Positions         []PositionView     `json:"positions"`
	RecentTrades      []TradeSummaryView `json:"recent_trades"`
}

// BuildPortfolioView loads (creating the portfolio if it doesn't exist yet)
// a full snapshot for one risk profile: cash, open positions valued at the
// latest known price, and recent trade history (last 10 trades).
func BuildPortfolioView(ctx context.Context, database *db.DB, profile RiskProfile) (*PortfolioView, error) {
	return BuildPortfolioViewRange(ctx, database, profile, nil, nil)
}

// BuildPortfolioViewRange is like BuildPortfolioView but, when from/to are
// both given, lists every trade executed within [from, to] instead of just
// the most recent 10 — used to inspect a specific period (e.g. one month).
func BuildPortfolioViewRange(ctx context.Context, database *db.DB, profile RiskProfile, from, to *time.Time) (*PortfolioView, error) {
	pf, err := database.EnsurePortfolio(ctx, profile.Name, time.Now().UTC(), profile.MonthlyAllowanceCents, profile.MinSignalStrength)
	if err != nil {
		return nil, err
	}
	positions, err := database.PositionsByPortfolio(ctx, pf.ID)
	if err != nil {
		return nil, err
	}
	positionsValueCents, err := database.PositionsMarketValueCents(ctx, pf.ID)
	if err != nil {
		return nil, err
	}
	var trades []db.Trade
	if from != nil && to != nil {
		trades, err = database.TradesByPortfolioRange(ctx, pf.ID, *from, *to)
	} else {
		trades, err = database.TradesByPortfolio(ctx, pf.ID, 10)
	}
	if err != nil {
		return nil, err
	}

	upperBound := time.Now().UTC().AddDate(0, 0, 1)
	if to != nil {
		upperBound = *to
	}
	allTrades, err := database.TradesByPortfolioRange(ctx, pf.ID, time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC), upperBound)
	if err != nil {
		return nil, err
	}
	cashAfter := cashAfterByTradeID(allTrades, profile.MonthlyAllowanceCents)

	v := &PortfolioView{
		RiskProfile:       profile.Name,
		Label:             profile.Label,
		CashUSD:           float64(pf.CashCents) / 100,
		PositionsValueUSD: float64(positionsValueCents) / 100,
		TotalEquityUSD:    float64(pf.CashCents+positionsValueCents) / 100,
		MinSignalStrength: pf.MinSignalStrength,
		// Siempre listas vacías, nunca nil: un slice nil en Go se serializa
		// a `null` en JSON (no `[]`), lo que rompe `.length`/`.map` en el
		// cliente cuando un periodo no tuvo posiciones u operaciones.
		Positions:    []PositionView{},
		RecentTrades: []TradeSummaryView{},
	}
	for _, p := range positions {
		v.Positions = append(v.Positions, PositionView{
			Ticker: p.Ticker, Quantity: p.Quantity, AvgCostUSD: float64(p.AvgCostCents) / 100,
		})
	}
	for _, t := range trades {
		v.RecentTrades = append(v.RecentTrades, TradeSummaryView{
			Ticker: t.Ticker, Side: t.Side, Quantity: t.Quantity,
			PriceUSD: float64(t.PriceCents) / 100, ExecutedAt: t.ExecutedAt.Format("2006-01-02"), Reasoning: t.Reasoning,
			CashAfterUSD: float64(cashAfter[t.ID]) / 100,
		})
	}
	return v, nil
}

// cashAfterByTradeID replays a portfolio's trade history in chronological
// order — interleaving the same monthly funding used in cmd/rewind-agents —
// and returns, for each trade, the portfolio's cash balance immediately
// after that trade executed. trades must be sorted oldest-first (as
// TradesByPortfolioRange returns them). Used to show "cash restante" under
// each trade in the web client.
func cashAfterByTradeID(trades []db.Trade, monthlyAllowanceCents int64) map[uuid.UUID]int64 {
	result := make(map[uuid.UUID]int64, len(trades))
	cash := int64(0)
	monthCursor := time.Date(SimulationStart.Year(), SimulationStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	for _, t := range trades {
		for !monthCursor.After(t.ExecutedAt) {
			cash += monthlyAllowanceCents
			monthCursor = monthCursor.AddDate(0, 1, 0)
		}
		amount := t.Quantity * t.PriceCents
		if t.Side == "buy" {
			cash -= amount
		} else {
			cash += amount
		}
		result[t.ID] = cash
	}
	return result
}

// RenderForPrompt turns a portfolio view into plain text for the feedback LLM prompt.
func (v *PortfolioView) RenderForPrompt() string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "%s (%s): cash $%.2f, positions worth $%.2f, total equity $%.2f\n", v.Label, v.RiskProfile, v.CashUSD, v.PositionsValueUSD, v.TotalEquityUSD)
	if len(v.Positions) == 0 {
		b.WriteString("  No open positions.\n")
	}
	for _, p := range v.Positions {
		fmt.Fprintf(b, "  Holding: %d %s @ avg $%.2f\n", p.Quantity, p.Ticker, p.AvgCostUSD)
	}
	if len(v.RecentTrades) == 0 {
		b.WriteString("  No trades yet.\n")
	}
	for _, t := range v.RecentTrades {
		fmt.Fprintf(b, "  Trade %s: %s %d %s @ $%.2f — %s\n", t.ExecutedAt, strings.ToUpper(t.Side), t.Quantity, t.Ticker, t.PriceUSD, t.Reasoning)
	}
	return b.String()
}

// FeedbackSystemPrompt is the persona/instructions for the third agent: a
// coach the user interacts with directly, reviewing the other two agents.
const FeedbackSystemPrompt = `You are a calm, experienced investing coach reviewing two SIMULATED portfolios (fictitious money, no real trades) that trade automatically based on an AI news analyst's daily output:
- "risky" (Agresivo): acts on lower-confidence signals, bets a bigger share of cash per trade, holds more positions.
- "conservative" (Conservador): only acts on high-confidence signals, bets smaller, holds fewer positions.
Both get the same fictitious $5,000/month allowance. You are talking directly to the person who built this, to help them learn about investing and risk management. Be concrete and honest — point out good and bad decisions, compare the two strategies, and answer their question directly. Write in clear paragraphs, not bullet lists. Keep it focused, not exhaustive. Reply in the same language the user asks in.
CRITICAL: only reference trades, prices, and dates that literally appear in the data given to you below. If a portfolio had zero trades in the period asked about, say so plainly — never describe a trade, price, or date that isn't in the data, even if it sounds plausible. If you're unsure whether something happened in this specific period, say you don't have that information rather than guessing.`

// BuildFeedbackUserPrompt assembles the user-turn prompt for the feedback coach.
func BuildFeedbackUserPrompt(risky, conservative *PortfolioView, question string) string {
	b := &strings.Builder{}
	b.WriteString(risky.RenderForPrompt())
	b.WriteString("\n")
	b.WriteString(conservative.RenderForPrompt())
	b.WriteString("\n")
	if strings.TrimSpace(question) != "" {
		fmt.Fprintf(b, "The user asks: %s\n", question)
	} else {
		b.WriteString("Give a short overall review of both portfolios' performance and strategy so far.\n")
	}
	return b.String()
}

// BenchmarkView is a passive "buy the whole month's allowance into one ETF
// and never sell" comparison, computed on the fly from the same monthly
// funding schedule and real price history the agents use — so you can tell
// whether either agent actually beats doing nothing.
type BenchmarkView struct {
	Label             string  `json:"label"`
	Ticker            string  `json:"ticker"`
	CashUSD           float64 `json:"cash_usd"`
	SharesHeld        int64   `json:"shares_held"`
	PositionsValueUSD float64 `json:"positions_value_usd"`
	TotalEquityUSD    float64 `json:"total_equity_usd"`
	TotalFundedUSD    float64 `json:"total_funded_usd"`
	ReturnPct         float64 `json:"return_pct"`
}

// BuildBenchmarkView simulates putting the same monthly allowance
// (monthlyAllowanceCents) entirely into `ticker` on the first of each month
// since SimulationStart, holding it, and never selling — the simplest
// possible passive baseline. Used to answer "would I have been better off
// just buying and holding SPY?"
func BuildBenchmarkView(ctx context.Context, database *db.DB, ticker string, monthlyAllowanceCents int64, asOf time.Time) (*BenchmarkView, error) {
	var cashCents, fundedCents, shares int64
	monthCursor := time.Date(SimulationStart.Year(), SimulationStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	for !monthCursor.After(asOf) {
		cashCents += monthlyAllowanceCents
		fundedCents += monthlyAllowanceCents
		if priceCents, err := database.LatestPriceOnOrBefore(ctx, ticker, monthCursor); err == nil && priceCents > 0 {
			if qty := cashCents / priceCents; qty > 0 {
				cashCents -= qty * priceCents
				shares += qty
			}
		}
		monthCursor = monthCursor.AddDate(0, 1, 0)
	}
	var positionsValueCents int64
	if priceCents, err := database.LatestPriceOnOrBefore(ctx, ticker, asOf); err == nil && priceCents > 0 {
		positionsValueCents = shares * priceCents
	}
	equityCents := cashCents + positionsValueCents
	returnPct := 0.0
	if fundedCents > 0 {
		returnPct = (float64(equityCents) - float64(fundedCents)) / float64(fundedCents) * 100
	}
	return &BenchmarkView{
		Label:             fmt.Sprintf("Benchmark: comprar y mantener %s", ticker),
		Ticker:            ticker,
		CashUSD:           float64(cashCents) / 100,
		SharesHeld:        shares,
		PositionsValueUSD: float64(positionsValueCents) / 100,
		TotalEquityUSD:    float64(equityCents) / 100,
		TotalFundedUSD:    float64(fundedCents) / 100,
		ReturnPct:         returnPct,
	}, nil
}

// Recommendation is a deterministic, explainable threshold suggestion —
// computed from real trade counts in Go, never parsed from the coach LLM's
// free text (a small local model's structured output isn't reliable enough
// to act on directly, as this project's own hallucinated-trades bug showed).
// The user applies or dismisses it explicitly; nothing changes automatically.
type Recommendation struct {
	Profile        string `json:"profile"`
	Field          string `json:"field"`
	CurrentValue   int    `json:"current_value"`
	SuggestedValue int    `json:"suggested_value"`
	Reason         string `json:"reason"`
}

// ComputeRecommendation looks at a profile's trade count over the trailing
// 14 days (as of asOf) and suggests nudging min_signal_strength when the
// pattern looks lopsided: no activity at all (the threshold may be
// filtering out everything), or unusually heavy activity for a profile
// meant to be selective. Returns nil if nothing stands out. This is
// intentionally a simple, auditable heuristic, not a model of "good"
// trading — the human decides whether to apply it.
func ComputeRecommendation(ctx context.Context, database *db.DB, profile RiskProfile, pf *db.Portfolio, asOf time.Time) (*Recommendation, error) {
	from := asOf.AddDate(0, 0, -14)
	trades, err := database.TradesByPortfolioRange(ctx, pf.ID, from, asOf)
	if err != nil {
		return nil, err
	}
	current := pf.MinSignalStrength
	switch {
	case len(trades) == 0 && current > 3:
		return &Recommendation{
			Profile: profile.Name, Field: "min_signal_strength",
			CurrentValue: current, SuggestedValue: current - 1,
			Reason: fmt.Sprintf("0 operaciones en los ultimos 14 dias (umbral actual: %d/10) - podria estar filtrando senales validas.", current),
		}, nil
	case profile.Name == Conservative.Name && len(trades) >= 10 && current < 9:
		return &Recommendation{
			Profile: profile.Name, Field: "min_signal_strength",
			CurrentValue: current, SuggestedValue: current + 1,
			Reason: fmt.Sprintf("%d operaciones en 14 dias es mucho para un perfil conservador (umbral actual: %d/10) - subir el umbral lo haria mas selectivo.", len(trades), current),
		}, nil
	default:
		return nil, nil
	}
}
