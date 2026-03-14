"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { format } from "date-fns";
import { es } from "date-fns/locale";
import {
  summaryByDayWithTranslations,
  downloadUrl,
  type DaySummary,
  type RankedItem,
} from "@/lib/api";
import Link from "next/link";

const EUROPA_KEYWORDS = [
  "spain", "españa", "español", "madrid", "barcelona", "eurozone", "euro area",
  "eu ", "e.u.", "european union", "unión europea", "ue ", "ecb", "bce", "europe",
  "europ", "brussels", "bruselas", "frankfurt", "lagarde", "euro ", "euros", "ftse",
  "cac ", "dax", "ibex", "banco central europeo", "commission europe",
  "merkel", "scholz", "macron", "sánchez", "sanch", "italy", "italia", "germany",
  "alemania", "france", "francia", "uk ", "reino unido", "brexit",
];

function getNewsScope(item: RankedItem): "europa" | "mundial" {
  const text = `${(item.title || "").toLowerCase()} ${(item.source || "").toLowerCase()}`;
  const isEuropa = EUROPA_KEYWORDS.some((k) => text.includes(k));
  return isEuropa ? "europa" : "mundial";
}

export default function DayPage() {
  const params = useParams();
  const day = params.day as string;
  const [summary, setSummary] = useState<DaySummary | null>(null);
  const [translations, setTranslations] = useState<Record<string, string>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [translationUnavailable, setTranslationUnavailable] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!day) return;
    setLoading(true);
    setError(null);
    setTranslationUnavailable(null);
    try {
      const data = await summaryByDayWithTranslations(day);
      if (data) {
        setSummary(data.summary);
        setTranslations(data.translations ?? {});
        if (data.quotaExceeded && data.retryAfter) {
          setTranslationUnavailable(data.retryAfter);
        }
      } else {
        setSummary(null);
      }
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load");
    } finally {
      setLoading(false);
    }
  }, [day]);

  useEffect(() => {
    load();
  }, [load]);

  if (loading) {
    return (
      <div className="py-12 text-center text-slate-500">Loading...</div>
    );
  }
  if (error || !summary) {
    return (
      <div className="space-y-4">
        <Link href="/" className="text-blue-400 hover:underline">← Back to calendar</Link>
        <p className="text-red-400">{error || "Summary not found."}</p>
      </div>
    );
  }

  const top10 = Array.isArray(summary.top10) ? summary.top10 : [];
  const other90 = Array.isArray(summary.other90) ? summary.other90 : [];

  return (
    <div className="space-y-8">
      {translationUnavailable && (
        <p className="rounded-lg border border-amber-500/40 bg-amber-500/10 px-4 py-2 text-sm text-amber-200/90">
          Traducciones no disponibles hasta dentro de {translationUnavailable}.
        </p>
      )}
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <Link href="/" className="text-blue-400 hover:underline">← Calendar</Link>
          <h1 className="mt-2 text-2xl font-bold">
            Daily Market Brief — {format(parseISO(summary.day), "PPP")}
          </h1>
          <p className="mt-1 text-lg font-medium text-amber-200/90">
            Resumen diario de mercados — {format(parseISO(summary.day), "d 'de' MMMM, yyyy", { locale: es })}
          </p>
          <p className="mt-1 text-sm text-slate-400">
            Generated {summary.generated_at} · {summary.items_analyzed} items analyzed
          </p>
        </div>
        <a
          href={downloadUrl(day)}
          download={`${day}.txt`}
          className="inline-flex rounded bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-500"
        >
          Download .txt
        </a>
      </div>

      <section className="rounded-lg border border-slate-700 bg-slate-800/50 p-6">
        <h2 className="mb-4 text-lg font-semibold">TOP 10 (Most Influential)</h2>
        <ol className="space-y-3">
          {top10.map((item: RankedItem) => {
            const scope = getNewsScope(item);
            return (
              <li key={item.rank} className="flex flex-col gap-0.5">
                <div className="flex flex-wrap items-center gap-2">
                  <span
                    className={`inline-flex rounded px-2 py-0.5 text-xs font-medium ${
                      scope === "europa"
                        ? "bg-amber-500/25 text-amber-200 border border-amber-500/40"
                        : "bg-slate-600/50 text-slate-300 border border-slate-500/50"
                    }`}
                  >
                    {scope === "europa" ? "Europa / UE" : "Mundial"}
                  </span>
                  <a
                    href={item.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="font-medium text-blue-400 hover:underline"
                  >
                    <span className="text-base text-amber-200/95">
                      {translations[item.url] || item.title}
                    </span>
                  </a>
                </div>
                {translations[item.url] && (
                  <span className="text-sm text-slate-400">
                    {item.title}
                  </span>
                )}
                <span className="text-xs text-slate-500">
                  {item.source} · {item.url}
                </span>
              </li>
            );
          })}
        </ol>
      </section>

      <section className="rounded-lg border border-slate-700 bg-slate-800/50 p-6">
        <h2 className="mb-4 text-lg font-semibold">Other 90 (Ranked)</h2>
        <ol className="space-y-2">
          {other90.slice(0, 30).map((item: RankedItem, i: number) => {
            const scope = getNewsScope(item);
            return (
              <li key={i} className="flex flex-col gap-0.5">
                <div className="flex flex-wrap items-center gap-2">
                  <span
                    className={`inline-flex rounded px-2 py-0.5 text-xs font-medium ${
                      scope === "europa"
                        ? "bg-amber-500/25 text-amber-200 border border-amber-500/40"
                        : "bg-slate-600/50 text-slate-300 border border-slate-500/50"
                    }`}
                  >
                    {scope === "europa" ? "Europa / UE" : "Mundial"}
                  </span>
                  <a
                    href={item.url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-slate-300 hover:text-white hover:underline"
                  >
                    <span className="text-sm text-amber-200/90">
                      {translations[item.url] || item.title}
                    </span>
                  </a>
                </div>
                {translations[item.url] && (
                  <span className="text-xs text-slate-400">
                    {item.title}
                  </span>
                )}
                <span className="text-xs text-slate-500">
                  {item.source} — {item.url}
                </span>
              </li>
            );
          })}
          {other90.length > 30 && (
            <li className="pt-2 text-slate-500">
              … and {other90.length - 30} more (see downloaded .txt for full list).
            </li>
          )}
        </ol>
      </section>
    </div>
  );
}

function parseISO(s: string): Date {
  return new Date(s + "T12:00:00Z");
}
