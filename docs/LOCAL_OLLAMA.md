# Investment analyst con Ollama local (gratis, sin API key)

En vez de pagar OpenAI, la API puede usar un modelo Llama corriendo con
[Ollama](https://ollama.com) en tu propia red — no requiere API key ni tiene
costo por token.

## 0. Qué necesitás

- Ollama instalado en alguna Mac/PC de tu red (macOS: la app `Ollama.app`).
- Un modelo descargado, por ejemplo:
  ```bash
  ollama pull llama3.2:3b
  ```
  (`ollama list` muestra los modelos que ya tenés).

## 1. Exponer Ollama en la red (si el modelo corre en OTRA máquina que la API)

Por defecto Ollama escucha solo en `127.0.0.1:11434` — no es alcanzable desde
otra PC. Para exponerlo en la LAN, en la máquina que corre Ollama:

```bash
# 1) Cerrá la app Ollama (ícono de la barra de menú → Quit)
# 2) Seteá la variable de entorno a nivel de sesión gráfica:
launchctl setenv OLLAMA_HOST "0.0.0.0:11434"
# 3) Volvé a abrir Ollama.app (o `ollama serve` desde terminal si no usás la app)
```

Si tu Mac tiene el firewall activo (Configuración → Red → Firewall), asegurate
de permitir conexiones entrantes para "Ollama".

Confirmá que escucha en todas las interfaces (no solo en `127.0.0.1`):

```bash
lsof -iTCP -sTCP:LISTEN -P -n | grep 11434
# esperado: algo como *:11434 (LISTEN), no 127.0.0.1:11434
```

Anotá la IP local de esa máquina (para usarla desde la API):

```bash
ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1 2>/dev/null
```

## 2. Probar que se puede llegar desde afuera

Desde la máquina donde vas a correr la API (puede ser la misma u otra):

```bash
curl http://<IP-DE-OLLAMA>:11434/api/tags
```

Debería devolver un JSON con la lista de modelos. Si da timeout/refused,
revisá el paso 1 (host bind) y el firewall.

## 3. Configurar la API para usar Ollama

En `.env` (raíz del repo, o exportado en la terminal donde corrés `run-api.sh`
/ `run-api-neon.sh`):

```bash
LLM_PROVIDER=ollama
OLLAMA_BASE_URL=http://<IP-DE-OLLAMA>:11434
OLLAMA_MODEL=llama3.2:3b
```

Si Ollama corre en la MISMA máquina que la API, `OLLAMA_BASE_URL` puede
quedar sin definir (usa `http://localhost:11434` por defecto).

Arrancá (o reiniciá) la API. En el log debería aparecer:

```
analyst: using Ollama (LLM_PROVIDER=ollama)
```

## 4. Probar el análisis

```bash
curl -X POST http://localhost:3090/api/analyze \
  -H "Content-Type: application/json" \
  -d '{"title":"Fed holds rates steady, signals cuts ahead","url":"https://example.com","source":"Reuters"}'
```

Debería devolver el JSON estructurado (relevance, category, impact_level, etc.)
generado por el modelo local.

## Notas

- `llama3.2:3b` es un modelo chico (2 GB) — es rápido pero menos preciso que
  GPT-4o-mini siguiendo instrucciones complejas o devolviendo JSON perfecto.
  Si ves errores de "parse json", probá con un modelo más grande
  (`ollama pull llama3.1:8b`) y ajustá `OLLAMA_MODEL`.
- No hay costo por token ni límite de uso — podés analizar todas las noticias
  del día sin preocuparte por la factura de OpenAI.
- `OLLAMA_API_KEY` es opcional: solo hace falta si pusiste un reverse proxy
  con autenticación delante de Ollama.
