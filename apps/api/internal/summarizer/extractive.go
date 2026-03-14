package summarizer

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/daily-market-brief/api/internal/db"
)

const (
	maxLineLen = 160
	fillLine   = "[FILL] Not enough unique items today. Consider widening sources."
)

// ExtractiveSummarizer implements Summarizer without LLM: ranking + strict 100-line format.
type ExtractiveSummarizer struct{}

func NewExtractive() *ExtractiveSummarizer {
	return &ExtractiveSummarizer{}
}

func (s *ExtractiveSummarizer) Summarize(ctx context.Context, day time.Time, items []db.NewsItem) (*Result, error) {
	if len(items) == 0 {
		return s.emptyResult(day), nil
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ImpactScore > items[j].ImpactScore
	})
	top10 := make([]RankedItem, 0, 10)
	other := make([]RankedItem, 0, 90)
	for i, it := range items {
		r := RankedItem{
			Rank:   i + 1,
			Title:  it.Title,
			Source: it.Source,
			URL:    it.URL,
			Score:  it.ImpactScore,
		}
		if i < 10 {
			top10 = append(top10, r)
		} else {
			other = append(other, r)
		}
	}
	lines := build100Lines(day, time.Now().UTC(), len(items), top10, other)
	return &Result{
		Top10:         top10,
		Other90:       other,
		Lines:         lines,
		ItemsAnalyzed: len(items),
	}, nil
}

func (s *ExtractiveSummarizer) emptyResult(day time.Time) *Result {
	lines := make([]string, 100)
	lines[0] = fmt.Sprintf("DAILY MARKET BRIEF - %s (US Markets)", day.Format("2006-01-02"))
	lines[1] = fmt.Sprintf("Generated at: %s | Sources: 0 | Items analyzed: 0", time.Now().UTC().Format(time.RFC3339))
	lines[2] = "============================================================"
	lines[3] = "TOP 10 (Most Influential)"
	for i := 4; i < 14; i++ {
		lines[i] = fmt.Sprintf("%02d) (No items)", i-3)
	}
	lines[14] = "------------------------------------------------------------"
	lines[15] = "OTHER 90 (Ranked)"
	for i := 16; i < 100; i++ {
		lines[i] = fillLine
	}
	return &Result{
		Top10:         nil,
		Other90:       nil,
		Lines:         lines,
		ItemsAnalyzed: 0,
	}
}

func build100Lines(day time.Time, generatedAt time.Time, itemsAnalyzed int, top10 []RankedItem, other []RankedItem) []string {
	lines := make([]string, 100)
	lines[0] = fmt.Sprintf("DAILY MARKET BRIEF - %s (US Markets)", day.Format("2006-01-02"))
	lines[1] = fmt.Sprintf("Generated at: %s | Sources: - | Items analyzed: %d", generatedAt.Format(time.RFC3339), itemsAnalyzed)
	lines[2] = "============================================================"
	lines[3] = "TOP 10 (Most Influential)"
	for i := 0; i < 10; i++ {
		if i < len(top10) {
			lines[4+i] = formatTopLine(i+1, top10[i])
		} else {
			lines[4+i] = fmt.Sprintf("%02d) (No item)", i+1)
		}
	}
	lines[14] = "------------------------------------------------------------"
	lines[15] = "OTHER 90 (Ranked)"
	needOther := 84
	for i := 0; i < needOther && i < len(other); i++ {
		lines[16+i] = formatOtherLine(other[i])
	}
	fillStart := 16 + min(needOther, len(other))
	for i := fillStart; i < 100; i++ {
		lines[i] = fillLine
	}
	return lines
}

func formatTopLine(rank int, r RankedItem) string {
	s := fmt.Sprintf("%02d) %s — %s — %s", rank, truncate(r.Title, 80), r.Source, r.URL)
	return truncate(s, maxLineLen)
}

func formatOtherLine(r RankedItem) string {
	s := fmt.Sprintf("%s — %s — %s", truncate(r.Title, 80), r.Source, r.URL)
	return truncate(s, maxLineLen)
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
