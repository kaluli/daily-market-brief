package api

import (
	"github.com/daily-market-brief/api/internal/config"
	"github.com/gofiber/fiber/v2"
)

// adminSourcesGet returns the current news sources config (sources + rss_sources).
func (s *Server) adminSourcesGet(c *fiber.Ctx) error {
	configDir := config.FindConfigDir()
	if configDir == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "no se encontró config (news_sources.json). Usa CONFIG_DIR o arranca desde la raíz del repo o apps/api.",
		})
	}
	cfg, err := config.LoadNewsSources(configDir)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to load news sources config: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"sources":     cfg.Sources,
		"rss_sources": cfg.RSSSources,
	})
}

type setEnabledBody struct {
	Enabled bool `json:"enabled"`
}

// adminSourceSetEnabled toggles the enabled flag for a source or rss_source by id.
func (s *Server) adminSourceSetEnabled(c *fiber.Ctx) error {
	configDir := config.FindConfigDir()
	if configDir == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "no se encontró config. Usa CONFIG_DIR o arranca desde la raíz del repo o apps/api.",
		})
	}
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing id"})
	}
	var body setEnabledBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body, expected { \"enabled\": true|false }"})
	}

	cfg, err := config.LoadNewsSources(configDir)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to load config: " + err.Error(),
		})
	}

	found := false
	for i := range cfg.Sources {
		if cfg.Sources[i].ID == id {
			cfg.Sources[i].Enabled = body.Enabled
			found = true
			break
		}
	}
	if !found {
		for i := range cfg.RSSSources {
			if cfg.RSSSources[i].ID == id {
				cfg.RSSSources[i].Enabled = body.Enabled
				found = true
				break
			}
		}
	}
	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "source not found: " + id})
	}

	if err := config.SaveNewsSources(configDir, cfg); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save config: " + err.Error(),
		})
	}
	return c.JSON(fiber.Map{"id": id, "enabled": body.Enabled})
}
