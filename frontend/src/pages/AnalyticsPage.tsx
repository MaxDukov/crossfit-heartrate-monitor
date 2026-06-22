import { useEffect, useState } from "react";
import { useParams, Link } from "react-router-dom";
import type { AthleteStats, SessionStats } from "../types";
import { api } from "../lib/api";
import { ZONE_COLORS, ZONE_NAMES } from "../lib/zones";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from "recharts";

export default function AnalyticsPage() {
  const { id } = useParams<{ id: string }>();
  const [stats, setStats] = useState<AthleteStats | null>(null);
  const [history, setHistory] = useState<SessionStats[]>([]);
  const [athleteName, setAthleteName] = useState("");

  useEffect(() => {
    if (!id) return;
    Promise.all([
      api.analytics.athleteStats(id),
      api.analytics.athleteHistory(id),
      api.athletes.list(),
    ]).then(([s, h, athletes]) => {
      setStats(s);
      setHistory(h);
      const a = athletes.find((a) => a.id === id);
      if (a) setAthleteName(a.name);
    });
  }, [id]);

  if (!stats) return <div className="p-6 text-slate-500">Загрузка...</div>;

  const pieData = history.reduce(
    (acc, s) => {
      acc[0].value += s.zones.zone_1_seconds;
      acc[1].value += s.zones.zone_2_seconds;
      acc[2].value += s.zones.zone_3_seconds;
      acc[3].value += s.zones.zone_4_seconds;
      return acc;
    },
    [
      { name: ZONE_NAMES[1], value: 0 },
      { name: ZONE_NAMES[2], value: 0 },
      { name: ZONE_NAMES[3], value: 0 },
      { name: ZONE_NAMES[4], value: 0 },
    ]
  );

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="flex items-center gap-4 mb-6">
        <Link to="/athletes" className="text-slate-400 hover:text-white text-sm">
          ← Спортсмены
        </Link>
        <h1 className="text-2xl font-bold">{athleteName}</h1>
      </div>

      <div className="grid grid-cols-4 gap-4 mb-8">
        {[
          { label: "Тренировок", value: stats.total_sessions },
          {
            label: "Общее время",
            value: `${Math.floor(stats.total_duration_seconds / 3600)}ч ${Math.floor(
              (stats.total_duration_seconds % 3600) / 60
            )}м`,
          },
          { label: "Средний пульс", value: `${stats.avg_hr} bpm` },
          { label: "Макс. пульс", value: `${stats.max_hr_ever} bpm` },
        ].map((c) => (
          <div
            key={c.label}
            className="bg-slate-800/50 border border-slate-700 rounded-lg p-4"
          >
            <div className="text-sm text-slate-400">{c.label}</div>
            <div className="text-2xl font-bold mt-1">{c.value}</div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-2 gap-6 mb-8">
        <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-4">
          <h2 className="text-lg font-semibold mb-4">Распределение по зонам</h2>
          <ResponsiveContainer width="100%" height={200}>
            <PieChart>
              <Pie data={pieData} dataKey="value" innerRadius={50} outerRadius={80}>
                {pieData.map((_, i) => (
                  <Cell key={i} fill={ZONE_COLORS[i + 1]} />
                ))}
              </Pie>
              <Tooltip />
            </PieChart>
          </ResponsiveContainer>
          <div className="flex justify-center gap-4 mt-2">
            {pieData.map((d, i) => (
              <div key={i} className="flex items-center gap-1 text-xs">
                <span
                  className="w-3 h-3 rounded-full"
                  style={{ backgroundColor: ZONE_COLORS[i + 1] }}
                />
                <span className="text-slate-400">{d.name}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-slate-800/50 border border-slate-700 rounded-lg p-4">
          <h2 className="text-lg font-semibold mb-4">Пульс по тренировкам</h2>
          <ResponsiveContainer width="100%" height={200}>
            <LineChart
              data={history.slice().reverse().map((s, i) => ({
                name: s.session_name || `#${i + 1}`,
                avg: Math.round(s.avg_hr),
                max: s.max_hr,
              }))}
            >
              <XAxis dataKey="name" tick={{ fill: "#94A3B8", fontSize: 12 }} />
              <YAxis tick={{ fill: "#94A3B8", fontSize: 12 }} />
              <Tooltip />
              <Line type="monotone" dataKey="avg" stroke="#3B82F6" strokeWidth={2} name="Средний" />
              <Line type="monotone" dataKey="max" stroke="#EF4444" strokeWidth={2} name="Макс" />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>

      <h2 className="text-lg font-semibold mb-3">История тренировок</h2>
      <div className="space-y-2">
        {history.map((s) => (
          <div
            key={s.session_id}
            className="bg-slate-800/50 border border-slate-700 rounded-lg p-3"
          >
            <div className="flex justify-between">
              <span className="font-medium">
                {s.session_name || "Тренировка"}
              </span>
              <span className="text-sm text-slate-400">
                {Math.floor(s.duration_seconds / 60)} мин · Средн: {Math.round(s.avg_hr)} · Макс:{" "}
                {s.max_hr} · Мин: {s.min_hr}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
