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

function polar(cx: number, cy: number, r: number, deg: number) {
  const rad = ((deg - 180) * Math.PI) / 180;
  return { x: cx + r * Math.cos(rad), y: cy + r * Math.sin(rad) };
}

function arc(cx: number, cy: number, r: number, startDeg: number, endDeg: number) {
  const s = polar(cx, cy, r, startDeg);
  const e = polar(cx, cy, r, endDeg);
  const large = endDeg - startDeg > 180 ? 1 : 0;
  return `M ${s.x} ${s.y} A ${r} ${r} 0 ${large} 1 ${e.x} ${e.y}`;
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
  const pct = Math.min(data.zone_percent, 120);
  const needleDeg = (pct / 120) * 180;

  const cx = 100;
  const cy = 100;
  const r = 88;

  const zones = [
    { from: 0, to: 50, color: colors[1] },
    { from: 50, to: 66.67, color: colors[2] },
    { from: 66.67, to: 83.33, color: colors[3] },
    { from: 83.33, to: 100, color: colors[4] },
  ];

  const chartData = history.map((p) => ({ hr: p.hr, idx: p.t }));

  const nameColor = theme === "dark" ? "#F1F5F9" : "#0F172A";
  const secondaryColor = theme === "dark" ? "#94A3B8" : "#64748B";

  return (
    <div
      className="border rounded-xl flex flex-col relative overflow-hidden min-h-0 p-2 sm:p-3"
      style={{ background: gradient, borderColor: theme === "dark" ? undefined : "rgba(0,0,0,0.08)" }}
    >
      <div className="text-center shrink-0 font-bold" style={{ fontSize: "clamp(3.2rem, 11.2vw, 14.4rem)", color: nameColor }}>
        {data.athlete_name || `Sensor #${data.device_id}`}
      </div>

      <div className="flex-1 flex items-center min-h-0">
        <div className="flex items-center justify-center h-full" style={{ width: "clamp(12rem, 25vw, 30rem)", flexShrink: 0 }}>
          <svg viewBox="0 0 200 115" className="w-full" preserveAspectRatio="xMidYMid meet">
            {zones.map((z, i) => (
              <path
                key={i}
                d={arc(cx, cy, r, z.from, z.to)}
                stroke={z.color}
                strokeWidth="14"
                fill="none"
                strokeLinecap="butt"
                opacity={zone === i + 1 ? 1 : 0.25}
              />
            ))}
            <line
              x1={cx}
              y1={cy}
              x2={polar(cx, cy, r - 4, needleDeg).x}
              y2={polar(cx, cy, r - 4, needleDeg).y}
              stroke={zoneColor}
              strokeWidth="4"
              strokeLinecap="round"
            />
            <circle cx={cx} cy={cy} r="6" fill={zoneColor} />
          </svg>
        </div>

        <div className="flex-1 flex items-center justify-center">
          <span
            className="font-black tabular-nums leading-none"
            style={{
              fontSize: "clamp(4rem, 14vw, 18rem)",
              color: zoneColor,
              textShadow: theme === "dark" ? "0 4px 12px rgba(0,0,0,0.5)" : "none",
            }}
          >
            {data.heart_rate}
          </span>
        </div>
      </div>

      <div className="shrink-0 flex items-center justify-between mb-1" style={{ fontSize: "clamp(0.6rem, 1vw, 1.2rem)" }}>
        <span style={{ color: zoneColor }} className="font-bold">
          {zoneName}
        </span>
        <span style={{ color: secondaryColor }}>
          {Math.round(data.zone_percent)}% · Max {maxHr}
        </span>
      </div>

      <div className="shrink-0" style={{ height: "clamp(3rem, 20%, 8rem)" }}>
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
          <div className="h-full flex items-center justify-center" style={{ color: secondaryColor, fontSize: "0.7rem" }}>
            Сбор данных...
          </div>
        )}
      </div>
    </div>
  );
}
