import { NextRequest, NextResponse } from "next/server";

/** Backend API base (origen sin /api/v1). Server-side usa API_BACKEND_URL o NEXT_PUBLIC_API_URL. */
function getBackendUrl(): string | null {
  const raw = process.env.API_BACKEND_URL || process.env.NEXT_PUBLIC_API_URL;
  if (raw) {
    const base = raw.replace(/\/api\/v1\/?$/, "").replace(/\/$/, "");
    return base || raw;
  }
  if (process.env.VERCEL_URL) return null;
  return "http://localhost:3090";
}

type RankedItem = { rank: number; title: string; source: string; url: string; score: number };
type DaySummary = {
  day: string;
  generated_at: string;
  top10: RankedItem[];
  other90: RankedItem[];
  items_analyzed: number;
};

/** GET /api/day/[day]: summary from Go API (sin traducciones para carga rápida). */
export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ day: string }> }
) {
  const { day } = await params;
  if (!day?.trim()) {
    return NextResponse.json({ error: "Missing day" }, { status: 400 });
  }
  const backendUrl = getBackendUrl();
  if (!backendUrl) {
    return NextResponse.json(
      { error: "API no configurada. En el proyecto web (Vercel) configurá API_BACKEND_URL = URL del proyecto API (ej. https://daily-market-brief-api.vercel.app)." },
      { status: 503 }
    );
  }
  // No llamar a la propia web (infinite loop o 404)
  try {
    const backendHost = new URL(backendUrl).hostname;
    if (process.env.VERCEL_URL && backendHost === process.env.VERCEL_URL) {
      return NextResponse.json(
        { error: "API_BACKEND_URL no puede ser la URL de la web. Usá la URL del proyecto API (ej. https://daily-market-brief-api.vercel.app)." },
        { status: 503 }
      );
    }
  } catch (_) {}
  try {
    const res = await fetch(`${backendUrl}/api/summaries/day/${day}`, {
      cache: "no-store",
    });
    if (res.status === 404) {
      return NextResponse.json({ error: "Not found" }, { status: 404 });
    }
    if (!res.ok) {
      return NextResponse.json(
        { error: "Failed to load summary" },
        { status: 502 }
      );
    }
    const summary = (await res.json()) as DaySummary;
    return NextResponse.json(
      { summary, translations: {} },
      { headers: { "Cache-Control": "no-store" } }
    );
  } catch (e) {
    console.error("Day API error:", e);
    return NextResponse.json(
      { error: "Server error" },
      { status: 500 }
    );
  }
}
