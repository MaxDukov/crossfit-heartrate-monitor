import { useHrStore } from "../lib/store";
import AthleteCard from "../components/AthleteCard";
import NewSensorAlert from "../components/NewSensorAlert";
import WodPanel from "../components/WodPanel";

export default function Dashboard() {
  const hrData = useHrStore((s) => s.hrData);
  const hrHistory = useHrStore((s) => s.hrHistory);
  const sensors = useHrStore((s) => s.sensors);

  const assignedDeviceIds = new Set(
    sensors.filter((s) => s.athlete_id).map((s) => s.device_id)
  );

  const entries = Object.values(hrData).filter(
    (d) => assignedDeviceIds.has(d.device_id)
  );

  const n = entries.length;

  const gridClass =
    n <= 1
      ? "grid-cols-1 grid-rows-1"
      : n === 2
      ? "grid-cols-2 grid-rows-1"
      : n === 3
      ? "grid-cols-3 grid-rows-1"
      : n === 4
      ? "grid-cols-2 grid-rows-2"
      : n <= 6
      ? "grid-cols-3 grid-rows-2"
      : "grid-cols-4 grid-rows-2";

  return (
    <div className="h-full flex flex-col">
      <WodPanel />
      <div className={`flex-1 grid ${gridClass} gap-2 sm:gap-3 p-2 sm:p-4`}>
        {entries.map((data) => (
          <AthleteCard
            key={data.device_id}
            data={data}
            history={hrHistory[data.device_id] || []}
          />
        ))}
        {entries.length === 0 && (
          <div className="col-span-full flex items-center justify-center text-slate-400 dark:text-slate-500 text-lg">
            Ожидание данных с датчиков...
          </div>
        )}
      </div>
      <NewSensorAlert />
    </div>
  );
}
