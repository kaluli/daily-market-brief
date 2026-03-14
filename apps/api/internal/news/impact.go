package news

import (
	"strings"
	"time"

	"github.com/daily-market-brief/api/internal/config"
)

// ImpactScore computes a non-LLM impact score from config weights.
func ImpactScore(item RawItem, cfg *config.RankingWeightsConfig, sourceWeight float64, duplicateCount int) float64 {
	if cfg == nil {
		return 1.0
	}
	score := 1.0
	if sourceWeight > 0 {
		score *= sourceWeight
	} else if w, ok := cfg.SourceWeights[item.Source]; ok {
		score *= w
	} else if w, ok := cfg.SourceWeights["default"]; ok {
		score *= w
	}
	titleLower := strings.ToLower(item.Title)
	for level, keywords := range cfg.Keywords {
		mult := cfg.KeywordScores[level]
		if mult == 0 {
			mult = 1.0
		}
		for _, kw := range keywords {
			if strings.Contains(titleLower, strings.ToLower(kw)) {
				score *= mult
				break
			}
		}
	}
	if len(item.Tickers) > 0 && cfg.TickerBoost > 0 {
		score *= cfg.TickerBoost
	}
	if duplicateCount > 0 && cfg.DuplicateBoost > 0 {
		for i := 0; i < duplicateCount && i < 3; i++ {
			score *= cfg.DuplicateBoost
		}
	}
	hours := time.Since(item.PublishedAt).Hours()
	if cfg.RecencyHoursDecay > 0 && hours > 0 {
		decay := 1.0 - (hours / float64(cfg.RecencyHoursDecay)*0.5)
		if decay < 0.5 {
			decay = 0.5
		}
		score *= decay
	}
	return score
}
