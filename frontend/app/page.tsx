"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

const MODULES = ["ports", "headers", "tls", "fuzz", "xss", "sqli", "cve"] as const;

export default function Home() {
  const [url, setUrl] = useState("");
  const [selectedModules, setSelectedModules] = useState<string[]>([...MODULES]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const router = useRouter();

  const toggleModule = (mod: string) => {
    setSelectedModules((prev) =>
      prev.includes(mod) ? prev.filter((m) => m !== mod) : [...prev, mod]
    );
  };

  const handleScan = async () => {
    if (!url) {
      setError("Lütfen bir URL girin.");
      return;
    }
    setError("");
    setLoading(true);

    try {
      const res = await fetch("/api/scan", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ url, modules: selectedModules }),
      });

      if (!res.ok) {
        const data = await res.json();
        throw new Error(data.error || "Tarama başlatılamadı.");
      }

      const data = await res.json();
      router.push(`/scan/${data.scan_id}`);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Bilinmeyen hata");
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="min-h-screen flex flex-col items-center justify-center px-4 py-16">
      {/* Hero */}
      <div className="text-center mb-12">
        <div className="inline-flex items-center gap-2 bg-sky-500/10 border border-sky-500/30 rounded-full px-4 py-1 text-sky-400 text-sm font-medium mb-6">
          <span className="w-2 h-2 rounded-full bg-sky-400 animate-pulse" />
          7 Güvenlik Modülü — Paralel Analiz
        </div>
        <h1 className="text-5xl md:text-6xl font-bold text-white mb-4 tracking-tight">
          Sec<span className="text-sky-400">Scan</span>
        </h1>
        <p className="text-slate-400 text-lg max-w-xl mx-auto">
          Bir URL girin — portlar, başlıklar, TLS, XSS, SQLi ve CVE&apos;ler
          aynı anda taransın.
        </p>
      </div>

      {/* Scan Card */}
      <div className="w-full max-w-2xl bg-slate-800/50 border border-slate-700 rounded-2xl p-8 shadow-2xl backdrop-blur-sm">
        {/* URL Input */}
        <div className="flex gap-3 mb-6">
          <input
            id="url-input"
            type="url"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleScan()}
            placeholder="https://example.com"
            className="flex-1 bg-slate-900 border border-slate-600 rounded-xl px-4 py-3 text-white placeholder-slate-500 font-mono text-sm focus:outline-none focus:border-sky-500 transition-colors"
          />
          <button
            id="scan-btn"
            onClick={handleScan}
            disabled={loading}
            className="bg-sky-500 hover:bg-sky-400 disabled:opacity-50 disabled:cursor-not-allowed text-white font-semibold px-6 py-3 rounded-xl transition-colors whitespace-nowrap"
          >
            {loading ? "Başlatılıyor…" : "Tara →"}
          </button>
        </div>

        {/* Module Toggles */}
        <div>
          <p className="text-slate-400 text-xs font-medium uppercase tracking-wider mb-3">
            Modüller
          </p>
          <div className="flex flex-wrap gap-2">
            {MODULES.map((mod) => (
              <button
                key={mod}
                id={`module-${mod}`}
                onClick={() => toggleModule(mod)}
                className={`px-3 py-1.5 rounded-lg text-xs font-mono font-medium transition-all ${
                  selectedModules.includes(mod)
                    ? "bg-sky-500/20 border border-sky-500/50 text-sky-300"
                    : "bg-slate-700 border border-slate-600 text-slate-400"
                }`}
              >
                {mod}
              </button>
            ))}
          </div>
        </div>

        {/* Error */}
        {error && (
          <p className="mt-4 text-red-400 text-sm bg-red-500/10 border border-red-500/30 rounded-lg px-4 py-2">
            ⚠ {error}
          </p>
        )}
      </div>

      {/* Feature Pills */}
      <div className="mt-10 flex flex-wrap justify-center gap-3 text-sm text-slate-500">
        {[
          "🔒 SSRF Korumalı",
          "⚡ Goroutine ile Paralel",
          "📊 Radar Raporu",
          "📄 PDF Export",
        ].map((f) => (
          <span
            key={f}
            className="bg-slate-800 border border-slate-700 rounded-full px-4 py-1"
          >
            {f}
          </span>
        ))}
      </div>
    </main>
  );
}
