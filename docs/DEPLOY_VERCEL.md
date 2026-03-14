# Desplegar la web en Vercel

Solo la **web** (Next.js). La API la desplegás aparte cuando quieras.

---

## 1. Probar la app sin API (recomendado primero)

Desplegás solo la web. El calendario carga bien y se ve vacío (sin datos); no hace falta base de datos ni API.

1. [vercel.com](https://vercel.com) → **Add New** → **Project** → importá el repo.
2. **Root Directory**: **`apps/web`**.
3. **Environment Variables**: no agregues nada (dejá vacío).
4. **Deploy**.

Listo. Abrís la URL que te da Vercel: verás el calendario y las vistas (día/semana/mes) vacías, sin errores.

---

## 2. Conectar la API (después)

Cuando tengas la API desplegada en algún lado:

1. En Vercel → tu proyecto → **Settings** → **Environment Variables**.
2. Agregá **NEXT_PUBLIC_API_URL** = URL de tu API + `/api/v1` (ej. `https://tu-api.ejemplo.com/api/v1`).
3. **Redeploy** (Deployments → ⋮ en el último → Redeploy).

La web usará esa API y mostrará los datos en el calendario.

---

## 3. Neon (solo si tu API usa Neon)

Si la API usará Neon como base de datos:

1. [neon.tech](https://neon.tech) → Sign up con GitHub → **New Project**.
2. Copiá la **Connection string**.
3. En local (una vez):  
   `cd apps/api && DATABASE_URL="TU_CONNECTION_STRING" go run cmd/migrate/main.go`  
   Debe salir `migrations ok`.

---

## 4. Datos en el calendario

Los resúmenes se generan con la API (ingest + summarize en local o cron), no en Vercel. Cuando la API esté en línea y `NEXT_PUBLIC_API_URL` esté configurado en Vercel, al recargar la app verás los datos.

---

## Resumen

| Qué | Dónde |
|-----|--------|
| **Web** | Vercel, Root = `apps/web` |
| **Sin API** | Deploy sin variables → calendario vacío, sin errores |
| **Con API** | En Vercel: `NEXT_PUBLIC_API_URL` = URL de tu API + `/api/v1` |
