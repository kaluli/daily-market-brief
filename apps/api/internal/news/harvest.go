package news

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/daily-market-brief/api/internal/config"
	"github.com/daily-market-brief/api/internal/db"
	"github.com/google/uuid"
)

const titleSimilarityThreshold = 0.6

// Harvest runs all enabled providers, normalizes, dedupes, scores, and persists.
func Harvest(ctx context.Context, database *db.DB, configDir string) (int, error) {
	sourcesCfg, err := config.LoadNewsSources(configDir)
	if err != nil {
		return 0, err
	}
	weightsCfg, err := config.LoadRankingWeights(configDir)
	if err != nil {
		return 0, err
	}
	providers := buildProviders(sourcesCfg)
	var all []RawItem
	for _, p := range providers {
		items, err := p.Fetch(ctx)
		if err != nil {
			log.Printf("provider %s fetch error: %v", p.ID(), err)
			continue
		}
		_ = sourceWeight(sourcesCfg, p) // used by sourceWeightByName via item.Source
		for i := range items {
			items[i].Source = p.Name()
			all = append(all, items[i])
		}
	}
	all = DedupeByURL(all)
	now := time.Now().UTC()
	inserted := 0
	for i := range all {
		item := &all[i]
		day := time.Date(item.PublishedAt.Year(), item.PublishedAt.Month(), item.PublishedAt.Day(), 0, 0, 0, 0, time.UTC)
		dupCount := CountSimilarTitles(all, item.Title, titleSimilarityThreshold)
		sw := sourceWeightByName(sourcesCfg, weightsCfg, item.Source)
		score := ImpactScore(*item, weightsCfg, sw, dupCount)
		rawJSON, _ := json.Marshal(item.Raw)
		if rawJSON == nil {
			rawJSON = []byte("{}")
		}
		exists, _ := database.URLExists(ctx, item.URL)
		if exists {
			continue
		}
		dbItem := &db.NewsItem{
			ID:          uuid.New(),
			PublishedAt: item.PublishedAt,
			Day:         day,
			Source:      item.Source,
			Title:       item.Title,
			URL:         item.URL,
			Tickers:     item.Tickers,
			Raw:         rawJSON,
			ImpactScore: score,
			CreatedAt:   now,
		}
		if err := database.InsertNewsItem(ctx, dbItem); err != nil {
			log.Printf("insert error %s: %v", item.URL, err)
			continue
		}
		inserted++
	}
	return inserted, nil
}

func buildProviders(sourcesCfg *config.NewsSourcesConfig) []Provider {
	var out []Provider
	for _, s := range sourcesCfg.Sources {
		if !s.Enabled {
			continue
		}
		switch s.ID {
		case "newsapi":
			out = append(out, NewNewsAPIProvider())
		case "finnhub":
			out = append(out, NewFinnhubProvider())
		}
	}
	for _, r := range sourcesCfg.RSSSources {
		if !r.Enabled {
			continue
		}
		out = append(out, NewRSSProvider(r.ID, r.Name, r.URL))
	}
	return out
}

func sourceWeight(sourcesCfg *config.NewsSourcesConfig, p Provider) float64 {
	for _, s := range sourcesCfg.Sources {
		if s.ID == p.ID() {
			return s.Weight
		}
	}
	for _, r := range sourcesCfg.RSSSources {
		if r.ID == p.ID() {
			return r.Weight
		}
	}
	return 1.0
}

func sourceWeightByName(sourcesCfg *config.NewsSourcesConfig, weightsCfg *config.RankingWeightsConfig, name string) float64 {
	if w, ok := weightsCfg.SourceWeights[name]; ok {
		return w
	}
	for _, s := range sourcesCfg.Sources {
		if s.Name == name {
			return s.Weight
		}
	}
	for _, r := range sourcesCfg.RSSSources {
		if r.Name == name {
			return r.Weight
		}
	}
	if w, ok := weightsCfg.SourceWeights["default"]; ok {
		return w
	}
	return 1.0
}
