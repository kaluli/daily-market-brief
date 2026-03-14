package news

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type RSSProvider struct {
	id     string
	name   string
	feedURL string
}

func NewRSSProvider(id, name, feedURL string) *RSSProvider {
	return &RSSProvider{id: id, name: name, feedURL: feedURL}
}

func (p *RSSProvider) ID() string   { return p.id }
func (p *RSSProvider) Name() string { return p.name }

func (p *RSSProvider) Fetch(ctx context.Context) ([]RawItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.feedURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rss status %d", resp.StatusCode)
	}
	var feed struct {
		Channel struct {
			Item []struct {
				Title   string `xml:"title"`
				Link    string `xml:"link"`
				PubDate string `xml:"pubDate"`
				Desc    string `xml:"description"`
			} `xml:"item"`
		} `xml:"channel"`
	}
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, err
	}
	items := make([]RawItem, 0, len(feed.Channel.Item))
	for _, it := range feed.Channel.Item {
		if it.Link == "" {
			continue
		}
		t, _ := time.Parse(time.RFC1123Z, it.PubDate)
		if t.IsZero() {
			t, _ = time.Parse(time.RFC1123, it.PubDate)
		}
		if t.IsZero() {
			t = time.Now().UTC()
		}
		title := strings.TrimSpace(it.Title)
		if title == "" {
			continue
		}
		items = append(items, RawItem{
			Title:       title,
			URL:         it.Link,
			Source:      p.name,
			PublishedAt: t,
			Raw:         map[string]interface{}{"description": it.Desc},
		})
	}
	return items, nil
}
