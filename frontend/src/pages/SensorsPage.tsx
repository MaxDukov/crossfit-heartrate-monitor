import { useEffect, useState, useCallback } from "react";
import type { Sensor, Athlete } from "../types";
import { api } from "../lib/api";

export default function SensorsPage() {
  const [sensors, setSensors] = useState<Sensor[]>([]);
  const [athletes, setAthletes] = useState<Athlete[]>([]);
  const [assignMap, setAssignMap] = useState<Record<number, string>>({});

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

  const activeSensors = sensors.filter((s) => !s.ignored);
  const ignoredSensors = sensors.filter((s) => s.ignored);

  const renderSensor = (s: Sensor) => (
    <div
      key={s.device_id}
      className={`border rounded-lg p-4 flex items-center justify-between ${
        s.ignored
          ? "bg-slate-100 dark:bg-slate-900/40 border-slate-200 dark:border-slate-800 opacity-60"
          : s.athlete_id
          ? "bg-white dark:bg-slate-800/50 border-slate-200 dark:border-slate-700"
          : "bg-amber-50 dark:bg-amber-950/30 border-amber-200 dark:border-amber-500/30"
      }`}
    >
      <div className="min-w-0">
        <div className="font-medium text-slate-900 dark:text-slate-100 flex items-center gap-2">
          ID: {s.device_id}
          {s.ignored && (
            <span className="text-xs bg-slate-400 text-white px-1.5 py-0.5 rounded">
              Чужой
            </span>
          )}
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
        {s.ignored ? (
          <button
            className="text-sm bg-emerald-600 hover:bg-emerald-500 px-3 py-1.5 rounded text-white"
            onClick={() => handleUnignore(s.device_id)}
          >
            Вернуть
          </button>
        ) : s.athlete_id ? (
          <>
            <button
              className="text-sm bg-slate-200 dark:bg-slate-700 hover:bg-slate-300 dark:hover:bg-slate-600 px-3 py-1.5 rounded text-slate-700 dark:text-slate-200"
              onClick={() => handleUnassign(s.device_id)}
            >
              Отвязать
            </button>
            <button
              className="text-sm text-slate-400 hover:text-red-500 px-2 py-1.5"
              title="Пометить как чужой"
              onClick={() => handleIgnore(s.device_id)}
            >
              🚫
            </button>
          </>
        ) : (
          <>
            <select
              className="bg-white dark:bg-slate-700 border border-slate-300 dark:border-slate-600 rounded px-2 py-1.5 text-sm text-slate-900 dark:text-white"
              value={assignMap[s.device_id] || ""}
              onChange={(e) =>
                setAssignMap((m) => ({
                  ...m,
                  [s.device_id]: e.target.value,
                }))
              }
            >
              <option value="">Привязать к...</option>
              {unassignedAthletes(s.device_id).map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name}
                </option>
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
              className="text-sm text-slate-400 hover:text-red-500 px-2 py-1.5"
              title="Пометить как чужой"
              onClick={() => handleIgnore(s.device_id)}
            >
              🚫
            </button>
          </>
        )}
      </div>
    </div>
  );

  return (
    <div className="max-w-3xl mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6 text-slate-900 dark:text-slate-100">Датчики</h1>

      <div className="space-y-3">
        {activeSensors.map(renderSensor)}

        {ignoredSensors.length > 0 && (
          <>
            <div className="text-sm text-slate-400 dark:text-slate-500 pt-4 pb-1 border-t border-slate-200 dark:border-slate-700">
              Проигнорированные ({ignoredSensors.length})
            </div>
            {ignoredSensors.map(renderSensor)}
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
