"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import {
  Radar,
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  ResponsiveContainer,
} from "recharts";

type ModuleResult = {
  module: string;
  status: string;
  summary: string;
  details?: unknown;
};

type ScanJob = {
  id: string;
  url: string;
  status: string;
  created_at: string;
  results: Record<string, ModuleResult>;
  progress: string[];
};

const STATUS_COLOR: Record<string, string> = {
  ok: "text-green-400",
  warn: "text-yellow-400",
  error: "text-red-400",
  stub: "text-slate-400",
  running: "text-sky-400",
  queued: "text-slate-400",
  done: "text-green-400",
};

// Calculate abstract security score based on status
function getScore(status: string) {
  if (status === "ok") return 100;
  if (status === "warn") return 50;
  if (status === "error") return 10;
  return 0;
}

export default function ScanPage() {
  const { id } = useParams<{ id: string }>();
  const [job, setJob] = useState<ScanJob | null>(null);
  const [progress, setProgress] = useState<string[]>([]);
  const [error, setError] = useState("");

  // Poll for result
  useEffect(() => {
    if (!id) return;

    const poll = async () => {
      try {
        const res = await fetch(`/api/scan/${id}`);
        if (!res.ok) throw new Error("Tarama bulunamadı.");
        const data: ScanJob = await res.json();
        setJob(data);
        if (data.status !== "done" && data.status !== "error") {
          setTimeout(poll, 1500);
        }
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : "Hata");
      }
    };

    poll();
  }, [id]);

  // SSE for live progress
  useEffect(() => {
    if (!id) return;
    const es = new EventSource(`/api/scan/${id}/stream`);
    es.onmessage = (e) => {
      if (e.data.startsWith("[DONE]")) {
        es.close();
        return;
      }
      setProgress((prev) => [...prev, e.data]);
    };
    es.onerror = () => es.close();
    return () => es.close();
  }, [id]);

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-red-400">⚠ {error}</p>
      </div>
    );
  }

  if (!job) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="w-8 h-8 border-2 border-sky-500 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-slate-400">Tarama başlatılıyor…</p>
        </div>
      </div>
    );
  }

  const modules = Object.values(job.results ?? {});

  // Prepare radar chart data
  const radarData = modules.map((m) => ({
    subject: m.module.toUpperCase(),
    score: getScore(m.status),
    fullMark: 100,
  }));

  return (
    <main className="min-h-screen px-4 py-12 max-w-4xl mx-auto">
      {/* Header */}
      <div className="mb-8">
        <a href="/" className="text-sky-400 text-sm hover:underline">
          ← Yeni Tarama
        </a>
        <h1 className="text-2xl font-bold text-white mt-3 font-mono break-all">
          {job.url}
        </h1>
        <div className="flex items-center gap-3 mt-2">
          <span
            className={`text-sm font-medium ${STATUS_COLOR[job.status] ?? "text-slate-400"}`}
          >
            ● {job.status.toUpperCase()}
          </span>
          <span className="text-slate-500 text-sm">
            {new Date(job.created_at).toLocaleString("tr-TR")}
          </span>
          <span className="text-slate-500 text-sm font-mono">#{id}</span>
        </div>
      </div>

      {/* Radar Chart (F09) */}
      {job.status === "done" && radarData.length > 2 && (
        <div className="bg-slate-900 border border-slate-700 rounded-xl p-6 mb-8 h-[350px] w-full flex flex-col items-center">
          <h2 className="text-slate-300 font-semibold mb-2">Güvenlik Skoru Dağılımı</h2>
          <ResponsiveContainer width="100%" height="100%">
            <RadarChart cx="50%" cy="50%" outerRadius="80%" data={radarData}>
              <PolarGrid stroke="#334155" />
              <PolarAngleAxis dataKey="subject" tick={{ fill: "#94a3b8", fontSize: 12 }} />
              <PolarRadiusAxis angle={30} domain={[0, 100]} tick={false} axisLine={false} />
              <Radar
                name="Score"
                dataKey="score"
                stroke="#38bdf8"
                fill="#38bdf8"
                fillOpacity={0.4}
              />
            </RadarChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Live Progress */}
      {progress.length > 0 && job.status !== "done" && (
        <div className="bg-slate-900 border border-slate-700 rounded-xl p-4 mb-8 font-mono text-xs text-slate-400 max-h-40 overflow-y-auto">
          {progress.map((line, i) => (
            <div key={i}>{line}</div>
          ))}
        </div>
      )}

      {/* Module Results */}
      {modules.length > 0 && (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-8">
          {modules.map((mod) => (
            <div
              key={mod.module}
              className="bg-slate-800/50 border border-slate-700 rounded-xl p-5"
            >
              <div className="flex items-center justify-between mb-2">
                <span className="font-mono font-semibold text-white text-sm">
                  {mod.module}
                </span>
                <span
                  className={`text-xs font-medium ${STATUS_COLOR[mod.status] ?? "text-slate-400"}`}
                >
                  {mod.status}
                </span>
              </div>
              <p className="text-slate-400 text-sm">{mod.summary}</p>
            </div>
          ))}
        </div>
      )}

      {/* Empty state while running */}
      {modules.length === 0 && job.status === "running" && (
        <div className="text-center py-16 text-slate-500">
          <div className="w-8 h-8 border-2 border-sky-500 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          Modüller çalışıyor… (SSE Stream Aktif)
        </div>
      )}

      {/* PDF Download (F09) */}
      {job.status === "done" && (
        <div className="mt-6 flex justify-end gap-3 border-t border-slate-800 pt-6">
          <a
            href={`/api/scan/${id}/report.pdf`}
            target="_blank"
            rel="noopener noreferrer"
            className="bg-slate-700 hover:bg-slate-600 text-white text-sm font-medium px-5 py-2.5 rounded-xl transition-colors shadow-lg"
          >
            📄 PDF İndir
          </a>
        </div>
      )}
    </main>
  );
}
