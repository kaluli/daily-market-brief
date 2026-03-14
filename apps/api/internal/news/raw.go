package news

import "time"

// RawItem is the normalized shape produced by providers before DB persistence.
type RawItem struct {
	Title       string
	URL         string
	Source      string
	PublishedAt time.Time
	Tickers     []string
	Raw         map[string]interface{}
}
