const API_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:3090";

/** Base URL of the API (for links like admin). Use this so dev goes to 3090 and prod uses NEXT_PUBLIC_API_URL. */
export const apiBaseUrl = API_URL;

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

/** Fetches day summary + Spanish translations from Next.js API (server-side translation). */
export async function summaryByDayWithTranslations(day: string): Promise<{
  summary: DaySummary;
  translations: Record<string, string>;
  quotaExceeded?: boolean;
  retryAfter?: string;
} | null> {
  const base =
    typeof window !== "undefined"
      ? window.location.origin
      : process.env.NEXT_PUBLIC_APP_URL || "";
  const r = await fetch(`${base}/api/day/${encodeURIComponent(day)}`);
  if (r.status === 404) return null;
  if (!r.ok) throw new Error("Failed to load day");
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

export type AnalysisResult = {
  relevance: string;
  category: string;
  summary: string;
  why_it_matters: string;
  impact_level: string;
  expected_market_reaction: string;
  affected_assets: string[];
  directional_bias: Record<string, string>;
  investment_signals: string[];
  time_horizon: string;
  signal_strength: string;
};

export async function analysisByDay(day: string): Promise<{
  day: string;
  items_analyzed: number;
  analyses: AnalysisResult[];
}> {
  const r = await fetch(`${API_URL}/api/analysis/day/${day}`);
  if (!r.ok) throw new Error("Failed to fetch analysis");
  return r.json();
}

// --- Admin: news sources ---

export type AdminSource = {
  id: string;
  name: string;
  enabled: boolean;
  weight: number;
  config?: Record<string, unknown>;
};

export type AdminRSSSource = {
  id: string;
  name: string;
  url: string;
  enabled: boolean;
  weight: number;
};

export type AdminSourcesResponse = {
  sources: AdminSource[];
  rss_sources: AdminRSSSource[];
};

export async function getAdminSources(): Promise<AdminSourcesResponse> {
  const r = await fetch(`${API_URL}/api/admin/sources`);
  if (!r.ok) {
    const err = await r.json().catch(() => ({}));
    const msg = (err as { error?: string }).error || "Error al cargar fuentes";
    throw new Error(msg);
  }
  return r.json();
}

export async function setSourceEnabled(
  id: string,
  enabled: boolean
): Promise<{ id: string; enabled: boolean }> {
  const r = await fetch(`${API_URL}/api/admin/sources/${encodeURIComponent(id)}/enabled`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ enabled }),
  });
  if (!r.ok) {
    const err = await r.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error || "Failed to update source");
  }
  return r.json();
}
