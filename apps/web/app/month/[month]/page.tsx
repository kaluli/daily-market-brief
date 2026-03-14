"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { summaryByMonth } from "@/lib/api";

export default function MonthPage() {
  const params = useParams();
  const month = params.month as string;
  const [data, setData] = useState<{ month: string; summaries: { day: string }[] } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!month) return;
    setLoading(true);
    setError(null);
    try {
      const res = await summaryByMonth(month);
      setData(res);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load");
    } finally {
      setLoading(false);
    }
  }, [month]);

  useEffect(() => {
    load();
  }, [load]);

  if (loading) return <div className="py-12 text-center text-slate-500">Loading...</div>;
  if (error || !data) {
    return (
      <div className="space-y-4">
        <Link href="/" className="text-blue-400 hover:underline">← Back</Link>
        <p className="text-red-400">{error || "Not found."}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Link href="/" className="text-blue-400 hover:underline">← Calendar</Link>
      <h1 className="text-2xl font-bold">Month {month}</h1>
      <ul className="grid gap-2 sm:grid-cols-2 md:grid-cols-3">
        {data.summaries.length === 0 ? (
          <li className="text-slate-500">No summaries for this month.</li>
        ) : (
          data.summaries.map((s) => (
            <li key={s.day}>
              <Link href={`/day/${s.day}`} className="text-blue-400 hover:underline">
                {s.day}
              </Link>
            </li>
          ))
        )}
      </ul>
    </div>
  );
}
