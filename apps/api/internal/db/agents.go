package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Portfolio is one investor agent's simulated brokerage account (fictitious
// money). One row per risk profile ("risky", "conservative").
type Portfolio struct {
	ID                    uuid.UUID
	RiskProfile           string
	StartedAt             time.Time
	InitialCashCents      int64
	CashCents             int64
	MonthlyAllowanceCents int64
	LastFundedMonth       *time.Time
	CreatedAt             time.Time
}

// Position is one open holding (ticker + quantity + average cost) in a portfolio.
type Position struct {
	ID           uuid.UUID
	PortfolioID  uuid.UUID
	Ticker       string
	Quantity     int64
	AvgCostCents int64
	UpdatedAt    time.Time
}

// Trade is one executed buy/sell for a portfolio, optionally tied to the
// news item that triggered it.
type Trade struct {
	ID          uuid.UUID
	PortfolioID uuid.UUID
	Ticker      string
	Side        string // "buy" | "sell"
	Quantity    int64
	PriceCents  int64
	NewsItemID  uuid.NullUUID
	Reasoning   string
	ExecutedAt  time.Time
}

// FeedbackEntry is one Q&A exchange with the feedback/coach agent.
type FeedbackEntry struct {
	ID       uuid.UUID `json:"id"`
	AskedAt  time.Time `json:"asked_at"`
	Question string    `json:"question"`
	Answer   string    `json:"answer"`
	Provider string    `json:"provider"`
}

// EnsurePortfolio returns the portfolio for a risk profile, creating it
// (with zero cash, to be funded by EnsureMonthlyFunding) if it doesn't exist yet.
func (db *DB) EnsurePortfolio(ctx context.Context, riskProfile string, startedAt time.Time, monthlyAllowanceCents int64) (*Portfolio, error) {
	if p, err := db.GetPortfolioByRiskProfile(ctx, riskProfile); err == nil {
		return p, nil
	} else if err != ErrNotFound {
		return nil, err
	}

	id := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO portfolios (id, agent_type, risk_profile, started_at, initial_cash_cents, cash_cents, monthly_allowance_cents)
		VALUES ($1, $2, $2, $3, 0, 0, $4)
		ON CONFLICT (risk_profile) DO NOTHING
	`, id, riskProfile, startedAt.Format("2006-01-02"), monthlyAllowanceCents)
	if err != nil {
		return nil, err
	}
	return db.GetPortfolioByRiskProfile(ctx, riskProfile)
}

// GetPortfolioByRiskProfile looks up a portfolio by its risk profile ("risky"/"conservative").
func (db *DB) GetPortfolioByRiskProfile(ctx context.Context, riskProfile string) (*Portfolio, error) {
	return db.scanPortfolio(db.QueryRowContext(ctx, `
		SELECT id, risk_profile, started_at, initial_cash_cents, cash_cents, monthly_allowance_cents, last_funded_month, created_at
		FROM portfolios WHERE risk_profile = $1
	`, riskProfile))
}

// GetPortfolio looks up a portfolio by id.
func (db *DB) GetPortfolio(ctx context.Context, id uuid.UUID) (*Portfolio, error) {
	return db.scanPortfolio(db.QueryRowContext(ctx, `
		SELECT id, risk_profile, started_at, initial_cash_cents, cash_cents, monthly_allowance_cents, last_funded_month, created_at
		FROM portfolios WHERE id = $1
	`, id))
}

func (db *DB) scanPortfolio(row *sql.Row) (*Portfolio, error) {
	var p Portfolio
	var lastFunded sql.NullTime
	err := row.Scan(&p.ID, &p.RiskProfile, &p.StartedAt, &p.InitialCashCents, &p.CashCents, &p.MonthlyAllowanceCents, &lastFunded, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if lastFunded.Valid {
		p.LastFundedMonth = &lastFunded.Time
	}
	return &p, nil
}

// ListPortfolios returns all agent portfolios, ordered by risk profile name.
func (db *DB) ListPortfolios(ctx context.Context) ([]Portfolio, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, risk_profile, started_at, initial_cash_cents, cash_cents, monthly_allowance_cents, last_funded_month, created_at
		FROM portfolios ORDER BY risk_profile
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Portfolio
	for rows.Next() {
		var p Portfolio
		var lastFunded sql.NullTime
		if err := rows.Scan(&p.ID, &p.RiskProfile, &p.StartedAt, &p.InitialCashCents, &p.CashCents, &p.MonthlyAllowanceCents, &lastFunded, &p.CreatedAt); err != nil {
			return nil, err
		}
		if lastFunded.Valid {
			p.LastFundedMonth = &lastFunded.Time
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// EnsureMonthlyFunding credits the portfolio's fictitious monthly allowance
// for the calendar month containing `day`, exactly once per month. Returns
// whether funding was applied (false if that month was already funded).
func (db *DB) EnsureMonthlyFunding(ctx context.Context, portfolioID uuid.UUID, day time.Time) (bool, error) {
	monthStart := time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, time.UTC)
	res, err := db.ExecContext(ctx, `
		UPDATE portfolios
		SET cash_cents = cash_cents + monthly_allowance_cents,
		    last_funded_month = $2
		WHERE id = $1 AND (last_funded_month IS NULL OR last_funded_month < $2)
	`, portfolioID, monthStart.Format("2006-01-02"))
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// PositionsByPortfolio returns the portfolio's currently open (quantity > 0) positions.
func (db *DB) PositionsByPortfolio(ctx context.Context, portfolioID uuid.UUID) ([]Position, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, portfolio_id, ticker, quantity, avg_cost_cents, updated_at
		FROM positions WHERE portfolio_id = $1 AND quantity > 0 ORDER BY ticker
	`, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Position
	for rows.Next() {
		var p Position
		if err := rows.Scan(&p.ID, &p.PortfolioID, &p.Ticker, &p.Quantity, &p.AvgCostCents, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// TradesByPortfolio returns the portfolio's most recent trades, newest first.
func (db *DB) TradesByPortfolio(ctx context.Context, portfolioID uuid.UUID, limit int) ([]Trade, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id, portfolio_id, ticker, side, quantity, price_cents, news_item_id, reasoning, executed_at
		FROM trades WHERE portfolio_id = $1 ORDER BY executed_at DESC LIMIT $2
	`, portfolioID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Trade
	for rows.Next() {
		var t Trade
		if err := rows.Scan(&t.ID, &t.PortfolioID, &t.Ticker, &t.Side, &t.Quantity, &t.PriceCents, &t.NewsItemID, &t.Reasoning, &t.ExecutedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// TradesByPortfolioRange returns all of a portfolio's trades executed within
// [from, to] (inclusive by day), oldest first — used to show a full period
// (e.g. a specific month) instead of just the most recent few.
func (db *DB) TradesByPortfolioRange(ctx context.Context, portfolioID uuid.UUID, from, to time.Time) ([]Trade, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, portfolio_id, ticker, side, quantity, price_cents, news_item_id, reasoning, executed_at
		FROM trades
		WHERE portfolio_id = $1 AND executed_at >= $2 AND executed_at < $3
		ORDER BY executed_at ASC, ticker ASC, (side = 'sell') ASC
	`, portfolioID, from.Format("2006-01-02"), to.AddDate(0, 0, 1).Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Trade
	for rows.Next() {
		var t Trade
		if err := rows.Scan(&t.ID, &t.PortfolioID, &t.Ticker, &t.Side, &t.Quantity, &t.PriceCents, &t.NewsItemID, &t.Reasoning, &t.ExecutedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ExecuteTrade atomically applies a buy/sell: adjusts the portfolio's cash,
// upserts the position (weighted-average cost on buy), and records the trade.
// side must be "buy" or "sell". newsItemID may be uuid.Nil if not tied to a
// specific news item. executedAt is the SIMULATED trading day the trade
// belongs to (RunDay's `day`) — not necessarily "now": cmd/simulate can
// process March's news on a September real-world run, and trades.executed_at
// must reflect the simulated day so date-range queries (the /agents
// calendar filter, cmd/rewind-agents) work correctly.
func (db *DB) ExecuteTrade(ctx context.Context, portfolioID uuid.UUID, ticker, side string, quantity, priceCents int64, newsItemID uuid.UUID, reasoning string, executedAt time.Time) (*Trade, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	amountCents := quantity * priceCents
	var cashDelta int64
	if side == "buy" {
		cashDelta = -amountCents
	} else {
		cashDelta = amountCents
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE portfolios SET cash_cents = cash_cents + $1
		WHERE id = $2 AND ($1 >= 0 OR cash_cents >= $3)
	`, cashDelta, portfolioID, amountCents)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, fmt.Errorf("insufficient cash or portfolio not found")
	}

	var existingQty, existingAvg int64
	err = tx.QueryRowContext(ctx, `
		SELECT quantity, avg_cost_cents FROM positions WHERE portfolio_id = $1 AND ticker = $2 FOR UPDATE
	`, portfolioID, ticker).Scan(&existingQty, &existingAvg)
	hasPosition := true
	if err == sql.ErrNoRows {
		hasPosition = false
	} else if err != nil {
		return nil, err
	}

	var newQty, newAvg int64
	if side == "buy" {
		newQty = existingQty + quantity
		newAvg = existingAvg
		if newQty > 0 {
			newAvg = (existingQty*existingAvg + quantity*priceCents) / newQty
		}
	} else {
		newQty = existingQty - quantity
		if newQty < 0 {
			newQty = 0
		}
		newAvg = existingAvg
	}

	if hasPosition {
		_, err = tx.ExecContext(ctx, `
			UPDATE positions SET quantity = $1, avg_cost_cents = $2, updated_at = NOW()
			WHERE portfolio_id = $3 AND ticker = $4
		`, newQty, newAvg, portfolioID, ticker)
	} else {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO positions (portfolio_id, ticker, quantity, avg_cost_cents, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
		`, portfolioID, ticker, newQty, newAvg)
	}
	if err != nil {
		return nil, err
	}

	id := uuid.New()
	nullNewsItemID := uuid.NullUUID{UUID: newsItemID, Valid: newsItemID != uuid.Nil}
	if executedAt.IsZero() {
		executedAt = time.Now().UTC()
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO trades (id, portfolio_id, ticker, side, quantity, price_cents, news_item_id, reasoning, executed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, id, portfolioID, ticker, side, quantity, priceCents, nullNewsItemID, reasoning, executedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &Trade{
		ID: id, PortfolioID: portfolioID, Ticker: ticker, Side: side,
		Quantity: quantity, PriceCents: priceCents, NewsItemID: nullNewsItemID, Reasoning: reasoning, ExecutedAt: executedAt,
	}, nil
}

// PositionsMarketValueCents sums quantity * latest known price across a
// portfolio's open positions. Tickers with no price data yet are skipped.
func (db *DB) PositionsMarketValueCents(ctx context.Context, portfolioID uuid.UUID) (int64, error) {
	var cents sql.NullInt64
	err := db.QueryRowContext(ctx, `
		SELECT SUM(p.quantity * lp.close_cents)
		FROM positions p
		JOIN LATERAL (
			SELECT close_cents FROM asset_prices ap WHERE ap.ticker = p.ticker ORDER BY ap.day DESC LIMIT 1
		) lp ON true
		WHERE p.portfolio_id = $1 AND p.quantity > 0
	`, portfolioID).Scan(&cents)
	if err != nil {
		return 0, err
	}
	return cents.Int64, nil
}

// LatestPriceOnOrBefore returns the last known close price (in cents) for a
// ticker on or before the given day.
func (db *DB) LatestPriceOnOrBefore(ctx context.Context, ticker string, day time.Time) (int64, error) {
	var cents int64
	err := db.QueryRowContext(ctx, `
		SELECT close_cents FROM asset_prices
		WHERE ticker = $1 AND day <= $2
		ORDER BY day DESC LIMIT 1
	`, ticker, day.Format("2006-01-02")).Scan(&cents)
	if err == sql.ErrNoRows {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	return cents, nil
}

// UpsertAssetPrice records (or overwrites) one ticker's daily close price.
func (db *DB) UpsertAssetPrice(ctx context.Context, ticker string, day time.Time, closeCents int64) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO asset_prices (ticker, day, close_cents)
		VALUES ($1, $2, $3)
		ON CONFLICT (ticker, day) DO UPDATE SET close_cents = EXCLUDED.close_cents
	`, ticker, day.Format("2006-01-02"), closeCents)
	return err
}

// CountAssetPrices returns how many price rows are loaded (sanity check for
// whether the price seed step has been run).
func (db *DB) CountAssetPrices(ctx context.Context) (int, error) {
	var n int
	err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM asset_prices`).Scan(&n)
	return n, err
}

// InsertFeedback records one Q&A exchange with the feedback/coach agent.
// askedAt lets a weekly review generated during a historical simulation run
// be stamped with the simulated period it reviews, instead of the real wall
// clock time it happened to run at (see cmd/simulate's weeklyFeedback). Zero
// value falls back to now — used for live interactive coach questions.
func (db *DB) InsertFeedback(ctx context.Context, question, answer, provider string, askedAt time.Time) (*FeedbackEntry, error) {
	if askedAt.IsZero() {
		askedAt = time.Now().UTC()
	}
	id := uuid.New()
	_, err := db.ExecContext(ctx, `
		INSERT INTO agent_feedback (id, asked_at, question, answer, provider)
		VALUES ($1, $2, $3, $4, $5)
	`, id, askedAt, question, answer, provider)
	if err != nil {
		return nil, err
	}
	return &FeedbackEntry{ID: id, AskedAt: askedAt, Question: question, Answer: answer, Provider: provider}, nil
}

// RewindPortfolio atomically deletes every trade executed on or after
// `from`, replaces the portfolio's open positions with `positions`, and sets
// cash_cents to `cashCents`. Used by cmd/rewind-agents to safely redo a
// period of the simulation (instead of re-running cmd/simulate over
// already-processed days, which has no idempotency guard and would
// duplicate trades).
func (db *DB) RewindPortfolio(ctx context.Context, portfolioID uuid.UUID, from time.Time, cashCents int64, positions []Position) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM trades WHERE portfolio_id = $1 AND executed_at >= $2
	`, portfolioID, from.Format("2006-01-02")); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM positions WHERE portfolio_id = $1`, portfolioID); err != nil {
		return err
	}
	for _, p := range positions {
		if p.Quantity <= 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO positions (portfolio_id, ticker, quantity, avg_cost_cents, updated_at)
			VALUES ($1, $2, $3, $4, NOW())
		`, portfolioID, p.Ticker, p.Quantity, p.AvgCostCents); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE portfolios SET cash_cents = $1 WHERE id = $2
	`, cashCents, portfolioID); err != nil {
		return err
	}
	return tx.Commit()
}

// FeedbackInRange returns feedback entries asked within [from, to]
// (inclusive by day), newest first — used to inspect a specific period.
func (db *DB) FeedbackInRange(ctx context.Context, from, to time.Time) ([]FeedbackEntry, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, asked_at, question, answer, provider FROM agent_feedback
		WHERE asked_at >= $1 AND asked_at < $2
		ORDER BY asked_at DESC
	`, from.Format("2006-01-02"), to.AddDate(0, 0, 1).Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FeedbackEntry
	for rows.Next() {
		var f FeedbackEntry
		if err := rows.Scan(&f.ID, &f.AskedAt, &f.Question, &f.Answer, &f.Provider); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// RecentFeedback returns the most recent feedback Q&A entries, newest first.
func (db *DB) RecentFeedback(ctx context.Context, limit int) ([]FeedbackEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id, asked_at, question, answer, provider FROM agent_feedback
		ORDER BY asked_at DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FeedbackEntry
	for rows.Next() {
		var f FeedbackEntry
		if err := rows.Scan(&f.ID, &f.AskedAt, &f.Question, &f.Answer, &f.Provider); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}
