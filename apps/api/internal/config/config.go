package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type NewsSourcesConfig struct {
	Sources    []SourceConfig    `json:"sources"`
	RSSSources []RSSSourceConfig `json:"rss_sources"`
}

type SourceConfig struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Enabled bool                   `json:"enabled"`
	Weight  float64                `json:"weight"`
	Config  map[string]interface{} `json:"config"`
}

type RSSSourceConfig struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	URL     string  `json:"url"`
	Enabled bool    `json:"enabled"`
	Weight  float64 `json:"weight"`
}

type RankingWeightsConfig struct {
	SourceWeights    map[string]float64 `json:"source_weights"`
	Keywords         map[string][]string `json:"keywords"`
	KeywordScores    map[string]float64 `json:"keyword_scores"`
	RecencyHoursDecay int               `json:"recency_hours_decay"`
	DuplicateBoost   float64            `json:"duplicate_boost"`
	TickerBoost      float64            `json:"ticker_boost"`
}

func LoadNewsSources(configDir string) (*NewsSourcesConfig, error) {
	b, err := os.ReadFile(filepath.Join(configDir, "news_sources.json"))
	if err != nil {
		return nil, err
	}
	var c NewsSourcesConfig
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func LoadRankingWeights(configDir string) (*RankingWeightsConfig, error) {
	b, err := os.ReadFile(filepath.Join(configDir, "ranking_weights.json"))
	if err != nil {
		return nil, err
	}
	var c RankingWeightsConfig
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if c.RecencyHoursDecay == 0 {
		c.RecencyHoursDecay = 24
	}
	if c.DuplicateBoost == 0 {
		c.DuplicateBoost = 1.3
	}
	if c.TickerBoost == 0 {
		c.TickerBoost = 1.1
	}
	return &c, nil
}
