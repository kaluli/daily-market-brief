package news

import (
	"strings"
	"unicode"
)

// NormalizeTitle for similarity: lowercase, collapse spaces, remove punctuation.
func NormalizeTitle(s string) string {
	var b strings.Builder
	lastSpace := true
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(unicode.ToLower(r))
			lastSpace = false
		} else if !lastSpace && (r == ' ' || r == '\t') {
			b.WriteByte(' ')
			lastSpace = true
		}
	}
	return strings.TrimSpace(b.String())
}

// TitleSimilarity returns a value in [0,1]. 1 = same.
func TitleSimilarity(a, b string) float64 {
	an := NormalizeTitle(a)
	bn := NormalizeTitle(b)
	if an == bn {
		return 1.0
	}
	if len(an) == 0 || len(bn) == 0 {
		return 0
	}
	// Jaccard-like: words in common / total unique words
	wa := wordSet(an)
	wb := wordSet(bn)
	if len(wa) == 0 && len(wb) == 0 {
		return 1.0
	}
	common := 0
	for w := range wa {
		if wb[w] {
			common++
		}
	}
	total := len(wa) + len(wb) - common
	if total == 0 {
		return 0
	}
	return float64(common) / float64(total)
}

func wordSet(s string) map[string]bool {
	m := make(map[string]bool)
	for _, w := range strings.Fields(s) {
		if len(w) > 1 {
			m[w] = true
		}
	}
	return m
}

// DedupeByURL returns unique items by URL; first occurrence wins.
func DedupeByURL(items []RawItem) []RawItem {
	seen := make(map[string]bool)
	out := make([]RawItem, 0, len(items))
	for _, it := range items {
		u := strings.TrimSpace(it.URL)
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, it)
	}
	return out
}

// CountSimilarTitles returns how many items in list have title similar to the given title (>= threshold), excluding exact match.
func CountSimilarTitles(items []RawItem, title string, threshold float64) int {
	n := 0
	for _, it := range items {
		if it.Title == title {
			continue
		}
		if TitleSimilarity(it.Title, title) >= threshold {
			n++
		}
	}
	return n
}
