#!/usr/bin/env bash
# Borra los trades de los agentes desde -from en adelante y recalcula
# cash/posiciones con el historial anterior, usando la misma .env/.env.neon
# que el resto de los scripts (Neon). Por defecto es -dry-run=true: no toca
# nada hasta que le pases -dry-run=false explicitamente.
#
# Ejemplos:
#   ./scripts/rewind-agents-neon.sh -from=2026-03-22
#   ./scripts/rewind-agents-neon.sh -from=2026-03-22 -dry-run=false
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
go run cmd/rewind-agents/main.go "$@"
