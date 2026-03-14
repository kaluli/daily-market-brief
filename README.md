# Daily Market Brief

Daily market brief: financial news ingestion, impact-based ranking, and summaries (top 10 + rest). Web app with calendar and views by day, week, and month.

---

## How to run (local, without Docker)

You need **PostgreSQL** running and **Go** and **Node** installed.

### 1. Start the API

In a terminal, from the **project root**:

```bash
./scripts/run-api.sh
```

Keep the terminal open. When you see `listening on http://localhost:3090`, the API is ready.

- **API:** http://localhost:3090  
- **Health:** http://localhost:3090/api/health  

### 2. Start the web app

In **another terminal**, from the project root:

```bash
cd apps/web
npm install   # first time only
npm run dev
```

Open in your browser: **http://localhost:3000** (calendar and day/week/month views).

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

The web app uses the API on port **3090** by default. To use the API through the same origin (port 3000), use the **`/api/v1`** path (API versioning): Next.js rewrites those requests to the Go API. Admin: **http://localhost:3000/api/v1/admin**. If you use a different port for the API, start the web with:

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

## Deploy todo en Vercel (web + API, gratis sin tarjeta)

Puedes desplegar **web y API** en un solo proyecto de Vercel. La API en Go corre como función serverless; la base de datos puede ser **Neon** (Postgres gratis, sin tarjeta).

**Guía paso a paso:** [docs/DEPLOY_VERCEL.md](docs/DEPLOY_VERCEL.md)

### 1. Base de datos en Neon

1. Entra en [neon.tech](https://neon.tech) y crea una cuenta (GitHub).
2. Crea un proyecto y anota la **connection string** (Postgres).
3. Ejecuta las migraciones una vez (desde tu máquina con esa URL):
   ```bash
   cd apps/api && DATABASE_URL="postgresql://..." go run cmd/migrate/main.go
   ```

### 2. Desplegar en Vercel

1. [vercel.com/new](https://vercel.com/new) → Import el repo **daily-market-brief**.
2. **Root Directory:** `apps/web`.
3. **Variables de entorno:**
   - **`DATABASE_URL`** = connection string de Neon (con `?sslmode=require` si hace falta).
   - **`NEXT_PUBLIC_API_URL`** = `https://tu-dominio.vercel.app/api/v1` (la URL de tu deploy en Vercel + `/api/v1`).
   - **`NEWS_SOURCES_JSON`** (opcional) = contenido completo del JSON de `config/news_sources.json` (para el panel admin en serverless; si no, admin dará "no config").
4. Deploy. Vercel construye Next.js y la función Go en `api/`; las peticiones a `/api/v1/*` van a la API.

### 3. Resumen

- **Web:** calendario, día, semana, mes, traducción, etc.
- **API:** resúmenes, health, admin (solo lectura si usas `NEWS_SOURCES_JSON`).
- **Descarga .txt** por día no está disponible en serverless (no hay disco); el resto funciona con los datos en Neon.

---

## Deploy solo web en Vercel (API aparte)

Solo la **app web** (Next.js) en Vercel. La **API en Go** desplegada aparte (Railway, Render, etc.) y su URL en `NEXT_PUBLIC_API_URL`.

### Pasos

1. **Importar el repo en Vercel**  
   [vercel.com/new](https://vercel.com/new) → Import Git Repository → elige `daily-market-brief`.

2. **Configurar el proyecto**
   - **Root Directory:** haz clic en *Edit* y selecciona **`apps/web`** (monorepo).
   - **Framework Preset:** Next.js (se detecta solo).
   - **Build Command:** `npm run build`
   - **Output Directory:** (vacío; Next.js por defecto)

3. **Variables de entorno**
   - **`NEXT_PUBLIC_API_URL`** = URL pública de tu API (ej. `https://tu-api.railway.app` o `https://daily-market-brief-api.fly.dev`).  
   Si la API va detrás del mismo dominio que la web, puede ser `https://tudominio.com/api/v1` (y configurar rewrites en Vercel).

4. **Deploy**  
   Pulsa *Deploy*. Vercel construye `apps/web` y publica la app.

La API (Go + Postgres) se despliega por separado. Opción **gratuita**: Render (ver abajo).

---

## Desplegar la API en Render (gratis)

[Render](https://render.com) tiene plan gratuito para Web Service y Postgres (con límites). No hace falta tarjeta para empezar.

### Pasos

1. **Entra en Render**  
   [dashboard.render.com](https://dashboard.render.com) → Sign up / Login (con GitHub).

2. **Nuevo Blueprint**  
   - **New** → **Blueprint**.  
   - Conecta el repo **daily-market-brief** (GitHub).  
   - Render detectará el `render.yaml` en la raíz del repo.

3. **Aplicar el Blueprint**  
   - Se crearán: 1 base de datos PostgreSQL (free) y 1 Web Service (API en Go).  
   - Revisa que el servicio use el Dockerfile de `apps/api` y que `DATABASE_URL` venga de la base de datos.  
   - Pulsa **Apply**.

4. **Migraciones**  
   La primera vez hay que crear las tablas. En el servicio de la API, **Shell** (o **Settings** → comando de inicio temporal), ejecuta:
   ```bash
   /app/migrate
   ```
   (El Dockerfile ya incluye el binario `migrate`.)

5. **URL de la API**  
   En el dashboard, el Web Service tendrá una URL tipo `https://daily-market-brief-api.onrender.com`. Cópiala.

6. **Vercel (web)**  
   En el proyecto de Vercel de la app web, añade la variable de entorno:
   - **`NEXT_PUBLIC_API_URL`** = `https://daily-market-brief-api.onrender.com`  
   y vuelve a desplegar.

**Nota:** En plan free, el Web Service se **duerme** tras ~15 min sin peticiones; la primera petición tras eso puede tardar unos segundos (cold start). Los summaries (archivos `.txt`) no persisten entre despliegues; para tener datos puedes ejecutar `ingest` y `summarize` en local y subir los archivos, o añadir un cron job en Render que los genere.

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

- **"database: ..."** when starting the API → Check that Postgres is running and your database is configured (see `.env.example`).
- **"address already in use"** → Port 3090 is taken. Use `PORT=3091 ./scripts/run-api.sh` and for the web: `NEXT_PUBLIC_API_URL=http://localhost:3091 npm run dev`.
- **Web shows no data** → Run `make ingest` then `make summarize` (or `make summarize DAY=YYYY-MM-DD`).
