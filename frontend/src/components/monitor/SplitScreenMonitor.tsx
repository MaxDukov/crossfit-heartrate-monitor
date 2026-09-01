import { useHrStore } from "../../lib/store";
import { useActiveWod } from "../../lib/useActiveWod";
import NewSensorAlert from "../NewSensorAlert";
import SplitScreenCard from "./SplitScreenCard";
import MainPanel from "./MainPanel";
import { computeMedals } from "./medals";

export default function SplitScreenMonitor() {
  const hrData = useHrStore((s) => s.hrData);
  const hrHistory = useHrStore((s) => s.hrHistory);
  const sensors = useHrStore((s) => s.sensors);
  const wod = useActiveWod();

  const assignedDeviceIds = new Set(
    sensors.filter((s) => s.athlete_id).map((s) => s.device_id)
  );

  const entries = Object.values(hrData)
    .filter((d) => assignedDeviceIds.has(d.device_id))
    .sort((a, b) => (a.device_id - b.device_id));

  const medals = computeMedals(hrData);
  const half = Math.ceil(entries.length / 2);
  const sidebar = entries.slice(0, half);
  const bottom = entries.slice(half);

  return (
    <>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "3fr 1fr",
          gridTemplateRows: "4fr 1fr",
          gridTemplateAreas: '"main right" "bottom bottom"',
          gap: 6,
          padding: 6,
          height: "100%",
        }}
      >
        <div style={{ gridArea: "main", minHeight: 0 }}>
          <MainPanel variant="split" entries={entries} wod={wod} />
        </div>

        <aside
          style={{
            gridArea: "right",
            display: "flex",
            flexDirection: "column",
            gap: 6,
            minHeight: 0,
          }}
        >
          {sidebar.map((data) => (
            <div key={data.device_id} style={{ flex: "1 1 0", minHeight: 0 }}>
              <SplitScreenCard
                data={data}
                history={hrHistory[data.device_id] || []}
                medal={medals.get(data.device_id)}
              />
            </div>
          ))}
        </aside>

        <section
          style={{
            gridArea: "bottom",
            display: "grid",
            gridTemplateColumns: `repeat(${Math.max(bottom.length, 1)}, 1fr)`,
            gap: 6,
          }}
        >
          {bottom.map((data) => (
            <SplitScreenCard
              key={data.device_id}
              data={data}
              history={hrHistory[data.device_id] || []}
              medal={medals.get(data.device_id)}
            />
          ))}
        </section>
      </div>
      <NewSensorAlert />
    </>
  );
}
