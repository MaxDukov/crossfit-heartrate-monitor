import { useEffect, useMemo, useState } from "react";
import { api } from "../../lib/api";
import type { Equipment, Movement } from "../../types";
import {
  MOVEMENT_MODALITY_LABELS,
  MOVEMENT_MUSCLE_LABELS,
  MOVEMENT_DIFFICULTY_LABELS,
  THEME_LABELS,
} from "../../types";

// Единый пикер упражнений: поиск + фильтры-чипы + выбор в один клик.
// Используется в конструкторе тренировки (браузинг каталога без ввода имени).
export default function ExercisePicker({
  movements,
  onClose,
  onPick,
}: {
  movements: Movement[];
  onClose: () => void;
  onPick: (m: Movement) => void;
}) {
  const [search, setSearch] = useState("");
  const [modality, setModality] = useState("");
  const [muscle, setMuscle] = useState("");
  const [difficulty, setDifficulty] = useState("");
  const [themes, setThemes] = useState<string[]>([]);
  const [equipmentKeys, setEquipmentKeys] = useState<string[]>([]);
  const [equipment, setEquipment] = useState<Equipment[]>([]);

  useEffect(() => {
    api.equipment.list().then(setEquipment).catch(() => setEquipment([]));
  }, []);

  const chip = (active: boolean) =>
    `px-2.5 py-1 rounded text-xs transition-colors ${
      active
        ? "bg-emerald-600 text-white font-medium"
        : "bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
    }`;

  const toggle = (list: string[], setList: (v: string[]) => void, v: string) =>
    setList(list.includes(v) ? list.filter((x) => x !== v) : [...list, v]);

  const single = (current: string, set: (v: string) => void, v: string) =>
    set(current === v ? "" : v);

  const parseCSV = (s: string): string[] =>
    (s || "").split(",").map((x) => x.trim()).filter(Boolean);

  const filtered = useMemo(() => {
    const q = search.trim().toLowerCase();
    return movements.filter(
      (m) =>
        (!q || m.name.toLowerCase().includes(q) || m.key.includes(q)) &&
        (!modality || m.modality === modality) &&
        (!muscle || m.muscle_group === muscle) &&
        (!difficulty || m.difficulty === difficulty) &&
        (themes.length === 0 || themes.some((t) => parseCSV(m.themes).includes(t))) &&
        (equipmentKeys.length === 0 || equipmentKeys.some((e) => parseCSV(m.equipment_keys).includes(e)))
    );
  }, [movements, search, modality, muscle, difficulty, themes, equipmentKeys]);

  const eqName = (key: string) => equipment.find((e) => e.key === key)?.name || key;

  const group = (
    label: string,
    entries: [string, string][],
    isActive: (id: string) => boolean,
    onToggle: (id: string) => void
  ) => (
    <div className="mb-2.5">
      <span className="text-xs text-slate-500 dark:text-slate-400">{label}</span>
      <div className="flex gap-1.5 mt-1 flex-wrap">
        {entries.map(([id, name]) => (
          <button key={id} className={chip(isActive(id))} onClick={() => onToggle(id)}>
            {name}
          </button>
        ))}
      </div>
    </div>
  );

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div
        className="bg-white dark:bg-slate-900 rounded-xl max-w-2xl w-full p-6 max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-4">
          Выбор упражнения
        </h3>

        <input
          className="w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white mb-4"
          placeholder="Поиск по названию или ключу…"
          autoFocus
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />

        {group("Модальность", Object.entries(MOVEMENT_MODALITY_LABELS), (id) => modality === id, (id) => single(modality, setModality, id))}
        {group("Группа мышц", Object.entries(MOVEMENT_MUSCLE_LABELS), (id) => muscle === id, (id) => single(muscle, setMuscle, id))}
        {group("Сложность", Object.entries(MOVEMENT_DIFFICULTY_LABELS), (id) => difficulty === id, (id) => single(difficulty, setDifficulty, id))}
        {group("Темы", Object.entries(THEME_LABELS), (id) => themes.includes(id), (id) => toggle(themes, setThemes, id))}
        {group("Инвентарь", equipment.map((e) => [e.key, e.name] as [string, string]), (id) => equipmentKeys.includes(id), (id) => toggle(equipmentKeys, setEquipmentKeys, id))}

        <p className="text-xs text-slate-400 mb-2">
          Найдено: {filtered.length} из {movements.length}
        </p>

        <div className="border border-slate-200 dark:border-slate-700 rounded-lg divide-y divide-slate-100 dark:divide-slate-800 max-h-72 overflow-y-auto mb-4">
          {filtered.map((m) => (
            <button
              key={m.key}
              className="w-full text-left px-4 py-2.5 hover:bg-emerald-50 dark:hover:bg-emerald-950/20 transition-colors"
              onClick={() => onPick(m)}
            >
              <div className="flex items-center gap-2 flex-wrap">
                <span className="text-sm font-medium text-slate-900 dark:text-white">{m.name}</span>
                {m.is_custom && (
                  <span className="text-[10px] font-semibold px-1.5 py-0.5 rounded bg-emerald-500/15 text-emerald-600 dark:text-emerald-400">
                    своё
                  </span>
                )}
                <span className="text-xs text-slate-400">{MOVEMENT_MUSCLE_LABELS[m.muscle_group] || m.muscle_group}</span>
                <span className="text-xs text-slate-400">· {MOVEMENT_MODALITY_LABELS[m.modality] || m.modality}</span>
                <span className="text-xs text-slate-400">· {MOVEMENT_DIFFICULTY_LABELS[m.difficulty] || m.difficulty}</span>
                <span className="flex-1" />
                <span className="text-xs text-slate-500 dark:text-slate-400">
                  {parseCSV(m.equipment_keys).length === 0 ? "без инвентаря" : parseCSV(m.equipment_keys).map(eqName).join(", ")}
                </span>
              </div>
            </button>
          ))}
          {filtered.length === 0 && (
            <p className="text-sm text-slate-400 text-center py-6">Ничего не найдено.</p>
          )}
        </div>

        <div className="flex justify-end">
          <button
            className="px-4 py-2 rounded text-sm text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
            onClick={onClose}
          >
            Отмена
          </button>
        </div>
      </div>
    </div>
  );
}
