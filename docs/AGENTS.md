# Agentes inversores (Fase 4)

Tres agentes:

1. **Agresivo ("risky")** — recibe $5,000 ficticios por mes, actúa con señales de
   menor confianza y apuesta una porción más grande del cash disponible por operación.
2. **Conservador ("conservative")** — recibe los mismos $5,000 ficticios por mes,
   solo actúa con señales de alta confianza y apuesta menos por operación.
3. **Feedback / coach** — no opera. Vos hablás con él (vía `/api/agents/feedback`):
   revisa las carteras de los otros dos y te da su devolución.

Todo es dinero ficticio — no hay operaciones reales en ningún broker.

## Cómo deciden los agentes 1 y 2

No usan el LLM para decidir la operación en sí — usan el análisis que ya produce el
**investment analyst** existente (`/api/analyze`, el framework de 10 pasos) y aplican
reglas mecánicas según su perfil de riesgo:

0. Antes de analizar nada: solo se mandan al LLM las `AGENTS_MAX_NEWS_PER_DAY`
   noticias de mayor `impact_score` del día (default: 25 — configurable en `.env`).
   `impact_score` ya se calcula en la ingesta (fuente, palabras clave, tickers,
   duplicados, antigüedad) y `NewsItemsByDay` devuelve las noticias ordenadas por
   ese score. Esto evita analizar decenas de noticias irrelevantes con un modelo
   local lento — un día con 80+ noticias sin este corte puede tardar mucho.
1. Se ignoran noticias que no sean `relevance = "Market Moving"`.
2. Se ignoran si `signal_strength` (1–10) es menor al mínimo del perfil
   (agresivo: 5, conservador: 7).
3. Para cada activo afectado (`affected_assets`) que se pueda mapear al universo fijo
   de tickers (ver abajo), se mira `directional_bias`:
   - **Bullish** → compra, si hay cash y no se superó el máximo de posiciones abiertas
     del perfil. El tamaño es un % fijo del cash disponible (agresivo: 30%,
     conservador: 12%).
   - **Bearish** → vende (cierra) la posición existente en ese ticker, si la hay.
     No se hace short-selling (no se abren posiciones en corto sin tenerlas).

Esto corre una vez por día, para todos los agentes, con `POST /api/agents/run-day`
(ver más abajo) — o `make agents-run-day DAY=YYYY-MM-DD`.

## Universo de activos (ETFs líquidos reales)

Las noticias mencionan activos en texto libre ("US Treasury Yields", "Dollar Index"...).
Para poder operar con precios reales, eso se mapea (por palabras clave, ver
`internal/agents/universe.go`) a uno de estos 10 tickers:

| Ticker | Qué representa |
|---|---|
| SPY | S&P 500 (acciones grandes EE.UU.) |
| QQQ | Nasdaq 100 (tech) |
| IWM | Russell 2000 (small caps) |
| XLE | Sector energía |
| XLF | Sector financiero / bancos |
| TLT | Bonos del Tesoro EE.UU. 20+ años |
| GLD | Oro |
| USO | Petróleo crudo |
| UUP | Índice del dólar |
| BITO | Bitcoin (vía futuros) |

Si una noticia menciona un activo que no matchea ninguna palabra clave, se ignora
(no se fuerza una operación).

## Precios: datos históricos reales

Los precios vienen de datos históricos reales (API de gráficos de Yahoo Finance,
gratis, sin API key), no son sintéticos. Como el entorno de esta sesión de Claude no tiene acceso general a
internet, este paso se corre **en tu propia terminal**:

```bash
./scripts/fetch-prices.sh
```

Esto descarga un CSV por ticker en `data/prices/` (desde el 1 de marzo hasta hoy;
para otro rango: `D1=20260101 D2=20260601 ./scripts/fetch-prices.sh`).

Después, cargalos en Neon:

```bash
./scripts/seed-prices-neon.sh
```

(o manualmente: `cd apps/api && DATABASE_URL=... go run cmd/seed-prices/main.go -dir=../../data/prices`)

Los agentes solo operan un ticker en un día si hay un precio cargado para ese
ticker en esa fecha o antes — si `run-day` no genera operaciones, lo primero a
revisar es si corriste este seed.

## Endpoints

- `GET /api/agents/portfolios` — snapshot de ambas carteras (cash, posiciones
  abiertas valuadas al último precio conocido, trades recientes).
- `GET /api/agents/portfolios/:id` — una cartera (`:id` = `risky` o `conservative`).
- `POST /api/agents/run-day` — corre un día de trading para ambos agentes.
  Body opcional `{"day":"YYYY-MM-DD"}` (default: hoy UTC).
- `POST /api/agents/feedback` — le preguntás algo al agente de feedback.
  Body opcional `{"question":"..."}` (si no mandás pregunta, da una revisión general).
- `GET /api/agents/feedback` — historial de las últimas devoluciones pedidas.

El agente de feedback usa el mismo proveedor de LLM configurado para el investment
analyst (`LLM_PROVIDER` / `OLLAMA_*` / `OPENAI_*` en `.env`) — con tu setup actual,
tu Ollama local.

## Financiamiento mensual

Cada cartera arranca en $0 y se financia sola con `$5,000` ficticios la primera vez
que se corre `run-day` dentro de un mes calendario nuevo (una sola vez por mes,
controlado por `portfolios.last_funded_month`). No hace falta un paso manual aparte.

## Simulación con el histórico (marzo → hoy)

Una vez cargados los precios (`fetch-prices.sh` + `seed-prices-neon.sh`) y con la API
corriendo, se puede "recorrer" cada día del histórico llamando `run-day` en orden:

```bash
for d in $(seq -f "2026-03-%02g" 1 31); do
  curl -s -X POST http://localhost:3090/api/agents/run-day \
    -H "Content-Type: application/json" -d "{\"day\":\"$d\"}" \
    | jq -c '.results[] | {profile: .risk_profile, trades: (.trades | length), cash: .cash_after_usd}'
done
```

Pero para la corrida real conviene usar `cmd/simulate` (más abajo), que además
hace el corte por `impact_score`, financia cada mes automáticamente y pide
feedback semanal.

## Corredor de simulación (`cmd/simulate`)

Recorre un rango de fechas día por día: cada día llama a la misma lógica que
`run-day` (financia el mes si corresponde, analiza noticias, deja operar a
ambos agentes), y **cada 7 días le pide al agente de feedback una revisión
semanal** de ambas carteras (la guarda en `agent_feedback` y la imprime en
el momento).

```bash
./scripts/simulate-neon.sh -from=2026-03-01 -to=2026-03-31
```

Como cada noticia analizada con un modelo local puede tardar bastante (ver
más arriba), para un rango largo conviene dejarlo corriendo en background:

```bash
nohup ./scripts/simulate-neon.sh -from=2026-03-01 -to=2026-03-31 > simulate-marzo.log 2>&1 &
disown
tail -f simulate-marzo.log   # para ver el progreso en vivo
```

**Se puede correr en tramos**: el estado (cash, posiciones, trades) vive en
la base, no en el proceso — así que `-from=2026-04-01 -to=2026-04-30` más
adelante sigue exactamente donde quedó marzo. Eso sí: **no vuelvas a correr
un rango de fechas que ya corriste** — no hay protección contra duplicados,
volvería a ejecutar operaciones para esos días.

Al terminar (o al final de cada tramo) imprime el resumen final de ambas
carteras: cash, valor de posiciones, equity total.
