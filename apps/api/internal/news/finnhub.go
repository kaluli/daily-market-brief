package news

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

const finnhubBase = "https://finnhub.io/api/v1"

type FinnhubProvider struct {
	apiKey string
}

func NewFinnhubProvider() *FinnhubProvider {
	return &FinnhubProvider{apiKey: os.Getenv("FINNHUB_API_KEY")}
}

func (p *FinnhubProvider) ID() string   { return "finnhub" }
func (p *FinnhubProvider) Name() string { return "Finnhub" }

func (p *FinnhubProvider) Fetch(ctx context.Context) ([]RawItem, error) {
	if p.apiKey == "" {
		return nil, nil
	}
	u, _ := url.Parse(finnhubBase + "/news")
	q := u.Query()
	q.Set("token", p.apiKey)
	q.Set("category", "general")
	u.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("finnhub status %d", resp.StatusCode)
	}
	var list []struct {
		Headline    string   `json:"headline"`
		URL        string   `json:"url"`
		Source     string   `json:"source"`
		Datetime   int64    `json:"datetime"`
		Related    string   `json:"related"`
		Summary    string   `json:"summary"`
		Symbols    []string `json:"symbols"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}
	items := make([]RawItem, 0, len(list))
	for _, a := range list {
		if a.URL == "" {
			continue
		}
		ts := time.Unix(a.Datetime, 0)
		src := a.Source
		if src == "" {
			src = "Finnhub"
		}
		tickers := a.Symbols
		if tickers == nil {
			tickers = []string{}
		}
		if a.Related != "" {
			tickers = append(tickers, a.Related)
		}
		items = append(items, RawItem{
			Title:       a.Headline,
			URL:         a.URL,
			Source:      src,
			PublishedAt: ts,
			Tickers:     tickers,
			Raw:         map[string]interface{}{"summary": a.Summary},
		})
	}
	return items, nil
}
