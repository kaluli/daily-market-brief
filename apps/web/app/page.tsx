"use client";

import { useCallback, useEffect, useState } from "react";
import {
  format,
  startOfMonth,
  endOfMonth,
  startOfWeek,
  endOfWeek,
  addMonths,
  subMonths,
  isSameMonth,
  isSameDay,
  addDays,
  parseISO,
  getISOWeek,
  getYear,
} from "date-fns";
import { es } from "date-fns/locale";
import { summariesRange, summaryByDay, analysisByDay, type AnalysisResult, type RankedItem } from "@/lib/api";
import Link from "next/link";

function getTranslateUrl(): string {
  if (typeof window === "undefined") return "/api/translate";
  return `${window.location.origin}/api/translate`;
}

async function translateToSpanish(text: string): Promise<string> {
  if (!text?.trim()) return "";
  try {
    const res = await fetch(
      `${getTranslateUrl()}?text=${encodeURIComponent(text.trim())}`
    );
    if (!res.ok) return "";
    const data = await res.json();
    if (data?.translated && typeof data.translated === "string") return data.translated;
  } catch {
    // ignore
  }
  return "";
}

/** Builds 2 paragraphs in Spanish. Uses translated title when available, otherwise original (so we never block). */
function buildDayBriefFromHeadlines(
  items: RankedItem[],
  day: string,
  translatedTitles: Record<string, string>
): { p1: string; p2: string } | null {
  const titles = items
    .slice(0, 12)
    .map((i) => (i.title && translatedTitles[i.title]) || i.title?.trim())
    .filter(Boolean) as string[];
  if (titles.length === 0) return null;
  const dateStr = format(new Date(day + "T12:00:00Z"), "d 'de' MMMM 'de' yyyy", { locale: es });
  const half = Math.min(5, Math.ceil(titles.length / 2));
  const first = titles.slice(0, half);
  const rest = titles.slice(half, 12);
  const p1 =
    first.length > 0
      ? `El ${dateStr}, los titulares más relevantes incluyeron: «${first.join("», «")}».`
      : "";
  const p2 =
    rest.length > 0
      ? `También destacaron noticias como «${rest.join("», «")}». La jornada estuvo marcada por la atención a la actualidad financiera y de mercados.`
      : "La jornada estuvo marcada por la atención a la actualidad financiera y de mercados.";
  if (!p1) return null;
  return { p1, p2 };
}

export default function HomePage() {
  const [currentMonth, setCurrentMonth] = useState(new Date());
  const [availableDays, setAvailableDays] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);
  const [analyses, setAnalyses] = useState<AnalysisResult[]>([]);
  const [analysisDay, setAnalysisDay] = useState<string | null>(null);
  const [dayHeadlines, setDayHeadlines] = useState<RankedItem[]>([]);
  const [headlineTranslations, setHeadlineTranslations] = useState<Record<string, string>>({});

  const loadAvailableDays = useCallback(async () => {
    setLoading(true);
    try {
      const start = startOfMonth(subMonths(currentMonth, 1));
      const end = endOfMonth(addMonths(currentMonth, 1));
      const list = await summariesRange(
        format(start, "yyyy-MM-dd"),
        format(end, "yyyy-MM-dd")
      );
      const days = new Set(list.map((s) => s.day));
      setAvailableDays(days);
      if (list.length > 0) {
        const sorted = [...list].sort((a, b) => b.day.localeCompare(a.day));
        const latestDay = sorted[0].day;
        setAnalysisDay(latestDay);
        const [summaryRes, analysisRes] = await Promise.all([
          summaryByDay(latestDay).catch(() => null),
          analysisByDay(latestDay).catch(() => ({ analyses: [] })),
        ]);
        if (summaryRes?.top10?.length || summaryRes?.other90?.length) {
          const top10 = Array.isArray(summaryRes.top10) ? summaryRes.top10 : [];
          const other90 = Array.isArray(summaryRes.other90) ? summaryRes.other90 : [];
          setDayHeadlines([...top10, ...other90.slice(0, 20)]);
        } else {
          setDayHeadlines([]);
        }
        setAnalyses(Array.isArray(analysisRes.analyses) ? analysisRes.analyses : []);
      } else {
        setAnalysisDay(null);
        setAnalyses([]);
        setDayHeadlines([]);
        setHeadlineTranslations({});
      }
    } catch {
      setAvailableDays(new Set());
      setAnalyses([]);
      setAnalysisDay(null);
      setDayHeadlines([]);
      setHeadlineTranslations({});
    } finally {
      setLoading(false);
    }
  }, [currentMonth]);

  useEffect(() => {
    loadAvailableDays();
  }, [loadAvailableDays]);

  // Translate headlines to Spanish in the background (first 12). If API fails, brief still shows with originals.
  useEffect(() => {
    const items = dayHeadlines.slice(0, 12);
    if (items.length === 0) return;
    let cancelled = false;
    (async () => {
      for (const item of items) {
        if (cancelled || !item.title?.trim()) continue;
        const spanish = await translateToSpanish(item.title);
        if (cancelled) break;
        if (spanish) {
          setHeadlineTranslations((prev) => ({ ...prev, [item.title!]: spanish }));
        }
        await new Promise((r) => setTimeout(r, 250));
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [analysisDay, dayHeadlines.length]);

  const monthStart = startOfMonth(currentMonth);
  const monthEnd = endOfMonth(currentMonth);
  const calStart = startOfWeek(monthStart, { weekStartsOn: 1 });
  const calEnd = endOfWeek(monthEnd, { weekStartsOn: 1 });
  const rows: Date[][] = [];
  let row: Date[] = [];
  let d = calStart;
  while (d <= calEnd) {
    row.push(d);
    if (row.length === 7) {
      rows.push(row);
      row = [];
    }
    d = addDays(d, 1);
  }
  if (row.length) rows.push(row);

  return (
    <div className="space-y-8">
      <h1 className="text-2xl font-bold">Daily Market Brief</h1>
      <div className="flex items-center justify-between">
        <button
          type="button"
          onClick={() => setCurrentMonth((m) => subMonths(m, 1))}
          className="rounded bg-slate-700 px-4 py-2 text-sm font-medium text-white hover:bg-slate-600"
        >
          Previous
        </button>
        <span className="text-lg font-medium">
          {format(currentMonth, "MMMM yyyy")}
        </span>
        <button
          type="button"
          onClick={() => setCurrentMonth((m) => addMonths(m, 1))}
          className="rounded bg-slate-700 px-4 py-2 text-sm font-medium text-white hover:bg-slate-600"
        >
          Next
        </button>
      </div>
      <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-4">
        <div className="grid grid-cols-7 gap-1 text-center text-sm text-slate-400">
          {["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"].map((w) => (
            <div key={w} className="py-2 font-medium">
              {w}
            </div>
          ))}
        </div>
        {loading ? (
          <div className="py-12 text-center text-slate-500">Loading...</div>
        ) : (
          rows.map((week, wi) => (
            <div key={wi} className="grid grid-cols-7 gap-1">
              {week.map((date) => {
                const dayStr = format(date, "yyyy-MM-dd");
                const hasSummary = availableDays.has(dayStr);
                const isCurrentMonth = isSameMonth(date, currentMonth);
                return (
                  <div
                    key={dayStr}
                    className={`flex min-h-[44px] items-center justify-center rounded text-sm ${
                      !isCurrentMonth ? "text-slate-600" : "text-slate-200"
                    }`}
                  >
                    {hasSummary ? (
                      <Link
                        href={`/day/${dayStr}`}
                        className="flex h-9 w-9 items-center justify-center rounded-full bg-blue-600 font-medium text-white hover:bg-blue-500"
                      >
                        {format(date, "d")}
                      </Link>
                    ) : (
                      <span>{format(date, "d")}</span>
                    )}
                  </div>
                );
              })}
            </div>
          ))
        )}
      </div>
      <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-4">
        <h2 className="mb-2 text-lg font-semibold">Views</h2>
        <ul className="flex flex-wrap gap-4">
          <li>
            <Link
              href={`/week/${format(currentMonth, "yyyy")}-W${String(getISOWeek(monthStart)).padStart(2, "0")}`}
              className="text-blue-400 hover:underline"
            >
              This week
            </Link>
          </li>
          <li>
            <Link
              href={`/month/${format(currentMonth, "yyyy-MM")}`}
              className="text-blue-400 hover:underline"
            >
              This month
            </Link>
          </li>
          <li>
            <Link href="/last-10" className="text-blue-400 hover:underline">
              Last 10 days
            </Link>
          </li>
        </ul>
      </div>

      {analysisDay && (
        <section className="rounded-lg border border-slate-700 bg-slate-800/50 p-6">
          <h2 className="mb-3 text-lg font-semibold">
            Resumen del día —{" "}
            <Link href={`/day/${analysisDay}`} className="text-blue-400 hover:underline">
              {format(new Date(analysisDay + "T12:00:00Z"), "d 'de' MMMM, yyyy", { locale: es })}
            </Link>
          </h2>
          {(() => {
            const brief = buildDayBriefFromHeadlines(dayHeadlines, analysisDay, headlineTranslations);
            if (dayHeadlines.length === 0) {
              return (
                <p className="text-slate-500 italic">
                  No hay titulares para este día. Revisa el reporte diario cuando se haya generado el resumen.
                </p>
              );
            }
            if (!brief) {
              return (
                <p className="text-slate-400 italic">Cargando resumen…</p>
              );
            }
            return (
              <>
                <p className="mb-4 text-slate-300 leading-relaxed">{brief.p1}</p>
                <p className="text-slate-300 leading-relaxed">{brief.p2}</p>
              </>
            );
          })()}
        </section>
      )}

      {analysisDay && analyses.length > 0 && (
        <section className="rounded-lg border border-slate-700 bg-slate-800/50 p-6">
          <h2 className="mb-2 text-lg font-semibold">Investment analysis</h2>
          <p className="mb-4 text-sm text-slate-400">
            Relevance, impact, and signals for the latest report
            {analysisDay && (
              <> · <Link href={`/day/${analysisDay}`} className="text-blue-400 hover:underline">Report {analysisDay}</Link></>
            )}
          </p>
          <div className="space-y-4">
            {analyses.slice(0, 10).map((a, i) => (
              <AnalysisCard key={i} analysis={a} />
            ))}
            {analyses.length > 10 && (
              <p className="text-sm text-slate-500">
                <Link href={`/day/${analysisDay}`} className="text-blue-400 hover:underline">
                  View all {analyses.length} analyses in the daily report →
                </Link>
              </p>
            )}
          </div>
        </section>
      )}
    </div>
  );
}

function AnalysisCard({ analysis }: { analysis: AnalysisResult }) {
  const relevanceColor =
    analysis.relevance === "Market Moving"
      ? "bg-amber-500/20 text-amber-300 border-amber-500/40"
      : analysis.relevance === "Potentially Relevant"
        ? "bg-blue-500/20 text-blue-300 border-blue-500/40"
        : "bg-slate-500/20 text-slate-400 border-slate-500/40";
  const impactColor =
    analysis.impact_level === "High"
      ? "bg-red-500/20 text-red-300"
      : analysis.impact_level === "Medium"
        ? "bg-amber-500/20 text-amber-300"
        : "bg-slate-500/20 text-slate-400";

  return (
    <div className="rounded-lg border border-slate-600 bg-slate-900/50 p-4">
      <div className="mb-2 flex flex-wrap items-center gap-2">
        <span className={`rounded px-2 py-0.5 text-xs font-medium border ${relevanceColor}`}>
          {analysis.relevance}
        </span>
        <span className="rounded bg-slate-600/50 px-2 py-0.5 text-xs text-slate-300">
          {analysis.category}
        </span>
        <span className={`rounded px-2 py-0.5 text-xs ${impactColor}`}>
          Impact: {analysis.impact_level}
        </span>
        {analysis.signal_strength && (
          <span className="text-xs text-slate-500">Signal strength: {analysis.signal_strength}/10</span>
        )}
      </div>
      <p className="mb-1 font-medium text-slate-200">{analysis.summary}</p>
      {analysis.why_it_matters && !analysis.why_it_matters.startsWith("Stub") && (
        <p className="mb-2 text-sm text-slate-400">{analysis.why_it_matters}</p>
      )}
      {analysis.expected_market_reaction && (
        <p className="mb-2 text-sm">
          <span className="text-slate-500">Market reaction: </span>
          <span className="text-slate-300">{analysis.expected_market_reaction}</span>
        </p>
      )}
      {analysis.affected_assets && analysis.affected_assets.length > 0 && (
        <div className="mb-2">
          <span className="text-xs text-slate-500">Affected assets: </span>
          <span className="text-xs text-slate-300">{analysis.affected_assets.join(", ")}</span>
        </div>
      )}
      {analysis.directional_bias && Object.keys(analysis.directional_bias).length > 0 && (
        <div className="mb-2 flex flex-wrap gap-1">
          {Object.entries(analysis.directional_bias).map(([asset, bias]) => (
            <span
              key={asset}
              className={`rounded px-2 py-0.5 text-xs ${
                bias === "Bullish"
                  ? "bg-emerald-500/20 text-emerald-300"
                  : bias === "Bearish"
                    ? "bg-red-500/20 text-red-300"
                    : "bg-slate-500/20 text-slate-400"
              }`}
            >
              {asset}: {bias}
            </span>
          ))}
        </div>
      )}
      {analysis.investment_signals && analysis.investment_signals.length > 0 && (
        <div className="mb-2">
          <span className="text-xs font-medium text-slate-400">Signals: </span>
          <ul className="mt-1 list-inside list-disc text-sm text-slate-300">
            {analysis.investment_signals.map((s, j) => (
              <li key={j}>{s}</li>
            ))}
          </ul>
        </div>
      )}
      {analysis.time_horizon && (
        <p className="text-xs text-slate-500">Horizon: {analysis.time_horizon}</p>
      )}
    </div>
  );
}
