import { useHrStore } from "../lib/store";
import { api } from "../lib/api";
import { useState } from "react";

export default function NewSensorAlert() {
  const { newSensors, dismissNewSensor, fetchSensors } = useHrStore();
  const [assignMap, setAssignMap] = useState<Record<number, string>>({});
  const [athletesList, setAthletesList] = useState<
    { id: string; name: string }[]
  >([]);

  if (newSensors.length === 0) return null;

  const loadAthletes = async () => {
    const list = await api.athletes.list();
    setAthletesList(list);
  };

  const handleAssign = async (deviceId: number) => {
    const athleteId = assignMap[deviceId];
    if (!athleteId) return;
    await api.sensors.assign(deviceId, athleteId);
    dismissNewSensor(deviceId);
    fetchSensors();
  };

  return (
    <div className="fixed bottom-4 right-4 flex flex-col gap-2 z-50 max-w-sm">
      {newSensors.map((deviceId) => (
        <div
          key={deviceId}
          className="bg-white dark:bg-slate-800 border border-amber-500/40 rounded-lg p-4 shadow-xl"
        >
          <div className="flex items-center gap-2 mb-2">
            <span className="w-2 h-2 bg-amber-400 rounded-full animate-pulse" />
            <span className="text-sm font-medium text-slate-900 dark:text-slate-100">
              Новый датчик: ID {deviceId}
            </span>
          </div>
          <div className="flex gap-2">
            <select
              className="flex-1 bg-slate-100 dark:bg-slate-700 text-slate-900 dark:text-slate-100 text-sm rounded px-2 py-1 border border-slate-300 dark:border-slate-600"
              value={assignMap[deviceId] || ""}
              onClick={loadAthletes}
              onChange={(e) =>
                setAssignMap((m) => ({ ...m, [deviceId]: e.target.value }))
              }
            >
              <option value="">Привязать к...</option>
              {athletesList.map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name}
                </option>
              ))}
            </select>
            <button
              className="bg-amber-600 hover:bg-amber-500 text-sm px-3 py-1 rounded font-medium disabled:opacity-40"
              disabled={!assignMap[deviceId]}
              onClick={() => handleAssign(deviceId)}
            >
              OK
            </button>
            <button
              className="text-slate-400 hover:text-slate-900 dark:hover:text-white text-sm px-2"
              onClick={() => dismissNewSensor(deviceId)}
            >
              ✕
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}
