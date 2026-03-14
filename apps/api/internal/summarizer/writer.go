package summarizer

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daily-market-brief/api/internal/db"
)

// WriteResult persists the result to a file and returns path and SHA256.
func WriteResult(summariesPath string, day time.Time, res *Result) (path, sha256Hex string, err error) {
	if len(res.Lines) != 100 {
		return "", "", fmt.Errorf("expected 100 lines, got %d", len(res.Lines))
	}
	if err := os.MkdirAll(summariesPath, 0755); err != nil {
		return "", "", err
	}
	fname := day.Format("2006-01-02") + ".txt"
	fpath := filepath.Join(summariesPath, fname)
	content := []byte(strings.Join(res.Lines, "\n") + "\n")
	if err := os.WriteFile(fpath, content, 0644); err != nil {
		return "", "", err
	}
	h := sha256.Sum256(content)
	return fpath, hex.EncodeToString(h[:]), nil
}

// ToDailySummary builds db.DailySummary from Result (for persistence).
func ToDailySummary(day time.Time, res *Result, textPath, textSHA256 string) *db.DailySummary {
	top10, _ := json.Marshal(res.Top10)
	other90, _ := json.Marshal(res.Other90)
	return &db.DailySummary{
		Day:           day,
		GeneratedAt:   time.Now().UTC(),
		Top10:         top10,
		Other90:       other90,
		TextPath:      textPath,
		TextSHA256:    textSHA256,
		ItemsAnalyzed: res.ItemsAnalyzed,
	}
}
