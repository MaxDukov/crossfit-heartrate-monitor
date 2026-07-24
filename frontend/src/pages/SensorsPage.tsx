import { useEffect, useState, useCallback } from "react";
import type { Sensor, Athlete } from "../types";
import { api } from "../lib/api";

type Tab = "detected" | "assigned";

export default function SensorsPage() {
  const [sensors, setSensors] = useState<Sensor[]>([]);
  const [athletes, setAthletes] = useState<Athlete[]>([]);
  const [assignMap, setAssignMap] = useState<Record<number, string>>({});
  const [tab, setTab] = useState<Tab>("detected");

  const load = useCallback(async () => {
    const [s, a] = await Promise.all([api.sensors.list(), api.athletes.list()]);
    setSensors(s);
    setAthletes(a);
  }, []);

  useEffect(() => { load(); }, [load]);

  const handleAssign = async (deviceId: number) => {
    const athleteId = assignMap[deviceId];
    if (!athleteId) return;
    await api.sensors.assign(deviceId, athleteId);
    setAssignMap((m) => ({ ...m, [deviceId]: "" }));
    load();
  };

  const handleUnassign = async (deviceId: number) => {
    await api.sensors.unassign(deviceId);
    load();
  };

  const handleIgnore = async (deviceId: number) => {
    await api.sensors.ignore(deviceId);
    load();
  };

  const handleUnignore = async (deviceId: number) => {
    await api.sensors.unignore(deviceId);
    load();
  };

  const unassignedAthletes = (currentDeviceId: number) =>
    athletes.filter(
      (a) =>
        !sensors.some(
          (s) => s.athlete_id === a.id && s.device_id !== currentDeviceId
        )
    );

  const detected = sensors.filter((s) => !s.ignored && !s.athlete_id);
  const assigned = sensors.filter((s) => !s.ignored && s.athlete_id);
  const ignored = sensors.filter((s) => s.ignored);

  const renderDetected = (s: Sensor) => (
    <div
      key={s.device_id}
      className="border rounded-lg p-4 flex items-center justify-between bg-amber-50 dark:bg-amber-950/30 border-amber-200 dark:border-amber-500/30"
    >
      <div className="min-w-0">
        <div className="font-medium text-slate-900 dark:text-slate-100">
          ID: {s.device_id}
        </div>
        <div className="text-sm text-slate-500 dark:text-slate-400">
          {s.last_hr != null ? `${s.last_hr} bpm` : "нет данных"}
          {s.battery_level != null && ` · Батарея: ${s.battery_level}%`}
        </div>
      </div>
      <div className="flex items-center gap-2 shrink-0">
        <select
          className="bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded px-2 py-1.5 text-sm text-slate-900 dark:text-white"
          value={assignMap[s.device_id] || ""}
          onChange={(e) =>
            setAssignMap((m) => ({ ...m, [s.device_id]: e.target.value }))
          }
        >
          <option value="">Привязать к...</option>
          {unassignedAthletes(s.device_id).map((a) => (
            <option key={a.id} value={a.id}>{a.name}</option>
          ))}
        </select>
        <button
          className="bg-emerald-600 hover:bg-emerald-500 px-3 py-1.5 rounded text-sm disabled:opacity-40"
          disabled={!assignMap[s.device_id]}
          onClick={() => handleAssign(s.device_id)}
        >
          OK
        </button>
        <button
          className="text-sm bg-slate-200 dark:bg-slate-700 hover:bg-red-500 hover:text-white text-slate-600 dark:text-slate-300 px-3 py-1.5 rounded"
          onClick={() => handleIgnore(s.device_id)}
        >
          Отключить
        </button>
      </div>
    </div>
  );

  const renderAssigned = (s: Sensor) => (
    <div
      key={s.device_id}
      className="border rounded-lg p-4 flex items-center justify-between bg-white dark:bg-slate-800/50 border-slate-200 dark:border-slate-700"
    >
      <div className="min-w-0">
        <div className="font-medium text-slate-900 dark:text-slate-100">
          ID: {s.device_id}
        </div>
        <div className="text-sm text-slate-500 dark:text-slate-400">
          {s.last_hr != null ? `${s.last_hr} bpm` : "нет данных"}
          {s.battery_level != null && ` · Батарея: ${s.battery_level}%`}
          {s.athlete_name && (
            <span className="text-emerald-600 dark:text-emerald-400 ml-2">
              → {s.athlete_name}
            </span>
          )}
        </div>
      </div>
      <div className="flex items-center gap-2 shrink-0">
        <button
          className="text-sm bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 px-3 py-1.5 rounded text-slate-700 dark:text-slate-200"
          onClick={() => handleUnassign(s.device_id)}
        >
          Отвязать
        </button>
        <button
          className="text-sm bg-slate-200 dark:bg-slate-700 hover:bg-red-500 hover:text-white text-slate-600 dark:text-slate-300 px-3 py-1.5 rounded"
          onClick={() => handleIgnore(s.device_id)}
        >
          Отключить
        </button>
      </div>
    </div>
  );

  const renderIgnored = (s: Sensor) => (
    <div
      key={s.device_id}
      className="border rounded-lg p-4 flex items-center justify-between bg-slate-100 dark:bg-slate-900/40 border-slate-200 dark:border-slate-800 opacity-60"
    >
      <div className="min-w-0">
        <div className="font-medium text-slate-900 dark:text-slate-100 flex items-center gap-2">
          ID: {s.device_id}
          <span className="text-xs bg-slate-400 text-white px-1.5 py-0.5 rounded">Чужой</span>
        </div>
        <div className="text-sm text-slate-500 dark:text-slate-400">
          {s.last_hr != null ? `${s.last_hr} bpm` : "нет данных"}
        </div>
      </div>
      <button
        className="text-sm bg-emerald-600 hover:bg-emerald-500 px-3 py-1.5 rounded text-white"
        onClick={() => handleUnignore(s.device_id)}
      >
        Вернуть
      </button>
    </div>
  );

  const tabBtn = (id: Tab, label: string, count: number) => (
    <button
      className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
        tab === id
          ? "bg-blue-600 text-white"
          : "bg-slate-200 dark:bg-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-300 dark:hover:bg-slate-600"
      }`}
      onClick={() => setTab(id)}
    >
      {label} ({count})
    </button>
  );

  return (
    <div className="max-w-3xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-4 text-slate-900 dark:text-slate-100">Датчики</h1>

      <div className="flex gap-2 mb-6">
        {tabBtn("detected", "Обнаруженные", detected.length)}
        {tabBtn("assigned", "Привязанные", assigned.length)}
      </div>

      <div className="space-y-3">
        {tab === "detected" && (
          <>
            {detected.map(renderDetected)}
            {detected.length === 0 && (
              <div className="text-slate-400 dark:text-slate-500 text-center py-12">
                Нет свободных датчиков.
              </div>
            )}

            {ignored.length > 0 && (
              <>
                <div className="text-sm text-slate-400 dark:text-slate-500 pt-4 pb-1 border-t border-slate-200 dark:border-slate-700">
                  Отключённые ({ignored.length})
                </div>
                {ignored.map(renderIgnored)}
              </>
            )}
          </>
        )}

        {tab === "assigned" && (
          <>
            {assigned.map(renderAssigned)}
            {assigned.length === 0 && (
              <div className="text-slate-400 dark:text-slate-500 text-center py-12">
                Нет привязанных датчиков. Перейдите на вкладку «Обнаруженные».
              </div>
            )}
          </>
        )}

        {sensors.length === 0 && (
          <div className="text-slate-400 dark:text-slate-500 text-center py-12">
            Датчики не обнаружены. Подключите ANT+ стик и включите датчики.
          </div>
        )}
      </div>
    </div>
  );
}
