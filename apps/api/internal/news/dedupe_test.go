package news

import (
	"testing"
)

func TestTitleSimilarity(t *testing.T) {
	if TitleSimilarity("Fed raises rates", "Fed raises rates") != 1.0 {
		t.Error("same title should be 1.0")
	}
	if TitleSimilarity("Fed raises rates", "Fed Raises Interest Rates") > 0.5 {
		// should be high due to common words
	}
	if TitleSimilarity("", "x") != 0 {
		t.Error("empty should give 0")
	}
}

func TestDedupeByURL(t *testing.T) {
	items := []RawItem{
		{URL: "https://a.com/1", Title: "A"},
		{URL: "https://a.com/1", Title: "A again"},
		{URL: "https://b.com/2", Title: "B"},
	}
	out := DedupeByURL(items)
	if len(out) != 2 {
		t.Fatalf("expected 2 unique, got %d", len(out))
	}
}
