#!/usr/bin/env bash
# Loads the CSVs from data/prices/ (see scripts/fetch-prices.sh) into the
# asset_prices table in your Neon database.
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$ROOT/.env.neon"

if [ ! -f "$ENV_FILE" ]; then
  echo "Falta $ENV_FILE — copia desde .env.neon.example y pega tu DATABASE_URL de Neon."
  exit 1
fi

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL esta vacio en $ENV_FILE"
  exit 1
fi

cd "$ROOT/apps/api"
go run cmd/seed-prices/main.go -dir="$ROOT/data/prices"
