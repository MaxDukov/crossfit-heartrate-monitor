import { useHrStore } from "../lib/store";
import AthleteCard from "../components/AthleteCard";
import NewSensorAlert from "../components/NewSensorAlert";
import WodPanel from "../components/WodPanel";
import { Fragment } from "react";

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

  const gridClass =
    entries.length <= 1
      ? "grid-cols-1 grid-rows-1"
      : entries.length === 2
      ? "grid-cols-2 grid-rows-1"
      : entries.length === 3
      ? "grid-cols-3 grid-rows-1"
      : entries.length === 4
      ? "grid-cols-2 grid-rows-2"
      : "grid-cols-3 grid-rows-3";

  return (
    <div className="h-full flex flex-col">
      <WodPanel />
      <div className={`flex-1 grid ${gridClass} gap-3 p-4`}>
        {entries.map((data, i) => (
          <Fragment key={data.device_id}>
            {entries.length > 4 && i === 4 && <div />}
            <AthleteCard
              data={data}
              history={hrHistory[data.device_id] || []}
            />
          </Fragment>
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
