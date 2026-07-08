import { useEffect, useState, useCallback } from "react";
import type { Session } from "../types";
import { api } from "../lib/api";

export default function SessionsPage() {
  const [sessions, setSessions] = useState<Session[]>([]);
  const [active, setActive] = useState<Session | null>(null);
  const [name, setName] = useState("");

  const load = useCallback(async () => {
    const [s, , act] = await Promise.all([
      api.sessions.list(),
      api.athletes.list(),
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
      <h1 className="text-2xl font-bold mb-6 text-slate-900 dark:text-slate-100">Тренировки</h1>

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
            Начать тренировку
          </button>
        </div>
      )}

      <h2 className="text-lg font-semibold mb-3 text-slate-600 dark:text-slate-300">История</h2>
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
                  ? ` → ${new Date(s.ended_at).toLocaleString("ru")}`
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
