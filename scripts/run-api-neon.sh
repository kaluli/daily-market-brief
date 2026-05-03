#!/usr/bin/env bash
# Arranca la API en local usando DATABASE_URL de Neon (desde .env.neon).
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$ROOT/.env.neon"

if [ ! -f "$ENV_FILE" ]; then
  echo "Falta $ENV_FILE — copiá desde .env.neon.example y pegá tu DATABASE_URL de Neon."
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL está vacío en $ENV_FILE"
  exit 1
fi

export DATABASE_URL
exec "$ROOT/scripts/run-api.sh"
