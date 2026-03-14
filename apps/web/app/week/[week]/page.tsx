"use client";

import { useCallback, useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { summaryByWeek } from "@/lib/api";

export default function WeekPage() {
  const params = useParams();
  const week = params.week as string;
  const [data, setData] = useState<{ week: string; summaries: { day: string }[] } | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!week) return;
    setLoading(true);
    setError(null);
    try {
      const res = await summaryByWeek(week);
      setData(res);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load");
    } finally {
      setLoading(false);
    }
  }, [week]);

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
      <h1 className="text-2xl font-bold">Week {week}</h1>
      <ul className="space-y-2">
        {data.summaries.length === 0 ? (
          <li className="text-slate-500">No summaries for this week.</li>
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
