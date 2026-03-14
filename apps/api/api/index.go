// Package handler is the Vercel serverless entry point for the Go API.
// Deploy this project with Root Directory = apps/api; the function handles all /api/* routes.
package handler

import (
	"net/http"

	"github.com/daily-market-brief/api/pkg/vercel"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	vercel.Handler(w, r)
}
