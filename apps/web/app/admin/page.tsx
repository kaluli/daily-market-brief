"use client";

import { useCallback, useEffect, useState } from "react";
import Link from "next/link";
import {
  getAdminSources,
  setSourceEnabled,
  type AdminSourcesResponse,
} from "@/lib/api";

export default function AdminSourcesPage() {
  const [data, setData] = useState<AdminSourcesResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [togglingId, setTogglingId] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await getAdminSources();
      setData(res);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load sources");
      setData(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const handleToggle = useCallback(
    async (id: string, currentEnabled: boolean) => {
      if (togglingId) return;
      setTogglingId(id);
      try {
        await setSourceEnabled(id, !currentEnabled);
        await load();
      } catch (e) {
        setError(e instanceof Error ? e.message : "Failed to update");
      } finally {
        setTogglingId(null);
      }
    },
    [load, togglingId]
  );

  if (loading) {
    return (
      <div className="space-y-4">
        <Link href="/" className="text-blue-400 hover:underline">
          ← Calendar
        </Link>
        <p className="text-slate-500">Loading sources...</p>
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="space-y-4">
        <Link href="/" className="text-blue-400 hover:underline">
          ← Calendar
        </Link>
        <p className="text-red-400">
          {error ?? "No se pudieron cargar las fuentes."}
        </p>
      </div>
    );
  }

  const sources = data.sources ?? [];
  const rssSources = data.rss_sources ?? [];

  return (
    <div className="space-y-8">
      <div>
        <Link href="/" className="text-blue-400 hover:underline">
          ← Calendar
        </Link>
        <h1 className="mt-2 text-2xl font-bold">Fuentes de noticias</h1>
        <p className="mt-1 text-sm text-slate-400">
          Activa o desactiva fuentes para la ingesta. Los cambios se guardan en{" "}
          <code className="rounded bg-slate-700 px-1">config/news_sources.json</code>.
        </p>
      </div>

      <section className="rounded-lg border border-slate-700 bg-slate-800/50 p-6">
        <h2 className="mb-4 text-lg font-semibold">Fuentes API (NewsAPI, Finnhub)</h2>
        <ul className="space-y-3">
          {sources.length === 0 ? (
            <li className="text-slate-500">No hay fuentes API configuradas.</li>
          ) : (
            sources.map((src) => (
              <SourceRow
                key={src.id}
                id={src.id}
                name={src.name}
                enabled={src.enabled}
                extra={`Peso: ${src.weight}`}
                togglingId={togglingId}
                onToggle={() => handleToggle(src.id, src.enabled)}
              />
            ))
          )}
        </ul>
      </section>

      <section className="rounded-lg border border-slate-700 bg-slate-800/50 p-6">
        <h2 className="mb-4 text-lg font-semibold">Fuentes RSS</h2>
        <ul className="space-y-3">
          {rssSources.length === 0 ? (
            <li className="text-slate-500">No hay fuentes RSS configuradas.</li>
          ) : (
            rssSources.map((src) => (
              <SourceRow
                key={src.id}
                id={src.id}
                name={src.name}
                enabled={src.enabled}
                extra={src.url}
                togglingId={togglingId}
                onToggle={() => handleToggle(src.id, src.enabled)}
              />
            ))
          )}
        </ul>
      </section>
    </div>
  );
}

function SourceRow({
  id,
  name,
  enabled,
  extra,
  togglingId,
  onToggle,
}: {
  id: string;
  name: string;
  enabled: boolean;
  extra: string;
  togglingId: string | null;
  onToggle: () => void;
}) {
  const busy = togglingId === id;
  return (
    <li className="flex flex-wrap items-center justify-between gap-4 rounded border border-slate-700 bg-slate-800 p-4">
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="font-medium text-white">{name}</span>
          <span
            className={`inline-flex rounded px-2 py-0.5 text-xs font-medium ${
              enabled
                ? "bg-emerald-900/50 text-emerald-300"
                : "bg-slate-700 text-slate-400"
            }`}
          >
            {enabled ? "Activa" : "Inactiva"}
          </span>
        </div>
        <p className="mt-1 truncate text-sm text-slate-400" title={extra}>
          {id} · {extra}
        </p>
      </div>
      <button
        type="button"
        onClick={onToggle}
        disabled={busy}
        className={`rounded px-4 py-2 text-sm font-medium transition ${
          enabled
            ? "bg-slate-600 text-white hover:bg-slate-500"
            : "bg-blue-600 text-white hover:bg-blue-500"
        } disabled:opacity-50`}
      >
        {busy ? "..." : enabled ? "Desactivar" : "Activar"}
      </button>
    </li>
  );
}
