package handler

import (
	"encoding/json"
	"net/http"
)

// Handler responds to GET /api/health with {"status":"ok","service":"market-brief-api"}.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "market-brief-api",
	})
}
