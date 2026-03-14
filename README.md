# Daily Market Brief

Brief diario de mercados: ingesta de noticias financieras, ranking por impacto y resúmenes (top 10 + resto). Web con calendario y vistas por día/semana/mes.

---

## Cómo correr la app (local, sin Docker)

Necesitás **PostgreSQL** corriendo, **Go** y **Node** instalados.

### 1. Base de datos (una sola vez)

Creá usuario y base en Postgres:

```sql
CREATE USER marketbrief WITH PASSWORD 'marketbrief_secret';
CREATE DATABASE marketbrief OWNER marketbrief;
```

Ejecutá las migraciones desde la raíz del proyecto:

```bash
make migrate
```

### 2. Arrancar la API

En una terminal, desde la **raíz del proyecto**:

```bash
./scripts/run-api.sh
```

Dejá la terminal abierta. Cuando veas `listening on http://localhost:3090`, la API está lista.

- **API:** http://localhost:3090  
- **Health:** http://localhost:3090/api/health  

### 3. Arrancar la web

En **otra terminal**, desde la raíz del proyecto:

```bash
cd apps/web
npm install   # solo la primera vez
npm run dev
```

Abrí en el navegador: **http://localhost:3000** (calendario y vistas por día/semana/mes).

### 4. (Opcional) Tener datos en el calendario

Para ver resúmenes en la web, hace falta ingestar noticias y generar al menos un día:

```bash
# Desde la raíz del proyecto
make ingest      # trae noticias (RSS sin API keys; NewsAPI/Finnhub opcionales)
make summarize   # genera resumen del día (hoy UTC)
# O un día concreto: make summarize DAY=2026-03-14
```

---

## Resumen de URLs

| Servicio | URL |
|----------|-----|
| **Web (calendario)** | http://localhost:3000 |
| **API** | http://localhost:3090 |
| **API health** | http://localhost:3090/api/health |

La web usa la API en el puerto **3090** por defecto. Si usás otro puerto para la API, arrancá la web con:

```bash
NEXT_PUBLIC_API_URL=http://localhost:PUERTO npm run dev
```

---

## Correr con Docker

Si preferís todo en contenedores:

```bash
mkdir -p summaries
docker compose up -d --build
docker compose exec api /app/migrate
# Opcional: docker compose exec api /app/ingest
# Opcional: docker compose exec api /app/summarize
```

- **Web:** http://localhost:3000  
- **API:** http://localhost:3090  

---

## Requisitos

- **Local:** Go 1.22+, Node 20+, PostgreSQL 16+
- **Docker:** Docker y Docker Compose

---

## Estructura del proyecto

| Parte | Descripción |
|-------|-------------|
| `apps/api` | API en Go (Fiber): resúmenes, noticias, health. Comandos: server, migrate, ingest, summarize. |
| `apps/web` | Front en Next.js: calendario, día, semana, mes, last-10. |
| `config/` | `news_sources.json`, `ranking_weights.json`. |
| `summaries/` | Archivos .txt generados por `summarize`. |
| `scripts/run-api.sh` | Script para arrancar la API en el puerto 3090. |

---

## Makefile (desde la raíz)

| Comando | Descripción |
|---------|-------------|
| `make migrate` | Ejecutar migraciones (Postgres local). |
| `make ingest` | Ingesta de noticias. |
| `make summarize` | Resumen del día; `make summarize DAY=YYYY-MM-DD` para una fecha. |
| `make run-api` | Arrancar la API (Go). |
| `make run-web` | Arrancar la web (`npm run dev`). |
| `make up` / `make down` | Docker compose up/down. |

---

## Analista de inversiones (AI)

El módulo **investment analyst** aplica un marco de 10 pasos a cada noticia: filtro de relevancia, categoría, resumen, impacto de mercado, activos afectados, sesgo direccional, señales de inversión, horizonte temporal y fuerza de la señal. La salida es JSON estructurado.

- **GET /api/analyst/prompt** — Devuelve el system prompt y las instrucciones para usar con un LLM (OpenAI, Anthropic, etc.).
- **POST /api/analyze** — Analiza una noticia (body: `{"title","url","source","summary"}`). Devuelve el JSON del esquema.
- **GET /api/analysis/day/:day** — Analiza todas las noticias del día y devuelve un array de análisis.

- **Con OpenAI:** Definí `OPENAI_API_KEY` (y opcionalmente `OPENAI_MODEL`, por defecto `gpt-4o-mini`). Al arrancar la API se usará el analista por LLM. Sin key se usa el stub.

---

## Problemas frecuentes

- **"database: ..."** al arrancar la API → Postgres no está corriendo o usuario/contraseña incorrectos.
- **"address already in use"** → El puerto 3090 está ocupado. Usá `PORT=3091 ./scripts/run-api.sh` y en la web `NEXT_PUBLIC_API_URL=http://localhost:3091 npm run dev`.
- **La web no muestra datos** → Ejecutá `make ingest` y después `make summarize` (o `make summarize DAY=YYYY-MM-DD`).
