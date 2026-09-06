import { useEffect, useState, useCallback } from "react";
import { api } from "../../lib/api";
import type { SlotDetail, Recommendation, WodTemplateItem, Athlete, WorkoutResult } from "../../types";
import {
  FORMAT_LABELS, LEVEL_LABELS, wodSummary,
  SLOT_WARMUP_MIN, SLOT_COOLDOWN_MIN, SLOT_WOD_CAP, SLOT_DENSE_TOTAL, SLOT_FREE_MIN, SLOT_DAY_MAX_MIN,
} from "../../types";

// Панель слота: рекомендации → назначение → проведение → результаты
// (Экраны 3, 5, 6 draft1.MD).
export default function SlotPanel({
  slotId,
  onClose,
  onChanged,
}: {
  slotId: string;
  onClose: () => void;
  onChanged: () => void;
}) {
  const [slot, setSlot] = useState<SlotDetail | null>(null);
  const [level, setLevel] = useState("intermediate");
  const [recs, setRecs] = useState<Recommendation[]>([]);
  const [libraryMode, setLibraryMode] = useState(false);
  const [librarySearch, setLibrarySearch] = useState("");
  const [library, setLibrary] = useState<WodTemplateItem[]>([]);
  const [warnings, setWarnings] = useState<string[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [prBanner, setPrBanner] = useState<string | null>(null);
  const [pendingAssign, setPendingAssign] = useState<WodTemplateItem | Recommendation | null>(null);
  const [expandedWod, setExpandedWod] = useState<string | null>(null);

  const load = useCallback(async () => {
    try {
      setSlot(await api.slots.get(slotId));
      setError(null);
    } catch (e) {
      setError((e as Error).message);
    }
  }, [slotId]);

  useEffect(() => { load(); }, [load]);

  const dayTotal = slot?.wods?.reduce((s, w) => s + w.duration_min, 0) ?? 0;
  const dayLeft = SLOT_WOD_CAP - dayTotal;
  // В день можно добавлять, пока есть свободные минуты (пустой или planned).
  const canAdd = slot != null
    && (slot.status === "empty" || slot.status === "planned")
    && dayLeft > 0;
  const showFill = canAdd && dayLeft > SLOT_FREE_MIN;

  const loadRecs = useCallback(async () => {
    try {
      setRecs(await api.slots.recommendations(slotId, level));
    } catch {
      setRecs([]);
    }
  }, [slotId, level]);

  useEffect(() => {
    if (!slot) return;
    const total = (slot.wods ?? []).reduce((s, w) => s + w.duration_min, 0);
    const canFill = (slot.status === "empty" || slot.status === "planned")
      && SLOT_WOD_CAP - total > SLOT_FREE_MIN;
    if (canFill) loadRecs();
  }, [slot, loadRecs]);

  useEffect(() => {
    if (!libraryMode) return;
    const t = setTimeout(() => {
      api.wods.templates({ search: librarySearch || undefined, limit: 30 }).then(setLibrary);
    }, 250);
    return () => clearTimeout(t);
  }, [libraryMode, librarySearch]);

  const doAssign = async (templateId: string) => {
    setError(null);
    try {
      const res = await api.slots.assign(slotId, templateId, level);
      setWarnings(res.warnings);
      setLibraryMode(false);
      await load();
      onChanged();
    } catch (e) {
      setError((e as Error).message);
    }
  };

  // Плотное расписание: сумма тренировок + разминка/заминка > 55 мин —
  // сохраняем только после подтверждения.
  const assign = (t: WodTemplateItem | Recommendation) => {
    if (dayTotal + t.duration_min + SLOT_WARMUP_MIN + SLOT_COOLDOWN_MIN > SLOT_DENSE_TOTAL) {
      setPendingAssign(t);
      return;
    }
    doAssign(t.template_id);
  };

  // Тренировка выводит день за 65 минут — в группу «Выходит за временной лимит».
  const overLimit = (t: WodTemplateItem | Recommendation) =>
    dayTotal + t.duration_min + SLOT_WARMUP_MIN + SLOT_COOLDOWN_MIN > SLOT_DAY_MAX_MIN;
  const overBy = (t: WodTemplateItem | Recommendation) =>
    dayTotal + t.duration_min + SLOT_WARMUP_MIN + SLOT_COOLDOWN_MIN - SLOT_DAY_MAX_MIN;

  const confirmAssign = async () => {
    if (!pendingAssign) return;
    const id = pendingAssign.template_id;
    setPendingAssign(null);
    await doAssign(id);
  };

  const removeWod = async (wodId: string) => {
    setError(null);
    try {
      await api.slots.unassignWod(slotId, wodId);
      setWarnings(null);
      await load();
      onChanged();
    } catch (e) {
      setError((e as Error).message);
    }
  };

  const unassign = async () => {
    await api.slots.unassign(slotId);
    setWarnings(null);
    await load();
    onChanged();
  };

  const start = async () => {
    setError(null);
    try {
      await api.slots.start(slotId);
      await load();
      onChanged();
    } catch (e) {
      setError((e as Error).message);
    }
  };

  const complete = async () => {
    await api.slots.complete(slotId);
    await load();
    onChanged();
  };

  if (error && !slot) {
    return (
      <Drawer onClose={onClose}>
        <div className="text-red-500 p-4">{error}</div>
      </Drawer>
    );
  }
  if (!slot) return null;

  return (
    <Drawer onClose={onClose}>
      {/* Шапка слота */}
      <div className="flex items-start justify-between mb-4">
        <div>
          <div className="text-xs text-slate-400 mb-1">
            {slot.group_name} · День {slot.day_number} · {slot.slot_date}
          </div>
          <h2 className="text-xl font-bold text-slate-900 dark:text-white">
            {slot.wod?.name || "Подбор тренировки"}
          </h2>
        </div>
        <span className="text-xs font-semibold px-2 py-1 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-500">
          {levelLabel(level)} · {slot.status === "empty" ? "свободен" : slot.status === "planned" ? "запланирован" : slot.status === "in_progress" ? "идёт сейчас" : "проведён"}
        </span>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-500/30 rounded-lg p-2 text-sm text-red-600 dark:text-red-300 mb-3">
          {error}
        </div>
      )}
      {warnings && warnings.length > 0 && (
        <div className="bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-500/30 rounded-lg p-3 mb-4 text-sm text-amber-700 dark:text-amber-300">
          {warnings.map((w, i) => <p key={i}>⚠ {w}</p>)}
        </div>
      )}
      {prBanner && (
        <div className="bg-emerald-100 dark:bg-emerald-900/50 border border-emerald-300 dark:border-emerald-500/50 rounded-lg p-3 mb-4 text-sm font-medium text-emerald-700 dark:text-emerald-300">
          🏆 {prBanner}
        </div>
      )}

      {/* Тренировки дня: 60 мин = разминка 10 + WOD + заминка 5 */}
      {slot.wods && slot.wods.length > 0 && (
        <div className="bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg p-3 mb-4">
          <div className="flex items-center justify-between mb-2">
            <h3 className="text-xs font-semibold uppercase tracking-wide text-slate-400">
              Тренировки дня
            </h3>
            <span className={`text-xs ${dayLeft > SLOT_FREE_MIN ? "text-slate-400" : "text-amber-600 dark:text-amber-400"}`}>
              {dayTotal}/{SLOT_WOD_CAP} мин · осталось {dayLeft}
            </span>
          </div>
          <div className="text-[10px] text-slate-400 mb-2">
            разминка {SLOT_WARMUP_MIN}м · заминка {SLOT_COOLDOWN_MIN}м · день 60м · лимит {SLOT_DAY_MAX_MIN}м
          </div>
          <div className="space-y-1">
            {slot.wods.map((w, i) => {
              const open = expandedWod === w.id;
              return (
                <div key={w.id} className="border border-slate-200 dark:border-slate-700 rounded">
                  <button
                    className="w-full flex items-center gap-2 text-sm px-2 py-1.5 hover:bg-slate-50 dark:hover:bg-slate-800/60"
                    title={wodSummary(w)}
                    onClick={() => setExpandedWod(open ? null : w.id)}
                  >
                    <span className="text-[10px] text-slate-400 w-4 shrink-0">{open ? "▾" : "▸"}</span>
                    <span className="text-[10px] text-slate-400 shrink-0">{i + 1}.</span>
                    <span className="flex-1 text-slate-800 dark:text-slate-200 truncate text-left">{w.name}</span>
                    <span className="text-xs text-slate-400 shrink-0">{w.duration_min} мин</span>
                    {slot.status === "planned" && (
                      <span
                        role="button"
                        className="text-slate-300 hover:text-red-500 px-1 shrink-0"
                        title="Снять тренировку"
                        onClick={(e) => {
                          e.stopPropagation();
                          removeWod(w.id);
                        }}
                      >
                        ✕
                      </span>
                    )}
                  </button>
                  {open && (
                    <div className="px-3 pb-2 pt-0.5 space-y-0.5">
                      {(w.movements ?? []).length === 0 && (
                        <p className="text-xs text-slate-400">Состав не задан</p>
                      )}
                      {(w.movements ?? []).map((m, mi) => (
                        <div key={mi} className="flex items-center gap-2 text-xs">
                          <span className="flex-1 text-slate-600 dark:text-slate-300">
                            {m.rounds_note && <span className="text-slate-400">{m.rounds_note} </span>}
                            {m.reps ? `${m.reps} × ` : ""}{m.movement_name}
                          </span>
                          <span className="text-slate-400">
                            {m.weight_male != null && `М ${m.weight_male}`}
                            {m.weight_female != null && ` Ж ${m.weight_female}`}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* Уровень группы */}
      {slot.status === "empty" && (
        <div className="flex items-center gap-2 mb-4 text-sm">
          <span className="text-slate-400">Уровень:</span>
          {Object.entries(LEVEL_LABELS).map(([id, label]) => (
            <button
              key={id}
              onClick={() => setLevel(id)}
              className={`px-2 py-1 rounded text-xs ${
                level === id
                  ? "bg-slate-800 dark:bg-slate-200 text-white dark:text-slate-900 font-semibold"
                  : "bg-slate-100 dark:bg-slate-800 text-slate-500"
              }`}
            >
              {label}
            </button>
          ))}
        </div>
      )}

      {/* ── empty: Экран 3 — рекомендации ── */}
      {slot.status === "empty" && !libraryMode && slot.kind === "off_cycle" && (
        <div className="bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-500/30 rounded-lg p-3 mb-4">
          <p className="text-sm font-semibold text-amber-700 dark:text-amber-300 mb-1">
            День вне цикла — планируется отдельно
          </p>
          {slot.notes && (
            <p className="text-sm text-amber-700/90 dark:text-amber-300/90">{slot.notes}</p>
          )}
          <p className="text-xs text-amber-600/70 dark:text-amber-400/70 mt-2">
            Штатные рекомендации не показаны. Назначьте тренировку вручную из библиотеки.
          </p>
          <button
            className="mt-2 text-sm text-amber-700 dark:text-amber-300 underline"
            onClick={() => setLibraryMode(true)}
          >
            вся библиотека →
          </button>
        </div>
      )}
      {showFill && !libraryMode && (slot.status === "planned" || (slot.status === "empty" && slot.kind !== "off_cycle")) && (
        <div>
          <div className="flex items-center justify-between mb-3">
            <h3 className="text-sm font-semibold text-slate-600 dark:text-slate-300">
              {slot.status === "empty" ? "Рекомендации системы" : "Добавить ещё тренировку"}
            </h3>
            <div className="flex gap-2">
              <button
                className="text-sm text-emerald-600 dark:text-emerald-400"
                onClick={loadRecs}
              >
                обновить
              </button>
              <button
                className="text-sm text-blue-600 dark:text-blue-400"
                onClick={() => setLibraryMode(true)}
              >
                вся библиотека →
              </button>
            </div>
          </div>
          <div className="space-y-2 mb-4">
            {recs.filter((r) => !overLimit(r)).length === 0 && (
              <p className="text-sm text-slate-400">Рекомендаций нет — посмотрите библиотеку.</p>
            )}
            {recs.filter((r) => !overLimit(r)).map((r) => (
              <button
                key={r.template_id}
                onClick={() => assign(r)}
                className="w-full text-left bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg p-3 hover:border-emerald-400 dark:hover:border-emerald-500/50"
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="font-medium text-slate-900 dark:text-white">
                    {r.is_benchmark && <span className="text-amber-500 mr-1">★</span>}
                    {r.name}
                  </span>
                  <span className="text-xs text-slate-400">
                    {r.format_name} · {r.duration_min} мин · {r.intensity_name}
                  </span>
                </div>
                <p className="text-xs text-emerald-600/80 dark:text-emerald-400/80 mb-1">
                  {r.reason}
                </p>
                <p className="text-xs text-slate-500 dark:text-slate-400">
                  {r.movements.map((m) => m.movement_name + (m.reps ? ` ×${m.reps}` : "")).join(", ")}
                </p>
              </button>
            ))}
            <OverLimitList
              items={recs.filter(overLimit)}
              overBy={overBy}
            />
          </div>
          <p className="text-xs text-slate-400">
            Нужна уникальная тренировка? Создайте её в разделе WoD → Конструктор, она появится
            в библиотеке.
          </p>
        </div>
      )}

      {/* ── библиотека ── */}
      {showFill && libraryMode && (
        <div>
          <div className="flex items-center gap-2 mb-3">
            <input
              className="flex-1 bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
              placeholder="Поиск…"
              value={librarySearch}
              onChange={(e) => setLibrarySearch(e.target.value)}
              autoFocus
            />
            <button className="text-sm text-slate-400" onClick={() => setLibraryMode(false)}>
              ← рекомендации
            </button>
          </div>
          <div className="space-y-2">
            {library.filter((t) => !overLimit(t)).map((t) => (
              <button
                key={t.template_id}
                onClick={() => assign(t)}
                className="w-full text-left bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg p-3 hover:border-emerald-400"
              >
                <span className="font-medium text-slate-900 dark:text-white">
                  {t.is_benchmark && <span className="text-amber-500 mr-1">★</span>}
                  {t.name}
                </span>
                <span className="text-xs text-slate-400 ml-2">
                  {FORMAT_LABELS[t.format] || t.format} · {t.duration_min} мин
                </span>
              </button>
            ))}
            <OverLimitList items={library.filter(overLimit)} overBy={overBy} />
          </div>
        </div>
      )}

      {/* ── planned/completed: Экран 5 — карточка WoD ── */}
      {slot.wod && slot.status !== "in_progress" && (
        <WodCard slot={slot} onUnassign={slot.status === "planned" ? unassign : undefined} />
      )}
      {slot.status === "planned" && slot.wod && (
        <div className="flex gap-3 mt-4">
          <button
            className="flex-1 bg-emerald-600 hover:bg-emerald-500 px-4 py-2.5 rounded text-sm font-semibold"
            onClick={start}
          >
            ▶ Начать тренировку
          </button>
          <button
            className="px-4 py-2.5 rounded text-sm text-red-500 hover:bg-red-50 dark:hover:bg-red-950/30"
            onClick={unassign}
          >
            Снять
          </button>
        </div>
      )}

      {/* ── in_progress: Экран 6 — проведение ── */}
      {slot.status === "in_progress" && slot.wod && (
        <LiveRun
          slot={slot}
          onResult={async (banner) => {
            setPrBanner(banner);
            await load();
            onChanged();
          }}
          onComplete={complete}
        />
      )}

      {/* ── completed: результаты ── */}
      {slot.status === "completed" && (
        <ResultsList results={slot.results} />
      )}

      {/* ── подтверждение плотного расписания ── */}
      {pendingAssign && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-60 p-4" onClick={() => setPendingAssign(null)}>
          <div
            className="bg-white dark:bg-slate-900 rounded-xl max-w-sm w-full p-6"
            onClick={(e) => e.stopPropagation()}
          >
            <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-2">
              Плотное расписание, уверены?
            </h3>
            <p className="text-sm text-slate-600 dark:text-slate-300 mb-1">
              С тренировкой «{pendingAssign.name}» ({pendingAssign.duration_min} мин) день займёт{" "}
              <span className="font-semibold">
                {dayTotal + pendingAssign.duration_min + SLOT_WARMUP_MIN + SLOT_COOLDOWN_MIN} мин
              </span>{" "}
              из 60: разминка {SLOT_WARMUP_MIN} + тренировки {dayTotal + pendingAssign.duration_min} + заминка {SLOT_COOLDOWN_MIN}.
            </p>
            <p className="text-xs text-slate-400 mb-5">Добавить эту тренировку в день?</p>
            <div className="flex gap-2">
              <button
                className="flex-1 bg-emerald-600 hover:bg-emerald-500 px-4 py-2 rounded text-sm font-medium"
                onClick={confirmAssign}
              >
                Добавить
              </button>
              <button
                className="px-4 py-2 rounded text-sm text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
                onClick={() => setPendingAssign(null)}
              >
                Отмена
              </button>
            </div>
          </div>
        </div>
      )}
    </Drawer>
  );
}

function Drawer({ children, onClose }: { children: React.ReactNode; onClose: () => void }) {
  return (
    <div className="fixed inset-0 bg-black/40 z-50 flex justify-end" onClick={onClose}>
      <div
        className="bg-slate-50 dark:bg-slate-950 w-full max-w-lg h-full overflow-y-auto p-6 shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <button
          className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 mb-4"
          onClick={onClose}
        >
          ✕ закрыть
        </button>
        {children}
      </div>
    </div>
  );
}

function WodCard({
  slot,
  onUnassign,
}: {
  slot: SlotDetail;
  onUnassign?: () => void;
}) {
  const w = slot.wod!;
  const [editing, setEditing] = useState(false);
  const [movs, setMovs] = useState(w.movements);

  const save = async () => {
    await api.slots.updateMovements(slot.id, movs.map((m) => ({
      movement_key: m.movement_key,
      reps: m.reps,
      weight_male: m.weight_male,
      weight_female: m.weight_female,
    })));
    setEditing(false);
  };

  return (
    <div className="bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-xl p-4">
      <div className="flex items-center justify-between mb-2">
        <div className="text-sm text-slate-500 dark:text-slate-400">
          {FORMAT_LABELS[w.format] || w.format} · {w.duration_min} мин · {w.theme}
        </div>
        {onUnassign && (
          <button className="text-xs text-slate-400 hover:text-slate-600" onClick={() => setEditing(!editing)}>
            {editing ? "отмена" : "изменить веса"}
          </button>
        )}
      </div>
      {w.description && (
        <p className="text-sm text-slate-500 dark:text-slate-400 mb-3">{w.description}</p>
      )}
      <div className="space-y-1.5">
        {w.movements.map((m, i) => (
          <div key={i} className="flex items-center gap-2 text-sm">
            <span className="flex-1 text-slate-800 dark:text-slate-200">
              {m.reps ? `${m.reps} × ` : ""}{m.movement_name}
              {m.rounds_note && <span className="text-xs text-slate-400"> ({m.rounds_note})</span>}
            </span>
            {editing ? (
              <>
                <input
                  type="number"
                  className="w-16 bg-slate-50 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded px-1.5 py-1 text-xs"
                  placeholder="повт"
                  value={movs[i]?.reps ?? ""}
                  onChange={(e) => setMovs(movs.map((x, xi) => xi === i ? { ...x, reps: e.target.value ? Number(e.target.value) : null } : x))}
                />
                <input
                  type="number"
                  className="w-16 bg-slate-50 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded px-1.5 py-1 text-xs"
                  placeholder="М кг"
                  value={movs[i]?.weight_male ?? ""}
                  onChange={(e) => setMovs(movs.map((x, xi) => xi === i ? { ...x, weight_male: e.target.value ? Number(e.target.value) : null } : x))}
                />
                <input
                  type="number"
                  className="w-16 bg-slate-50 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded px-1.5 py-1 text-xs"
                  placeholder="Ж кг"
                  value={movs[i]?.weight_female ?? ""}
                  onChange={(e) => setMovs(movs.map((x, xi) => xi === i ? { ...x, weight_female: e.target.value ? Number(e.target.value) : null } : x))}
                />
              </>
            ) : (
              <span className="text-slate-400 text-xs">
                {m.weight_male != null && `М ${m.weight_male}`}
                {m.weight_female != null && ` Ж ${m.weight_female}`}
              </span>
            )}
            {m.scaling_note && (
              <span className="text-[10px] text-amber-500 bg-amber-50 dark:bg-amber-950/30 px-1.5 py-0.5 rounded">
                {m.scaling_note}
              </span>
            )}
          </div>
        ))}
      </div>
      {editing && (
        <button
          className="mt-3 bg-emerald-600 hover:bg-emerald-500 px-4 py-1.5 rounded text-sm font-medium"
          onClick={save}
        >
          Сохранить для этой группы
        </button>
      )}
    </div>
  );
}

function LiveRun({
  slot,
  onResult,
  onComplete,
}: {
  slot: SlotDetail;
  onResult: (prBanner: string) => void;
  onComplete: () => void;
}) {
  const [athletes, setAthletes] = useState<Athlete[]>([]);
  const [athleteId, setAthleteId] = useState("");
  const [timeMin, setTimeMin] = useState("");
  const [timeSec, setTimeSec] = useState("");
  const [rounds, setRounds] = useState("");
  const [reps, setReps] = useState("");
  const [rpe, setRpe] = useState("7");
  const [version, setVersion] = useState("rx");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    api.athletes.list().then(setAthletes).catch(() => setAthletes([]));
  }, []);

  const submit = async () => {
    if (!athleteId) return;
    setBusy(true);
    setErr(null);
    try {
      const totalSec =
        timeMin || timeSec ? (Number(timeMin || 0) * 60 + Number(timeSec || 0)) : null;
      const res = await api.slots.saveResult(slot.id, {
        athlete_id: athleteId,
        time_seconds: totalSec,
        rounds: rounds ? Number(rounds) : null,
        reps: reps ? Number(reps) : null,
        rpe: Number(rpe) || null,
        scaled_version: version,
      });
      onResult(res.pr.is_pr ? res.pr.description || "Новый личный рекорд!" : "Результат записан");
      setTimeMin(""); setTimeSec(""); setRounds(""); setReps("");
    } catch (e) {
      setErr((e as Error).message);
    } finally {
      setBusy(false);
    }
  };

  return (
    <div>
      <div className="bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-300 dark:border-emerald-500/40 rounded-xl p-4 mb-4">
        <p className="text-sm font-semibold text-emerald-700 dark:text-emerald-300 mb-1">
          ● Тренировка идёт — WoD на мониторе
        </p>
        <p className="text-xs text-slate-500 dark:text-slate-400 mb-3">
          Сессия: {slot.session_id ? "привязана" : "—"} · участников: {slot.participants.length}
        </p>
        {slot.wod && (
          <div className="text-sm text-slate-600 dark:text-slate-300">
            {slot.wod.movements.map((m) => (m.reps ? `${m.reps}×` : "") + m.movement_name).join(" → ")}
          </div>
        )}
      </div>

      <h3 className="text-sm font-semibold text-slate-600 dark:text-slate-300 mb-2">
        Записать результат
      </h3>
      <div className="bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-xl p-4 mb-4 space-y-3">
        <select
          className="w-full bg-slate-50 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded px-3 py-2 text-sm text-slate-900 dark:text-white"
          value={athleteId}
          onChange={(e) => setAthleteId(e.target.value)}
        >
          <option value="">— спортсмен —</option>
          {athletes.map((a) => (
            <option key={a.id} value={a.id}>{a.name}</option>
          ))}
        </select>
        <div className="grid grid-cols-3 gap-2">
          <label className="text-xs text-slate-400">
            Время (мин:сек)
            <div className="flex gap-1 mt-1">
              <input type="number" placeholder="мин" className="w-full bg-slate-50 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded px-2 py-1.5 text-sm" value={timeMin} onChange={(e) => setTimeMin(e.target.value)} />
              <input type="number" placeholder="сек" className="w-full bg-slate-50 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded px-2 py-1.5 text-sm" value={timeSec} onChange={(e) => setTimeSec(e.target.value)} />
            </div>
          </label>
          <label className="text-xs text-slate-400">
            Раунды / повторения
            <div className="flex gap-1 mt-1">
              <input type="number" placeholder="раунды" className="w-full bg-slate-50 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded px-2 py-1.5 text-sm" value={rounds} onChange={(e) => setRounds(e.target.value)} />
              <input type="number" placeholder="репсы" className="w-full bg-slate-50 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded px-2 py-1.5 text-sm" value={reps} onChange={(e) => setReps(e.target.value)} />
            </div>
          </label>
          <label className="text-xs text-slate-400">
            RPE (1–10)
            <input type="number" min={1} max={10} className="mt-1 w-full bg-slate-50 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded px-2 py-1.5 text-sm" value={rpe} onChange={(e) => setRpe(e.target.value)} />
          </label>
        </div>
        <div className="flex items-center gap-2 text-xs">
          <span className="text-slate-400">Версия:</span>
          {[["rx", "RX"], ["intermediate", "Intermediate"], ["beginner", "Beginner"]].map(([id, label]) => (
            <button
              key={id}
              onClick={() => setVersion(id)}
              className={`px-2 py-1 rounded ${version === id ? "bg-slate-800 dark:bg-slate-200 text-white dark:text-slate-900 font-semibold" : "bg-slate-100 dark:bg-slate-800 text-slate-500"}`}
            >
              {label}
            </button>
          ))}
        </div>
        {err && <p className="text-xs text-red-500">{err}</p>}
        <button
          className="bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 px-4 py-2 rounded text-sm font-medium"
          disabled={!athleteId || busy}
          onClick={submit}
        >
          {busy ? "Запись…" : "Сохранить результат"}
        </button>
      </div>

      <ResultsList results={slot.results} />

      <button
        className="w-full bg-red-600 hover:bg-red-500 px-4 py-2.5 rounded text-sm font-semibold"
        onClick={onComplete}
      >
        ■ Завершить тренировку
      </button>
    </div>
  );
}

function ResultsList({ results }: { results: WorkoutResult[] }) {
  if (results.length === 0) {
    return (
      <p className="text-sm text-slate-400 mb-4">
        Результатов пока нет.
      </p>
    );
  }
  return (
    <div className="mb-4">
      <h3 className="text-sm font-semibold text-slate-600 dark:text-slate-300 mb-2">
        Результаты ({results.length})
      </h3>
      <div className="space-y-1">
        {results.map((r) => (
          <div
            key={r.id}
            className="flex items-center justify-between bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2 text-sm"
          >
            <span className="font-medium text-slate-800 dark:text-slate-200">{r.athlete_name}</span>
            <span className="text-slate-500 dark:text-slate-400">
              {r.time_seconds != null && fmtTime(r.time_seconds)}
              {r.rounds != null && ` ${r.rounds} р.`}
              {r.reps != null && ` ${r.reps} повт.`}
              {r.weight_kg != null && ` ${r.weight_kg} кг`}
              {r.scaled_version && r.scaled_version !== "rx" && ` (${r.scaled_version})`}
              {r.rpe != null && ` · RPE ${r.rpe}`}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

// Тренировки, выводящие день за 65 минут: отдельная группа, выбор недоступен.
function OverLimitList({
  items,
  overBy,
}: {
  items: (WodTemplateItem | Recommendation)[];
  overBy: (t: WodTemplateItem | Recommendation) => number;
}) {
  if (items.length === 0) return null;
  return (
    <div className="pt-2">
      <h4 className="text-xs font-semibold uppercase tracking-wide text-slate-400 mb-1.5">
        Выходит за временной лимит
      </h4>
      <div className="space-y-1">
        {items.map((t) => (
          <div
            key={t.template_id}
            className="flex items-center gap-2 text-sm px-3 py-2 rounded-lg bg-slate-100 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-700 opacity-70"
            title={`День займёт ${SLOT_DAY_MAX_MIN + overBy(t)} из ${SLOT_DAY_MAX_MIN} мин`}
          >
            <span className="flex-1 truncate text-slate-500 dark:text-slate-400">
              {t.name}
            </span>
            <span className="text-xs text-slate-400 shrink-0">{t.duration_min} мин</span>
            <span className="text-[10px] text-red-400 shrink-0">+{overBy(t)} мин перебор</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function fmtTime(sec: number) {
  return `${String(Math.floor(sec / 60)).padStart(2, "0")}:${String(sec % 60).padStart(2, "0")}`;
}

function levelLabel(id: string) {
  return LEVEL_LABELS[id] || id;
}
