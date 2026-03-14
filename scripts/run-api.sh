#!/usr/bin/env bash
# Run the API server (default port 3090 to avoid conflicts).
set -e
cd "$(dirname "$0")/.."
ROOT="$(pwd)"
API_DIR="$ROOT/apps/api"

export DATABASE_URL="${DATABASE_URL:-postgres://marketbrief:marketbrief_secret@localhost:5432/marketbrief?sslmode=disable}"
export SUMMARIES_PATH="${SUMMARIES_PATH:-$ROOT/summaries}"
export PORT="${PORT:-3090}"

echo "API: http://localhost:$PORT"
echo "Health: http://localhost:$PORT/api/health"
echo ""

cd "$API_DIR"
exec go run cmd/server/main.go
