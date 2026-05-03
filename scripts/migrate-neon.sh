#!/usr/bin/env bash
# Ejecuta migraciones contra Postgres en Neon (misma idea que otros proyectos Go).
# Requiere `.env.neon` en la raíz del repo con DATABASE_URL=postgresql://...
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="$ROOT/.env.neon"

if [ ! -f "$ENV_FILE" ]; then
  echo "Falta $ENV_FILE"
  echo "  cp .env.neon.example .env.neon"
  echo "  Editá .env.neon y pegá DATABASE_URL desde Neon (Dashboard)."
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

echo "Migrando contra Neon (solo DDL / migraciones del repo)..."
cd "$ROOT/apps/api"
go run cmd/migrate/main.go
echo "Listo."
