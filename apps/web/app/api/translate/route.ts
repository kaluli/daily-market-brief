import { NextRequest, NextResponse } from "next/server";
import { translateToSpanish } from "@/lib/translate-server";

const headers = { "Cache-Control": "no-store" } as const;

/** Server-side proxy to MyMemory so browser avoids CORS. */
export async function GET(request: NextRequest) {
  const raw = request.nextUrl.searchParams.get("text");
  const text = typeof raw === "string" ? raw.trim() : "";
  const result = text ? await translateToSpanish(text) : { translated: "" };
  return NextResponse.json(
    {
      translated: result.translated,
      ...(result.quotaExceeded && {
        quotaExceeded: true,
        retryAfter: result.retryAfter ?? "unas horas",
      }),
    },
    { headers }
  );
}
