import Link from "next/link";

const TICKERS: { symbol: string; name: string; description: string }[] = [
  {
    symbol: "QQQ",
    name: "Invesco QQQ Trust",
    description:
      "ETF del índice Nasdaq-100. Invierte en las 100 empresas no financieras más grandes de EE. UU. (mayormente tecnología como Apple, Microsoft, Nvidia).",
  },
  {
    symbol: "USO",
    name: "United States Oil Fund",
    description:
      "ETF de Petróleo Crudo. Rastrea el precio spot del petróleo West Texas Intermediate (WTI). Es la principal herramienta que usan los agentes para posicionarse bullish ante el shock petrolero.",
  },
  {
    symbol: "XLE",
    name: "Energy Select Sector SPDR Fund",
    description:
      "ETF del Sector Energía. Invierte en grandes empresas petroleras y energéticas de EE. UU. (como ExxonMobil, Chevron, ConocoPhillips).",
  },
  {
    symbol: "SPY",
    name: "SPDR S&P 500 ETF Trust",
    description:
      "ETF del índice S&P 500. Representa las 500 empresas más grandes de la bolsa estadounidense (referencia general del mercado).",
  },
  {
    symbol: "UUP",
    name: "Invesco DB US Dollar Index Bullish",
    description:
      "ETF del Dólar Estadounidense. Rastrea la fortaleza del USD frente a una cesta de divisas internacionales principales.",
  },
  {
    symbol: "TLT",
    name: "iShares 20+ Year Treasury Bond ETF",
    description:
      "ETF de Bonos del Tesoro a largo plazo. Activo refugio tradicional que invierte en deuda pública de EE. UU. a más de 20 años.",
  },
];

const PORTFOLIO_METRICS: { term: string; description: string }[] = [
  {
    term: "Cash / Cash restante",
    description:
      "Dinero en efectivo disponible y sin operar en la cuenta. Es un cálculo histórico: se reconstruye repitiendo todo el fondeo mensual y cada operación en orden cronológico, así que refleja el saldo real en ese momento de la simulación, no solo el saldo de hoy.",
  },
  {
    term: "Posiciones",
    description: "Valor total actual de los activos/acciones que el agente tiene guardados en cartera.",
  },
  {
    term: "Equity total",
    description: "Patrimonio o valor total neto de la cartera: Equity = Cash + Posiciones.",
  },
  {
    term: "avg (Average Price)",
    description: "Precio promedio ponderado al que se compraron las unidades actuales de un activo.",
  },
];

const SENTIMENT_TERMS: { term: string; description: string }[] = [
  {
    term: "Bullish (Alcista)",
    description:
      "El agente interpreta que la noticia hará que el precio del activo o mercado suba (ej. comprar petróleo tras amenazas de guerra o problemas de suministro).",
  },
  {
    term: "Bearish (Bajista)",
    description:
      "El agente interpreta que la noticia hará que el mercado o activo caiga (ej. vender acciones ante temores de recesión o inflación).",
  },
  {
    term: "Strength (ej. 8/10)",
    description: "Nivel de convicción o fuerza que la IA le asigna a su predicción sobre la noticia.",
  },
  {
    term: "CPI-Linked Notes",
    description: "Bonos o títulos de deuda vinculados al índice de precios al consumo (IPC/inflación).",
  },
];

function SectionTable({
  headers,
  rows,
}: {
  headers: string[];
  rows: { key: string; cols: string[] }[];
}) {
  return (
    <div className="overflow-x-auto rounded-lg border border-slate-700">
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b border-slate-700 bg-slate-800/80">
            {headers.map((h) => (
              <th key={h} className="px-4 py-2.5 text-xs font-medium uppercase tracking-wide text-slate-400">
                {h}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map((r, i) => (
            <tr
              key={r.key}
              className={`border-b border-slate-700/60 last:border-b-0 ${i % 2 === 1 ? "bg-slate-800/30" : ""}`}
            >
              {r.cols.map((c, j) => (
                <td
                  key={j}
                  className={
                    j === 0
                      ? "px-4 py-3 align-top font-semibold text-white"
                      : "px-4 py-3 align-top text-slate-300"
                  }
                >
                  {c}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default function GlosarioPage() {
  return (
    <div className="space-y-10">
      <div>
        <Link href="/agents" className="text-blue-400 hover:underline">
          ← Agentes
        </Link>
        <h1 className="mt-2 text-2xl font-bold">Glosario</h1>
        <p className="mt-1 text-sm text-slate-400">
          Referencia rápida de los tickers y términos que usan los agentes en sus operaciones y análisis.
        </p>
      </div>

      <section className="space-y-3">
        <h2 className="text-lg font-semibold">Significado de los tickers (activos / ETFs)</h2>
        <SectionTable
          headers={["Símbolo", "Nombre completo", "¿Qué es y qué rastrea?"]}
          rows={TICKERS.map((t) => ({
            key: t.symbol,
            cols: [t.symbol, t.name, t.description],
          }))}
        />
      </section>

      <section className="space-y-3">
        <h2 className="text-lg font-semibold">Métricas de la cartera</h2>
        <SectionTable
          headers={["Término", "Significado"]}
          rows={PORTFOLIO_METRICS.map((m) => ({
            key: m.term,
            cols: [m.term, m.description],
          }))}
        />
      </section>

      <section className="space-y-3">
        <h2 className="text-lg font-semibold">Términos de análisis de noticias y sentimiento</h2>
        <SectionTable
          headers={["Término", "Significado"]}
          rows={SENTIMENT_TERMS.map((t) => ({
            key: t.term,
            cols: [t.term, t.description],
          }))}
        />
      </section>
    </div>
  );
}
