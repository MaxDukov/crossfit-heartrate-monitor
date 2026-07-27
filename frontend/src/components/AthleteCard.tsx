import type { HrUpdate } from "../types";
import { ZONE_NAMES, getZoneColors, getZoneGradient } from "../lib/zones";
import { useTheme } from "../lib/theme";
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  YAxis,
  ReferenceLine,
} from "recharts";

interface HrPoint {
  t: number;
  hr: number;
}

interface Props {
  data: HrUpdate;
  history: HrPoint[];
}

export default function AthleteCard({ data, history }: Props) {
  const { theme } = useTheme();
  const zone = data.zone;
  const colors = getZoneColors(theme);
  const gradients = getZoneGradient(theme);
  const zoneName = ZONE_NAMES[zone] || "—";
  const zoneColor = colors[zone] || "#94A3B8";
  const gradient = gradients[zone] || gradients[1];
  const maxHr = data.max_hr || 190;

  const chartData = history.map((p) => ({ hr: p.hr, idx: p.t }));

  const nameColor = theme === "dark" ? "#F1F5F9" : "#0F172A";
  const secondaryColor = theme === "dark" ? "#94A3B8" : "#64748B";

  return (
    <div
      className="border rounded-xl flex flex-col relative overflow-hidden min-h-0 p-2 sm:p-3"
      style={{
        background: gradient,
        borderColor: theme === "dark" ? "rgba(255,255,255,0.08)" : "rgba(0,0,0,0.08)",
        containerType: "inline-size",
        transition: "background 0.8s ease, border-color 0.8s ease",
      }}
    >
      {/* Имя спортсмена */}
      <div
        className="text-center shrink-0 font-bold truncate px-1"
        style={{
          fontSize: "clamp(1.2rem, 16cqw, 12rem)",
          lineHeight: 1.1,
          color: nameColor,
        }}
      >
        {data.athlete_name || `Sensor #${data.device_id}`}
      </div>

      {/* Пульс и калории */}
      <div className="flex-1 flex flex-col items-center justify-center min-h-0">
        <span
          className="font-black tabular-nums leading-none"
          style={{
            fontSize: "clamp(2.5rem, 28cqw, 18rem)",
            color: zoneColor,
            textShadow: theme === "dark" ? "0 4px 12px rgba(0,0,0,0.5)" : "none",
          }}
        >
          {data.heart_rate}<span style={{ fontSize: "0.8em", opacity: 0.6 }}>♥</span>
        </span>
        <span
          className="font-bold tabular-nums leading-none mt-1"
          style={{
            fontSize: "clamp(0.75rem, 8.4cqw, 5.4rem)",
            color: zoneColor,
          }}
        >
          {Math.round(data.calories)}{" "}
          <svg
            width="0.8em"
            height="0.8em"
            viewBox="0 0 24 24"
            fill="currentColor"
            style={{ opacity: 0.6, display: "inline-block", verticalAlign: "middle" }}
          >
            <path d="M13.5.67s.74 2.65.74 4.8c0 2.06-1.35 3.73-3.41 3.73-2.07 0-3.63-1.67-3.63-3.73l.03-.36C5.21 7.51 4 10.62 4 14c0 4.42 3.58 8 8 8s8-3.58 8-8C20 8.61 17.41 3.8 13.5.67zM11.71 19c-1.78 0-3.22-1.4-3.22-3.14 0-1.62 1.05-2.76 2.81-3.12 1.77-.36 3.6-1.21 4.62-2.58.39 1.29.59 2.65.59 4.04 0 2.65-2.15 4.8-4.8 4.8z"/>
          </svg>
        </span>
      </div>

      {/* Зона и % от максимума */}
      <div
        className="shrink-0 flex items-center justify-between mb-1 px-1"
        style={{ fontSize: "clamp(0.5rem, 1.4cqw, 0.95rem)" }}
      >
        <span style={{ color: zoneColor }} className="font-bold truncate">
          {zoneName}
        </span>
        <span style={{ color: secondaryColor }} className="shrink-0 ml-2">
          {Math.round(data.zone_percent)}% · Max {maxHr}
        </span>
      </div>

      {/* График истории пульса */}
      <div
        className="shrink-0"
        style={{ height: "clamp(1.5rem, 16cqw, 6rem)" }}
      >
        {chartData.length > 1 ? (
          <ResponsiveContainer width="100%" height="100%">
            <AreaChart data={chartData} margin={{ top: 2, right: 2, bottom: 0, left: 2 }}>
              <defs>
                <linearGradient id={`grad-${data.device_id}`} x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor={zoneColor} stopOpacity={0.5} />
                  <stop offset="100%" stopColor={zoneColor} stopOpacity={0.02} />
                </linearGradient>
              </defs>
              <YAxis hide domain={[40, maxHr]} />
              <ReferenceLine y={maxHr} stroke="#EF4444" strokeDasharray="4 4" strokeOpacity={0.4} />
              <Area
                type="monotone"
                dataKey="hr"
                stroke={zoneColor}
                strokeWidth={2}
                fill={`url(#grad-${data.device_id})`}
                isAnimationActive={false}
                dot={false}
              />
            </AreaChart>
          </ResponsiveContainer>
        ) : (
          <div className="h-full flex items-center justify-center" style={{ color: secondaryColor, fontSize: "0.65rem" }}>
            Сбор данных...
          </div>
        )}
      </div>
    </div>
  );
}
