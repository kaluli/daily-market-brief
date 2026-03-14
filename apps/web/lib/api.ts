const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3090";

export type SummaryMeta = {
  day: string;
  generated_at: string;
  items_analyzed: number;
};

export type RankedItem = {
  rank: number;
  title: string;
  source: string;
  url: string;
  score: number;
};

export type DaySummary = {
  day: string;
  generated_at: string;
  top10: RankedItem[];
  other90: RankedItem[];
  items_analyzed: number;
};

export async function health(): Promise<{ status: string }> {
  const r = await fetch(`${API_URL}/api/health`);
  if (!r.ok) throw new Error("API health check failed");
  return r.json();
}

export async function summariesRange(from: string, to: string): Promise<SummaryMeta[]> {
  const r = await fetch(`${API_URL}/api/summaries?from=${from}&to=${to}`);
  if (!r.ok) throw new Error("Failed to fetch summaries");
  return r.json();
}

export async function summaryByDay(day: string): Promise<DaySummary | null> {
  const r = await fetch(`${API_URL}/api/summaries/day/${day}`);
  if (r.status === 404) return null;
  if (!r.ok) throw new Error("Failed to fetch summary");
  return r.json();
}

export async function summaryByWeek(week: string): Promise<{ week: string; summaries: { day: string }[] }> {
  const r = await fetch(`${API_URL}/api/summaries/week/${week}`);
  if (!r.ok) throw new Error("Failed to fetch week");
  return r.json();
}

export async function summaryByMonth(month: string): Promise<{ month: string; summaries: { day: string }[] }> {
  const r = await fetch(`${API_URL}/api/summaries/month/${month}`);
  if (!r.ok) throw new Error("Failed to fetch month");
  return r.json();
}

export function downloadUrl(day: string): string {
  return `${API_URL}/api/summaries/day/${day}/download`;
}
