package news

import "context"

// Provider fetches and returns raw news items. Each implementation (NewsAPI, Finnhub, RSS) implements this.
type Provider interface {
	ID() string
	Name() string
	Fetch(ctx context.Context) ([]RawItem, error)
}
