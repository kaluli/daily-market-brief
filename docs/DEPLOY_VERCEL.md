# Guía paso a paso: desplegar todo en Vercel (web + API)

Todo en un solo proyecto de Vercel, sin tarjeta. Base de datos en Neon (Postgres gratis).

---

## Dónde está la API y checklist (API + Neon)

### Dónde está la API

| Qué | Dónde |
|-----|--------|
| **API Go (lógica, DB, rutas)** | `apps/api/` — `internal/api/`, `internal/db/`, `pkg/vercel/` |
| **Entrada serverless en Vercel** | Raíz del repo: `api/index.go` y `api/health.go` + `go.mod` (usa `replace ... => ./apps/api`) |
| **Proxy Next.js** | `apps/web/app/api/v1/[[...path]]/route.ts` — recibe `/api/v1/*` y hace fetch a `/api` con header `X-Forwarded-Path` |
| **Frontend llama a la API** | `apps/web/lib/api.ts` — usa `NEXT_PUBLIC_API_URL` (ej. `https://tu-app.vercel.app/api/v1`) y arma rutas como `/api/health`, `/api/summaries/day/:day` |

Rutas de la API (Fiber, en `apps/api/internal/api/server.go`): `/`, `/api/health`, `/api/summaries`, `/api/summaries/day/:day`, `/api/news/day/:day`, `/api/admin/sources`, etc.

### Checklist: API bien configurada

- [ ] **Vercel → Environment Variables**
  - `DATABASE_URL` = connection string de Neon (ej. `postgresql://user:pass@ep-xxx.neon.tech/neondb?sslmode=require`).
  - `NEXT_PUBLIC_API_URL` = `https://TU-DOMINIO.vercel.app/api/v1` (sin barra final).
- [ ] **Vercel → General**
  - **Root Directory** = vacío (raíz del repo), para que se construyan la web y las funciones Go en `api/`.
- [ ] **Build**
  - En el log del deploy aparece el paso `go build ... ./api/index.go` sin error.
  - En **Functions** del deployment aparece una función para la ruta `/api`.

### Checklist: Neon bien configurado

- [ ] **Neon**
  - Proyecto creado en [neon.tech](https://neon.tech); connection string copiada (con `?sslmode=require` si hace falta).
- [ ] **Migraciones**
  - Ejecutadas una vez desde local:  
    `cd apps/api && DATABASE_URL="TU_CONNECTION_STRING" go run cmd/migrate/main.go`  
    Salida esperada: `migrations ok`.
- [ ] **Vercel**
  - La misma connection string está en **Environment Variables** como `DATABASE_URL` (Production y/o Preview).

Si la API devuelve 503 con `"api not configured (DATABASE_URL)"`, la función Go no tiene `DATABASE_URL` en el entorno (revisa variables en Vercel y redeploy).

Si el build falla con **`go mod tidy`**: hay un `go.mod` y `go.sum` dentro de `api/` (replace a `../apps/api`) para que la build de Go use ese directorio. Además en `vercel.json` está `GOFLAGS=-mod=readonly`. Si sigue fallando, en Vercel → Settings → Environment Variables añade `GOFLAGS` = `-mod=readonly` (Production y Preview).

**Si el build tarda mucho o falla (ej. "No Output Directory named public"):** podés volver a una configuración más simple: en Vercel → Settings → General pon **Root Directory** = **`apps/web`**. Así el build es solo Next.js (rápido) y la web despliega bien. La API en `/api/v1` no estará disponible en ese deploy (404); para tener web + API en el mismo proyecto hace falta Root vacío y que la build de Go en `api/` funcione. En `vercel.json` de la raíz está `outputDirectory: ".next"` para que Vercel encuentre el output cuando Root es la raíz del repo.

---

## Paso 1: Crear la base de datos en Neon

1. Abre **[neon.tech](https://neon.tech)** en el navegador.
2. Pulsa **Sign up** e inicia sesión con **GitHub**.
3. Una vez dentro, pulsa **New Project**.
4. Elige un nombre (ej. `daily-market-brief`) y la región que prefieras. Pulsa **Create project**.
5. En la pantalla del proyecto verás **Connection string**. Cópiala (empieza por `postgresql://...`).
   - Si tiene el formato “Pooled connection”, úsala.
   - Si Neon te pide SSL, añade al final: `?sslmode=require`
   - Ejemplo: `postgresql://user:pass@ep-xxx.region.aws.neon.tech/neondb?sslmode=require`
6. **Guarda esa URL** en un sitio seguro; la usarás en los siguientes pasos.

---

## Paso 2: Crear las tablas en Neon (migraciones)

Solo se hace **una vez**, desde tu ordenador.

1. Abre una terminal y ve a la raíz del proyecto:
   ```bash
   cd /ruta/a/daily-market-brief
   ```
2. Sustituye `TU_CONNECTION_STRING` por la URL que copiaste de Neon y ejecuta:
   ```bash
   cd apps/api && DATABASE_URL="TU_CONNECTION_STRING" go run cmd/migrate/main.go
   ```
3. Si todo va bien, no verás errores. Las tablas ya están creadas en Neon.

---

## Paso 3: Crear el proyecto en Vercel

1. Abre **[vercel.com](https://vercel.com)** e inicia sesión (con GitHub si puedes).
2. Pulsa **Add New…** → **Project**.
3. En “Import Git Repository” elige el repo **daily-market-brief** (o el nombre que tenga en tu GitHub).
   - Si no lo ves, pulsa **Configure** en GitHub y autoriza a Vercel para ese repo.
4. Pulsa **Import** en el repo.

---

## Paso 4: Configurar el proyecto en Vercel

1. **Root Directory**
   - Pon **Root Directory** = **`apps/web`** (recomendado).
   - Así el build es solo Next.js, más rápido y estable. La **API no se despliega** en este proyecto; la web funcionará y podés desplegar la API por otro medio (otro proyecto Vercel, Railway, etc.).
2. **Framework Preset**: debe detectar Next.js.
3. **No** pulses aún **Deploy**. Primero añadimos las variables.

---

## Paso 5: Añadir variables de entorno en Vercel

1. En la misma pantalla del proyecto, busca la sección **Environment Variables**.
2. Añade **una por una** estas variables (Name y Value):

   | Name | Value |
   |------|--------|
   | `DATABASE_URL` | La connection string de Neon (la del Paso 1). |
   | `NEXT_PUBLIC_API_URL` | Deja este valor para después del primer deploy (ver Paso 7). |

3. **DATABASE_URL**
   - Name: `DATABASE_URL`
   - Value: pega la URL de Neon (ej. `postgresql://user:pass@ep-xxx.neon.tech/neondb?sslmode=require`).
   - Environment: marca **Production** (y si quieres también Preview).
   - Pulsa **Save**.

4. **NEXT_PUBLIC_API_URL** (primera vez)
   - Name: `NEXT_PUBLIC_API_URL`
   - Value: escribe algo temporal, por ejemplo: `https://placeholder.vercel.app/api/v1`
   - Lo corregiremos en el Paso 7 con la URL real de tu deploy.
   - Pulsa **Save**.

5. **(Opcional) NEWS_SOURCES_JSON**
   - Si quieres que el panel Admin muestre las fuentes de noticias en serverless:
   - Abre en tu proyecto el archivo `config/news_sources.json`.
   - Copia **todo** el contenido (desde `{` hasta `}`).
   - En Vercel, nueva variable:
     - Name: `NEWS_SOURCES_JSON`
     - Value: pega el JSON completo (una sola línea o con saltos, da igual).
   - Pulsa **Save**.

---

## Paso 6: Primer deploy

1. Pulsa **Deploy**.
2. Espera a que el build termine (puede tardar 1–2 minutos).
3. Si algo falla, revisa:
   - Root Directory = **vacío** (raíz del repo).
   - `DATABASE_URL` bien pegada (sin espacios al inicio/final).
4. Cuando termine, verás algo como: **Visit** o una URL tipo `https://daily-market-brief-xxx.vercel.app`. **Cópiala**.

---

## Paso 7: Poner la URL de la API (cuando la tengas desplegada)

Cuando despliegues la API en otro sitio (Railway, otro proyecto Vercel, etc.):

1. En Vercel → **Settings** → **Environment Variables**.
2. **NEXT_PUBLIC_API_URL** = URL base de tu API + `/api/v1`, por ejemplo:
   - Si la API está en Railway: `https://tu-app.railway.app/api/v1`
   - Si más adelante la API está en el mismo Vercel: `https://TU-DOMINIO.vercel.app/api/v1`
3. Guarda y **Redeploy** para que la web use esa API.

---

## Paso 8: Comprobar que todo funciona

1. Abre la URL de tu proyecto (la del Paso 6).
2. Deberías ver el **calendario** (puede estar vacío si aún no hay datos).
3. Prueba el health de la API:
   - **Opción A (función Go mínima):** `https://TU-URL.vercel.app/api/health` → `{"status":"ok","service":"market-brief-api"}`.
   - **Opción B (API completa vía proxy):** `https://TU-URL.vercel.app/api/v1/api/health` → mismo JSON.
   - Con **Root = apps/web** la API no está en este deploy: `/api/v1/*` dará 404. Es esperado; desplegá la API por otro medio y configurá `NEXT_PUBLIC_API_URL` con esa URL.

**Nota:** Con Root = apps/web, en este proyecto solo se despliega la web. La API la desplegás aparte; luego configurás `NEXT_PUBLIC_API_URL` para que la web la consuma.

---

## Paso 9: Tener datos (resúmenes) en la app

La API en Vercel usa la base de datos de Neon, pero **ingest** y **summarize** se ejecutan en tu máquina (o en un cron externo). Una vez hechas las migraciones (Paso 2), puedes generar datos así:

1. En tu ordenador, en la raíz del proyecto:
   ```bash
   cd /ruta/a/daily-market-brief
   ```
2. Ingesta de noticias (usa tu connection string de Neon):
   ```bash
   cd apps/api && DATABASE_URL="TU_CONNECTION_STRING" CONFIG_DIR=../../config go run cmd/ingest/main.go
   ```
3. Generar resumen del día:
   ```bash
   cd apps/api && DATABASE_URL="TU_CONNECTION_STRING" SUMMARIES_PATH=../../summaries go run cmd/summarize/main.go -day=2026-03-14
   ```
   (Cambia la fecha por la que quieras.)
4. Los metadatos y el TOP 10 / other 90 se guardan en **Neon**. Al recargar la app en Vercel deberías ver ese día en el calendario y en la vista por día.

**Nota:** En serverless no hay “Descargar .txt” por día (no hay disco). El calendario, la vista por día y las traducciones sí funcionan con los datos de Neon.

---

## Resumen rápido

| Paso | Qué haces |
|------|-----------|
| 1 | Crear proyecto en Neon y copiar connection string |
| 2 | Ejecutar migraciones con esa URL (una vez) |
| 3 | Crear proyecto en Vercel e importar el repo |
| 4 | Root Directory = **`apps/web`** (solo web; API aparte) |
| 5 | Añadir `DATABASE_URL`, `NEXT_PUBLIC_API_URL` (temporal) y opcionalmente `NEWS_SOURCES_JSON` |
| 6 | Deploy y copiar la URL del deploy |
| 7 | Cambiar `NEXT_PUBLIC_API_URL` a `https://TU-URL.vercel.app/api/v1` y redeploy |
| 8 | Probar la web y `/api/v1/api/health` |
| 9 | Ejecutar ingest + summarize en local con `DATABASE_URL` de Neon para tener datos |

Si algo falla en un paso concreto, indica en cuál y el mensaje de error y lo afinamos.
