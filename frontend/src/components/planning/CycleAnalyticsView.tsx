import type { CycleAnalytics } from "../../types";

// Аналитика цикла — Экран 7 draft1.MD.
export default function CycleAnalyticsView({
  analytics,
  cycleId,
}: {
  analytics: CycleAnalytics | null;
  cycleId: string;
}) {
  if (!analytics) return <div className="p-6 text-slate-400">Загрузка аналитики…</div>;
  void cycleId;

  return (
    <div className="max-w-4xl mx-auto p-6 space-y-6">
      {/* План/факт */}
      <div className="grid grid-cols-4 gap-3">
        <Stat label="Слотов" value={analytics.slots_total} />
        <Stat label="Заполнено" value={`${Math.round(analytics.fill_percent)}%`} accent="blue" />
        <Stat label="Проведено" value={`${analytics.slots_completed}`} accent="emerald" />
        <Stat label="Пропущено" value={analytics.slots_skipped} accent="red" />
      </div>

      {/* Группы */}
      {analytics.groups.length > 0 && (
        <Section title="Группы: план → факт">
          <div className="space-y-2">
            {analytics.groups.map((g) => (
              <div key={g.group_id} className="flex items-center gap-3">
                <span className="w-28 text-sm font-medium text-slate-700 dark:text-slate-300 truncate">
                  {g.name}
                </span>
                <div className="flex-1 h-3 bg-slate-100 dark:bg-slate-800 rounded-full overflow-hidden flex">
                  <div className="h-full bg-emerald-500" style={{ width: `${pct(g.slots_completed, g.slots_total)}%` }} />
                  <div className="h-full bg-blue-400" style={{ width: `${pct(g.slots_filled - g.slots_completed, g.slots_total)}%` }} />
                </div>
                <span className="text-xs text-slate-400 w-24 text-right">
                  {g.slots_completed}/{g.slots_filled}/{g.slots_total}
                </span>
              </div>
            ))}
          </div>
        </Section>
      )}

      {/* Темы: план vs факт */}
      {analytics.themes.length > 0 && (
        <Section title="Модальности: запланировано / проведено">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
            {analytics.themes.map((t) => (
              <div
                key={t.theme}
                className={`rounded-lg p-3 border text-sm ${
                  t.done >= t.planned
                    ? "border-emerald-300 dark:border-emerald-500/40 bg-emerald-50 dark:bg-emerald-950/20"
                    : t.planned >= 2 && t.done === 0
                      ? "border-red-200 dark:border-red-500/30 bg-red-50 dark:bg-red-950/20"
                      : "border-slate-200 dark:border-slate-700"
                }`}
              >
                <div className="font-medium text-slate-700 dark:text-slate-300">{t.theme_name}</div>
                <div className="text-xs text-slate-400">
                  план {t.planned} · факт {t.done}
                </div>
              </div>
            ))}
          </div>
        </Section>
      )}

      {/* Спортсмены */}
      {analytics.athletes.length > 0 && (
        <Section title="Спортсмены">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-xs text-slate-400 border-b border-slate-200 dark:border-slate-700">
                  <th className="py-2">Имя</th>
                  <th className="py-2">Тренировок</th>
                  <th className="py-2">Объём, кг</th>
                  <th className="py-2">Ср. RPE</th>
                  <th className="py-2">PR</th>
                </tr>
              </thead>
              <tbody>
                {analytics.athletes.map((a) => (
                  <tr key={a.athlete_id} className="border-b border-slate-100 dark:border-slate-800">
                    <td className="py-2 font-medium text-slate-800 dark:text-slate-200">{a.name}</td>
                    <td className="py-2 text-slate-500">{a.workouts_done}</td>
                    <td className="py-2 text-slate-500">{Math.round(a.total_volume_kg).toLocaleString("ru")}</td>
                    <td className="py-2 text-slate-500">{a.avg_rpe != null ? a.avg_rpe.toFixed(1) : "—"}</td>
                    <td className="py-2 text-emerald-600 dark:text-emerald-400 font-semibold">{a.pr_count}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Section>
      )}

      {/* Динамика 1ПМ */}
      {analytics.rm_progress.length > 0 && (
        <Section title="Динамика 1ПМ (Epley)">
          <div className="space-y-1">
            {analytics.rm_progress.map((r, i) => (
              <div key={i} className="flex items-center justify-between text-sm bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-lg px-3 py-2">
                <span className="text-slate-700 dark:text-slate-300">
                  {r.athlete_name} · {r.movement_name}
                </span>
                <span className="font-semibold text-emerald-600 dark:text-emerald-400">
                  {r.value} кг <span className="text-xs text-slate-400">({r.date})</span>
                </span>
              </div>
            ))}
          </div>
        </Section>
      )}

      {/* Советы системы */}
      <Section title="Рекомендации системы">
        <ul className="space-y-1">
          {analytics.advice.map((a, i) => (
            <li key={i} className="text-sm text-slate-600 dark:text-slate-300 flex gap-2">
              <span className="text-emerald-500">→</span> {a}
            </li>
          ))}
        </ul>
      </Section>
    </div>
  );
}

function pct(v: number, total: number) {
  return total > 0 ? (v * 100) / total : 0;
}

function Stat({ label, value, accent }: { label: string; value: string | number; accent?: string }) {
  const colors =
    accent === "emerald"
      ? "text-emerald-600 dark:text-emerald-400"
      : accent === "blue"
        ? "text-blue-600 dark:text-blue-400"
        : accent === "red"
          ? "text-red-500"
          : "text-slate-900 dark:text-white";
  return (
    <div className="bg-white dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 rounded-xl p-4 text-center">
      <div className={`text-2xl font-bold ${colors}`}>{value}</div>
      <div className="text-xs text-slate-400 mt-1">{label}</div>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <h3 className="text-sm font-semibold text-slate-600 dark:text-slate-300 mb-3">{title}</h3>
      {children}
    </div>
  );
}
