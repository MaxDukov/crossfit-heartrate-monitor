import { useEffect, useState } from "react";
import { api } from "../../lib/api";
import type { Movement } from "../../types";
import { THEME_LABELS, THEME_ICONS, FORMAT_LABELS, LEVEL_LABELS } from "../../types";

interface Row {
  movement_key: string;
  movement_name: string;
  reps: string;
  weight_male: string;
  weight_female: string;
}

const EMPTY_ROW: Row = { movement_key: "", movement_name: "", reps: "", weight_male: "", weight_female: "" };

// Конструктор тренировки с нуля — Экран 4 draft1.MD.
export default function Constructor() {
  const [name, setName] = useState("");
  const [format, setFormat] = useState("amrap");
  const [duration, setDuration] = useState(15);
  const [intensity, setIntensity] = useState("medium");
  const [theme, setTheme] = useState("full_body");
  const [description, setDescription] = useState("");
  const [rows, setRows] = useState<Row[]>([{ ...EMPTY_ROW }]);
  const [movements, setMovements] = useState<Movement[]>([]);
  const [warnings, setWarnings] = useState<string[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [level, setLevel] = useState("intermediate");

  useEffect(() => {
    api.movements.list().then(setMovements).catch(() => setMovements([]));
  }, []);

  const setRow = (i: number, patch: Partial<Row>) =>
    setRows((rs) => rs.map((r, idx) => (idx === i ? { ...r, ...patch } : r)));

  const suggestions = (row: Row): Movement[] => {
    const q = row.movement_name.trim().toLowerCase();
    if (!q) return [];
    return movements
      .filter((m) => m.name.toLowerCase().includes(q) || m.key.includes(q))
      .slice(0, 6);
  };

  const submit = async () => {
    setBusy(true);
    setError(null);
    setWarnings(null);
    setSaved(null);
    try {
      const payload = {
        name: name || undefined,
        format,
        duration_min: duration,
        intensity,
        theme,
        description: description || undefined,
        movements: rows
          .filter((r) => r.movement_key)
          .map((r) => ({
            movement_key: r.movement_key,
            reps: r.reps ? Number(r.reps) : null,
            weight_male: r.weight_male ? Number(r.weight_male) : null,
            weight_female: r.weight_female ? Number(r.weight_female) : null,
          })),
      };
      const res = await api.wods.custom(payload);
      setWarnings(res.warnings);
      setSaved(res.template_id);
      setRows([{ ...EMPTY_ROW }]);
      setName("");
      setDescription("");
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  };

  const filled = rows.filter((r) => r.movement_key).length;
  const canSubmit = filled > 0 && !busy;

  return (
    <div className="max-w-4xl mx-auto p-6">
      <p className="text-slate-500 dark:text-slate-400 mb-4">
        Соберите тренировку по протоколу: система проверит инвентарь, баланс паттернов и стимул.
        Сохранённая тренировка попадает в библиотеку и доступна для назначения на слоты.
      </p>

      <div className="grid grid-cols-4 gap-3 mb-4">
        <label className="block">
          <span className="text-xs text-slate-500 dark:text-slate-400">Протокол</span>
          <select
            className="mt-1 w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-2 py-2 text-sm text-slate-900 dark:text-white"
            value={format}
            onChange={(e) => setFormat(e.target.value)}
          >
            {Object.entries(FORMAT_LABELS).map(([id, label]) => (
              <option key={id} value={id}>{label}</option>
            ))}
          </select>
        </label>
        <label className="block">
          <span className="text-xs text-slate-500 dark:text-slate-400">Длительность, мин</span>
          <input
            type="number"
            min={1}
            max={120}
            className="mt-1 w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-2 py-2 text-sm text-slate-900 dark:text-white"
            value={duration}
            onChange={(e) => setDuration(Number(e.target.value) || 15)}
          />
        </label>
        <label className="block">
          <span className="text-xs text-slate-500 dark:text-slate-400">Интенсивность</span>
          <select
            className="mt-1 w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-2 py-2 text-sm text-slate-900 dark:text-white"
            value={intensity}
            onChange={(e) => setIntensity(e.target.value)}
          >
            <option value="low">Низкая</option>
            <option value="medium">Средняя</option>
            <option value="high">Высокая</option>
          </select>
        </label>
        <label className="block">
          <span className="text-xs text-slate-500 dark:text-slate-400">Имя (необязательно)</span>
          <input
            className="mt-1 w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-2 py-2 text-sm text-slate-900 dark:text-white"
            placeholder="авто: протокол · тема"
            value={name}
            onChange={(e) => setName(e.target.value)}
          />
        </label>
      </div>

      <div className="grid grid-cols-4 gap-2 mb-6">
        {Object.keys(THEME_LABELS).map((t) => (
          <button
            key={t}
            onClick={() => setTheme(t)}
            className={`rounded-lg px-2 py-2 text-xs font-medium transition-colors ${
              theme === t
                ? "bg-emerald-600 text-white"
                : "bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300"
            }`}
          >
            {THEME_ICONS[t]} {THEME_LABELS[t]}
          </button>
        ))}
      </div>

      <div className="mb-4">
        <div className="flex items-center justify-between mb-2">
          <span className="text-sm font-medium text-slate-700 dark:text-slate-300">
            Упражнения
          </span>
          <button
            className="text-sm text-emerald-600 dark:text-emerald-400"
            onClick={() => setRows((rs) => [...rs, { ...EMPTY_ROW }])}
          >
            + строка
          </button>
        </div>
        <div className="space-y-2">
          {rows.map((r, i) => (
            <div key={i} className="relative flex gap-2">
              <div className="flex-1 relative">
                <input
                  className="w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
                  placeholder="Упражнение (начните вводить)"
                  value={r.movement_name}
                  onChange={(e) => setRow(i, { movement_name: e.target.value, movement_key: "" })}
                />
                {r.movement_name && !r.movement_key && suggestions(r).length > 0 && (
                  <div className="absolute z-10 left-0 right-0 mt-1 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded shadow-lg overflow-hidden">
                    {suggestions(r).map((m) => (
                      <button
                        key={m.key}
                        className="block w-full text-left px-3 py-2 text-sm text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700"
                        onClick={() => setRow(i, { movement_key: m.key, movement_name: m.name })}
                      >
                        {m.name}
                        <span className="text-xs text-slate-400 ml-2">
                          {m.difficulty} · {m.modality}
                        </span>
                      </button>
                    ))}
                  </div>
                )}
              </div>
              <input
                className="w-20 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-2 py-2 text-sm text-slate-900 dark:text-white"
                placeholder="повт."
                type="number"
                value={r.reps}
                onChange={(e) => setRow(i, { reps: e.target.value })}
              />
              <input
                className="w-24 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-2 py-2 text-sm text-slate-900 dark:text-white"
                placeholder="кг М"
                type="number"
                value={r.weight_male}
                onChange={(e) => setRow(i, { weight_male: e.target.value })}
              />
              <input
                className="w-24 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-2 py-2 text-sm text-slate-900 dark:text-white"
                placeholder="кг Ж"
                type="number"
                value={r.weight_female}
                onChange={(e) => setRow(i, { weight_female: e.target.value })}
              />
              {rows.length > 1 && (
                <button
                  className="text-red-400 hover:text-red-300 px-2"
                  onClick={() => setRows((rs) => rs.filter((_, idx) => idx !== i))}
                >
                  ✕
                </button>
              )}
            </div>
          ))}
        </div>
      </div>

      <label className="block mb-6">
        <span className="text-xs text-slate-500 dark:text-slate-400">Описание / стандарты</span>
        <textarea
          className="mt-1 w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
          rows={2}
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
      </label>

      {error && (
        <div className="bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-500/30 rounded-lg p-3 mb-4 text-sm text-red-700 dark:text-red-300">
          {error}
        </div>
      )}
      {warnings && (
        <div className="bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-500/30 rounded-lg p-3 mb-4 text-sm text-amber-700 dark:text-amber-300 space-y-1">
          <p className="font-semibold">Проверки системы:</p>
          {warnings.length === 0 ? (
            <p>✓ Замечаний нет — сбалансированная тренировка</p>
          ) : (
            warnings.map((w, i) => <p key={i}>⚠ {w}</p>)
          )}
        </div>
      )}
      {saved && (
        <div className="bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-500/30 rounded-lg p-3 mb-4 text-sm text-emerald-700 dark:text-emerald-300 flex items-center justify-between">
          <span>✓ Тренировка сохранена в библиотеку</span>
          <div className="flex gap-2">
            <button
              className="text-emerald-600 dark:text-emerald-400 font-medium"
              onClick={async () => {
                await api.wods.select(saved, level);
                window.location.href = "/";
              }}
            >
              Запустить сейчас
            </button>
          </div>
        </div>
      )}

      <div className="flex items-center gap-3">
        <button
          className="bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 px-5 py-2 rounded text-sm font-medium"
          disabled={!canSubmit}
          onClick={submit}
        >
          {busy ? "Сохранение…" : "Сохранить в библиотеку"}
        </button>
        <span className="text-xs text-slate-400">
          Уровень для запуска:{" "}
          <select
            className="bg-transparent border-none text-xs text-emerald-500"
            value={level}
            onChange={(e) => setLevel(e.target.value)}
          >
            {Object.entries(LEVEL_LABELS).map(([id, label]) => (
              <option key={id} value={id} className="text-slate-900">{label}</option>
            ))}
          </select>
        </span>
      </div>
    </div>
  );
}
