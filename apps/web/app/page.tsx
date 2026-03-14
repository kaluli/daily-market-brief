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
import { summaryByDay, summariesRange, type SummaryMeta } from "@/lib/api";
import Link from "next/link";

export default function HomePage() {
  const [currentMonth, setCurrentMonth] = useState(new Date());
  const [availableDays, setAvailableDays] = useState<Set<string>>(new Set());
  const [loading, setLoading] = useState(true);

  const loadAvailableDays = useCallback(async () => {
    setLoading(true);
    try {
      const start = startOfMonth(subMonths(currentMonth, 1));
      const end = endOfMonth(addMonths(currentMonth, 1));
      const list = await summariesRange(
        format(start, "yyyy-MM-dd"),
        format(end, "yyyy-MM-dd")
      );
      setAvailableDays(new Set(list.map((s) => s.day)));
    } catch {
      setAvailableDays(new Set());
    } finally {
      setLoading(false);
    }
  }, [currentMonth]);

  useEffect(() => {
    loadAvailableDays();
  }, [loadAvailableDays]);

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
      <h1 className="text-2xl font-bold">Calendar</h1>
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
    </div>
  );
}
