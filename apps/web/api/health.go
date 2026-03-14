// Package handler - minimal health endpoint to verify Vercel builds Go (no external deps).
package handler

import (
	"encoding/json"
	"net/http"
)

// Handler responds to GET /api/health with {"status":"ok","service":"market-brief-api"}.
// Minimal endpoint to verify Vercel builds Go; full API is in index.go.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "market-brief-api",
	})
}
