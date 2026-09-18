#!/usr/bin/env bash
# Run the API server (default port 3090 to avoid conflicts).
set -e
cd "$(dirname "$0")/.."
ROOT="$(pwd)"
API_DIR="$ROOT/apps/api"

# Load ./.env if present (e.g. LLM_PROVIDER/OLLAMA_* for the local analyst).
# Values already exported in the shell take precedence over the file.
if [ -f "$ROOT/.env" ]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

export DATABASE_URL="${DATABASE_URL:-postgres://marketbrief:marketbrief_secret@localhost:5432/marketbrief?sslmode=disable}"
export SUMMARIES_PATH="${SUMMARIES_PATH:-$ROOT/summaries}"
export CONFIG_DIR="${CONFIG_DIR:-$ROOT/config}"
export PORT="${PORT:-3090}"

echo "API: http://localhost:$PORT"
echo "Health: http://localhost:$PORT/api/health"
echo ""

cd "$API_DIR"
exec go run cmd/server/main.go
