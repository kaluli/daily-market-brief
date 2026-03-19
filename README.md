# Daily Market Brief

Daily market brief: financial news ingestion, impact-based ranking, and summaries (top 10 + rest). Web app with calendar and views by day, week, and month.

---

## How to run (local, without Docker)

You need **PostgreSQL** running and **Go** and **Node** installed.

### 0. Base de datos (obligatorio)

La API usa Postgres. Por defecto: `postgres://marketbrief:marketbrief_secret@localhost:5432/marketbrief`.

- **Con Docker:** desde la raíz del repo:
  ```bash
  docker compose up -d postgres
  cd apps/api && DATABASE_URL="postgres://marketbrief:marketbrief_secret@localhost:5432/marketbrief?sslmode=disable" go run cmd/migrate/main.go
  ```
- **Sin Docker:** ten Postgres en 5432 con usuario `marketbrief`, DB `marketbrief`, y ejecuta las migraciones:
  ```bash
  cd apps/api && DATABASE_URL="postgres://marketbrief:marketbrief_secret@localhost:5432/marketbrief?sslmode=disable" go run cmd/migrate/main.go
  ```
  Debe salir `migrations ok`. Si la API arranca y falla con `database: ...`, revisa que Postgres esté arriba y las migraciones hechas.

### 1. Start the API

En una terminal, desde la **raíz del proyecto**:

```bash
./scripts/run-api.sh
```

Déjala abierta. Cuando veas `listening on http://localhost:3090`, la API está lista.

- **API:** http://localhost:3090  
- **Health:** http://localhost:3090/api/health  

### 2. Start the web app

En **otra terminal**, desde la raíz:

```bash
cd apps/web
npm install   # solo la primera vez
npm run dev
```

Abre en el navegador **http://localhost:3000** (calendario y vistas por día/semana/mes). La web usa por defecto el proxy `/api/v1` cuando corre en localhost:3000, así que no hace falta definir `NEXT_PUBLIC_API_URL`.

### 3. (Optional) Populate the calendar with data

To see summaries in the web app, ingest news and generate at least one day:

```bash
# From the project root
make ingest      # fetches news (RSS works without API keys; NewsAPI/Finnhub optional)
make summarize   # generates summary for today (UTC)
# Or a specific day: make summarize DAY=2026-03-14
```

---

## URL summary

| Service | URL |
|---------|-----|
| **Web (calendar)** | http://localhost:3000 |
| **API** | http://localhost:3090 |
| **API via web (same origin)** | http://localhost:3000/api/v1 |
| **Admin (sources)** | http://localhost:3000/api/v1/admin |

En local, si abrís la web en **http://localhost:3000**, el frontend usa el proxy **/api/v1** (mismo origen). La API debe estar en 3090; las peticiones desde el navegador pasan por Next.js y este reenvía al backend. Admin: **http://localhost:3000/api/v1/admin**. Si la API va en otro puerto, arrancá la web con:

```bash
NEXT_PUBLIC_API_URL=http://localhost:PORT npm run dev
```

---

## Run with Docker

To run everything in containers:

```bash
mkdir -p summaries
docker compose up -d --build
docker compose exec api /app/migrate
# Optional: docker compose exec api /app/ingest
# Optional: docker compose exec api /app/summarize
```

- **Web:** http://localhost:3000  
- **API:** http://localhost:3090  

---

## Deploy en Vercel (solo web)

**Usar cuenta de GitHub kaluli en Vercel (no karinap-eb).**

Solo la **app web** (Next.js) en Vercel. La **API en Go** desplegada aparte; su URL se configura en `NEXT_PUBLIC_API_URL`.

### Pasos

1. **Importar el repo en Vercel**  
   [vercel.com/new](https://vercel.com/new) → iniciar sesión con **kaluli** → Import Git Repository → elige `daily-market-brief`.

2. **Configurar el proyecto**
   - **Root Directory:** ⚠️ **OBLIGATORIO** — haz clic en *Edit* y selecciona **`apps/web`**. Si está vacío o en la raíz, el build fallará con "No Output Directory named .next found".
   - **Framework Preset:** Next.js (se detecta solo).
   - **Build Command:** (vacío; usa `npm run build` por defecto)
   - **Output Directory:** (vacío; Next.js por defecto)

3. **Variables de entorno**
   - **`NEXT_PUBLIC_API_URL`** = URL pública de tu API + `/api/v1`.

4. **Deploy**  
   Pulsa *Deploy*. Vercel construye `apps/web` y publica la app.

La API (Go + Postgres) se despliega por separado. Ver `docs/DEPLOY_VERCEL.md` para la web en Vercel.

---

## Production: API behind the same domain (port 3000)

To have everything go through the same domain/port in production (e.g. only expose 3000):

1. **Rewrite:** Next.js is already configured so that any request under **`/api/v1`** is proxied to the Go API (API versioning, good practice). No code changes needed.

2. **Environment variables on the Next.js server (build or runtime):**
   - **`API_BACKEND_URL`** = Internal URL of the Go service (e.g. `http://api:3090` in Docker/Kubernetes, or `http://localhost:3090` if the API runs on the same machine).

3. **Frontend variables (build time):**
   - **`NEXT_PUBLIC_API_URL`** = `''` or the public URL with the prefix, e.g. `https://yourdomain.com/api/v1`. This way the calendar and the rest of the web app call the API from the same origin.

4. **Production URLs:**
   - Web: `https://yourdomain.com/`
   - Admin (sources): `https://yourdomain.com/api/v1/admin`
   - Health API: `https://yourdomain.com/api/v1/api/health`

You only need to expose the Next.js entry point; the API can stay internal.

---

## Requirements

- **Local:** Go 1.22+, Node 20+, PostgreSQL 16+
- **Docker:** Docker and Docker Compose

---

## Project structure

| Part | Description |
|------|-------------|
| `apps/api` | Go API (Fiber): summaries, news, health. Commands: server, migrate, ingest, summarize. |
| `apps/web` | Next.js frontend: calendar, day, week, month, last-10. |
| `config/` | `news_sources.json`, `ranking_weights.json`. |
| `summaries/` | `.txt` files generated by `summarize`. |
| `scripts/run-api.sh` | Script to start the API on port 3090. |

---

## Makefile (from project root)

| Command | Description |
|---------|-------------|
| `make migrate` | Run database migrations (local Postgres). |
| `make ingest` | Ingest news from configured sources. |
| `make summarize` | Generate summary for the day; `make summarize DAY=YYYY-MM-DD` for a specific date. |
| `make run-api` | Start the API (Go). |
| `make run-web` | Start the web app (`npm run dev`). |
| `make up` / `make down` | Docker compose up/down. |

---

## Investment analyst (AI)

The **investment analyst** module applies a 10-step framework to each news item: relevance filter, category, summary, market impact, affected assets, directional bias, investment signals, time horizon, and signal strength. Output is structured JSON.

- **GET /api/analyst/prompt** — Returns the system prompt and instructions for use with an LLM (OpenAI, Anthropic, etc.).
- **POST /api/analyze** — Analyze a single news item (body: `{"title","url","source","summary"}`). Returns the schema JSON.
- **GET /api/analysis/day/:day** — Analyze all news items for the day and return an array of analyses.

**With OpenAI:** Set `OPENAI_API_KEY` (and optionally `OPENAI_MODEL`, default `gpt-4o-mini`). When the API starts, the LLM analyst is used. Without the key, a stub returns placeholder values.

---

## Troubleshooting

- **"No Output Directory named .next found"** en Vercel → En el proyecto Vercel: Settings → General → Root Directory. Debe estar en **`apps/web`**, no vacío ni en la raíz del repo.
- **Calendario vacío / no trae noticias** en producción → En Vercel (proyecto web): Settings → Environment Variables. Debe tener **`API_BACKEND_URL`** y **`NEXT_PUBLIC_API_URL`** = URL de tu API (ej. `https://daily-market-brief-api-xxx.vercel.app`). Redeploy después de agregar variables.
- **"database: ..."** when starting the API → Check that Postgres is running and your database is configured (see `.env.example`).
- **"address already in use"** → Port 3090 is taken. Use `PORT=3091 ./scripts/run-api.sh` and for the web: `NEXT_PUBLIC_API_URL=http://localhost:3091 npm run dev`.
- **Web shows no data** → Run `make ingest` then `make summarize` (or `make summarize DAY=YYYY-MM-DD`).
