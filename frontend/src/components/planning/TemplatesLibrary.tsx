import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../../lib/api";
import type { WodTemplateItem, TemplateDeleteConflict } from "../../types";
import { THEME_LABELS, FORMAT_LABELS } from "../../types";

// Библиотека тренировок — карточки как в быстром выборе; клик открывает
// редактирование, корзина удаляет (с архивацией, если тренировка выполнялась).
export default function TemplatesLibrary({
  initialSearch = "",
  onEdit,
}: {
  initialSearch?: string;
  onEdit: (templateId: string) => void;
}) {
  const [theme, setTheme] = useState("");
  const [search, setSearch] = useState(initialSearch);
  const [items, setItems] = useState<WodTemplateItem[]>([]);
  const [notice, setNotice] = useState<string | null>(null);
  const [conflict, setConflict] = useState<TemplateDeleteConflict | null>(null);
  const navigate = useNavigate();

  const load = useCallback(() => {
    api.wods
      .templates({ theme: theme || undefined, search: search || undefined, limit: 200, include_archived: true })
      .then(setItems)
      .catch(() => setItems([]));
  }, [theme, search]);

  useEffect(() => {
    const t = setTimeout(load, 250);
    return () => clearTimeout(t);
  }, [load]);

  const remove = async (t: WodTemplateItem) => {
    setNotice(null);
    setConflict(null);
    if (!confirm(`Удалить «${t.name}» из библиотеки?\n\nЕсли тренировка уже выполнялась — она будет перемещена в архив.`)) return;
    try {
      const res = await api.wods.deleteTemplate(t.template_id);
      if (res.archived) {
        setNotice(`«${t.name}» выполнялась ранее — перемещена в архив. Её можно редактировать и вернуть в активные.`);
      }
      load();
    } catch (e) {
      const err = e as Error & { body?: TemplateDeleteConflict };
      if (err.body?.cycles?.length) {
        setConflict(err.body);
      } else {
        setNotice(err.message);
      }
    }
  };

  const restore = async (t: WodTemplateItem) => {
    setNotice(null);
    setConflict(null);
    await api.wods.restoreTemplate(t.template_id);
    setNotice(`«${t.name}» возвращена в активные`);
    load();
  };

  const active = items.filter((t) => !t.archived);
  const archived = items.filter((t) => t.archived);

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

      <p className="text-xs text-slate-400 dark:text-slate-500 mb-4">
        Клик по карточке — редактирование тренировки.
      </p>

      {notice && (
        <div className="bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-500/30 rounded-lg p-3 mb-4 text-sm text-blue-700 dark:text-blue-300">
          {notice}
        </div>
      )}
      {conflict && (
        <div className="bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-500/30 rounded-lg p-3 mb-4 text-sm text-red-700 dark:text-red-300">
          <p className="font-semibold mb-2">{conflict.detail}:</p>
          <div className="flex flex-wrap gap-2">
            {conflict.cycles.map((c) => (
              <button
                key={c.cycle_id + c.slot_id}
                className="underline hover:no-underline"
                onClick={() => {
                  setConflict(null);
                  navigate(`/wod/cycles/${c.cycle_id}`);
                }}
              >
                {c.cycle_name} · {String(c.slot_date).slice(0, 10)}
              </button>
            ))}
          </div>
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {active.map((t) => (
          <div
            key={t.template_id}
            className="bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-xl p-4 flex flex-col cursor-pointer hover:border-emerald-400 dark:hover:border-emerald-500/50 transition-colors"
            onClick={() => onEdit(t.template_id)}
          >
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                {t.is_benchmark && <span className="text-xs font-bold text-amber-500">★ Benchmark</span>}
                <span className="text-xs text-slate-400">{THEME_LABELS[t.theme] || t.theme}</span>
              </div>
              <button
                title="Удалить из библиотеки"
                className="text-slate-300 hover:text-red-500 dark:text-slate-600 dark:hover:text-red-400 text-base leading-none px-1 transition-colors shrink-0"
                onClick={(e) => {
                  e.stopPropagation();
                  remove(t);
                }}
              >
                🗑
              </button>
            </div>
            <div className="font-bold text-slate-900 dark:text-white mb-1">{t.name}</div>
            <div className="text-sm text-slate-500 dark:text-slate-400 mb-2">
              {FORMAT_LABELS[t.format] || t.format} · {t.duration_min} мин ·{" "}
              {t.intensity === "low" ? "Низкая" : t.intensity === "high" ? "Высокая" : "Средняя"}
            </div>
            {t.description && (
              <p className="text-xs text-slate-500 dark:text-slate-400 mb-2">{t.description}</p>
            )}
            <ul className="text-sm text-slate-600 dark:text-slate-300 space-y-1 flex-1">
              {(t.movements || []).map((m, i) => (
                <li key={i}>
                  {m.reps ? `${m.reps} × ` : ""}{m.movement_name}
                  {m.weight_male != null && ` (М ${m.weight_male}кг`}
                  {m.weight_female != null && ` / Ж ${m.weight_female}кг`}
                  {m.weight_male != null && ")"}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
      {active.length === 0 && (
        <p className="text-slate-400 text-center py-8">Ничего не найдено</p>
      )}

      {archived.length > 0 && (
        <div className="mt-8">
          <h2 className="text-lg font-semibold text-slate-600 dark:text-slate-300 mb-3">
            Архивные
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {archived.map((t) => (
              <div
                key={t.template_id}
                className="bg-slate-100 dark:bg-slate-800/30 border border-slate-200 dark:border-slate-700 rounded-xl p-4 flex flex-col opacity-80 cursor-pointer hover:opacity-100 transition-all"
                onClick={() => onEdit(t.template_id)}
              >
                <div className="flex items-center justify-between mb-2">
                  <span className="text-xs font-semibold px-2 py-0.5 rounded-full bg-slate-200 dark:bg-slate-700 text-slate-500 dark:text-slate-400">
                    Архивный
                  </span>
                  <button
                    className="text-emerald-600 dark:text-emerald-400 text-xs font-medium px-1 shrink-0"
                    onClick={(e) => {
                      e.stopPropagation();
                      restore(t);
                    }}
                  >
                    Вернуть в активные
                  </button>
                </div>
                <div className="font-bold text-slate-700 dark:text-slate-200 mb-1">{t.name}</div>
                <div className="text-sm text-slate-500 dark:text-slate-400 mb-2">
                  {FORMAT_LABELS[t.format] || t.format} · {t.duration_min} мин
                </div>
                {t.description && (
                  <p className="text-xs text-slate-500 dark:text-slate-400 mb-2">{t.description}</p>
                )}
                <ul className="text-sm text-slate-500 dark:text-slate-400 space-y-1 flex-1">
                  {(t.movements || []).map((m, i) => (
                    <li key={i}>
                      {m.reps ? `${m.reps} × ` : ""}{m.movement_name}
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
