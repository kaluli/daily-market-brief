.PHONY: up down migrate ingest summarize test install-all bootstrap-local run-api run-web

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
