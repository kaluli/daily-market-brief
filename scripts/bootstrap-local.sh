#!/usr/bin/env bash
# Instala dependencias (Go + npm) y aplica migraciones en Postgres local.
# No arranca servidores. Uso: ./scripts/bootstrap-local.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "→ Go modules (apps/api)..."
(cd "$ROOT/apps/api" && go mod download)

echo "→ npm (apps/web)..."
(cd "$ROOT/apps/web" && npm install)

echo "→ Migraciones Postgres..."
export DATABASE_URL="${DATABASE_URL:-postgres://marketbrief:marketbrief_secret@localhost:5432/marketbrief?sslmode=disable}"
(cd "$ROOT/apps/api" && go run cmd/migrate/main.go)

echo ""
echo "Listo. Postgres debe estar en localhost:5432 (usuario marketbrief, DB marketbrief)."
echo "Arrancá la API:  ./scripts/run-api.sh"
echo "Arrancá la web:  make run-web   o   make run-web WEB_PORT=3001"
