import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../../lib/api";
import type { WodTemplateItem } from "../../types";
import { THEME_LABELS, FORMAT_LABELS } from "../../types";

// Библиотека тренировок — все шаблоны (включая созданные конструктором).
export default function TemplatesLibrary() {
  const [theme, setTheme] = useState("");
  const [search, setSearch] = useState("");
  const [items, setItems] = useState<WodTemplateItem[]>([]);

  useEffect(() => {
    const t = setTimeout(() => {
      api.wods
        .templates({ theme: theme || undefined, search: search || undefined, limit: 200 })
        .then(setItems)
        .catch(() => setItems([]));
    }, 250);
    return () => clearTimeout(t);
  }, [theme, search]);

  const navigate = useNavigate();

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="flex gap-3 mb-4">
        <input
          className="flex-1 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
          placeholder="Поиск по названию…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <select
          className="bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
          value={theme}
          onChange={(e) => setTheme(e.target.value)}
        >
          <option value="">Все темы</option>
          {Object.entries(THEME_LABELS).map(([id, label]) => (
            <option key={id} value={id}>{label}</option>
          ))}
        </select>
      </div>

      <div className="space-y-2">
        {items.map((t) => (
          <button
            key={t.template_id}
            className="w-full text-left bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg p-3 hover:border-emerald-400 dark:hover:border-emerald-500/50 transition-colors"
            onClick={async () => {
              await api.wods.select(t.template_id, "intermediate");
              navigate("/");
            }}
          >
            <div className="flex items-center justify-between">
              <div className="font-medium text-slate-900 dark:text-slate-100">
                {t.is_benchmark && <span className="text-amber-500 mr-1">★</span>}
                {t.name}
              </div>
              <div className="text-xs text-slate-400">
                {FORMAT_LABELS[t.format] || t.format} · {t.duration_min} мин ·{" "}
                {t.movements_count} движ.
              </div>
            </div>
            <div className="text-xs text-slate-400 mt-1">{THEME_LABELS[t.theme] || t.theme}</div>
          </button>
        ))}
        {items.length === 0 && (
          <p className="text-slate-400 text-center py-8">Ничего не найдено</p>
        )}
      </div>
    </div>
  );
}
