import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../../lib/api";
import type { WodVariant } from "../../types";
import { THEME_LABELS, THEME_ICONS, LEVEL_LABELS } from "../../types";

const THEMES = Object.keys(THEME_LABELS);

// Быстрый выбор WoD без планирования — как прежний конструктор-подборщик.
export default function QuickPick() {
  const [theme, setTheme] = useState("full_body");
  const [level, setLevel] = useState("intermediate");
  const [variants, setVariants] = useState<WodVariant[]>([]);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const generate = async (t = theme) => {
    setBusy(true);
    setError(null);
    try {
      const v = await api.wods.generate(t, level);
      setVariants(v);
    } catch (e) {
      setError((e as Error).message);
      setVariants([]);
    } finally {
      setBusy(false);
    }
  };

  useEffect(() => { generate(); /* eslint-disable-next-line */ }, []);

  const pick = async (templateId: string) => {
    await api.wods.select(templateId, level);
    navigate("/");
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <p className="text-slate-500 dark:text-slate-400 mb-4">
        Мгновенный подбор тренировки на сегодня — 3 варианта по теме и уровню группы.
      </p>

      <div className="grid grid-cols-4 gap-2 mb-4">
        {THEMES.map((t) => (
          <button
            key={t}
            onClick={() => { setTheme(t); generate(t); }}
            className={`rounded-lg p-3 text-sm font-medium transition-colors ${
              theme === t
                ? "bg-emerald-600 text-white"
                : "bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
            }`}
          >
            <span className="mr-1">{THEME_ICONS[t]}</span> {THEME_LABELS[t]}
          </button>
        ))}
      </div>

      <div className="flex items-center gap-2 mb-6">
        <span className="text-sm text-slate-500 dark:text-slate-400">Уровень группы:</span>
        {Object.entries(LEVEL_LABELS).map(([id, label]) => (
          <button
            key={id}
            onClick={() => setLevel(id)}
            className={`px-3 py-1.5 rounded text-sm ${
              level === id
                ? "bg-slate-800 dark:bg-slate-200 text-white dark:text-slate-900 font-semibold"
                : "bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400"
            }`}
          >
            {label}
          </button>
        ))}
      </div>

      {error && <div className="text-red-500 text-sm mb-4">{error}</div>}
      {busy && <div className="text-slate-400">Генерация…</div>}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        {variants.map((v) => (
          <div
            key={v.template_id}
            className="bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-xl p-4 flex flex-col"
          >
            <div className="flex items-center gap-2 mb-2">
              {v.is_benchmark && (
                <span className="text-xs font-bold text-amber-500">★ Benchmark</span>
              )}
              <span className="text-xs text-slate-400">{THEME_LABELS[v.theme]}</span>
            </div>
            <div className="font-bold text-slate-900 dark:text-white mb-1">{v.name}</div>
            <div className="text-sm text-slate-500 dark:text-slate-400 mb-3">
              {v.format_name} · {v.duration_min} мин · {v.intensity_name}
            </div>
            <ul className="text-sm text-slate-600 dark:text-slate-300 space-y-1 mb-4 flex-1">
              {v.movements.map((m, i) => (
                <li key={i}>
                  {m.reps ? `${m.reps} × ` : ""}{m.movement_name}
                  {m.weight_male != null && ` (М ${m.weight_male}кг`}
                  {m.weight_female != null && ` / Ж ${m.weight_female}кг`}
                  {m.weight_male != null && ")"}
                </li>
              ))}
            </ul>
            <button
              className="bg-emerald-600 hover:bg-emerald-500 px-4 py-2 rounded text-sm font-medium"
              onClick={() => pick(v.template_id)}
            >
              Запустить сейчас
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
