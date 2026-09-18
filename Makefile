.PHONY: up down migrate ingest summarize test install-all bootstrap-local run-api run-web seed-prices agents-run-day agents-feedback simulate

# Default day for summarize (today UTC)
DAY ?= $(shell date -u +%Y-%m-%d)

up:
	docker compose up -d

down:
	docker compose down

migrate:
	cd apps/api && go run cmd/migrate/main.go

ingest:
	cd apps/api && CONFIG_DIR=../../config go run cmd/ingest/main.go

summarize:
	cd apps/api && SUMMARIES_PATH=../../summaries go run cmd/summarize/main.go -day=$(DAY)

# Load historical prices (run scripts/fetch-prices.sh first, in your own terminal)
seed-prices:
	cd apps/api && go run cmd/seed-prices/main.go -dir=../../data/prices

# Trigger one day of trading for both investor agents (API must be running)
agents-run-day:
	curl -sS -X POST http://localhost:$${PORT:-3090}/api/agents/run-day -H "Content-Type: application/json" -d "{\"day\":\"$(DAY)\"}"

# Ask the feedback agent for a general review (API must be running)
agents-feedback:
	curl -sS -X POST http://localhost:$${PORT:-3090}/api/agents/feedback -H "Content-Type: application/json" -d "{}"

# Run the full day-by-day simulation for a date range (API does NOT need to
# be running; talks to the DB directly). Example: make simulate FROM=2026-03-01 TO=2026-03-31
FROM ?= 2026-03-01
TO ?=
simulate:
	./scripts/simulate-neon.sh -from=$(FROM) -to=$(TO)

# "Pisa" (borra) los trades de los agentes desde FROM en adelante y
# recalcula cash/posiciones con el historial anterior, para volver a correr
# cmd/simulate sobre ese tramo sin duplicar operaciones. Por defecto es
# dry-run (no toca nada). Ejemplo:
#   make rewind-agents FROM=2026-03-22            (solo muestra que haria)
#   make rewind-agents FROM=2026-03-22 APPLY=1    (aplica los cambios)
DRYRUN := $(if $(APPLY),false,true)
rewind-agents:
	./scripts/rewind-agents-neon.sh -from=$(FROM) -dry-run=$(DRYRUN)

test:
	cd apps/api && go test ./...
	cd apps/web && npm run build

# Dependencias locales (Go + Node) sin levantar servidores
install-all:
	cd apps/api && go mod download
	cd apps/web && npm install

bootstrap-local:
	chmod +x scripts/bootstrap-local.sh
	./scripts/bootstrap-local.sh

# Run API locally (needs Postgres)
run-api:
	cd apps/api && CONFIG_DIR=../../config SUMMARIES_PATH=../../summaries go run cmd/server/main.go

# Run Web locally (WEB_PORT=3001 si el 3000 está ocupado por otra app)
WEB_PORT ?= 3000
run-web:
	cd apps/web && npm run dev -- -p $(WEB_PORT)
