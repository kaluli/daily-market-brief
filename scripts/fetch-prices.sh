#!/usr/bin/env bash
# Downloads free daily historical close prices (Yahoo Finance chart API, no
# API key) for the fixed ETF universe the investor agents trade, from March 1
# to today.
#
# Run this in YOUR OWN terminal (needs normal internet access — this can't
# run inside the Claude session's restricted shells). See docs/AGENTS.md.
#
# (We originally tried Stooq's CSV export, but it now serves a JS bot-check
# page instead of data when fetched with plain curl, so we use Yahoo's chart
# JSON API instead and convert it to the same Date,Close CSV shape.)
set -e
cd "$(dirname "$0")/.."
OUT_DIR="data/prices"
mkdir -p "$OUT_DIR"

D1="${D1:-2026-03-01}"
D2="${D2:-$(date -u +%Y-%m-%d)}"

to_epoch() {
  date -j -f "%Y-%m-%d" "$1" +%s 2>/dev/null || date -d "$1" +%s
}
P1=$(to_epoch "$D1")
P2=$(( $(to_epoch "$D2") + 86400 )) # include the end day

TICKERS=(SPY QQQ IWM XLE XLF TLT GLD USO UUP BITO)
UA="Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36"

for T in "${TICKERS[@]}"; do
  echo "Descargando $T ..."
  URL="https://query1.finance.yahoo.com/v8/finance/chart/${T}?period1=${P1}&period2=${P2}&interval=1d"
  curl -sS -A "$UA" "$URL" > "$OUT_DIR/$T.json"
  python3 - "$T" "$OUT_DIR/$T.json" "$OUT_DIR/$T.csv" <<'PYEOF'
import json, sys, datetime

ticker, in_path, out_path = sys.argv[1], sys.argv[2], sys.argv[3]
with open(in_path) as f:
    data = json.load(f)

result = data.get("chart", {}).get("result")
if not result:
    err = data.get("chart", {}).get("error")
    print(f"  !! {ticker}: sin datos en la respuesta ({err})")
    sys.exit(0)

r = result[0]
timestamps = r.get("timestamp") or []
closes = (r.get("indicators", {}).get("quote", [{}]) or [{}])[0].get("close") or []
offset = r.get("meta", {}).get("gmtoffset", 0)

rows = []
for ts, close in zip(timestamps, closes):
    if close is None:
        continue
    day = datetime.datetime.utcfromtimestamp(ts + offset).strftime("%Y-%m-%d")
    rows.append((day, close))

with open(out_path, "w") as f:
    f.write("Date,Close\n")
    for day, close in rows:
        f.write(f"{day},{close:.4f}\n")

print(f"  -> {out_path} ({len(rows)} filas)")
PYEOF
  rm -f "$OUT_DIR/$T.json"
  sleep 1
done

echo ""
echo "Listo. Ahora corre (con DATABASE_URL de Neon en el entorno):"
echo "  ./scripts/seed-prices-neon.sh"
