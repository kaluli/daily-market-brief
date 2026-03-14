"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { format } from "date-fns";
import { summaryByDay, downloadUrl, type DaySummary, type RankedItem } from "@/lib/api";
import Link from "next/link";

export default function DayPage() {
  const params = useParams();
  const day = params.day as string;
  const [summary, setSummary] = useState<DaySummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!day) return;
    setLoading(true);
    setError(null);
    try {
      const s = await summaryByDay(day);
      setSummary(s);
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
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <Link href="/" className="text-blue-400 hover:underline">← Calendar</Link>
          <h1 className="mt-2 text-2xl font-bold">
            Daily Market Brief — {format(parseISO(summary.day), "PPP")}
          </h1>
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
          {top10.map((item: RankedItem) => (
            <li key={item.rank} className="flex flex-col gap-1">
              <a
                href={item.url}
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium text-blue-400 hover:underline"
              >
                {item.title}
              </a>
              <span className="text-sm text-slate-500">
                {item.source} · {item.url}
              </span>
            </li>
          ))}
        </ol>
      </section>

      <section className="rounded-lg border border-slate-700 bg-slate-800/50 p-6">
        <h2 className="mb-4 text-lg font-semibold">Other 90 (Ranked)</h2>
        <ol className="space-y-2">
          {other90.slice(0, 30).map((item: RankedItem, i: number) => (
            <li key={i} className="flex flex-col gap-0.5">
              <a
                href={item.url}
                target="_blank"
                rel="noopener noreferrer"
                className="text-slate-300 hover:text-white hover:underline"
              >
                {item.title}
              </a>
              <span className="text-xs text-slate-500">
                {item.source} — {item.url}
              </span>
            </li>
          ))}
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
