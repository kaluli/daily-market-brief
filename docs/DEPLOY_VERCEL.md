# Desplegar en Vercel

**Importante:** Usar la cuenta de GitHub **kaluli** para Vercel (no usar karinap-eb).

Tenés **dos opciones**: desplegar solo la **web** (Next.js) en Vercel y conectar una API en otro lado, o desplegar **web y API** en Vercel (dos proyectos).

---

## 1. Probar la app sin API (recomendado primero)

Desplegás solo la web. El calendario carga bien y se ve vacío (sin datos); no hace falta base de datos ni API.

1. [vercel.com](https://vercel.com) → iniciar sesión con GitHub **kaluli** → **Add New** → **Project** → importá el repo.
2. **Root Directory**: **`apps/web`**.
3. **Environment Variables**: no agregues nada (dejá vacío).
4. **Deploy**.

Listo. Abrís la URL que te da Vercel: verás el calendario y las vistas (día/semana/mes) vacías, sin errores.

---

## 2. Desplegar la API también en Vercel (segundo proyecto)

Podés tener la API (Go) en un **segundo proyecto** de Vercel. Así todo queda en Vercel.

1. En [vercel.com](https://vercel.com) → **Add New** → **Project** → importá el **mismo repo**.
2. **Root Directory**: **`apps/api`** (no apps/web).
3. **Environment Variables** (obligatorias para la API):
   - **DATABASE_URL** = connection string de Postgres (ej. [Neon](https://neon.tech): crear proyecto, copiar connection string).
   - **NEWS_SOURCES_JSON** = contenido completo del archivo `config/news_sources.json` como texto (una línea, sin saltos). Así la página Admin puede listar fuentes. Si no lo ponés, `/api/admin/sources` devuelve error.
4. Opcional: **OPENAI_API_KEY** si querés el analista con LLM.
5. **Deploy.**

Cuando termine, Vercel te da una URL tipo `https://daily-market-brief-api-xxx.vercel.app`. Esa es la URL de tu API.

6. En el **proyecto de la web** (Root = `apps/web`) → **Settings** → **Environment Variables**:
   - **NEXT_PUBLIC_API_URL** = URL del proyecto API, ej. `https://daily-market-brief-api.vercel.app` (sin `/api` ni trailing slash).
   - **API_BACKEND_URL** = la **misma** URL. La ruta `/api/day/[day]` usa esta variable para pedir el resumen a la API; si falta o apunta a la web, no se ve la información del día.
7. **Redeploy** del proyecto web (las variables se aplican en el próximo deploy).

**Migraciones:** Antes del primer deploy de la API, ejecutá las migraciones una vez con la misma `DATABASE_URL` (en local):  
`cd apps/api && DATABASE_URL="postgres://..." go run cmd/migrate/main.go` → debe salir `migrations ok`.

---

## 3. Conectar la API desplegada en otro servicio (Railway, Fly, etc.)

Si en cambio desplegás la API en otro servicio:

1. En Vercel → **Settings** → **Environment Variables**.
2. **NEXT_PUBLIC_API_URL** = URL de tu API (origen), ej. `https://tu-api.railway.app`. Si ponés `https://tu-api.railway.app/api/v1`, la app quita `/api/v1` y construye bien las rutas (`/api/health`, `/api/summaries/day/...`).
3. **API_BACKEND_URL** = la misma URL (para que el servidor también llame a tu API).
4. **Redeploy.**

**Importante:** Ambas variables tienen que ser la URL de tu API (Railway, Fly, etc.), **no** la URL de la app en Vercel. Si ponés la URL de Vercel se produce un "infinite loop".

La web usará esa API y mostrará los datos en el calendario.

**Si ves "Failed to fetch" o "Failed to load" al abrir un día:** la API no está configurada o no es alcanzable. Revisá que `NEXT_PUBLIC_API_URL` esté bien y que la API esté en línea y permita CORS desde tu dominio Vercel.

---

## 4. Neon (para la API en producción)

1. [neon.tech](https://neon.tech) → **Connection string** del proyecto (formato `postgresql://...?sslmode=require`).
2. **En el proyecto de la API en Vercel** → **Settings** → **Environment Variables** → agregá:
   - **DATABASE_URL** = tu connection string de Neon (pegá el valor completo).
3. **Migraciones (una sola vez):** En local, con esa misma URL:
   ```bash
   cd apps/api && DATABASE_URL="postgresql://..." go run cmd/migrate/main.go
   ```
   Debe salir `migrations ok`. Así la API en Vercel encuentra las tablas al conectarse a Neon.

---

## 5. Información del día (resúmenes en el calendario)

La API en Vercel **sirve** los datos que ya están en la base de datos; **no genera** resúmenes sola. Para que aparezca la "información del día" (resumen con top 10, etc.) tenés que **generar** esos resúmenes y guardarlos en la misma DB que usa la API.

**Por qué en local tenés noticias y en producción no:** En local la API usa una base (ej. Postgres en tu máquina) donde ya corriste `make ingest` y `make summarize`. En producción la API usa **otra** base (Neon); si nunca corriste ingest + summarize con la `DATABASE_URL` de Neon, esa base no tiene resúmenes. Hay que correr ingest y summarize apuntando a Neon (ver abajo).

**Una vez por día (o cuando quieras actualizar), en tu máquina**, con la **misma DATABASE_URL** que tiene el proyecto de la API en Vercel (ej. la connection string de Neon):

```bash
# Desde la raíz del repo
export DATABASE_URL="postgres://...tu-connection-string-de-neon..."

# 1. Ingest: trae noticias y las guarda en la DB
make ingest
# o: cd apps/api && go run cmd/ingest/main.go

# 2. Summarize: arma el resumen del día y lo guarda en la DB
make summarize
# o para una fecha: make summarize DAY=2026-03-14
# o: cd apps/api && go run cmd/summarize/main.go -day 2026-03-14
```

Después de eso, al recargar la web (con `NEXT_PUBLIC_API_URL` apuntando a tu API) verás el resumen de ese día en el calendario.

**Resumen:** Las fuentes y la lista de noticias salen de la API/DB. La **información del día** (resumen top 10) solo aparece después de correr **ingest** + **summarize** con la misma `DATABASE_URL` (en local o con un cron).

---

## 6. Automatizar ingest + summarize (GitHub Actions)

Para no tener que ejecutar `make ingest` y `make summarize` a mano cada día, el repo incluye un workflow que corre **una vez al día** (y también se puede lanzar a mano).

1. **Secret en GitHub:** En el repo → **Settings** → **Secrets and variables** → **Actions** → **New repository secret**:
   - **Nombre:** `DATABASE_URL`
   - **Valor:** la misma connection string de Neon que usa la API en Vercel (ej. `postgresql://usuario:password@host.neon.tech/neondb?sslmode=require`).

2. **Horario:** Por defecto el workflow corre a las **13:00 UTC** (9:00 en Argentina). Podés cambiarlo en `.github/workflows/daily-ingest.yml` (cron `0 13 * * *`).

3. **Ejecución manual:** En el repo → pestaña **Actions** → **Daily ingest and summarize** → **Run workflow**.

Las fuentes de noticias se leen del `config/` del repo (igual que en local); no hace falta configurar nada más. Después de cada ejecución, la web en producción mostrará el resumen del día cuando recargues.

---

## Resumen

| Qué | Dónde |
|-----|--------|
| **Web** | Vercel, Root = `apps/web` |
| **API en Vercel** | Segundo proyecto, Root = `apps/api`, DATABASE_URL + NEWS_SOURCES_JSON |
| **Sin API** | Sin variables → calendario vacío |
| **Con API** | NEXT_PUBLIC_API_URL = URL de la API (Vercel o Railway, etc.) |
| **Automatizar resúmenes** | GitHub Actions: secret DATABASE_URL, workflow daily-ingest (sección 6) |
