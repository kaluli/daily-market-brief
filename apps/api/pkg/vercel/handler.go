// Package vercel exposes a net/http handler for Vercel serverless.
package vercel

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/daily-market-brief/api/internal/analyst"
	"github.com/daily-market-brief/api/internal/api"
	"github.com/daily-market-brief/api/internal/db"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

var (
	once   sync.Once
	server *api.Server
)

func initServer() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Print("vercel api: DATABASE_URL not set")
		return
	}
	d, err := db.New(databaseURL)
	if err != nil {
		log.Printf("vercel api: database: %v", err)
		return
	}
	var a analyst.Analyzer
	if os.Getenv("OPENAI_API_KEY") != "" {
		a = analyst.NewOpenAIAnalyzer("", "")
	} else {
		a = analyst.NewStub()
	}
	server = api.New(d, "/tmp/summaries", a, "")
}

// Handler returns the net/http handler for Vercel serverless.
// Request path /api/v1/... is stripped to /api/... for Fiber routes.
// If X-Forwarded-Path is set (e.g. by Next.js proxy), that path is used and /api/v1 stripped.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initServer)
	if server == nil {
		http.Error(w, `{"error":"api not configured (DATABASE_URL)"}`, http.StatusServiceUnavailable)
		return
	}
	path := r.URL.Path
	if p := r.Header.Get("X-Forwarded-Path"); p != "" {
		path = p
	}
	if strings.HasPrefix(path, "/api/v1") {
		path = strings.TrimPrefix(path, "/api/v1")
		if path == "" {
			path = "/"
		}
	}
	r.URL.Path = path
	r.RequestURI = r.URL.String()
	adaptor.FiberApp(server.App())(w, r)
}
