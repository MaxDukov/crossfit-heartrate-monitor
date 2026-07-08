import type { HrUpdate } from "../types";
import { ZONE_GRADIENT, ZONE_NAMES, ZONE_COLORS } from "../lib/zones";
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
  const zone = data.zone;
  const gradient = ZONE_GRADIENT[zone] || ZONE_GRADIENT[1];
  const zoneName = ZONE_NAMES[zone] || "—";
  const zoneColor = ZONE_COLORS[zone] || "#94A3B8";
  const maxHr = data.max_hr || 190;
  const pct = Math.min(data.zone_percent, 120);
  const needleDeg = (pct / 120) * 180;

  const cx = 100;
  const cy = 100;
  const r = 88;

  const zones = [
    { from: 0, to: 50, color: ZONE_COLORS[1] },
    { from: 50, to: 66.67, color: ZONE_COLORS[2] },
    { from: 66.67, to: 83.33, color: ZONE_COLORS[3] },
    { from: 83.33, to: 100, color: ZONE_COLORS[4] },
  ];

  const chartData = history.map((p) => ({ hr: p.hr, idx: p.t }));

  return (
    <div
      className={`bg-gradient-to-br ${gradient} border rounded-xl flex flex-col relative overflow-hidden min-h-0 p-2 sm:p-3`}
    >
      <div className="text-center shrink-0 font-bold text-slate-100" style={{ fontSize: "clamp(3.2rem, 11.2vw, 14.4rem)" }}>
        {data.athlete_name || `Sensor #${data.device_id}`}
      </div>

      <div className="flex-1 flex items-center justify-center min-h-0 relative">
        <span
          className="font-black tabular-nums leading-none"
          style={{
            fontSize: "clamp(4rem, 14vw, 18rem)",
            color: zoneColor,
            textShadow: "0 4px 12px rgba(0,0,0,0.5)",
          }}
        >
          {data.heart_rate}
        </span>

        <div style={{ width: "15%", minWidth: "3rem", maxWidth: "5rem" }} className="absolute bottom-0 left-3">
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
              strokeWidth="3"
              strokeLinecap="round"
            />
            <circle cx={cx} cy={cy} r="5" fill={zoneColor} />
          </svg>
        </div>
      </div>

      <div className="shrink-0 flex items-center justify-between mb-1" style={{ fontSize: "clamp(0.6rem, 1vw, 1.2rem)" }}>
        <span style={{ color: zoneColor }} className="font-bold">
          {zoneName}
        </span>
        <span className="text-slate-400">
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
          <div className="h-full flex items-center justify-center text-slate-500" style={{ fontSize: "0.7rem" }}>
            Сбор данных...
          </div>
        )}
      </div>
    </div>
  );
}
