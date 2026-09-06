import { useState } from "react";
import { api } from "../../lib/api";

// Форма создания макроцикла — Экран 1 draft1.MD.
export default function CycleForm({
  onClose,
  onCreated,
}: {
  onClose: () => void;
  onCreated: (cycleId: string) => void;
}) {
  const [name, setName] = useState("");
  const [goal, setGoal] = useState("");
  const [weeks, setWeeks] = useState(8);
  const [startDate, setStartDate] = useState(() => {
    const d = new Date();
    d.setDate(d.getDate() + ((8 - d.getDay()) % 7 || 7)); // ближайший понедельник
    return d.toISOString().slice(0, 10);
  });
  const [modality, setModality] = useState("strength");
  const [groups, setGroups] = useState<{ name: string; weekdays: number[]; thirdDayOff: boolean }[]>([
    { name: "Группа А", weekdays: [2, 4, 6], thirdDayOff: false },
  ]);
  const [warnings, setWarnings] = useState<string[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const toggleWeekday = (gi: number, wd: number) => {
    setGroups((gs) =>
      gs.map((g, i) =>
        i === gi
          ? {
              ...g,
              weekdays: g.weekdays.includes(wd)
                ? g.weekdays.filter((w) => w !== wd)
                : [...g.weekdays, wd].sort((a, b) => a - b),
            }
          : g,
      ),
    );
  };

  const submit = async () => {
    setBusy(true);
    setError(null);
    try {
      const res = await api.cycles.create({
        name,
        goal: goal || undefined,
        weeks,
        start_date: startDate,
        modality_priority: modality,
        groups: groups.filter((g) => g.weekdays.length > 0).map((g) => ({
          name: g.name,
          weekdays: g.weekdays,
          third_day_off_cycle: g.thirdDayOff,
        })),
      });
      if (res.warnings && res.warnings.length > 0) {
        setWarnings(res.warnings);
        setTimeout(() => onCreated(res.id), 1800);
      } else {
        onCreated(res.id);
      }
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  };

  const canSubmit =
    name.trim() && groups.every((g) => g.weekdays.length > 0) && groups.length > 0 && !busy;

  const WD = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-slate-900 rounded-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto p-6">
        <h2 className="text-xl font-bold text-slate-900 dark:text-white mb-4">
          Новый макроцикл
        </h2>

        <div className="grid grid-cols-2 gap-4 mb-4">
          <label className="block">
            <span className="text-sm text-slate-500 dark:text-slate-400">Название</span>
            <input
              className="mt-1 w-full bg-slate-50 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
              placeholder="Например: Подготовка к Open"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </label>
          <label className="block">
            <span className="text-sm text-slate-500 dark:text-slate-400">Цель</span>
            <input
              className="mt-1 w-full bg-slate-50 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
              placeholder="Стратегическая цель цикла"
              value={goal}
              onChange={(e) => setGoal(e.target.value)}
            />
          </label>
          <label className="block">
            <span className="text-sm text-slate-500 dark:text-slate-400">Длительность, недель</span>
            <input
              type="number"
              min={1}
              max={52}
              className="mt-1 w-full bg-slate-50 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
              value={weeks}
              onChange={(e) => setWeeks(Number(e.target.value) || 8)}
            />
          </label>
          <label className="block">
            <span className="text-sm text-slate-500 dark:text-slate-400">Дата старта</span>
            <input
              type="date"
              className="mt-1 w-full bg-slate-50 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
            />
          </label>
          <label className="block col-span-2">
            <span className="text-sm text-slate-500 dark:text-slate-400">
              Приоритетная модальность первых недель
            </span>
            <div className="mt-2 flex gap-2">
              {[
                ["strength", "Сила"],
                ["gymnastics", "Гимнастика"],
                ["cardio", "Кардио"],
              ].map(([id, label]) => (
                <button
                  key={id}
                  onClick={() => setModality(id)}
                  className={`px-4 py-2 rounded-lg text-sm font-medium ${
                    modality === id
                      ? "bg-emerald-600 text-white"
                      : "bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300"
                  }`}
                >
                  {label}
                </button>
              ))}
            </div>
          </label>
        </div>

        <div className="mb-4">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium text-slate-700 dark:text-slate-300">
              Группы и дни недели
            </span>
            <button
              className="text-sm text-emerald-600 dark:text-emerald-400"
              onClick={() =>
                setGroups((gs) => [...gs, { name: `Группа ${String.fromCharCode(65 + gs.length)}`, weekdays: [], thirdDayOff: false }])
              }
            >
              + группа
            </button>
          </div>
          {groups.map((g, gi) => (
            <div
              key={gi}
              className="flex items-center gap-2 mb-2 bg-slate-50 dark:bg-slate-800/50 rounded-lg p-3"
            >
              <input
                className="w-32 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-2 py-1.5 text-sm text-slate-900 dark:text-white"
                value={g.name}
                onChange={(e) =>
                  setGroups((gs) => gs.map((x, i) => (i === gi ? { ...x, name: e.target.value } : x)))
                }
              />
              <div className="flex gap-1">
                {WD.map((label, idx) => {
                  const wd = idx + 1;
                  const on = g.weekdays.includes(wd);
                  const isThirdDay =
                    g.thirdDayOff && g.weekdays.filter((w) => w <= wd).length === 3 && on;
                  return (
                    <button
                      key={wd}
                      onClick={() => toggleWeekday(gi, wd)}
                      title={
                        isThirdDay
                          ? "3-й тренировочный день — вне цикла (техника, тесты 1ПМ)"
                          : undefined
                      }
                      className={`w-9 h-9 rounded text-xs font-semibold ${
                        isThirdDay
                          ? "bg-amber-500 text-white ring-2 ring-amber-300"
                          : on
                          ? "bg-emerald-600 text-white"
                          : "bg-slate-200 dark:bg-slate-700 text-slate-500 dark:text-slate-400"
                      }`}
                    >
                      {label}
                    </button>
                  );
                })}
                {/* Переключатель «третий день вне цикла» — колонкой третьего дня */}
                <button
                  onClick={() =>
                    setGroups((gs) =>
                      gs.map((x, i) => (i === gi ? { ...x, thirdDayOff: !x.thirdDayOff } : x))
                    )
                  }
                  title="Каждый 3-й тренировочный день цикла планируется отдельно: отработка техники, тесты 1ПМ"
                  className={`w-14 h-9 rounded text-[10px] font-bold leading-tight ${
                    g.thirdDayOff
                      ? "bg-amber-500 text-white"
                      : "bg-slate-200 dark:bg-slate-700 text-slate-500 dark:text-slate-400"
                  }`}
                >
                  3-й день<br />вне цикла
                </button>
              </div>
              {groups.length > 1 && (
                <button
                  className="ml-auto text-red-400 hover:text-red-300 text-sm"
                  onClick={() => setGroups((gs) => gs.filter((_, i) => i !== gi))}
                >
                  ✕
                </button>
              )}
            </div>
          ))}
          <p className="text-xs text-slate-400 mt-1">
            Слоты создаются автоматически по дням недели. Система предупредит, если между
            тренировками меньше 48 часов. «3-й день вне цикла»: каждый третий тренировочный день
            группы планируется отдельно — отработка техники из текущего цикла, тестирование 1ПМ.
          </p>
        </div>

        {warnings && (
          <div className="bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-500/30 rounded-lg p-3 mb-4 text-sm text-amber-700 dark:text-amber-300 space-y-1">
            {warnings.map((w, i) => (
              <p key={i}>⚠ {w}</p>
            ))}
            <p className="text-xs opacity-70">Переходим к календарю…</p>
          </div>
        )}
        {error && (
          <div className="bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-500/30 rounded-lg p-3 mb-4 text-sm text-red-700 dark:text-red-300">
            {error}
          </div>
        )}

        <div className="flex justify-end gap-3">
          <button
            className="px-4 py-2 rounded text-sm text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
            onClick={onClose}
          >
            Отмена
          </button>
          <button
            className="bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 px-5 py-2 rounded text-sm font-medium"
            disabled={!canSubmit}
            onClick={submit}
          >
            Создать цикл
          </button>
        </div>
      </div>
    </div>
  );
}
