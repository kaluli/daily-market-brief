/** URL pública válida (no placeholders tipo "temp"). */
function isValidHttpUrl(s: string | undefined): boolean {
  if (!s?.trim()) return false;
  const t = s.trim();
  return t.startsWith("http://") || t.startsWith("https://");
}

/** Cliente en localhost: siempre usamos el proxy del propio Next (cualquier puerto). */
function isLocalhostBrowser(): boolean {
  return typeof window !== "undefined" && window.location?.hostname === "localhost";
}

/** True si la app puede llamar a la API (localhost con proxy o NEXT_PUBLIC_API_URL válida en prod). */
export function hasApi(): boolean {
  if (isLocalhostBrowser()) return true;
  if (isValidHttpUrl(process.env.NEXT_PUBLIC_API_URL)) return true;
  return false;
}

/**
 * Base URL de la API. En localhost (browser) usa el proxy /api/v1 del mismo origen (cualquier puerto).
 * En producción: NEXT_PUBLIC_API_URL debe ser http(s) válido; valores como "temp" se ignoran en dev local.
 */
export function getApiUrl(): string {
  if (isLocalhostBrowser()) {
    return `${window.location.origin}/api/v1`;
  }
  const pub = process.env.NEXT_PUBLIC_API_URL?.trim();
  if (isValidHttpUrl(pub)) {
    const base = pub!.replace(/\/api\/v1\/?$/, "").replace(/\/$/, "");
    return base || pub!;
  }
  return "";
}

/** Panel admin de la API (enlace en layout). Mismo origen vía proxy en local. */
export function apiAdminHref(): string {
  if (isLocalhostBrowser()) return "/api/v1/admin";
  const pub = process.env.NEXT_PUBLIC_API_URL?.trim();
  if (isValidHttpUrl(pub)) {
    const base = pub!.replace(/\/api\/v1\/?$/, "").replace(/\/$/, "") || pub!;
    return `${base}/admin`;
  }
  return "/api/v1/admin";
}

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
  if (!hasApi()) return { status: "no-api" };
  const r = await fetch(`${getApiUrl()}/api/health`);
  if (!r.ok) throw new Error("API health check failed");
  return r.json();
}

export async function summariesRange(from: string, to: string): Promise<SummaryMeta[]> {
  if (!hasApi()) return [];
  const r = await fetch(`${getApiUrl()}/api/summaries?from=${from}&to=${to}`);
  if (!r.ok) throw new Error("Failed to fetch summaries");
  return r.json();
}

export async function summaryByDay(day: string): Promise<DaySummary | null> {
  if (!hasApi()) return null;
  const r = await fetch(`${getApiUrl()}/api/summaries/day/${day}`);
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
  if (!hasApi()) return null;
  const base =
    typeof window !== "undefined"
      ? window.location.origin
      : process.env.NEXT_PUBLIC_APP_URL ||
        (process.env.VERCEL_URL ? `https://${process.env.VERCEL_URL}` : "");
  const r = await fetch(`${base}/api/day/${encodeURIComponent(day)}`);
  if (r.status === 404) return null;
  if (!r.ok) {
    const body = await r.json().catch(() => ({}));
    const msg = (body as { error?: string }).error || "Failed to load day";
    throw new Error(msg);
  }
  return r.json();
}

export async function summaryByWeek(week: string): Promise<{ week: string; summaries: { day: string }[] }> {
  if (!hasApi()) return { week, summaries: [] };
  const r = await fetch(`${getApiUrl()}/api/summaries/week/${week}`);
  if (!r.ok) throw new Error("Failed to fetch week");
  return r.json();
}

export async function summaryByMonth(month: string): Promise<{ month: string; summaries: { day: string }[] }> {
  if (!hasApi()) return { month, summaries: [] };
  const r = await fetch(`${getApiUrl()}/api/summaries/month/${month}`);
  if (!r.ok) throw new Error("Failed to fetch month");
  return r.json();
}

export function downloadUrl(day: string): string {
  if (!hasApi()) return "#";
  return `${getApiUrl()}/api/summaries/day/${day}/download`;
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
  if (!hasApi()) return { day, items_analyzed: 0, analyses: [] };
  const r = await fetch(`${getApiUrl()}/api/analysis/day/${day}`);
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
  if (!hasApi()) return { sources: [], rss_sources: [] };
  const r = await fetch(`${getApiUrl()}/api/admin/sources`);
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
  if (!hasApi()) throw new Error("API no configurada");
  const r = await fetch(`${getApiUrl()}/api/admin/sources/${encodeURIComponent(id)}/enabled`, {
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

// --- Investor agents ---

export type AgentPosition = {
  ticker: string;
  quantity: number;
  avg_cost_usd: number;
};

export type AgentTrade = {
  ticker: string;
  side: "buy" | "sell";
  quantity: number;
  price_usd: number;
  executed_at: string;
  reasoning: string;
  cash_after_usd: number;
};

export type AgentPortfolio = {
  risk_profile: string;
  label: string;
  cash_usd: number;
  positions_value_usd: number;
  total_equity_usd: number;
  positions: AgentPosition[];
  recent_trades: AgentTrade[];
};

export type AgentDateRange = { from: string; to: string };

export async function getAgentPortfolios(range?: AgentDateRange): Promise<AgentPortfolio[]> {
  if (!hasApi()) return [];
  const qs = range ? `?from=${range.from}&to=${range.to}` : "";
  const r = await fetch(`${getApiUrl()}/api/agents/portfolios${qs}`);
  if (!r.ok) throw new Error("Failed to fetch agent portfolios");
  const data = await r.json();
  return (data as { portfolios?: AgentPortfolio[] }).portfolios ?? [];
}

export type AgentFeedbackEntry = {
  id: string;
  asked_at: string;
  question: string;
  answer: string;
  provider: string;
};

export async function getAgentFeedback(range?: AgentDateRange): Promise<AgentFeedbackEntry[]> {
  if (!hasApi()) return [];
  const qs = range ? `?from=${range.from}&to=${range.to}` : "";
  const r = await fetch(`${getApiUrl()}/api/agents/feedback${qs}`);
  if (!r.ok) throw new Error("Failed to fetch agent feedback");
  const data = await r.json();
  return (data as { feedback?: AgentFeedbackEntry[] }).feedback ?? [];
}

export async function askAgentCoach(question: string): Promise<{
  provider: string;
  question: string;
  answer: string;
}> {
  if (!hasApi()) throw new Error("API no configurada");
  const r = await fetch(`${getApiUrl()}/api/agents/feedback`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ question }),
  });
  if (!r.ok) {
    const err = await r.json().catch(() => ({}));
    throw new Error((err as { error?: string }).error || "Failed to ask coach");
  }
  return r.json();
}
