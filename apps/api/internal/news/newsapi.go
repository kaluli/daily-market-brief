package news

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const newsAPIBase = "https://newsapi.org/v2"

type NewsAPIProvider struct {
	apiKey string
}

func NewNewsAPIProvider() *NewsAPIProvider {
	return &NewsAPIProvider{apiKey: os.Getenv("NEWSAPI_KEY")}
}

func (p *NewsAPIProvider) ID() string   { return "newsapi" }
func (p *NewsAPIProvider) Name() string { return "NewsAPI" }

func (p *NewsAPIProvider) Fetch(ctx context.Context) ([]RawItem, error) {
	if p.apiKey == "" {
		return nil, nil
	}
	u, _ := url.Parse(newsAPIBase + "/everything")
	q := u.Query()
	q.Set("apiKey", p.apiKey)
	q.Set("language", "en")
	q.Set("sortBy", "publishedAt")
	q.Set("pageSize", "50")
	q.Set("q", "stock market OR Fed OR inflation OR earnings OR economy OR Wall Street")
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
		return nil, fmt.Errorf("newsapi status %d", resp.StatusCode)
	}
	var out struct {
		Articles []struct {
			Title       string    `json:"title"`
			URL         string    `json:"url"`
			Source      struct{ Name string } `json:"source"`
			PublishedAt time.Time `json:"publishedAt"`
			Description string    `json:"description"`
		} `json:"articles"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	items := make([]RawItem, 0, len(out.Articles))
	for _, a := range out.Articles {
		if a.URL == "" || strings.HasPrefix(a.URL, "https://removed.") {
			continue
		}
		sourceName := a.Source.Name
		if sourceName == "" {
			sourceName = "NewsAPI"
		}
		items = append(items, RawItem{
			Title:       a.Title,
			URL:         a.URL,
			Source:      sourceName,
			PublishedAt: a.PublishedAt,
			Raw:         map[string]interface{}{"description": a.Description},
		})
	}
	return items, nil
}
