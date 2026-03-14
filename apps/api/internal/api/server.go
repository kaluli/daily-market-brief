package api

import (
	"github.com/daily-market-brief/api/internal/analyst"
	"github.com/daily-market-brief/api/internal/db"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Server struct {
	db             *db.DB
	summariesPath  string
	analyst        analyst.Analyzer
	app            *fiber.App
}

// New creates the API server. If a is nil, a stub analyst is used.
func New(database *db.DB, summariesPath string, a analyst.Analyzer) *Server {
	if a == nil {
		a = analyst.NewStub()
	}
	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
		JSONEncoder:           nil,
	})
	app.Use(logger.New())
	app.Use(cors.New())
	s := &Server{
		db:            database,
		summariesPath: summariesPath,
		analyst:       a,
		app:           app,
	}
	s.routes()
	return s
}

func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

func (s *Server) routes() {
	s.app.Get("/", s.root)
	s.app.Get("/api/health", s.health)
	s.app.Get("/api/summaries", s.summariesRange)
	s.app.Get("/api/summaries/day/:day/download", s.downloadSummary)
	s.app.Get("/api/summaries/day/:day", s.summaryByDay)
	s.app.Get("/api/summaries/week/:week", s.summaryByWeek)
	s.app.Get("/api/summaries/month/:month", s.summaryByMonth)
	s.app.Get("/api/news/day/:day", s.newsByDay)

	// Investment analyst (10-step framework, JSON schema)
	s.app.Get("/api/analyst/prompt", s.analystPrompt)
	s.app.Post("/api/analyze", s.analyzeOne)
	s.app.Get("/api/analysis/day/:day", s.analysisByDay)

	// Phase 4 stubs
	s.app.Get("/api/agents/portfolios", s.stubPortfolios)
	s.app.Get("/api/agents/portfolios/:id", s.stubPortfolioByID)

	// 404: evita "Cannot GET /"
	s.app.Use(s.notFound)
}
