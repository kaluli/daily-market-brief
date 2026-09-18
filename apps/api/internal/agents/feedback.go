package agents

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	Ticker     string  `json:"ticker"`
	Side       string  `json:"side"`
	Quantity   int64   `json:"quantity"`
	PriceUSD   float64 `json:"price_usd"`
	ExecutedAt string  `json:"executed_at"`
	Reasoning  string  `json:"reasoning"`
}

// PortfolioView is a full snapshot of one agent's simulated portfolio.
type PortfolioView struct {
	RiskProfile       string             `json:"risk_profile"`
	Label             string             `json:"label"`
	CashUSD           float64            `json:"cash_usd"`
	PositionsValueUSD float64            `json:"positions_value_usd"`
	TotalEquityUSD    float64            `json:"total_equity_usd"`
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
	pf, err := database.EnsurePortfolio(ctx, profile.Name, time.Now().UTC(), profile.MonthlyAllowanceCents)
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

	v := &PortfolioView{
		RiskProfile:       profile.Name,
		Label:             profile.Label,
		CashUSD:           float64(pf.CashCents) / 100,
		PositionsValueUSD: float64(positionsValueCents) / 100,
		TotalEquityUSD:    float64(pf.CashCents+positionsValueCents) / 100,
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
		})
	}
	return v, nil
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
Both get the same fictitious $5,000/month allowance. You are talking directly to the person who built this, to help them learn about investing and risk management. Be concrete and honest — point out good and bad decisions, compare the two strategies, and answer their question directly. Write in clear paragraphs, not bullet lists. Keep it focused, not exhaustive. Reply in the same language the user asks in.`

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
