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
	maxLineLen   = 160
	fillLine     = "[FILL] Not enough unique items today. Consider widening sources."
	minEuropeTop = 3 // guarantee at least 3 Europe/UE items in TOP 10
)

var europaKeywords = []string{
	"spain", "españa", "español", "madrid", "barcelona", "eurozone", "euro area",
	"eu ", "e.u.", "european union", "unión europea", "ue ", "ecb", "bce", "europe",
	"europ", "brussels", "bruselas", "frankfurt", "lagarde", "euro ", "euros", "ftse",
	"cac ", "dax", "ibex", "banco central europeo", "commission europe",
	"merkel", "scholz", "macron", "sánchez", "sanch", "italy", "italia", "germany",
	"alemania", "france", "francia", "uk ", "reino unido", "brexit",
}

func isEuropeItem(it *db.NewsItem) bool {
	text := strings.ToLower(it.Title + " " + it.Source)
	for _, k := range europaKeywords {
		if strings.Contains(text, k) {
			return true
		}
	}
	return false
}

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
	top10 := buildTop10WithEuropeQuota(items)
	other := make([]RankedItem, 0, 90)
	seenURL := make(map[string]bool)
	for _, r := range top10 {
		seenURL[r.URL] = true
	}
	rank := len(top10) + 1
	for _, it := range items {
		if seenURL[it.URL] {
			continue
		}
		other = append(other, RankedItem{
			Rank:   rank,
			Title:  it.Title,
			Source: it.Source,
			URL:    it.URL,
			Score:  it.ImpactScore,
		})
		rank++
	}
	lines := build100Lines(day, time.Now().UTC(), len(items), top10, other)
	return &Result{
		Top10:         top10,
		Other90:       other,
		Lines:         lines,
		ItemsAnalyzed: len(items),
	}, nil
}

// buildTop10WithEuropeQuota returns 10 items sorted by impact, with at least 3 from Europe/UE.
func buildTop10WithEuropeQuota(items []db.NewsItem) []RankedItem {
	var europe, rest []db.NewsItem
	for i := range items {
		if isEuropeItem(&items[i]) {
			europe = append(europe, items[i])
		} else {
			rest = append(rest, items[i])
		}
	}
	// Take up to 3 from Europe (by impact, already sorted)
	takeEurope := minEuropeTop
	if takeEurope > len(europe) {
		takeEurope = len(europe)
	}
	// Take the rest from all items, excluding the Europe ones we already took
	needRest := 10 - takeEurope
	usedURL := make(map[string]bool)
	var top10 []RankedItem
	for i := 0; i < takeEurope && i < len(europe); i++ {
		it := &europe[i]
		usedURL[it.URL] = true
		top10 = append(top10, RankedItem{
			Rank:   len(top10) + 1,
			Title:  it.Title,
			Source: it.Source,
			URL:    it.URL,
			Score:  it.ImpactScore,
		})
	}
	for i := range items {
		if needRest <= 0 {
			break
		}
		if usedURL[items[i].URL] {
			continue
		}
		usedURL[items[i].URL] = true
		top10 = append(top10, RankedItem{
			Rank:   len(top10) + 1,
			Title:  items[i].Title,
			Source: items[i].Source,
			URL:    items[i].URL,
			Score:  items[i].ImpactScore,
		})
		needRest--
	}
	// Re-sort by score so order reflects impact
	sort.Slice(top10, func(i, j int) bool {
		return top10[i].Score > top10[j].Score
	})
	for i := range top10 {
		top10[i].Rank = i + 1
	}
	return top10
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
