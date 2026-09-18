package db

const (
	MigrationNewsItems = `
CREATE TABLE IF NOT EXISTS news_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    published_at TIMESTAMPTZ NOT NULL,
    day DATE NOT NULL,
    source TEXT NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    tickers TEXT[],
    raw JSONB,
    impact_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_news_items_day ON news_items(day);
CREATE INDEX IF NOT EXISTS idx_news_items_published ON news_items(published_at);
CREATE INDEX IF NOT EXISTS idx_news_items_impact ON news_items(day, impact_score DESC);
`

	MigrationDailySummaries = `
CREATE TABLE IF NOT EXISTS daily_summaries (
    day DATE PRIMARY KEY,
    generated_at TIMESTAMPTZ NOT NULL,
    top10 JSONB NOT NULL,
    other90 JSONB NOT NULL,
    text_path TEXT NOT NULL,
    text_sha256 TEXT,
    items_analyzed INT NOT NULL DEFAULT 0
);
`

	// Phase 4 stubs: placeholder tables for investor agents
	MigrationPortfoliosStub = `
CREATE TABLE IF NOT EXISTS portfolios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_type TEXT NOT NULL,
    started_at DATE NOT NULL,
    initial_cash_cents BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS trades (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id UUID REFERENCES portfolios(id),
    ticker TEXT NOT NULL,
    side TEXT NOT NULL,
    quantity INT NOT NULL,
    price_cents BIGINT NOT NULL,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id UUID REFERENCES portfolios(id),
    ticker TEXT NOT NULL,
    quantity INT NOT NULL,
    avg_cost_cents BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

	// Phase 4: the actual investor agents (risky / conservative trading
	// profiles + a feedback coach). Builds on top of the stub tables above
	// with additive, idempotent changes so it's safe to run against a
	// database that already has the stub tables.
	MigrationAgentsSchema = `
ALTER TABLE portfolios ADD COLUMN IF NOT EXISTS risk_profile TEXT NOT NULL DEFAULT '';
ALTER TABLE portfolios ADD COLUMN IF NOT EXISTS cash_cents BIGINT NOT NULL DEFAULT 0;
ALTER TABLE portfolios ADD COLUMN IF NOT EXISTS monthly_allowance_cents BIGINT NOT NULL DEFAULT 500000;
ALTER TABLE portfolios ADD COLUMN IF NOT EXISTS last_funded_month DATE;
CREATE UNIQUE INDEX IF NOT EXISTS idx_portfolios_risk_profile ON portfolios(risk_profile);

ALTER TABLE trades ADD COLUMN IF NOT EXISTS news_item_id UUID REFERENCES news_items(id);
ALTER TABLE trades ADD COLUMN IF NOT EXISTS reasoning TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_trades_portfolio_executed ON trades(portfolio_id, executed_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_positions_portfolio_ticker ON positions(portfolio_id, ticker);

CREATE TABLE IF NOT EXISTS asset_prices (
    ticker TEXT NOT NULL,
    day DATE NOT NULL,
    close_cents BIGINT NOT NULL,
    PRIMARY KEY (ticker, day)
);

CREATE TABLE IF NOT EXISTS agent_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    asked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    question TEXT NOT NULL DEFAULT '',
    answer TEXT NOT NULL,
    provider TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_agent_feedback_asked_at ON agent_feedback(asked_at DESC);
`
)

func (db *DB) Migrate() error {
	for _, m := range []string{MigrationNewsItems, MigrationDailySummaries, MigrationPortfoliosStub, MigrationAgentsSchema} {
		if _, err := db.Exec(m); err != nil {
			return err
		}
	}
	return nil
}
