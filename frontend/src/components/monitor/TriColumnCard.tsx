import type { HrUpdate } from "../../types";

const ZONE_COLORS: Record<number, string> = {
  1: "#15FD07",
  2: "#15FD07",
  3: "#FDE507",
  4: "#FD0707",
};

interface Props {
  data: HrUpdate;
  medal?: string;
  stripeSide?: "left" | "right";
}

export default function TriColumnCard({ data, medal, stripeSide = "left" }: Props) {
  const zc = ZONE_COLORS[data.zone] || "#94A3B8";

  return (
    <div
      style={{
        containerType: "inline-size",
        position: "relative",
        flex: "1 1 0",
        minHeight: 0,
        display: "grid",
        gridTemplateRows: "auto 1fr auto",
        alignItems: "center",
        gap: 8,
        padding: "14px 12px",
        background: `linear-gradient(135deg, color-mix(in srgb, ${zc} 15%, transparent), var(--mon-card-surface))`,
        border: "1px solid var(--mon-border)",
        borderLeft: stripeSide === "left" ? `9px solid ${zc}` : undefined,
        borderRight: stripeSide === "right" ? `9px solid ${zc}` : undefined,
        borderRadius: 12,
        overflow: "hidden",
      }}
    >
      {medal && (
        <span
          style={{
            position: "absolute",
            top: 4,
            left: 8,
            fontSize: "clamp(1.5rem, 2.2vw, 2.8rem)",
            lineHeight: 1,
            zIndex: 2,
            filter: "drop-shadow(0 2px 4px rgba(0,0,0,0.3))",
          }}
        >
          {medal}
        </span>
      )}

      <div
        style={{
          fontWeight: 700,
          fontSize: "clamp(2rem, 2.5vw, 3.2rem)",
          textAlign: "center",
          lineHeight: 1,
          overflow: "hidden",
          letterSpacing: "-0.02em",
          color: "var(--mon-text)",
        }}
      >
        {data.athlete_name || `Sensor #${data.device_id}`}
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "baseline",
          justifyContent: "center",
          gap: 6,
        }}
      >
        <span
          style={{
            fontVariantNumeric: "tabular-nums",
            fontWeight: 800,
            color: zc,
            fontSize: "clamp(2.5rem, 4vw, 5rem)",
            lineHeight: 1,
            letterSpacing: "-0.03em",
          }}
        >
          {data.heart_rate}
        </span>
        <span
          style={{
            color: zc,
            fontSize: "clamp(1.2rem, 1.8vw, 2.2rem)",
          }}
        >
          ♥
        </span>
      </div>

      <div
        style={{
          fontVariantNumeric: "tabular-nums",
          fontWeight: 700,
          color: "var(--mon-text-dim)",
          fontSize: "clamp(1.8rem, 2vw, 2.6rem)",
          textAlign: "center",
          lineHeight: 1,
        }}
      >
        <span style={{ color: "#F97316", marginRight: 4 }}>
          <svg width="1em" height="1em" viewBox="0 0 24 24" fill="currentColor" style={{ display: "inline-block", verticalAlign: "middle" }}>
            <path d="M13.5.67s.74 2.65.74 4.8c0 2.06-1.35 3.73-3.41 3.73-2.07 0-3.63-1.67-3.63-3.73l.03-.36C5.21 7.51 4 10.62 4 14c0 4.42 3.58 8 8 8s8-3.58 8-8C20 8.61 17.41 3.8 13.5.67zM11.71 19c-1.78 0-3.22-1.4-3.22-3.14 0-1.62 1.05-2.76 2.81-3.12 1.77-.36 3.6-1.21 4.62-2.58.39 1.29.59 2.65.59 4.04 0 2.65-2.15 4.8-4.8 4.8z"/>
          </svg>
        </span>
        {Math.round(data.calories)}
      </div>
    </div>
  );
}
