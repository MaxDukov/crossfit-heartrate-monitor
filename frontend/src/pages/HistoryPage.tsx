import { useEffect, useState, useCallback } from "react";
import type { Session, Wod } from "../types";
import { api } from "../lib/api";
import { FORMAT_LABELS, THEME_LABELS } from "../types";

type HistoryTab = "sessions" | "wods";

export default function HistoryPage() {
  const [tab, setTab] = useState<HistoryTab>("sessions");

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="px-6 pt-4 pb-2 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shrink-0">
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-3">История</h1>
        <div className="flex gap-1">
          {([
            { id: "sessions", label: "Сессии" },
            { id: "wods", label: "Тренировки (WoD)" },
          ] as { id: HistoryTab; label: string }[]).map((t) => (
            <button
              key={t.id}
              onClick={() => setTab(t.id)}
              className={`px-4 py-2 rounded-t-lg text-sm font-medium transition-colors ${
                tab === t.id
                  ? "bg-slate-100 dark:bg-slate-800 text-slate-900 dark:text-white"
                  : "text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>
      <div className="flex-1 min-h-0 overflow-y-auto">
        {tab === "sessions" ? <SessionsHistory /> : <WodsHistory />}
      </div>
    </div>
  );
}

function SessionsHistory() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [active, setActive] = useState<Session | null>(null);
  const [name, setName] = useState("");

  const load = useCallback(async () => {
    const [s, act] = await Promise.all([
      api.sessions.list(),
      api.sessions.active(),
    ]);
    setSessions(s);
    setActive(act);
  }, []);

  useEffect(() => { load(); }, [load]);

  const handleCreate = async () => {
    await api.sessions.create(name || undefined);
    setName("");
    load();
  };

  const handleEnd = async () => {
    if (!active) return;
    await api.sessions.end(active.id);
    load();
  };

  return (
    <div className="max-w-2xl mx-auto p-6">
      {active ? (
        <div className="bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-500/30 rounded-lg p-4 mb-6">
          <div className="flex items-center justify-between mb-2">
            <div>
              <span className="font-medium text-emerald-600 dark:text-emerald-400">
                Активная сессия
              </span>
              <span className="text-slate-500 dark:text-slate-400 ml-2">
                {active.name || "Без названия"}
              </span>
            </div>
            <button
              className="bg-red-600 hover:bg-red-500 px-4 py-1.5 rounded text-sm font-medium"
              onClick={handleEnd}
            >
              Завершить
            </button>
          </div>
          <div className="text-sm text-slate-500 dark:text-slate-400">
            Участников: {active.athlete_count} · Начата:{" "}
            {new Date(active.started_at).toLocaleString("ru")}
          </div>
        </div>
      ) : (
        <div className="flex gap-3 mb-6">
          <input
            className="flex-1 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
            placeholder="Название тренировки"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
          <button
            className="bg-emerald-600 hover:bg-emerald-500 px-4 py-2 rounded text-sm font-medium"
            onClick={handleCreate}
          >
            Начать свободную сессию
          </button>
        </div>
      )}

      <h2 className="text-lg font-semibold mb-3 text-slate-600 dark:text-slate-300">Все сессии</h2>
      <div className="space-y-2">
        {sessions.map((s) => (
          <div
            key={s.id}
            className="bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg p-3 flex items-center justify-between"
          >
            <div>
              <div className="font-medium text-slate-900 dark:text-slate-100">{s.name || "Тренировка"}</div>
              <div className="text-sm text-slate-500 dark:text-slate-400">
                {new Date(s.started_at).toLocaleString("ru")}
                {s.ended_at
                  ? ` → ${new Date(s.ended_at).toLocaleTimeString("ru")}`
                  : " (активна)"}
                {" · "}Участников: {s.athlete_count}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function WodsHistory() {
  const [wods, setWods] = useState<Wod[]>([]);

  useEffect(() => {
    api.wods.history(50).then(setWods).catch(() => setWods([]));
  }, []);

  return (
    <div className="max-w-2xl mx-auto p-6 space-y-2">
      {wods.length === 0 && (
        <p className="text-slate-500 dark:text-slate-400">Тренировок пока не было.</p>
      )}
      {wods.map((w) => (
        <div
          key={w.id}
          className={`bg-white dark:bg-slate-800/50 border rounded-lg p-3 ${
            w.is_active
              ? "border-emerald-400 dark:border-emerald-500/50"
              : "border-slate-200 dark:border-slate-700"
          }`}
        >
          <div className="flex items-center justify-between">
            <div className="font-medium text-slate-900 dark:text-slate-100">
              {w.name}
              {w.is_active && (
                <span className="ml-2 text-xs font-semibold text-emerald-600 dark:text-emerald-400">
                  ● активна
                </span>
              )}
            </div>
            <div className="text-xs text-slate-400">
              {new Date(w.created_at).toLocaleDateString("ru")}
            </div>
          </div>
          <div className="text-sm text-slate-500 dark:text-slate-400 mt-1">
            {FORMAT_LABELS[w.format] || w.format} · {w.duration_min} мин ·{" "}
            {THEME_LABELS[w.theme] || w.theme}
          </div>
          <div className="text-sm text-slate-500 dark:text-slate-400 mt-1">
            {w.movements
              .map((m) => m.movement_name + (m.reps ? ` ×${m.reps}` : ""))
              .join(", ")}
          </div>
        </div>
      ))}
    </div>
  );
}
