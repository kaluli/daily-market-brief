package summarizer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/daily-market-brief/api/internal/db"
	"github.com/google/uuid"
)

func TestExtractiveSummarizer_Exactly100Lines(t *testing.T) {
	day := time.Date(2025, 3, 10, 0, 0, 0, 0, time.UTC)
	items := make([]db.NewsItem, 95)
	for i := range items {
		items[i] = db.NewsItem{
			ID: uuid.New(), Title: "Headline", Source: "Test", URL: fmt.Sprintf("https://example.com/%d", i),
			ImpactScore: float64(100 - i), PublishedAt: day, Day: day, CreatedAt: time.Now(),
			Raw: []byte("{}"),
		}
	}
	s := NewExtractive()
	res, err := s.Summarize(context.Background(), day, items)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Lines) != 100 {
		t.Fatalf("expected 100 lines, got %d", len(res.Lines))
	}
	if len(res.Top10) != 10 {
		t.Fatalf("expected 10 top, got %d", len(res.Top10))
	}
}
