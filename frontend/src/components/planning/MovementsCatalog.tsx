import { useEffect, useState, useCallback } from "react";
import { Link } from "react-router-dom";
import { api, type MovementPayload } from "../../lib/api";
import type { Movement, Equipment } from "../../types";
import {
  MOVEMENT_MODALITY_LABELS,
  MOVEMENT_MUSCLE_LABELS,
  MOVEMENT_DIFFICULTY_LABELS,
  THEME_LABELS,
} from "../../types";

// Каталог движений: просмотр, редактирование, создание (признаки draft1.MD).
export default function MovementsCatalog() {
  const [items, setItems] = useState<Movement[]>([]);
  const [equipment, setEquipment] = useState<Equipment[]>([]);
  const [search, setSearch] = useState("");
  const [editing, setEditing] = useState<{ mode: "create" } | { mode: "edit"; movement: Movement } | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [savedKey, setSavedKey] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      const list = await api.movements.list();
      setItems(list);
      setError(null);
    } catch (e) {
      setError((e as Error).message);
    }
  }, []);

  useEffect(() => {
    load();
    api.equipment.list().then(setEquipment).catch(() => setEquipment([]));
  }, [load]);

  const q = search.trim().toLowerCase();
  const filtered = q
    ? items.filter((m) => m.name.toLowerCase().includes(q) || m.key.includes(q))
    : items;

  return (
    <div className="max-w-5xl mx-auto p-6">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h2 className="text-lg font-bold text-slate-800 dark:text-slate-200">Движения</h2>
          <p className="text-sm text-slate-400">
            Признаки используются рекомендациями, подбором инвентаря и масштабированием · {items.length} шт.
          </p>
        </div>
        <button
          className="bg-emerald-600 hover:bg-emerald-500 px-4 py-2 rounded text-sm font-medium"
          onClick={() => setEditing({ mode: "create" })}
        >
          + Новое движение
        </button>
      </div>

      <input
        className="w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white mb-4"
        placeholder="Поиск по названию или ключу…"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
      />

      {error && (
        <div className="bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-500/30 rounded-lg p-3 mb-4 text-sm text-red-600 dark:text-red-300">
          {error}
        </div>
      )}
      {savedKey && (
        <div className="bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-500/30 rounded-lg p-3 mb-4 text-sm text-emerald-600 dark:text-emerald-300">
          ✓ Сохранено: {savedKey}
        </div>
      )}

      <div className="space-y-1.5">
        {filtered.map((m) => (
          <button
            key={m.key}
            onClick={() => setEditing({ mode: "edit", movement: m })}
            className="w-full text-left bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg px-4 py-3 hover:border-emerald-400 dark:hover:border-emerald-500/50 transition-colors"
          >
            <div className="flex items-center gap-2 flex-wrap">
              <span className="font-medium text-slate-900 dark:text-white">{m.name}</span>
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
                {parseCSV(m.equipment_keys).length === 0 ? "без инвентаря" : parseCSV(m.equipment_keys).map(eqName(equipment)).join(", ")}
              </span>
            </div>
          </button>
        ))}
        {filtered.length === 0 && (
          <p className="text-sm text-slate-400 text-center py-6">Ничего не найдено.</p>
        )}
      </div>

      {editing && (
        <MovementEditor
          movement={editing.mode === "edit" ? editing.movement : null}
          equipment={equipment}
          onClose={() => setEditing(null)}
          onSaved={async (key) => {
            setEditing(null);
            if (key) setSavedKey(key);
            await load();
          }}
        />
      )}
    </div>
  );
}

function parseCSV(s: string): string[] {
  return (s || "").split(",").map((x) => x.trim()).filter(Boolean);
}

function eqName(equipment: Equipment[]) {
  return (key: string) => equipment.find((e) => e.key === key)?.name || key;
}

function MovementEditor({
  movement,
  equipment,
  onClose,
  onSaved,
}: {
  movement: Movement | null;
  equipment: Equipment[];
  onClose: () => void;
  onSaved: (key: string) => void;
}) {
  const [name, setName] = useState(movement?.name ?? "");
  const [modality, setModality] = useState(movement?.modality ?? "gymnastics");
  const [muscle, setMuscle] = useState(movement?.muscle_group ?? "full_body");
  const [difficulty, setDifficulty] = useState(movement?.difficulty ?? "beginner");
  const [themes, setThemes] = useState<string[]>(parseCSV(movement?.themes ?? ""));
  const [equipmentKeys, setEquipmentKeys] = useState<string[]>(parseCSV(movement?.equipment_keys ?? ""));
  const [scalingBeginner, setScalingBeginner] = useState(movement?.scaling_beginner ?? "");
  const [scalingIntermediate, setScalingIntermediate] = useState(movement?.scaling_intermediate ?? "");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [usedIn, setUsedIn] = useState<{ template_id: string; name: string }[]>([]);

  const toggle = (list: string[], setList: (v: string[]) => void, v: string) =>
    setList(list.includes(v) ? list.filter((x) => x !== v) : [...list, v]);

  const chip = (active: boolean) =>
    `px-2.5 py-1 rounded text-xs transition-colors ${
      active
        ? "bg-emerald-600 text-white font-medium"
        : "bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700"
    }`;

  const save = async () => {
    setBusy(true);
    setError(null);
    try {
      const payload: MovementPayload = {
        name,
        modality,
        muscle_group: muscle,
        themes,
        equipment_keys: equipmentKeys,
        difficulty,
        scaling_beginner: scalingBeginner || undefined,
        scaling_intermediate: scalingIntermediate || undefined,
      };
      const res = movement
        ? await api.movements.update(movement.key, payload)
        : await api.movements.create({ ...payload, key: undefined });
      onSaved(res.key);
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setBusy(false);
    }
  };

  const remove = async () => {
    if (!movement) return;
    setBusy(true);
    setError(null);
    setUsedIn([]);
    try {
      await api.movements.delete(movement.key);
      onSaved("");
    } catch (e) {
      const err = e as Error & { body?: { templates?: { template_id: string; name: string }[] } };
      setError(err.message);
      setUsedIn(err.body?.templates ?? []);
      setConfirmDelete(false);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div
        className="bg-white dark:bg-slate-900 rounded-xl max-w-lg w-full p-6 max-h-[90vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-4">
          {movement ? `Редактирование · ${movement.name}` : "Новое движение"}
        </h3>

        <label className="block mb-3">
          <span className="text-xs text-slate-500 dark:text-slate-400">Название</span>
          <input
            className="mt-1 w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Например: Строгие подтягивания"
          />
        </label>
        {!movement && (
          <p className="text-xs text-slate-400 mb-3">
            Ключ сгенерируется автоматически из названия (латиницей).
          </p>
        )}
        {movement && (
          <p className="text-xs text-slate-400 mb-3">Ключ: {movement.key}</p>
        )}

        <div className="mb-3">
          <span className="text-xs text-slate-500 dark:text-slate-400">Модальность</span>
          <div className="flex gap-2 mt-1">
            {Object.entries(MOVEMENT_MODALITY_LABELS).map(([id, label]) => (
              <button key={id} className={chip(modality === id)} onClick={() => setModality(id)}>
                {label}
              </button>
            ))}
          </div>
        </div>

        <div className="mb-3">
          <span className="text-xs text-slate-500 dark:text-slate-400">Группа мышц</span>
          <div className="flex gap-2 mt-1 flex-wrap">
            {Object.entries(MOVEMENT_MUSCLE_LABELS).map(([id, label]) => (
              <button key={id} className={chip(muscle === id)} onClick={() => setMuscle(id)}>
                {label}
              </button>
            ))}
          </div>
        </div>

        <div className="mb-3">
          <span className="text-xs text-slate-500 dark:text-slate-400">Сложность</span>
          <div className="flex gap-2 mt-1">
            {Object.entries(MOVEMENT_DIFFICULTY_LABELS).map(([id, label]) => (
              <button key={id} className={chip(difficulty === id)} onClick={() => setDifficulty(id)}>
                {label}
              </button>
            ))}
          </div>
        </div>

        <div className="mb-3">
          <span className="text-xs text-slate-500 dark:text-slate-400">Темы</span>
          <div className="flex gap-2 mt-1 flex-wrap">
            {Object.entries(THEME_LABELS).map(([id, label]) => (
              <button key={id} className={chip(themes.includes(id))} onClick={() => toggle(themes, setThemes, id)}>
                {label}
              </button>
            ))}
          </div>
        </div>

        <div className="mb-3">
          <span className="text-xs text-slate-500 dark:text-slate-400">Инвентарь</span>
          <div className="flex gap-2 mt-1 flex-wrap">
            {equipment.map((e) => (
              <button
                key={e.key}
                className={chip(equipmentKeys.includes(e.key))}
                onClick={() => toggle(equipmentKeys, setEquipmentKeys, e.key)}
              >
                {e.name}
              </button>
            ))}
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3 mb-4">
          <label className="block">
            <span className="text-xs text-slate-500 dark:text-slate-400">Масштабирование · новичок</span>
            <input
              className="mt-1 w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
              value={scalingBeginner}
              onChange={(e) => setScalingBeginner(e.target.value)}
            />
          </label>
          <label className="block">
            <span className="text-xs text-slate-500 dark:text-slate-400">Масштабирование · средний</span>
            <input
              className="mt-1 w-full bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
              value={scalingIntermediate}
              onChange={(e) => setScalingIntermediate(e.target.value)}
            />
          </label>
        </div>

        {error && (
          <div className="bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-500/30 rounded-lg p-2 mb-3 text-sm text-red-600 dark:text-red-300">
            {error}
          </div>
        )}

        {usedIn.length > 0 && (
          <div className="bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700 rounded-lg p-2 mb-3">
            <p className="text-xs text-slate-500 dark:text-slate-400 mb-1">
              Используется в шаблонах:
            </p>
            <div className="flex flex-wrap gap-1.5">
              {usedIn.map((t) => (
                <Link
                  key={t.template_id}
                  className="text-xs bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 rounded px-2 py-1 text-blue-600 dark:text-blue-400 hover:border-blue-400"
                  to={`/wod?tab=library&search=${encodeURIComponent(t.name)}`}
                  onClick={onClose}
                >
                  {t.name} →
                </Link>
              ))}
            </div>
          </div>
        )}

        {confirmDelete ? (
          <div className="bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-500/30 rounded-lg p-3 mb-3">
            <p className="text-sm text-red-700 dark:text-red-300 mb-2">
              Удалить «{movement?.name}» из каталога?
            </p>
            <div className="flex gap-2">
              <button
                className="bg-red-600 hover:bg-red-500 disabled:opacity-50 px-3 py-1.5 rounded text-sm font-medium text-white"
                disabled={busy}
                onClick={remove}
              >
                Удалить
              </button>
              <button
                className="px-3 py-1.5 rounded text-sm text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800"
                onClick={() => setConfirmDelete(false)}
              >
                Не удалять
              </button>
            </div>
          </div>
        ) : null}

        <div className="flex gap-2">
          <button
            className="flex-1 bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 px-4 py-2 rounded text-sm font-medium"
            disabled={!name.trim() || busy}
            onClick={save}
          >
            {busy ? "Сохранение…" : "Сохранить"}
          </button>
          {movement && !confirmDelete && (
            <button
              className="px-4 py-2 rounded text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-950/30"
              onClick={() => setConfirmDelete(true)}
            >
              Удалить
            </button>
          )}
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
