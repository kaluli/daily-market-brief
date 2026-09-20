"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import {
  getAgentPortfolios,
  getAgentFeedback,
  getAgentBenchmark,
  askAgentCoach,
  respondToRecommendation,
  type AgentPortfolio,
  type AgentFeedbackEntry,
  type AgentBenchmark,
  type AgentDateRange,
} from "@/lib/api";

const REFRESH_MS = 30_000;

const PRESETS: { label: string; from: string; to: string }[] = [
  { label: "Marzo 2026", from: "2026-03-01", to: "2026-03-31" },
  { label: "Abril 2026", from: "2026-04-01", to: "2026-04-30" },
  { label: "Mayo 2026", from: "2026-05-01", to: "2026-05-31" },
];

function usd(n: number): string {
  return n.toLocaleString("en-US", { style: "currency", currency: "USD", minimumFractionDigits: 2 });
}

function isWeeklyReview(question: string): boolean {
  const q = question.trim().toLowerCase();
  return q.startsWith("revision semanal") || q.startsWith("revisión semanal");
}

export default function AgentsPage() {
  const [portfolios, setPortfolios] = useState<AgentPortfolio[]>([]);
  const [feedback, setFeedback] = useState<AgentFeedbackEntry[]>([]);
  const [benchmark, setBenchmark] = useState<AgentBenchmark | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [respondingId, setRespondingId] = useState<string | null>(null);
  const [respondError, setRespondError] = useState<string | null>(null);

  const [question, setQuestion] = useState("");
  const [asking, setAsking] = useState(false);
  const [askError, setAskError] = useState<string | null>(null);

  // Date range filter: when set, muestra TODAS las operaciones y feedback de
  // ese periodo (en vez del estado actual + los últimos 10/20).
  const [fromInput, setFromInput] = useState("");
  const [toInput, setToInput] = useState("");
  const [activeRange, setActiveRange] = useState<AgentDateRange | undefined>(undefined);

  const load = useCallback(async (showSpinner: boolean, range: AgentDateRange | undefined) => {
    if (showSpinner) setLoading(true);
    setError(null);
    try {
      const [p, f, b] = await Promise.all([
        getAgentPortfolios(range),
        getAgentFeedback(range),
        getAgentBenchmark().catch(() => null), // benchmark is a nice-to-have; don't block the page on it
      ]);
      setPortfolios(p);
      setFeedback(f);
      setBenchmark(b);
      setLastUpdated(new Date());
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load");
    } finally {
      if (showSpinner) setLoading(false);
    }
  }, []);

  async function handleRecommendation(feedbackId: string, action: "apply" | "dismiss") {
    setRespondingId(feedbackId);
    setRespondError(null);
    try {
      await respondToRecommendation(feedbackId, action);
      await load(false, activeRange);
    } catch (err) {
      setRespondError(err instanceof Error ? err.message : "Failed to respond");
    } finally {
      setRespondingId(null);
    }
  }

  useEffect(() => {
    load(true, activeRange);
    // Solo auto-refresca cuando estamos viendo el estado actual (sin filtro
    // de fecha) — si el usuario está mirando un periodo fijo, no lo pisamos.
    if (!activeRange) {
      const id = setInterval(() => load(false, undefined), REFRESH_MS);
      return () => clearInterval(id);
    }
  }, [load, activeRange]);

  function applyRange(from: string, to: string) {
    if (!from || !to) return;
    setFromInput(from);
    setToInput(to);
    setActiveRange({ from, to });
  }

  function clearRange() {
    setFromInput("");
    setToInput("");
    setActiveRange(undefined);
  }

  async function handleAsk(e: React.FormEvent) {
    e.preventDefault();
    setAsking(true);
    setAskError(null);
    try {
      await askAgentCoach(question.trim());
      setQuestion("");
      await load(false, activeRange);
    } catch (err) {
      setAskError(err instanceof Error ? err.message : "Failed to ask coach");
    } finally {
      setAsking(false);
    }
  }

  if (loading) {
    return <div className="py-12 text-center text-slate-500">Loading...</div>;
  }

  return (
    <div className="space-y-10">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <Link href="/" className="text-blue-400 hover:underline">← Calendar</Link>
          <h1 className="mt-2 text-2xl font-bold">Agentes inversores</h1>
          <p className="mt-1 text-sm text-slate-400">
            Dos agentes simulan trading según las noticias (dinero ficticio, $5.000/mes cada uno). Un tercer
            agente los revisa y te responde preguntas.
            {lastUpdated && (
              <span className="ml-2 text-slate-500">· Actualizado {lastUpdated.toLocaleTimeString()}</span>
            )}
          </p>
        </div>
        <button
          onClick={() => load(true, activeRange)}
          className="inline-flex rounded bg-slate-700 px-4 py-2 text-sm font-medium text-white hover:bg-slate-600"
        >
          Actualizar
        </button>
      </div>

      <section className="space-y-3 rounded-lg border border-slate-700 bg-slate-800/50 p-5">
        <h2 className="text-sm font-medium text-slate-300">Ver un periodo específico</h2>
        <div className="flex flex-wrap items-end gap-3">
          <div className="flex flex-col gap-1">
            <label className="text-xs text-slate-500">Desde</label>
            <input
              type="date"
              value={fromInput}
              onChange={(e) => setFromInput(e.target.value)}
              className="rounded border border-slate-700 bg-slate-800 px-2 py-1.5 text-sm text-white focus:border-blue-500 focus:outline-none"
            />
          </div>
          <div className="flex flex-col gap-1">
            <label className="text-xs text-slate-500">Hasta</label>
            <input
              type="date"
              value={toInput}
              onChange={(e) => setToInput(e.target.value)}
              className="rounded border border-slate-700 bg-slate-800 px-2 py-1.5 text-sm text-white focus:border-blue-500 focus:outline-none"
            />
          </div>
          <button
            onClick={() => applyRange(fromInput, toInput)}
            disabled={!fromInput || !toInput}
            className="rounded bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-500 disabled:opacity-40"
          >
            Ver periodo
          </button>
          {activeRange && (
            <button
              onClick={clearRange}
              className="rounded bg-slate-700 px-3 py-1.5 text-sm font-medium text-white hover:bg-slate-600"
            >
              Ver estado actual
            </button>
          )}
          <div className="flex flex-wrap gap-2 pl-2">
            {PRESETS.map((p) => (
              <button
                key={p.label}
                onClick={() => applyRange(p.from, p.to)}
                className="rounded border border-slate-600 px-2.5 py-1 text-xs text-slate-300 hover:border-blue-500 hover:text-white"
              >
                {p.label}
              </button>
            ))}
          </div>
        </div>
        {activeRange ? (
          <p className="text-xs text-slate-500">
            Mostrando operaciones y feedback del {activeRange.from} al {activeRange.to}.
          </p>
        ) : (
          <p className="text-xs text-slate-500">
            Mostrando el estado actual (últimas operaciones y últimos feedbacks). Elegí un periodo para ver todo
            lo que pasó en un mes específico.
          </p>
        )}
      </section>

      {error && (
        <p className="rounded-lg border border-red-500/40 bg-red-500/10 px-4 py-2 text-sm text-red-300">
          {error} — ¿está corriendo la API? (<code>./scripts/run-api-neon.sh</code>)
        </p>
      )}

      {portfolios.length === 0 && !error ? (
        <p className="text-slate-500">Todavía no hay carteras. Corré al menos un día de agentes.</p>
      ) : (
        <div className="space-y-3">
          <div className="inline-flex items-center gap-2 rounded-lg border border-blue-500/30 bg-blue-500/10 px-4 py-2 text-sm">
            <span className="text-blue-300">Periodo analizado:</span>
            <span className="font-semibold text-white">
              {activeRange ? `${activeRange.from} → ${activeRange.to}` : "Estado actual (últimas operaciones)"}
            </span>
          </div>
          <div className="grid gap-6 md:grid-cols-2">
            {portfolios.map((p) => (
              <PortfolioCard key={p.risk_profile} portfolio={p} filtered={!!activeRange} />
            ))}
          </div>
          {benchmark && <BenchmarkCard benchmark={benchmark} />}
        </div>
      )}

      {!activeRange && (
        <section className="space-y-4">
          <h2 className="text-lg font-semibold">Preguntale al agente coach</h2>
          <form onSubmit={handleAsk} className="flex flex-col gap-2 sm:flex-row">
            <input
              type="text"
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              placeholder="Ej: ¿por qué el agente arriesgado compró XLE esta semana?"
              className="flex-1 rounded border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white placeholder:text-slate-500 focus:border-blue-500 focus:outline-none"
            />
            <button
              type="submit"
              disabled={asking}
              className="inline-flex justify-center rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-500 disabled:opacity-50"
            >
              {asking ? "Preguntando..." : "Preguntar"}
            </button>
          </form>
          {askError && <p className="text-sm text-red-400">{askError}</p>}
          <p className="text-xs text-slate-500">
            Esto le pide una respuesta nueva al modelo (necesita Ollama corriendo). El feedback semanal de abajo
            ya está guardado en la base y se ve igual, esté o no Ollama disponible.
          </p>
        </section>
      )}

      <section className="space-y-4">
        <h2 className="text-lg font-semibold">
          {activeRange ? `Feedback del coach (${activeRange.from} a ${activeRange.to})` : "Feedback del coach"}
        </h2>
        {feedback.length === 0 ? (
          <p className="text-slate-500">Sin feedback registrado en este periodo.</p>
        ) : (
          <ul className="space-y-4">
            {feedback.map((f) => (
              <li key={f.id} className="rounded-lg border border-slate-700 bg-slate-800/50 p-5">
                <div className="mb-2 flex flex-wrap items-center gap-2">
                  <span
                    className={`inline-flex rounded px-2 py-0.5 text-xs font-medium ${
                      isWeeklyReview(f.question)
                        ? "border border-blue-500/40 bg-blue-500/20 text-blue-200"
                        : "border border-slate-500/50 bg-slate-600/50 text-slate-300"
                    }`}
                  >
                    {isWeeklyReview(f.question) ? "Feedback semanal" : f.question.trim() ? "Pregunta" : "Revisión general"}
                  </span>
                  <span className="text-xs text-slate-500">
                    {new Date(f.asked_at).toLocaleString()} · {f.provider}
                  </span>
                </div>
                {f.question.trim() && !isWeeklyReview(f.question) && (
                  <p className="mb-2 text-sm font-medium text-amber-200/90">{f.question}</p>
                )}
                <p className="whitespace-pre-wrap text-sm text-slate-200">{f.answer}</p>
                {f.recommendation_status === "pending" && f.recommendations && f.recommendations.length > 0 && (
                  <div className="mt-3 space-y-2 rounded-lg border border-amber-500/40 bg-amber-500/10 p-3">
                    <p className="text-xs font-medium uppercase tracking-wide text-amber-300">Recomendación del coach</p>
                    {f.recommendations.map((rec, i) => (
                      <p key={i} className="text-sm text-amber-100/90">
                        <span className="font-semibold">{rec.profile}</span>: {rec.field} {rec.current_value} → {rec.suggested_value}. {rec.reason}
                      </p>
                    ))}
                    <div className="flex gap-2 pt-1">
                      <button
                        onClick={() => handleRecommendation(f.id, "apply")}
                        disabled={respondingId === f.id}
                        className="rounded bg-amber-600 px-3 py-1 text-xs font-medium text-white hover:bg-amber-500 disabled:opacity-50"
                      >
                        {respondingId === f.id ? "..." : "Aplicar"}
                      </button>
                      <button
                        onClick={() => handleRecommendation(f.id, "dismiss")}
                        disabled={respondingId === f.id}
                        className="rounded bg-slate-700 px-3 py-1 text-xs font-medium text-white hover:bg-slate-600 disabled:opacity-50"
                      >
                        Descartar
                      </button>
                    </div>
                  </div>
                )}
                {f.recommendation_status === "applied" && (
                  <p className="mt-2 text-xs text-emerald-400">✓ Recomendación aplicada</p>
                )}
                {f.recommendation_status === "dismissed" && (
                  <p className="mt-2 text-xs text-slate-500">Recomendación descartada</p>
                )}
              </li>
            ))}
          </ul>
        )}
        {respondError && <p className="mt-2 text-sm text-red-400">{respondError}</p>}
      </section>
    </div>
  );
}

function BenchmarkCard({ benchmark }: { benchmark: AgentBenchmark }) {
  return (
    <div className="rounded-lg border border-dashed border-slate-600 bg-slate-800/30 p-6">
      <div className="mb-4 flex items-baseline justify-between">
        <h2 className="text-lg font-semibold text-slate-300">{benchmark.label}</h2>
        <span className="text-xs uppercase tracking-wide text-slate-500">pasivo, sin operar</span>
      </div>
      <div className="mb-2 grid grid-cols-3 gap-3 text-center">
        <div>
          <div className="text-xs text-slate-500">Cash</div>
          <div className="text-base font-semibold">{usd(benchmark.cash_usd)}</div>
        </div>
        <div>
          <div className="text-xs text-slate-500">Posiciones</div>
          <div className="text-base font-semibold">{usd(benchmark.positions_value_usd)}</div>
        </div>
        <div>
          <div className="text-xs text-slate-500">Equity total</div>
          <div className="text-base font-semibold text-emerald-400">{usd(benchmark.total_equity_usd)}</div>
        </div>
      </div>
      <p className="text-center text-xs text-slate-500">
        Retorno desde el inicio: <span className={benchmark.return_pct >= 0 ? "text-emerald-400" : "text-red-400"}>{benchmark.return_pct.toFixed(1)}%</span>
        {" "}· {benchmark.shares_held} {benchmark.ticker} · fondeado {usd(benchmark.total_funded_usd)}
      </p>
      <p className="mt-2 text-center text-xs text-slate-600">
        Comprar {benchmark.ticker} con el mismo aporte mensual y no operar nunca — el punto de referencia para saber si los agentes realmente agregan valor.
      </p>
    </div>
  );
}

function PortfolioCard({ portfolio, filtered }: { portfolio: AgentPortfolio; filtered: boolean }) {
  return (
    <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-6">
      <div className="mb-4 flex items-baseline justify-between">
        <h2 className="text-lg font-semibold">{portfolio.label}</h2>
        <span className="text-xs uppercase tracking-wide text-slate-500">{portfolio.risk_profile}</span>
      </div>
      <div className="mb-4 grid grid-cols-3 gap-3 text-center">
        <div>
          <div className="text-xs text-slate-500">Cash</div>
          <div className="text-base font-semibold">{usd(portfolio.cash_usd)}</div>
        </div>
        <div>
          <div className="text-xs text-slate-500">Posiciones</div>
          <div className="text-base font-semibold">{usd(portfolio.positions_value_usd)}</div>
        </div>
        <div>
          <div className="text-xs text-slate-500">Equity total</div>
          <div className="text-base font-semibold text-emerald-400">{usd(portfolio.total_equity_usd)}</div>
        </div>
      </div>
      {!filtered && (
        <div className="mb-4">
          <h3 className="mb-2 text-sm font-medium text-slate-300">Posiciones abiertas</h3>
          {(portfolio.positions ?? []).length === 0 ? (
            <p className="text-sm text-slate-500">Sin posiciones abiertas.</p>
          ) : (
            <ul className="space-y-1 text-sm">
              {(portfolio.positions ?? []).map((pos) => (
                <li key={pos.ticker} className="flex justify-between text-slate-300">
                  <span>{pos.quantity} {pos.ticker}</span>
                  <span className="text-slate-500">avg {usd(pos.avg_cost_usd)}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}

      <div>
        <h3 className="mb-2 text-sm font-medium text-slate-300">
          {filtered ? "Operaciones del periodo" : "Operaciones recientes"}
        </h3>
        {(portfolio.recent_trades ?? []).length === 0 ? (
          <p className="text-sm text-slate-500">Sin operaciones {filtered ? "en este periodo." : "todavía."}</p>
        ) : (
          <ul className="space-y-2 text-sm">
            {(portfolio.recent_trades ?? []).map((t, i) => (
              <li key={i} className="border-l-2 border-slate-600 pl-3">
                <div className="flex flex-wrap items-center gap-2">
                  <span
                    className={`inline-flex rounded px-1.5 py-0.5 text-xs font-medium ${
                      t.side === "buy"
                        ? "border border-emerald-500/40 bg-emerald-500/15 text-emerald-300"
                        : "border border-red-500/40 bg-red-500/15 text-red-300"
                    }`}
                  >
                    {t.side === "buy" ? "COMPRA" : "VENTA"}
                  </span>
                  <span className="text-slate-200">{t.quantity} {t.ticker} @ {usd(t.price_usd)}</span>
                  <span className="text-xs text-slate-500">{t.executed_at}</span>
                </div>
                <p className="mt-1 text-xs text-slate-500">Cash restante: <span className="text-slate-300">{usd(t.cash_after_usd)}</span></p>
                {t.reasoning && <p className="mt-1 text-xs text-slate-400">{t.reasoning}</p>}
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
