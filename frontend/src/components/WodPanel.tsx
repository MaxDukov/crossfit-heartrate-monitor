import { useEffect, useState } from "react";
import { api } from "../lib/api";
import type { Wod } from "../types";
import { FORMAT_LABELS, LEVEL_LABELS, THEME_LABELS } from "../types";
import WorkoutTimer from "./WorkoutTimer";

export default function WodPanel() {
  const [wod, setWod] = useState<Wod | null>(null);
  const [collapsed, setCollapsed] = useState(false);

  const fetchActive = () => api.wods.active().then(setWod);

  useEffect(() => {
    fetchActive();
    const id = setInterval(fetchActive, 5000);
    return () => clearInterval(id);
  }, []);

  const endWod = async () => {
    await api.wods.endActive();
    setWod(null);
  };

  if (!wod) return null;

  return (
    <div className="bg-slate-900 border-b border-slate-800 shrink-0">
      <div className="px-4 py-2 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span
            className="text-xs font-medium px-2 py-1 rounded bg-slate-700 text-slate-300 cursor-pointer"
            onClick={() => setCollapsed(!collapsed)}
          >
            {collapsed ? "▼" : "▲"} Свернуть
          </span>
          <h2 className="text-lg font-bold text-white">{wod.name}</h2>
          <span className="text-xs bg-blue-600/20 text-blue-400 px-2 py-1 rounded font-medium">
            {FORMAT_LABELS[wod.format] || wod.format}
          </span>
          <span className="text-xs bg-slate-800 text-slate-400 px-2 py-1 rounded">
            {THEME_LABELS[wod.theme] || wod.theme}
          </span>
          <span className="text-xs bg-slate-800 text-slate-400 px-2 py-1 rounded">
            {LEVEL_LABELS[wod.group_level] || wod.group_level}
          </span>
        </div>
        <button
          onClick={endWod}
          className="text-xs px-3 py-1.5 bg-red-600/20 hover:bg-red-600/40 text-red-400 rounded font-medium transition-colors"
        >
          Завершить
        </button>
      </div>

      {!collapsed && (
        <div className="px-4 pb-3">
          {/* Движения */}
          <div className="flex gap-2 flex-wrap mb-3">
            {wod.movements.map((m, idx) => (
              <div
                key={idx}
                className="flex items-center gap-2 bg-slate-800/60 px-3 py-1.5 rounded-lg"
              >
                <span className="text-slate-500 text-xs">{idx + 1}.</span>
                <span className="text-slate-200 text-sm font-medium">
                  {m.movement_name}
                </span>
                {(m.reps || m.rounds_note) && (
                  <span className="text-blue-400 text-sm font-mono">
                    {m.rounds_note || `${m.reps}`}
                  </span>
                )}
                {m.weight_male && (
                  <span className="text-slate-500 text-xs">
                    {m.weight_male}/{m.weight_female}кг
                  </span>
                )}
                {m.scaling_note && (
                  <span className="text-amber-500/70 text-xs">
                    ({m.scaling_note})
                  </span>
                )}
              </div>
            ))}
          </div>

          {/* Описание + таймер */}
          <div className="flex items-center justify-between gap-4">
            {wod.description && (
              <p className="text-sm text-slate-400 flex-1">{wod.description}</p>
            )}
            <div className="shrink-0">
              <WorkoutTimer
                format={wod.format}
                durationMin={wod.duration_min}
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
