package handler

import (
	"net/http"

	"github.com/daily-market-brief/api/pkg/vercel"
)

// Handler is the Vercel serverless entry point.
func Handler(w http.ResponseWriter, r *http.Request) {
	vercel.Handler(w, r)
}
