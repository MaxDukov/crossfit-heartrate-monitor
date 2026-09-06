import { useEffect, useState, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { api } from "../lib/api";
import type { CycleDetail, SlotView, CycleAnalytics } from "../types";
import { STATUS_LABELS, MODALITY_LABELS, SLOT_DAY_MIN, SLOT_WARMUP_MIN, SLOT_COOLDOWN_MIN, SLOT_WOD_CAP, SLOT_FREE_MIN, SLOT_DENSE_TOTAL, wodSummary } from "../types";
import SlotPanel from "../components/planning/SlotPanel";
import CycleAnalyticsView from "../components/planning/CycleAnalyticsView";

const STATUS_STYLES: Record<string, string> = {
  empty: "bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-slate-700",
  planned: "bg-blue-50 dark:bg-blue-950/30 border-blue-300 dark:border-blue-500/40",
  in_progress: "bg-emerald-50 dark:bg-emerald-950/30 border-emerald-400 dark:border-emerald-500/50 animate-pulse",
  completed: "bg-emerald-100 dark:bg-emerald-900/40 border-emerald-500 dark:border-emerald-600",
  skipped: "bg-red-50 dark:bg-red-950/20 border-red-200 dark:border-red-500/20 opacity-60",
};

// Третьи дни (вне цикла) выделяются тёмно-серым фоном.
const OFF_CYCLE_STYLE = "bg-slate-300 dark:bg-slate-800/60 border-slate-400 dark:border-slate-600";

export default function CycleDetailPage() {
  const { cycleId } = useParams<{ cycleId: string }>();
  const [cycle, setCycle] = useState<CycleDetail | null>(null);
  const [analytics, setAnalytics] = useState<CycleAnalytics | null>(null);
  const [tab, setTab] = useState<"calendar" | "analytics">("calendar");
  const [selectedSlot, setSelectedSlot] = useState<string | null>(null);
  const [replan, setReplan] = useState<{ groupId: string; groupName: string } | null>(null);
  const [toggleBusy, setToggleBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  const load = useCallback(async () => {
    if (!cycleId) return;
    try {
      const c = await api.cycles.get(cycleId);
      setCycle(c);
    } catch (e) {
      setError((e as Error).message);
    }
  }, [cycleId]);

  useEffect(() => { load(); }, [load]);

  useEffect(() => {
    if (cycleId && tab === "analytics") {
      api.cycles.analytics(cycleId).then(setAnalytics).catch(() => setAnalytics(null));
    }
  }, [cycleId, tab]);

  if (error) {
    return (
      <div className="p-6 text-center text-red-500">
        {error}{" "}
        <button className="underline" onClick={() => navigate("/wod")}>к списку циклов</button>
      </div>
    );
  }
  if (!cycle) return <div className="p-6 text-slate-400">Загрузка…</div>;

  const setStatus = async (status: string) => {
    await api.cycles.setStatus(cycle.id, status);
    load();
  };

  const toggleThirdDay = async (groupId: string, groupName: string, enabled: boolean) => {
    setToggleBusy(true);
    setError(null);
    try {
      await api.cycles.setThirdDayOff(cycle.id, groupId, enabled);
      load();
    } catch (e) {
      const err = e as Error & { status?: number };
      if (err.status === 409) {
        // Конфликт: на третьи дни уже назначены тренировки — спросить про перепланирование.
        setReplan({ groupId, groupName });
      } else {
        setError(err.message);
      }
    } finally {
      setToggleBusy(false);
    }
  };

  const doReplan = async (mode: "rollback" | "auto") => {
    if (!replan) return;
    setToggleBusy(true);
    try {
      await api.cycles.replanThirdDayOff(cycle.id, replan.groupId, mode);
      setReplan(null);
      load();
    } catch (e) {
      setError((e as Error).message);
    } finally {
      setToggleBusy(false);
    }
  };

  const removeWod = async (slotId: string, wodId: string) => {
    try {
      await api.slots.unassignWod(slotId, wodId);
      load();
    } catch (e) {
      setError((e as Error).message);
    }
  };

  return (
    <div className="h-full flex flex-col overflow-hidden">
      {/* Шапка */}
      <div className="px-6 pt-4 pb-3 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shrink-0">
        <div className="flex items-start justify-between mb-2">
          <div>
            <button
              className="text-sm text-slate-400 hover:text-slate-600 dark:hover:text-slate-300 mb-1"
              onClick={() => navigate("/wod")}
            >
              ← Циклы
            </button>
            <h1 className="text-2xl font-bold text-slate-900 dark:text-white">
              {cycle.name}
              <span className="ml-3 text-sm font-normal text-slate-400">
                {cycle.goal} · {cycle.weeks} нед. ·{" "}
                {cycle.modality_priority
                  ? `приоритет: ${MODALITY_LABELS[cycle.modality_priority] || cycle.modality_priority}`
                  : ""}
              </span>
            </h1>
          </div>
          <div className="flex items-center gap-2">
            {cycle.status === "planned" && (
              <button
                className="bg-emerald-600 hover:bg-emerald-500 px-4 py-2 rounded text-sm font-medium"
                onClick={() => setStatus("active")}
              >
                Запустить цикл
              </button>
            )}
            {cycle.status === "active" && (
              <button
                className="bg-slate-700 hover:bg-slate-600 px-4 py-2 rounded text-sm font-medium"
                onClick={() => setStatus("completed")}
              >
                Завершить цикл
              </button>
            )}
          </div>
        </div>
        <div className="flex items-center gap-4 text-sm">
          <div className="flex-1 h-2 bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden max-w-xs">
            <div className="h-full bg-emerald-500" style={{ width: `${Math.min(100, cycle.fill_percent)}%` }} />
          </div>
          <span className="text-slate-400">
            заполнено {Math.round(cycle.fill_percent)}% · проведено {cycle.slots_completed}/{cycle.slots_total}
          </span>
          <div className="flex gap-1 ml-auto">
            {(["calendar", "analytics"] as const).map((t) => (
              <button
                key={t}
                onClick={() => setTab(t)}
                className={`px-4 py-1.5 rounded text-sm font-medium ${
                  tab === t
                    ? "bg-slate-100 dark:bg-slate-800 text-slate-900 dark:text-white"
                    : "text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
                }`}
              >
                {t === "calendar" ? "Календарь" : "Аналитика"}
              </button>
            ))}
          </div>
        </div>
        {cycle.warnings && cycle.warnings.length > 0 && (
          <div className="mt-3 bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-500/30 rounded-lg p-2 text-sm text-amber-700 dark:text-amber-300">
            {cycle.warnings.map((w, i) => <p key={i}>⚠ {w}</p>)}
          </div>
        )}
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto">
        {tab === "calendar" ? (
          <div className="max-w-5xl mx-auto p-6 space-y-8">
            {cycle.groups.map((g) => (
              <GroupCalendar
                key={g.id}
                group={g}
                slots={cycle.slots.filter((s) => s.group_id === g.id)}
                onSelect={setSelectedSlot}
                onRemoveWod={removeWod}
                onToggleThirdDay={(gid, enabled) => toggleThirdDay(gid, g.name, enabled)}
                toggleBusy={toggleBusy}
              />
            ))}
          </div>
        ) : (
          <CycleAnalyticsView analytics={analytics} cycleId={cycle.id} />
        )}
      </div>

      {selectedSlot && (
        <SlotPanel
          slotId={selectedSlot}
          onClose={() => setSelectedSlot(null)}
          onChanged={() => {
            load();
          }}
        />
      )}

      {replan && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-slate-900 rounded-xl max-w-md w-full p-6">
            <h3 className="text-lg font-bold text-slate-900 dark:text-white mb-2">
              Перепланирование · {replan.groupName}
            </h3>
            <p className="text-sm text-slate-600 dark:text-slate-300 mb-5">
              Третий день был включён в тренировочный цикл. Начать перепланирование?
            </p>
            <div className="flex flex-col gap-2">
              <button
                disabled={toggleBusy}
                className="px-4 py-2 rounded text-sm font-medium bg-blue-600 hover:bg-blue-500 text-white disabled:opacity-50"
                onClick={() => doReplan("auto")}
              >
                Автоперепланирование — перенести тренировки на первые два дня недели
              </button>
              <button
                disabled={toggleBusy}
                className="px-4 py-2 rounded text-sm font-medium bg-red-600/90 hover:bg-red-500 text-white disabled:opacity-50"
                onClick={() => doReplan("rollback")}
              >
                Полный откат — снять все будущие запланированные тренировки
              </button>
              <button
                disabled={toggleBusy}
                className="px-4 py-2 rounded text-sm text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800 disabled:opacity-50"
                onClick={() => setReplan(null)}
              >
                Отмена — оставить всё как было
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function GroupCalendar({
  group,
  slots,
  onSelect,
  onRemoveWod,
  onToggleThirdDay,
  toggleBusy,
}: {
  group: CycleDetail["groups"][number];
  slots: SlotView[];
  onSelect: (id: string) => void;
  onRemoveWod: (slotId: string, wodId: string) => void;
  onToggleThirdDay: (groupId: string, enabled: boolean) => void;
  toggleBusy: boolean;
}) {
  // Группировка слотов по 7-дневным неделям от первого слота.
  const weeks: SlotView[][] = [];
  for (const s of slots) {
    const first = new Date(slots[0].slot_date + "T00:00:00");
    const cur = new Date(s.slot_date + "T00:00:00");
    const weekIdx = Math.floor((cur.getTime() - first.getTime()) / (7 * 86400000));
    if (!weeks[weekIdx]) weeks[weekIdx] = [];
    weeks[weekIdx].push(s);
  }

  return (
    <div>
      <div className="flex items-center gap-2 mb-3">
        <h2 className="text-lg font-bold text-slate-800 dark:text-slate-200">{group.name}</h2>
        <span className="text-sm text-slate-400">{group.weekdays_names.join(" · ")}</span>
      </div>
      {weeks.length > 0 && (
        <div className="flex gap-2 mb-2">
          <span className="w-14 shrink-0" />
          {weeks[0].map((s, idx) => (
            <div key={s.id} className="w-44 shrink-0">
              {idx === 2 && <ThirdDaySwitch group={group} onToggle={onToggleThirdDay} disabled={toggleBusy} />}
            </div>
          ))}
        </div>
      )}
      <div className="space-y-2">
        {weeks.map((week, wi) => (
          <div key={wi} className="flex gap-2">
            <span className="text-xs text-slate-400 w-14 shrink-0 self-center">
              нед. {wi + 1}
            </span>
            <div className="flex gap-2 flex-wrap">
              {week.map((s) => (
                <DayCard key={s.id} s={s} onSelect={onSelect} onRemoveWod={onRemoveWod} />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

const pctOf = (min: number) => `${(min / SLOT_DAY_MIN) * 100}%`;

const SEG_MUTED =
  "bg-black/[0.04] dark:bg-white/[0.06] border border-dashed border-slate-300/70 dark:border-slate-600/60 " +
  "text-slate-400 dark:text-slate-500 text-[9px] leading-none flex items-center justify-center overflow-hidden whitespace-nowrap";

// Карточка дня: 60 минут = разминка 10 + тренировки (по длительности) + заминка 5.
// Свободное время >10 мин — слот «Подобрать тренировку».
function DayCard({
  s,
  onSelect,
  onRemoveWod,
}: {
  s: SlotView;
  onSelect: (id: string) => void;
  onRemoveWod: (slotId: string, wodId: string) => void;
}) {
  const wods = s.wods ?? [];
  const total = wods.reduce((x, w) => x + w.duration_min, 0);
  const left = SLOT_WOD_CAP - total;

  if (wods.length === 0) {
    return (
      <button
        onClick={() => onSelect(s.id)}
        className={`w-44 min-h-[6.5rem] text-left border rounded-lg p-3 transition-all hover:shadow-md ${
          s.kind === "off_cycle" ? OFF_CYCLE_STYLE : (STATUS_STYLES[s.status] || STATUS_STYLES.empty)
        }`}
      >
        <div className="flex items-center justify-between mb-1">
          <span className="text-xs text-slate-400">
            День {s.day_number} · {fmtDate(s.slot_date)}
          </span>
          <span
            className={`text-[10px] font-semibold px-1.5 py-0.5 rounded ${
              s.kind === "off_cycle" ? "bg-amber-500/15 text-amber-600 dark:text-amber-400" : ""
            }`}
            title={s.kind === "off_cycle" ? "День вне цикла" : undefined}
          >
            {s.kind === "off_cycle" ? "ВЦ" : ""}
          </span>
        </div>
        <div className="text-sm text-slate-400">+ подобрать тренировку</div>
      </button>
    );
  }

  const dense = total + SLOT_WARMUP_MIN + SLOT_COOLDOWN_MIN > SLOT_DENSE_TOTAL;

  return (
    <div
      onClick={() => onSelect(s.id)}
      className={`w-44 h-40 flex flex-col text-left border rounded-lg overflow-hidden cursor-pointer transition-all hover:shadow-md ${
        s.kind === "off_cycle" ? OFF_CYCLE_STYLE : (STATUS_STYLES[s.status] || STATUS_STYLES.empty)
      }`}
    >
      <div className="flex items-center justify-between gap-1 px-2 pt-1.5 pb-1 shrink-0">
        <span className="text-[10px] text-slate-500 dark:text-slate-400 truncate">
          День {s.day_number} · {fmtDate(s.slot_date)}
        </span>
        <span className="flex items-center gap-1 shrink-0">
          {s.kind === "off_cycle" && (
            <span
              className="text-[9px] font-semibold px-1 py-0.5 rounded bg-amber-500/15 text-amber-600 dark:text-amber-400"
              title="День вне цикла"
            >
              ВЦ
            </span>
          )}
          <span
            className={`text-[9px] font-semibold ${
              s.status === "completed" || s.status === "in_progress"
                ? "text-emerald-600 dark:text-emerald-400"
                : "text-blue-600 dark:text-blue-400"
            }`}
          >
            {STATUS_LABELS[s.status]}
          </span>
        </span>
      </div>

      <div className="flex-1 flex flex-col gap-px px-1 pb-1 min-h-0">
        <div style={{ flexBasis: pctOf(SLOT_WARMUP_MIN), minHeight: 13 }} className={SEG_MUTED}>
          Разминка · {SLOT_WARMUP_MIN}м
        </div>
        {wods.map((w) => (
          <div
            key={w.id}
            style={{ flexBasis: pctOf(w.duration_min) }}
            className="min-h-[18px] bg-white/70 dark:bg-slate-900/50 border border-slate-200 dark:border-slate-700 rounded flex items-center gap-1 px-1.5 overflow-hidden"
            title={wodSummary(w, w.name)}
          >
            <span className="text-[11px] font-medium text-slate-800 dark:text-slate-200 truncate flex-1">
              {w.name}
            </span>
            <span className="text-[9px] text-slate-400 shrink-0">{w.duration_min}м</span>
            {s.status === "planned" && (
              <button
                className="text-slate-400 hover:text-red-500 text-[10px] shrink-0 leading-none"
                title="Снять тренировку"
                onClick={(e) => {
                  e.stopPropagation();
                  onRemoveWod(s.id, w.id);
                }}
              >
                ✕
              </button>
            )}
          </div>
        ))}
        {left > SLOT_FREE_MIN && (
          <button
            style={{ flexBasis: pctOf(left) }}
            className="min-h-[18px] border border-dashed border-emerald-500/50 text-emerald-600 dark:text-emerald-400 text-[10px] leading-none rounded flex items-center justify-center overflow-hidden whitespace-nowrap hover:bg-emerald-50 dark:hover:bg-emerald-950/30"
            onClick={(e) => {
              e.stopPropagation();
              onSelect(s.id);
            }}
          >
            Подобрать тренировку
          </button>
        )}
        <div style={{ flexBasis: pctOf(SLOT_COOLDOWN_MIN), minHeight: 12 }} className={SEG_MUTED}>
          Заминка · {SLOT_COOLDOWN_MIN}м
        </div>
      </div>

      {dense && (
        <div
          className="text-[9px] text-amber-600 dark:text-amber-400 bg-amber-500/10 px-2 py-0.5 shrink-0"
          title="Сумма длительностей с разминкой и заминкой превышает 55 минут"
        >
          ⚠ плотный день · {total + SLOT_WARMUP_MIN + SLOT_COOLDOWN_MIN}м
        </div>
      )}
    </div>
  );
}

function ThirdDaySwitch({
  group,
  onToggle,
  disabled,
}: {
  group: CycleDetail["groups"][number];
  onToggle: (groupId: string, enabled: boolean) => void;
  disabled: boolean;
}) {
  const on = group.third_day_off;
  return (
    <div
      className="border border-slate-200 dark:border-slate-700 rounded-lg bg-white dark:bg-slate-900 p-2"
      title="Каждый 3-й тренировочный день планируется отдельно (техника, тесты 1ПМ)"
    >
      <div className="text-[10px] font-semibold uppercase tracking-wide text-slate-400 mb-1.5">
        3-й день
      </div>
      <button
        role="switch"
        aria-checked={on}
        disabled={disabled}
        onClick={() => onToggle(group.id, !on)}
        className={`flex items-center gap-2 ${disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer"}`}
      >
        <span
          className={`relative inline-block w-9 h-5 rounded-full transition-colors shrink-0 ${
            on ? "bg-amber-500" : "bg-slate-300 dark:bg-slate-600"
          }`}
        >
          <span
            className={`absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white shadow transition-transform ${
              on ? "translate-x-4" : ""
            }`}
          />
        </span>
        <span className={`text-xs ${on ? "text-amber-600 dark:text-amber-400 font-medium" : "text-slate-500 dark:text-slate-400"}`}>
          {on ? "вне цикла" : "в цикле"}
        </span>
      </button>
    </div>
  );
}

function fmtDate(iso: string) {
  const d = new Date(iso + "T00:00:00");
  return d.toLocaleDateString("ru", { day: "numeric", month: "short" });
}
