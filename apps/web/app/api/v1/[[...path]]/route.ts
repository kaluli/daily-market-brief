import { NextRequest, NextResponse } from "next/server";

/**
 * Proxy /api/v1/* to the API Go (API_BACKEND_URL). En Vercel sin API_BACKEND_URL devuelve 503.
 */
export async function GET(request: NextRequest) {
  return proxyToGo(request);
}

export async function POST(request: NextRequest) {
  return proxyToGo(request);
}

export async function PUT(request: NextRequest) {
  return proxyToGo(request);
}

export async function PATCH(request: NextRequest) {
  return proxyToGo(request);
}

export async function DELETE(request: NextRequest) {
  return proxyToGo(request);
}

async function proxyToGo(request: NextRequest) {
  const pathname = request.nextUrl.pathname;
  const search = request.nextUrl.search;

  // En Vercel sin API_BACKEND_URL no hay backend: devolver 503 (evita bucle y fetch a localhost).
  const backend =
    process.env.API_BACKEND_URL ||
    (process.env.VERCEL_URL ? null : "http://localhost:3090");

  if (!backend) {
    return NextResponse.json(
      { error: "API no configurada. Configurá API_BACKEND_URL o NEXT_PUBLIC_API_URL en Vercel." },
      { status: 503 }
    );
  }

  const localPath = pathname.replace(/^\/api\/v1/, "") || "/";
  const url = new URL(localPath, backend.replace(/\/$/, ""));
  url.search = search;

  const headers = new Headers(request.headers);
  headers.delete("host");

  let body: string | undefined;
  try {
    body = await request.text();
  } catch {
    // no body
  }

  const res = await fetch(url.toString(), {
    method: request.method,
    headers,
    body: body && body.length > 0 ? body : undefined,
  });

  const resHeaders = new Headers(res.headers);
  resHeaders.delete("content-encoding");

  return new NextResponse(res.body, {
    status: res.status,
    statusText: res.statusText,
    headers: resHeaders,
  });
}
