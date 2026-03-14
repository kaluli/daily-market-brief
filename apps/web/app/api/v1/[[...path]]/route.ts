import { NextRequest, NextResponse } from "next/server";

const API_ROUTE = "/api";

/**
 * Proxy /api/v1/* to the Go serverless function at /api.
 * The Go handler uses X-Forwarded-Path to route correctly (Vercel rewrites can lose the path).
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

  // En Vercel: llamar a la función Go en /api con el path en header (el rewrite pierde el path).
  // En local: llamar al backend (puerto 3090) con el path completo.
  const isVercel = !!process.env.VERCEL_URL;
  const backend =
    process.env.API_BACKEND_URL || "http://localhost:3090";
  const origin = isVercel
    ? `https://${process.env.VERCEL_URL}`
    : backend;

  const url = new URL(isVercel ? API_ROUTE : pathname, origin);
  url.search = search;

  const headers = new Headers(request.headers);
  headers.delete("host");
  if (isVercel) {
    headers.set("X-Forwarded-Path", pathname);
  }

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
