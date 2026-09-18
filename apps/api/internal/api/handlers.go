package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/daily-market-brief/api/internal/agents"
	"github.com/daily-market-brief/api/internal/analyst"
	"github.com/daily-market-brief/api/internal/db"
	"github.com/gofiber/fiber/v2"
)

func (s *Server) root(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"service": "market-brief-api",
		"status":  "ok",
		"health":  c.BaseURL() + "/api/health",
	})
}

func (s *Server) notFound(c *fiber.Ctx) error {
	return c.Status(http.StatusNotFound).JSON(fiber.Map{
		"error":   "not found",
		"path":    c.Path(),
		"health":  c.BaseURL() + "/api/health",
	})
}

func (s *Server) health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "service": "market-brief-api"})
}

func (s *Server) summariesRange(c *fiber.Ctx) error {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" || toStr == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "from and to (YYYY-MM-DD) required"})
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid from date"})
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid to date"})
	}
	summaries, err := s.db.GetDailySummariesRange(c.Context(), from, to)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	type item struct {
		Day           string `json:"day"`
		GeneratedAt   string `json:"generated_at"`
		ItemsAnalyzed int    `json:"items_analyzed"`
	}
	out := make([]item, len(summaries))
	for i := range summaries {
		out[i] = item{
			Day:           summaries[i].Day.Format("2006-01-02"),
			GeneratedAt:   summaries[i].GeneratedAt.Format(time.RFC3339),
			ItemsAnalyzed: summaries[i].ItemsAnalyzed,
		}
	}
	return c.JSON(out)
}

func (s *Server) summaryByDay(c *fiber.Ctx) error {
	day, err := time.Parse("2006-01-02", c.Params("day"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid day"})
	}
	sum, err := s.db.GetDailySummary(c.Context(), day)
	if err != nil {
		if err == db.ErrNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "summary not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"day":             sum.Day.Format("2006-01-02"),
		"generated_at":    sum.GeneratedAt.Format(time.RFC3339),
		"top10":          json.RawMessage(sum.Top10),
		"other90":        json.RawMessage(sum.Other90),
		"items_analyzed": sum.ItemsAnalyzed,
	})
}

func (s *Server) summaryByWeek(c *fiber.Ctx) error {
	weekStr := c.Params("week") // YYYY-WW
	var y, w int
	if _, err := fmt.Sscanf(weekStr, "%d-%d", &y, &w); err != nil || y < 2000 || w < 1 || w > 53 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid week (use YYYY-WW)"})
	}
	start := firstDayOfISOWeek(y, w)
	end := start.AddDate(0, 0, 6)
	summaries, err := s.db.GetDailySummariesRange(c.Context(), start, end)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	type item struct {
		Day string `json:"day"`
	}
	out := make([]item, len(summaries))
	for i := range summaries {
		out[i] = item{Day: summaries[i].Day.Format("2006-01-02")}
	}
	return c.JSON(fiber.Map{"week": weekStr, "summaries": out})
}

func (s *Server) summaryByMonth(c *fiber.Ctx) error {
	monthStr := c.Params("month") // YYYY-MM
	start, err := time.Parse("2006-01", monthStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid month (use YYYY-MM)"})
	}
	end := start.AddDate(0, 1, -1)
	summaries, err := s.db.GetDailySummariesRange(c.Context(), start, end)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	type item struct {
		Day string `json:"day"`
	}
	out := make([]item, len(summaries))
	for i := range summaries {
		out[i] = item{Day: summaries[i].Day.Format("2006-01-02")}
	}
	return c.JSON(fiber.Map{"month": monthStr, "summaries": out})
}

func (s *Server) downloadSummary(c *fiber.Ctx) error {
	day, err := time.Parse("2006-01-02", c.Params("day"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid day"})
	}
	sum, err := s.db.GetDailySummary(c.Context(), day)
	if err != nil {
		if err == db.ErrNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "summary not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	fname := day.Format("2006-01-02") + ".txt"
	fpath := filepath.Join(s.summariesPath, fname)
	body, err := os.ReadFile(fpath)
	if err != nil && sum.TextPath != "" {
		body, err = os.ReadFile(sum.TextPath)
	}
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "file not found"})
	}
	c.Set("Content-Type", "text/plain; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=\""+fname+"\"")
	return c.Send(body)
}

func (s *Server) newsByDay(c *fiber.Ctx) error {
	day, err := time.Parse("2006-01-02", c.Params("day"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid day"})
	}
	items, err := s.db.NewsItemsByDay(c.Context(), day)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	type row struct {
		ID          string    `json:"id"`
		PublishedAt time.Time `json:"published_at"`
		Source      string    `json:"source"`
		Title       string    `json:"title"`
		URL         string    `json:"url"`
		Tickers     []string  `json:"tickers"`
		ImpactScore float64   `json:"impact_score"`
	}
	out := make([]row, len(items))
	for i := range items {
		out[i] = row{
			ID:          items[i].ID.String(),
			PublishedAt: items[i].PublishedAt,
			Source:      items[i].Source,
			Title:       items[i].Title,
			URL:         items[i].URL,
			Tickers:     items[i].Tickers,
			ImpactScore: items[i].ImpactScore,
		}
	}
	return c.JSON(out)
}

// analystPrompt returns the AI investment analyst system prompt and instructions (for use with an LLM).
func (s *Server) analystPrompt(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"system_prompt":     analyst.SystemPrompt,
		"instructions":      analyst.AnalysisInstructions,
		"json_schema_example": analyst.JSONSchemaExample,
	})
}

// analyzeOne runs the analyst on a single news item (POST body: title, url, source, summary optional).
func (s *Server) analyzeOne(c *fiber.Ctx) error {
	var input analyst.NewsInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid JSON body"})
	}
	if input.Title == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "title required"})
	}
	result, err := s.analyst.Analyze(c.Context(), input)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(result)
}

// analysisByDay returns analyses for all news items of the given day (stub or cached; run analyze per item).
func (s *Server) analysisByDay(c *fiber.Ctx) error {
	day, err := time.Parse("2006-01-02", c.Params("day"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid day (use YYYY-MM-DD)"})
	}
	items, err := s.db.NewsItemsByDay(c.Context(), day)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	results := make([]*analyst.AnalysisResult, 0, len(items))
	for i := range items {
		input := analyst.NewsInput{
			Title:  items[i].Title,
			URL:    items[i].URL,
			Source: items[i].Source,
		}
		res, err := s.analyst.Analyze(c.Context(), input)
		if err != nil {
			continue
		}
		results = append(results, res)
	}
	return c.JSON(fiber.Map{"day": day.Format("2006-01-02"), "items_analyzed": len(results), "analyses": results})
}

// agentsPortfolios returns a snapshot of both investor agents (risky, conservative):
// cash, open positions (valued at the latest known price), and recent trades.
// Optional ?from=YYYY-MM-DD&to=YYYY-MM-DD lists every trade in that period
// instead of just the most recent 10 (e.g. to inspect one specific month).
func (s *Server) agentsPortfolios(c *fiber.Ctx) error {
	from, to, err := parseOptionalDateRange(c)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	var out []*agents.PortfolioView
	for _, profile := range agents.All {
		v, err := agents.BuildPortfolioViewRange(c.Context(), s.db, profile, from, to)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		out = append(out, v)
	}
	return c.JSON(fiber.Map{"portfolios": out})
}

// parseOptionalDateRange reads ?from=YYYY-MM-DD&to=YYYY-MM-DD from the query
// string. Returns (nil, nil, nil) if neither is present.
func parseOptionalDateRange(c *fiber.Ctx) (*time.Time, *time.Time, error) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	if fromStr == "" && toStr == "" {
		return nil, nil, nil
	}
	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid from date (use YYYY-MM-DD)")
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid to date (use YYYY-MM-DD)")
	}
	return &from, &to, nil
}

// agentPortfolioByID returns one agent's portfolio. :id is the risk profile
// name ("risky" or "conservative").
func (s *Server) agentPortfolioByID(c *fiber.Ctx) error {
	profile, ok := agents.ByName(c.Params("id"))
	if !ok {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "unknown risk profile (use: risky, conservative)"})
	}
	v, err := agents.BuildPortfolioView(c.Context(), s.db, profile)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(v)
}

// agentsRunDay runs one day's trading cycle for every risk-profile agent:
// funds the monthly allowance if due, analyzes that day's news, and lets
// each profile decide and execute trades. Body: {"day":"YYYY-MM-DD"}
// (optional, defaults to today UTC).
func (s *Server) agentsRunDay(c *fiber.Ctx) error {
	var body struct {
		Day string `json:"day"`
	}
	_ = c.BodyParser(&body) // optional body; ignore parse errors on empty body

	day := time.Now().UTC()
	day = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	if body.Day != "" {
		parsed, err := time.Parse("2006-01-02", body.Day)
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid day (use YYYY-MM-DD)"})
		}
		day = parsed
	}

	results, err := agents.RunDay(c.Context(), s.db, s.analyst, day)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"day": day.Format("2006-01-02"), "results": results})
}

// agentsFeedback asks the third agent — a coach with no trading power — to
// review both portfolios and answer the user's question (or give a general
// review if none is given). Body: {"question": "..."} (optional).
func (s *Server) agentsFeedback(c *fiber.Ctx) error {
	var body struct {
		Question string `json:"question"`
	}
	_ = c.BodyParser(&body)

	risky, err := agents.BuildPortfolioView(c.Context(), s.db, agents.Risky)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	conservative, err := agents.BuildPortfolioView(c.Context(), s.db, agents.Conservative)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	complete, provider := analyst.NewChatCompleterFromEnv()
	userPrompt := agents.BuildFeedbackUserPrompt(risky, conservative, body.Question)
	answer, err := complete(c.Context(), agents.FeedbackSystemPrompt, userPrompt)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("feedback agent (%s): %v", provider, err)})
	}

	if _, err := s.db.InsertFeedback(c.Context(), body.Question, answer, provider); err != nil {
		// Non-fatal: still return the answer even if logging it failed.
		fmt.Println("agents feedback: failed to log:", err)
	}

	return c.JSON(fiber.Map{
		"provider": provider,
		"question": body.Question,
		"answer":   answer,
		"portfolios": fiber.Map{
			"risky":        risky,
			"conservative": conservative,
		},
	})
}

// agentsFeedbackHistory returns the most recent feedback Q&A exchanges.
// Optional ?from=YYYY-MM-DD&to=YYYY-MM-DD restricts to that period instead
// of just the most recent 20 entries.
func (s *Server) agentsFeedbackHistory(c *fiber.Ctx) error {
	from, to, err := parseOptionalDateRange(c)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	var entries []db.FeedbackEntry
	if from != nil && to != nil {
		entries, err = s.db.FeedbackInRange(c.Context(), *from, *to)
	} else {
		entries, err = s.db.RecentFeedback(c.Context(), 20)
	}
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"feedback": entries})
}

func firstDayOfISOWeek(year, week int) time.Time {
	t := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	for t.Weekday() != time.Monday {
		t = t.AddDate(0, 0, -1)
	}
	_, w := t.ISOWeek()
	t = t.AddDate(0, 0, (week-w)*7)
	return t
}

