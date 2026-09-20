"use client";

import { useState } from "react";
import { apiAdminHref } from "@/lib/api";

const LINKS: { href: string; label: string; external?: boolean }[] = [
  { href: "/", label: "Calendar" },
  { href: "/last-10", label: "Last 10 days" },
  { href: "/agents", label: "Agentes" },
  { href: "/glosario", label: "Glosario" },
];

export default function Header() {
  const [open, setOpen] = useState(false);

  return (
    <header className="border-b border-slate-700 px-4 py-3 sm:px-6 sm:py-4">
      <div className="flex items-center justify-between gap-4">
        <a
          href="/"
          className="truncate text-lg font-semibold text-white sm:text-xl"
          onClick={() => setOpen(false)}
        >
          Daily Market Brief
        </a>

        <nav className="hidden items-center gap-6 md:flex">
          {LINKS.map((link) => (
            <a key={link.href} href={link.href} className="text-slate-400 hover:text-white">
              {link.label}
            </a>
          ))}
          <a
            href={apiAdminHref()}
            className="text-slate-400 hover:text-white"
            target="_blank"
            rel="noopener noreferrer"
          >
            API status
          </a>
        </nav>

        <button
          type="button"
          onClick={() => setOpen((v) => !v)}
          aria-expanded={open}
          aria-label={open ? "Cerrar menú" : "Abrir menú"}
          className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-slate-700 text-slate-300 hover:bg-slate-800 md:hidden"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth={2}
            strokeLinecap="round"
            strokeLinejoin="round"
            className="h-5 w-5"
          >
            {open ? (
              <path d="M18 6 6 18M6 6l12 12" />
            ) : (
              <path d="M3 6h18M3 12h18M3 18h18" />
            )}
          </svg>
        </button>
      </div>

      {open && (
        <nav className="mt-3 flex flex-col gap-1 border-t border-slate-700 pt-3 md:hidden">
          {LINKS.map((link) => (
            <a
              key={link.href}
              href={link.href}
              onClick={() => setOpen(false)}
              className="rounded-md px-2 py-2 text-slate-300 hover:bg-slate-800 hover:text-white"
            >
              {link.label}
            </a>
          ))}
          <a
            href={apiAdminHref()}
            target="_blank"
            rel="noopener noreferrer"
            onClick={() => setOpen(false)}
            className="rounded-md px-2 py-2 text-slate-300 hover:bg-slate-800 hover:text-white"
          >
            API status
          </a>
        </nav>
      )}
    </header>
  );
}
