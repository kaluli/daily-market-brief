package summarizer

import (
	"context"
	"time"

	"github.com/daily-market-brief/api/internal/db"
)

// Summarizer generates a daily summary from news items. Replaceable by LLM implementation.
type Summarizer interface {
	// Summarize produces the 100-line text and metadata from items for the given day.
	Summarize(ctx context.Context, day time.Time, items []db.NewsItem) (*Result, error)
}

// Result holds the generated summary content and structured data for persistence.
type Result struct {
	Top10         []RankedItem
	Other90       []RankedItem
	Lines         []string // exactly 100 lines
	ItemsAnalyzed int
}

type RankedItem struct {
	Rank   int     `json:"rank"`
	Title  string  `json:"title"`
	Source string  `json:"source"`
	URL    string  `json:"url"`
	Score  float64 `json:"score"`
}
