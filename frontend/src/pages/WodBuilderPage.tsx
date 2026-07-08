import { useState } from "react";
import { api } from "../lib/api";
import type { WodVariant } from "../types";
import {
  THEME_LABELS,
  THEME_ICONS,
  LEVEL_LABELS,
  FORMAT_LABELS,
} from "../types";
import { useNavigate } from "react-router-dom";

const THEMES = Object.keys(THEME_LABELS);
const LEVELS = ["beginner", "intermediate", "advanced", "elite"];

export default function WodBuilderPage() {
  const navigate = useNavigate();
  const [selectedTheme, setSelectedTheme] = useState<string | null>(null);
  const [level, setLevel] = useState("intermediate");
  const [variants, setVariants] = useState<WodVariant[] | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const generate = async (theme: string) => {
    setSelectedTheme(theme);
    setLoading(true);
    setError(null);
    setVariants(null);
    try {
      const result = await api.wods.generate(theme, level);
      setVariants(result);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка генерации");
    } finally {
      setLoading(false);
    }
  };

  const selectWod = async (templateId: string) => {
    try {
      await api.wods.select(templateId, level);
      navigate("/");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка выбора");
    }
  };

  return (
    <div className="h-full overflow-y-auto p-6">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-1">Конструктор WoD</h1>
        <p className="text-slate-500 dark:text-slate-400 mb-6">
          Выберите тему дня — система предложит 3 варианта тренировки
        </p>

        {/* Уровень группы */}
        <div className="mb-6">
          <label className="block text-sm font-medium text-slate-600 dark:text-slate-300 mb-2">
            Уровень группы
          </label>
          <div className="flex gap-2">
            {LEVELS.map((l) => (
              <button
                key={l}
                onClick={() => setLevel(l)}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                  level === l
                    ? "bg-blue-600 text-white"
                    : "bg-white dark:bg-slate-800 text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-700"
                }`}
              >
                {LEVEL_LABELS[l]}
              </button>
            ))}
          </div>
        </div>

        {/* Темы */}
        <div className="mb-8">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            {THEMES.map((theme) => (
              <button
                key={theme}
                onClick={() => generate(theme)}
                className={`flex flex-col items-center gap-2 p-5 rounded-xl border-2 transition-all ${
                  selectedTheme === theme
                    ? "bg-blue-50 dark:bg-blue-600/20 border-blue-600 scale-105"
                    : "bg-white dark:bg-slate-900 border-slate-200 dark:border-slate-800 hover:border-slate-300 dark:hover:border-slate-600"
                }`}
              >
                <span className="text-4xl">{THEME_ICONS[theme]}</span>
                <span className="text-sm font-medium text-slate-700 dark:text-slate-200">
                  {THEME_LABELS[theme]}
                </span>
              </button>
            ))}
          </div>
        </div>

        {/* Загрузка */}
        {loading && (
          <div className="flex items-center justify-center py-12">
            <div className="text-slate-400 dark:text-slate-400 text-lg">
              Генерация тренировок...
            </div>
          </div>
        )}

        {/* Ошибка */}
        {error && (
          <div className="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-4 text-red-700 dark:text-red-300 mb-4">
            {error}
          </div>
        )}

        {/* Варианты */}
        {variants && variants.length > 0 && (
          <div>
            <h2 className="text-lg font-semibold text-slate-700 dark:text-slate-200 mb-4">
              3 варианта тренировки
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              {variants.map((v, i) => (
                <div
                  key={i}
                  className="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 p-5 flex flex-col"
                >
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      {v.is_benchmark && (
                        <span className="inline-block text-xs bg-amber-600/20 text-amber-400 px-2 py-0.5 rounded mb-1">
                          ★ Бенчмарк
                        </span>
                      )}
                      <h3 className="text-xl font-bold text-slate-900 dark:text-white">
                        {v.name}
                      </h3>
                    </div>
                  </div>

                  <div className="flex gap-2 mb-3 flex-wrap">
                    <span className="text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 px-2 py-1 rounded">
                      {FORMAT_LABELS[v.format] || v.format_name}
                    </span>
                    <span className="text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 px-2 py-1 rounded">
                      ~{v.duration_min} мин
                    </span>
                    <span className="text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 px-2 py-1 rounded">
                      {v.intensity_name}
                    </span>
                  </div>

                  <p className="text-sm text-slate-500 dark:text-slate-400 mb-4 flex-1">
                    {v.description}
                  </p>

                  {/* Движения */}
                  <div className="space-y-1.5 mb-4">
                    {v.movements.map((m, idx) => (
                      <div
                        key={idx}
                        className="flex items-center gap-2 text-sm"
                      >
                        <span className="text-slate-400 dark:text-slate-500 w-5 text-right">
                          {idx + 1}.
                        </span>
                        <span className="text-slate-800 dark:text-slate-200 flex-1">
                          {m.movement_name}
                        </span>
                        {(m.reps || m.rounds_note) && (
                          <span className="text-blue-600 dark:text-blue-400 font-mono text-xs">
                            {m.rounds_note || `${m.reps} повт.`}
                          </span>
                        )}
                        {m.weight_male && (
                          <span className="text-slate-400 dark:text-slate-500 text-xs">
                            {m.weight_male}/{m.weight_female}кг
                          </span>
                        )}
                      </div>
                    ))}
                  </div>

                  {/* Целевые зоны */}
                  <div className="flex items-center gap-2 mb-4">
                    <span className="text-xs text-slate-400 dark:text-slate-500">
                      Целевые зоны:
                    </span>
                    {v.target_zones.map((z) => (
                      <span
                        key={z}
                        className={`text-xs px-2 py-0.5 rounded ${
                          z === 1
                            ? "bg-blue-600/20 text-blue-400"
                            : z === 2
                            ? "bg-green-600/20 text-green-400"
                            : z === 3
                            ? "bg-amber-600/20 text-amber-400"
                            : "bg-red-600/20 text-red-400"
                        }`}
                      >
                        Z{z}
                      </span>
                    ))}
                  </div>

                  <button
                    onClick={() => selectWod(v.template_id)}
                    className="w-full py-2.5 bg-blue-600 hover:bg-blue-500 rounded-lg font-medium text-white transition-colors"
                  >
                    Выбрать эту тренировку
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
