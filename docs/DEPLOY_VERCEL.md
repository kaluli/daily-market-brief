# Guía paso a paso: desplegar todo en Vercel (web + API)

Todo en un solo proyecto de Vercel, sin tarjeta. Base de datos en Neon (Postgres gratis).

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
   - Donde pone “Root Directory” pulsa **Edit**.
   - Escribe: `apps/web`
   - Confirma (o elige la carpeta `apps/web` en el selector).
2. No cambies **Framework Preset** (debe detectar Next.js).
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
   - Root Directory = `apps/web`
   - `DATABASE_URL` bien pegada (sin espacios al inicio/final).
4. Cuando termine, verás algo como: **Visit** o una URL tipo `https://daily-market-brief-xxx.vercel.app`. **Cópiala**.

---

## Paso 7: Poner la URL real de la API

1. En Vercel, entra en tu proyecto → **Settings** → **Environment Variables**.
2. Busca **NEXT_PUBLIC_API_URL** y pulsa **Edit** (o los tres puntos → Edit).
3. Cambia el valor a (sustituye por tu URL real):
   ```text
   https://TU-URL-DEL-PASO-6.vercel.app/api/v1
   ```
   Ejemplo: si tu URL es `https://daily-market-brief-abc123.vercel.app`, el valor será:
   ```text
   https://daily-market-brief-abc123.vercel.app/api/v1
   ```
4. Guarda.
5. Ve a **Deployments** → en el último deployment pulsa los tres puntos → **Redeploy** (o haz un nuevo commit y deja que redepliegue solo). Así la web usará la API correcta.

---

## Paso 8: Comprobar que todo funciona

1. Abre la URL de tu proyecto (la del Paso 6).
2. Deberías ver el **calendario** (puede estar vacío si aún no hay datos).
3. Prueba el health de la API:
   - Abre en el navegador: `https://TU-URL.vercel.app/api/v1/api/health`
   - Deberías ver algo como: `{"status":"ok","service":"market-brief-api"}`.

**Nota técnica:** En Vercel, las peticiones a `/api/v1/*` las atiende un Route Handler de Next.js (`app/api/v1/[[...path]]/route.ts`) que hace proxy a la función Go en `/api` enviando el path original en el header `X-Forwarded-Path`. Así la API Go recibe bien la ruta aunque el rewrite de Vercel no la conserve.

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
| 4 | Root Directory = `apps/web` |
| 5 | Añadir `DATABASE_URL`, `NEXT_PUBLIC_API_URL` (temporal) y opcionalmente `NEWS_SOURCES_JSON` |
| 6 | Deploy y copiar la URL del deploy |
| 7 | Cambiar `NEXT_PUBLIC_API_URL` a `https://TU-URL.vercel.app/api/v1` y redeploy |
| 8 | Probar la web y `/api/v1/api/health` |
| 9 | Ejecutar ingest + summarize en local con `DATABASE_URL` de Neon para tener datos |

Si algo falla en un paso concreto, indica en cuál y el mensaje de error y lo afinamos.
