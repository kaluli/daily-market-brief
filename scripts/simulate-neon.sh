#!/usr/bin/env bash
# Runs the full agents simulation (day by day, monthly funding + weekly
# feedback) against your Neon database, using the same .env / .env.neon the
# API uses. Safe to run in date-range slices; see cmd/simulate's doc comment
# and docs/AGENTS.md for details (and the one important caveat: don't re-run
# a range you already simulated).
#
# Examples:
#   ./scripts/simulate-neon.sh -from=2026-03-01 -to=2026-03-31
#   nohup ./scripts/simulate-neon.sh -from=2026-04-01 -to=2026-04-30 > simulate-abril.log 2>&1 &
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

for ENV_FILE in "$ROOT/.env.neon" "$ROOT/.env"; do
  if [ -f "$ENV_FILE" ]; then
    set -a
    # shellcheck disable=SC1090
    source "$ENV_FILE"
    set +a
  fi
done

if [ -z "${DATABASE_URL:-}" ]; then
  echo "Falta DATABASE_URL (revisa $ROOT/.env.neon)"
  exit 1
fi

cd "$ROOT/apps/api"
go run cmd/simulate/main.go "$@"
