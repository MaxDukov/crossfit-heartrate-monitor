import { useStyle, type DashboardStyle } from "../lib/style";
import { useSearchParams } from "react-router-dom";
import AthletesPage from "./AthletesPage";
import SensorsPage from "./SensorsPage";

type SettingsTab = "general" | "athletes" | "sensors";

const TABS: { id: SettingsTab; label: string }[] = [
  { id: "general", label: "Общие" },
  { id: "athletes", label: "Спортсмены" },
  { id: "sensors", label: "Датчики" },
];

export default function SettingsPage() {
  const [params, setParams] = useSearchParams();
  const tabParam = params.get("tab") as SettingsTab | null;
  const tab: SettingsTab = TABS.some((t) => t.id === tabParam) ? tabParam! : "general";

  return (
    <div className="h-full flex flex-col overflow-hidden">
      <div className="px-6 pt-4 pb-2 border-b border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shrink-0">
        <h1 className="text-2xl font-bold text-slate-900 dark:text-white mb-3">Настройки</h1>
        <div className="flex gap-1">
          {TABS.map((t) => (
            <button
              key={t.id}
              onClick={() => setParams(t.id === "general" ? {} : { tab: t.id }, { replace: true })}
              className={`px-4 py-2 rounded-t-lg text-sm font-medium transition-colors ${
                tab === t.id
                  ? "bg-slate-100 dark:bg-slate-800 text-slate-900 dark:text-white"
                  : "text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>
      <div className="flex-1 min-h-0 overflow-hidden">
        {tab === "general" && <GeneralSettings />}
        {tab === "athletes" && <AthletesPage />}
        {tab === "sensors" && <SensorsPage />}
      </div>
    </div>
  );
}

const STYLES: {
  id: DashboardStyle;
  name: string;
  desc: string;
  preview: React.CSSProperties;
}[] = [
  {
    id: "basic",
    name: "Basic",
    desc: "Сетка карточек с графиком пульса. Классический вид.",
    preview: {
      display: "grid",
      gridTemplateColumns: "1fr 1fr",
      gridTemplateRows: "1fr 1fr",
      gap: 3,
      padding: 6,
    },
  },
  {
    id: "tri-column",
    name: "Tri-column",
    desc: "Две колонки спортсменов по краям, WOD и таймер в центре.",
    preview: {
      display: "grid",
      gridTemplateColumns: "1fr 2fr 1fr",
      gap: 3,
      padding: 6,
    },
  },
  {
    id: "split-screen",
    name: "Split-screen",
    desc: "Главный блок слева, карточки справа и снизу.",
    preview: {
      display: "grid",
      gridTemplateColumns: "3fr 1fr",
      gridTemplateRows: "1fr 40px",
      gap: 3,
      padding: 6,
    },
  },
];

function GeneralSettings() {
  const style = useStyle((s) => s.style);
  const setStyle = useStyle((s) => s.setStyle);

  return (
    <div className="h-full overflow-auto p-6 max-w-4xl mx-auto">
      <p className="text-slate-500 dark:text-slate-400 mb-8">
        Выберите стиль отображения монитора
      </p>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        {STYLES.map((s) => {
          const active = style === s.id;
          return (
            <button
              key={s.id}
              onClick={() => setStyle(s.id)}
              className={`rounded-xl border-2 p-4 text-left transition-all ${
                active
                  ? "border-emerald-500 bg-emerald-50 dark:bg-emerald-950/30"
                  : "border-slate-200 dark:border-slate-700 hover:border-slate-300 dark:hover:border-slate-600"
              }`}
            >
              <div
                className="rounded-lg mb-3 h-28 overflow-hidden"
                style={{
                  background: active
                    ? "rgba(16,185,129,0.08)"
                    : "rgba(148,163,184,0.08)",
                }}
              >
                <StylePreview id={s.id} previewStyle={s.preview} active={active} />
              </div>
              <div className="flex items-center gap-2 mb-1">
                <span
                  className={`font-bold ${
                    active
                      ? "text-emerald-600 dark:text-emerald-400"
                      : "text-slate-900 dark:text-white"
                  }`}
                >
                  {s.name}
                </span>
                {active && (
                  <span className="text-xs font-semibold text-emerald-600 dark:text-emerald-400 bg-emerald-100 dark:bg-emerald-900/40 px-2 py-0.5 rounded-full">
                    активно
                  </span>
                )}
              </div>
              <p className="text-sm text-slate-500 dark:text-slate-400">
                {s.desc}
              </p>
            </button>
          );
        })}
      </div>

      <p className="text-sm text-slate-400 dark:text-slate-500 mt-6">
        Стиль применяется мгновенно ко всем экранам, включая Demo режим.
      </p>
    </div>
  );
}

function StylePreview({
  id,
  previewStyle,
  active,
}: {
  id: DashboardStyle;
  previewStyle: React.CSSProperties;
  active: boolean;
}) {
  const cardBg = active ? "rgba(16,185,129,0.2)" : "rgba(148,163,184,0.2)";
  const centerBg = active ? "rgba(16,185,129,0.1)" : "rgba(148,163,184,0.1)";

  if (id === "basic") {
    return (
      <div style={previewStyle} className="h-full">
        {[0, 1, 2, 3].map((i) => (
          <div
            key={i}
            className="rounded"
            style={{ background: cardBg }}
          />
        ))}
      </div>
    );
  }

  if (id === "tri-column") {
    return (
      <div style={previewStyle} className="h-full">
        <div className="flex flex-col gap-1">
          {[0, 1].map((i) => (
            <div key={i} className="flex-1 rounded" style={{ background: cardBg }} />
          ))}
        </div>
        <div className="rounded flex items-center justify-center" style={{ background: centerBg }}>
          <span className="text-[8px] text-slate-400">WOD</span>
        </div>
        <div className="flex flex-col gap-1">
          {[0, 1].map((i) => (
            <div key={i} className="flex-1 rounded" style={{ background: cardBg }} />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div style={previewStyle} className="h-full">
      <div className="rounded flex items-center justify-center" style={{ background: centerBg, gridColumn: "1", gridRow: "1" }}>
        <span className="text-[8px] text-slate-400">WOD</span>
      </div>
      <div className="flex flex-col gap-1" style={{ gridColumn: "2", gridRow: "1" }}>
        {[0, 1].map((i) => (
          <div key={i} className="flex-1 rounded" style={{ background: cardBg }} />
        ))}
      </div>
      <div className="flex gap-1" style={{ gridColumn: "1 / 3", gridRow: "2" }}>
        {[0, 1, 2].map((i) => (
          <div key={i} className="flex-1 rounded" style={{ background: cardBg }} />
        ))}
      </div>
    </div>
  );
}
