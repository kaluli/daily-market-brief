# Desplegar en Vercel

**Importante:** Usar la cuenta de GitHub **kaluli** para Vercel (no usar karinap-eb).

En este proyecto en Vercel solo se despliega la **web** (Next.js). La API la desplegás aparte y la conectás con `NEXT_PUBLIC_API_URL`.

---

## 1. Probar la app sin API (recomendado primero)

Desplegás solo la web. El calendario carga bien y se ve vacío (sin datos); no hace falta base de datos ni API.

1. [vercel.com](https://vercel.com) → iniciar sesión con GitHub **kaluli** → **Add New** → **Project** → importá el repo.
2. **Root Directory**: **`apps/web`**.
3. **Environment Variables**: no agregues nada (dejá vacío).
4. **Deploy**.

Listo. Abrís la URL que te da Vercel: verás el calendario y las vistas (día/semana/mes) vacías, sin errores.

---

## 2. Conectar la API (desplegada aparte)

Si en cambio desplegás la API en otro servicio (Railway, Fly, etc.):

1. En Vercel → **Settings** → **Environment Variables**.
2. **NEXT_PUBLIC_API_URL** = URL base de tu API, ej. `https://tu-api.railway.app/api/v1` (la web llama a esta URL desde el navegador).
3. Opcional: **API_BACKEND_URL** = misma URL, para que las rutas del servidor (ej. resumen del día con traducción) también llamen a tu API. Si no la ponés, se usa `NEXT_PUBLIC_API_URL`.
4. **Redeploy.**

La web usará esa API y mostrará los datos en el calendario.

**Si ves "Failed to fetch" o "Failed to load" al abrir un día:** la API no está configurada o no es alcanzable. Revisá que `NEXT_PUBLIC_API_URL` esté bien y que la API esté en línea y permita CORS desde tu dominio Vercel.

---

## 3. Neon (para la API)

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
| **Sin API** | Sin variables → calendario vacío |
| **Con API** | NEXT_PUBLIC_API_URL = URL de tu API + `/api/v1` |
