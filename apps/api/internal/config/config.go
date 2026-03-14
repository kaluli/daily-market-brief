package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// FindConfigDir returns the directory containing news_sources.json.
// It checks CONFIG_DIR env, then tries paths relative to cwd and to the executable.
func FindConfigDir() string {
	if d := os.Getenv("CONFIG_DIR"); d != "" {
		if _, err := os.Stat(filepath.Join(d, "news_sources.json")); err == nil {
			return d
		}
	}
	// Relative to current working directory
	for _, rel := range []string{"config", "../config", "../../config"} {
		if _, err := os.Stat(filepath.Join(rel, "news_sources.json")); err == nil {
			if abs, err := filepath.Abs(rel); err == nil {
				return abs
			}
			return rel
		}
	}
	// Relative to executable (e.g. when run from apps/api, exe is apps/api/server)
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		for _, rel := range []string{"config", "../config", "../../config"} {
			p := filepath.Join(exeDir, rel, "news_sources.json")
			if _, err := os.Stat(p); err == nil {
				return filepath.Dir(p)
			}
		}
	}
	return ""
}

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

// LoadNewsSourcesFromEnv reads news sources from env NEWS_SOURCES_JSON (full JSON).
// Used on Vercel/serverless where there is no config dir.
func LoadNewsSourcesFromEnv() (*NewsSourcesConfig, error) {
	b := os.Getenv("NEWS_SOURCES_JSON")
	if b == "" {
		return nil, os.ErrNotExist
	}
	var c NewsSourcesConfig
	if err := json.Unmarshal([]byte(b), &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// SaveNewsSources writes the news sources config back to news_sources.json.
func SaveNewsSources(configDir string, c *NewsSourcesConfig) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(configDir, "news_sources.json"), b, 0644)
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
