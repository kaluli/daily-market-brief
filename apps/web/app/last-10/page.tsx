"use client";

import { useCallback, useEffect, useState } from "react";
import { format, subDays } from "date-fns";
import Link from "next/link";
import { summariesRange } from "@/lib/api";

export default function Last10Page() {
  const [list, setList] = useState<{ day: string; generated_at?: string }[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    const end = new Date();
    const start = subDays(end, 10);
    try {
      const res = await summariesRange(
        format(start, "yyyy-MM-dd"),
        format(end, "yyyy-MM-dd")
      );
      setList(res);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  if (loading) return <div className="py-12 text-center text-slate-500">Loading...</div>;
  if (error) {
    return (
      <div className="space-y-4">
        <Link href="/" className="text-blue-400 hover:underline">← Back</Link>
        <p className="text-red-400">{error}</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Link href="/" className="text-blue-400 hover:underline">← Calendar</Link>
      <h1 className="text-2xl font-bold">Last 10 days</h1>
      <div className="overflow-x-auto rounded-lg border border-slate-700">
        <table className="w-full text-left text-sm">
          <thead>
            <tr className="border-b border-slate-700 bg-slate-800/50">
              <th className="px-4 py-3 font-medium">Day</th>
              <th className="px-4 py-3 font-medium">Generated</th>
              <th className="px-4 py-3 font-medium">Link</th>
            </tr>
          </thead>
          <tbody>
            {list.length === 0 ? (
              <tr>
                <td colSpan={3} className="px-4 py-8 text-center text-slate-500">
                  No summaries in this range.
                </td>
              </tr>
            ) : (
              list.map((s) => (
                <tr key={s.day} className="border-b border-slate-700/50">
                  <td className="px-4 py-3">{s.day}</td>
                  <td className="px-4 py-3 text-slate-400">{s.generated_at || "—"}</td>
                  <td className="px-4 py-3">
                    <Link href={`/day/${s.day}`} className="text-blue-400 hover:underline">
                      View
                    </Link>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
