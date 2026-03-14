/**
 * Server-only: call MyMemory API to translate English text to Spanish.
 * Used by API routes (no CORS). Max 500 chars per request.
 */
const MYMEMORY_URL = "https://api.mymemory.translated.net/get";
const MAX_LENGTH = 500;

const QUOTA_MARKERS = [
  "MYMEMORY WARNING",
  "YOU USED ALL AVAILABLE FREE TRANSLATIONS",
  "quotareached",
  "quota",
];

/** Parses "NEXT AVAILABLE IN 17 HOURS 53 MINUTES 14 SECONDS" → "17 h 53 min" */
function parseRetryAfter(msg: string): string {
  const hoursMatch = msg.match(/(\d+)\s*HOURS?/i);
  const minsMatch = msg.match(/(\d+)\s*MINUTES?/i);
  const hours = hoursMatch ? parseInt(hoursMatch[1], 10) : 0;
  const mins = minsMatch ? parseInt(minsMatch[1], 10) : 0;
  if (hours > 0 && mins > 0) return `${hours} h ${mins} min`;
  if (hours > 0) return `${hours} h`;
  if (mins > 0) return `${mins} min`;
  return "unas horas";
}

function isQuotaMessage(text: string): boolean {
  const lower = text.toUpperCase();
  return QUOTA_MARKERS.some((m) => lower.includes(m.toUpperCase()));
}

export type TranslateResult = {
  translated: string;
  quotaExceeded?: boolean;
  retryAfter?: string;
};

export async function translateToSpanish(text: string): Promise<TranslateResult> {
  const t = typeof text === "string" ? text.trim() : "";
  if (!t) return { translated: "" };
  const toTranslate = t.length > MAX_LENGTH ? t.slice(0, MAX_LENGTH) : t;
  try {
    const res = await fetch(
      `${MYMEMORY_URL}?q=${encodeURIComponent(toTranslate)}&langpair=en|es`,
      { cache: "no-store" }
    );
    const data = (await res.json()) as {
      responseData?: { translatedText?: string };
      responseDetails?: string;
    };
    const raw = data.responseData?.translatedText ?? data.responseDetails ?? "";
    if (typeof raw !== "string") return { translated: "" };
    if (isQuotaMessage(raw)) {
      return {
        translated: "",
        quotaExceeded: true,
        retryAfter: parseRetryAfter(raw),
      };
    }
    return { translated: raw };
  } catch {
    return { translated: "" };
  }
}
