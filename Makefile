.PHONY: up down migrate ingest summarize test

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

test:
	cd apps/api && go test ./...
	cd apps/web && npm run build

# Run API locally (needs Postgres)
run-api:
	cd apps/api && CONFIG_DIR=../../config SUMMARIES_PATH=../../summaries go run cmd/server/main.go

# Run Web locally
run-web:
	cd apps/web && npm run dev
