import type { Metadata } from "next";
import { apiBaseUrl } from "@/lib/api";
import "./globals.css";

export const metadata: Metadata = {
  title: "Daily Market Brief",
  description: "US market news summaries",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="min-h-screen antialiased">
        <header className="border-b border-slate-700 px-6 py-4">
          <nav className="flex items-center gap-6">
            <a href="/" className="text-xl font-semibold text-white">
              Daily Market Brief
            </a>
            <a href="/" className="text-slate-400 hover:text-white">Calendar</a>
            <a href="/last-10" className="text-slate-400 hover:text-white">Last 10 days</a>
            <a href={`${apiBaseUrl}/admin`} className="text-slate-400 hover:text-white" target="_blank" rel="noopener noreferrer">API status</a>
          </nav>
        </header>
        <main className="mx-auto max-w-5xl px-6 py-8">{children}</main>
      </body>
    </html>
  );
}
