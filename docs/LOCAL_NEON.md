# API local contra Postgres en Neon

Mismo flujo que en otros proyectos Go (asistente, scraper idealista): **una sola variable `DATABASE_URL`** apuntando a Neon, migraciones una vez, después corrés el server.

## 1. Connection string en Neon

1. Entrá a [Neon](https://neon.tech) → tu proyecto (podés crear una base/branche nueva solo para Market Brief).
2. **Connection string** → copiá la URL (modo **transaction** o **session**, con `sslmode=require`).

## 2. Archivo `.env.neon` (local, no se sube a Git)

En la raíz del repo:

```bash
cp .env.neon.example .env.neon
```

Editá `.env.neon` y completá:

```bash
DATABASE_URL=postgresql://usuario:password@ep-xxxx.region.aws.neon.tech/neondb?sslmode=require
```

## 3. Migraciones (una vez por base nueva)

```bash
chmod +x scripts/migrate-neon.sh scripts/run-api-neon.sh   # si hace falta
./scripts/migrate-neon.sh
```

Debe imprimir `migrations ok`.

## 4. Arrancar la API contra Neon

```bash
./scripts/run-api-neon.sh
```

La API queda en **http://localhost:3090** igual que con Postgres local.

## 5. Web Next.js en local

En otra terminal (la web sigue usando el proxy a la API en 3090):

```bash
cd apps/web && npm run dev
```

Abrí **http://localhost:3000**. Si ese puerto está ocupado: `npm run dev:3001` o desde la raíz `make run-web WEB_PORT=3001` → **http://localhost:3001**.

## 6. Datos en Neon (opcional)

Con la misma `DATABASE_URL` en el entorno:

```bash
set -a && source .env.neon && set +a
make ingest
make summarize
```

O exportá `DATABASE_URL` manualmente antes de `make ingest` / `make summarize`.

## Notas

- **GitHub Actions** (`daily-ingest`) debe usar la **misma** `DATABASE_URL` de Neon si querés que el cron llene la misma base que la API en producción.
- No commitees `.env.neon` (está en `.gitignore`).
