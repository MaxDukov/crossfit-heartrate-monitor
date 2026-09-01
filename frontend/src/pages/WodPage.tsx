import { useEffect, useState, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../lib/api";
import type { CycleSummary } from "../types";
import CycleForm from "../components/planning/CycleForm";
import QuickPick from "../components/planning/QuickPick";
import Constructor from "../components/planning/Constructor";
import TemplatesLibrary from "../components/planning/TemplatesLibrary";

type WodTab = "cycles" | "quick" | "constructor" | "library";

const TABS: { id: WodTab; label: string }[] = [
  { id: "cycles", label: "Планирование" },
  { id: "quick", label: "Быстрый выбор" },
  { id: "constructor", label: "Конструктор" },
  { id: "library", label: "Библиотека" },
];

export default function WodPage() {
  const [tab, setTab] = useState<WodTab>("cycles");

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="px-6 pt-4 pb-2 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shrink-0">
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-3">WoD</h1>
        <div className="flex gap-1">
          {TABS.map((t) => (
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
        {tab === "cycles" && <CyclesList />}
        {tab === "quick" && <QuickPick />}
        {tab === "constructor" && <Constructor />}
        {tab === "library" && <TemplatesLibrary />}
      </div>
    </div>
  );
}

function CyclesList() {
  const [cycles, setCycles] = useState<CycleSummary[]>([]);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const load = useCallback(async () => {
    try {
      setCycles(await api.cycles.list());
    } catch (e) {
      setError((e as Error).message);
    }
  }, []);

  useEffect(() => { load(); }, [load]);

  const handleDelete = async (id: string, name: string) => {
    if (!confirm(`Удалить цикл «${name}»? Слоты и планы будут удалены, результаты останутся.`)) return;
    await api.cycles.delete(id);
    load();
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="flex items-center justify-between mb-6">
        <p className="text-slate-500 dark:text-slate-400">
          Макроциклы: стратегия → календарь → проведение → аналитика
        </p>
        <button
          className="bg-emerald-600 hover:bg-emerald-500 px-4 py-2 rounded text-sm font-medium shrink-0"
          onClick={() => setCreating(true)}
        >
          + Новый цикл
        </button>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-500/30 rounded-lg p-3 mb-4 text-sm text-red-700 dark:text-red-300">
          {error}
        </div>
      )}

      {cycles.length === 0 && !error && (
        <div className="text-center py-16 text-slate-400 dark:text-slate-500">
          <p className="text-lg mb-2">Циклов ещё нет</p>
          <p className="text-sm">
            Создайте первый макроцикл: цель, длительность, группы и дни недели.
          </p>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {cycles.map((c) => (
          <div
            key={c.id}
            className="bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-xl p-4 hover:border-emerald-400 dark:hover:border-emerald-500/50 cursor-pointer transition-colors"
            onClick={() => navigate(`/wod/cycles/${c.id}`)}
          >
            <div className="flex items-start justify-between mb-2">
              <div>
                <div className="font-bold text-slate-900 dark:text-white">{c.name}</div>
                <div className="text-sm text-slate-500 dark:text-slate-400">
                  {c.goal || "Без цели"}
                </div>
              </div>
              <span
                className={`text-xs font-semibold px-2 py-1 rounded-full shrink-0 ${
                  c.status === "active"
                    ? "bg-emerald-100 dark:bg-emerald-900/40 text-emerald-700 dark:text-emerald-300"
                    : c.status === "completed"
                      ? "bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300"
                      : "bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300"
                }`}
              >
                {c.status === "active" ? "активен" : c.status === "completed" ? "завершён" : "план"}
              </span>
            </div>
            <div className="text-sm text-slate-500 dark:text-slate-400 mb-3">
              {c.weeks} нед. · с {c.start_date.slice(0, 10)} · групп: {c.groups_count}
            </div>
            <div className="flex items-center gap-2">
              <div className="flex-1 h-2 bg-slate-100 dark:bg-slate-700 rounded-full overflow-hidden">
                <div
                  className="h-full bg-emerald-500"
                  style={{ width: `${Math.min(100, c.fill_percent)}%` }}
                />
              </div>
              <span className="text-xs text-slate-400 shrink-0">
                заполнено {Math.round(c.fill_percent)}% · проведено {c.slots_completed}/{c.slots_total}
              </span>
            </div>
            <div
              className="mt-3 text-xs text-red-500 hover:text-red-400"
              onClick={(e) => {
                e.stopPropagation();
                handleDelete(c.id, c.name);
              }}
            >
              удалить
            </div>
          </div>
        ))}
      </div>

      {creating && (
        <CycleForm
          onClose={() => setCreating(false)}
          onCreated={(id) => {
            setCreating(false);
            navigate(`/wod/cycles/${id}`);
          }}
        />
      )}
    </div>
  );
}
