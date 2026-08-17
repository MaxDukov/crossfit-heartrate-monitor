import { useHrStore } from "../../lib/store";
import NewSensorAlert from "../NewSensorAlert";
import TriColumnCard from "./TriColumnCard";
import MainPanel from "./MainPanel";
import { computeMedals } from "./medals";
import type { HrUpdate } from "../../types";

export default function TriColumnMonitor() {
  const hrData = useHrStore((s) => s.hrData);
  const sensors = useHrStore((s) => s.sensors);

  const assignedDeviceIds = new Set(
    sensors.filter((s) => s.athlete_id).map((s) => s.device_id)
  );

  const entries = Object.values(hrData)
    .filter((d) => assignedDeviceIds.has(d.device_id))
    .sort((a, b) => (a.device_id - b.device_id));

  const medals = computeMedals(hrData);
  const half = Math.ceil(entries.length / 2);
  const left = entries.slice(0, half);
  const right = entries.slice(half);

  return (
    <>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "20fr 60fr 20fr",
          gap: 8,
          padding: 8,
          height: "100%",
        }}
      >
        <AthleteCol entries={left} medals={medals} stripeSide="right" />

        <MainPanel variant="tri" entries={entries} />

        <AthleteCol entries={right} medals={medals} stripeSide="left" />
      </div>
      <NewSensorAlert />
    </>
  );
}

function AthleteCol({
  entries,
  medals,
  stripeSide,
}: {
  entries: HrUpdate[];
  medals: Map<number, string>;
  stripeSide: "left" | "right";
}) {
  if (entries.length === 0) {
    return (
      <aside
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          color: "var(--mon-text-muted)",
          fontSize: "0.9rem",
        }}
      >
        Ожидание данных...
      </aside>
    );
  }
  return (
    <aside
      style={{
        display: "flex",
        flexDirection: "column",
        gap: 8,
        minHeight: 0,
      }}
    >
      {entries.map((data) => (
        <TriColumnCard
          key={data.device_id}
          data={data}
          medal={medals.get(data.device_id)}
          stripeSide={stripeSide}
        />
      ))}
    </aside>
  );
}
