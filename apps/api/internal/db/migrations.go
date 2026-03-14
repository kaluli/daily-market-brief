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
)

func (db *DB) Migrate() error {
	for _, m := range []string{MigrationNewsItems, MigrationDailySummaries, MigrationPortfoliosStub} {
		if _, err := db.Exec(m); err != nil {
			return err
		}
	}
	return nil
}
