import type { HrUpdate } from "../../types";
import type { HrPoint } from "../../lib/store";

const ZONE_COLORS: Record<number, string> = {
  1: "#15FD07",
  2: "#15FD07",
  3: "#FDE507",
  4: "#FD0707",
};

const ZONE_FILL: Record<number, string> = {
  1: "rgba(21,253,7,0.12)",
  2: "rgba(21,253,7,0.12)",
  3: "rgba(253,229,7,0.12)",
  4: "rgba(253,7,7,0.12)",
};

interface Props {
  data: HrUpdate;
  history: HrPoint[];
  medal?: string;
}

export default function SplitScreenCard({ data, history, medal }: Props) {
  const zc = ZONE_COLORS[data.zone] || "#94A3B8";
  const zf = ZONE_FILL[data.zone] || "rgba(148,163,184,0.12)";

  const spark = buildSparkline(history, zc, zf);

  return (
    <div
      style={{
        containerType: "inline-size",
        position: "relative",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        textAlign: "center",
        gap: 2,
        padding: "8px 12px",
        height: "100%",
        background: `linear-gradient(135deg, color-mix(in srgb, ${zc} 15%, transparent), var(--mon-card-surface))`,
        border: "1px solid var(--mon-border)",
        borderLeft: `9px solid ${zc}`,
        borderRadius: 12,
        minWidth: 0,
        overflow: "hidden",
      }}
    >
      {medal && (
        <span
          style={{
            position: "absolute",
            top: 4,
            left: 8,
            fontSize: "clamp(1.2rem, 1.6vw, 2rem)",
            lineHeight: 1,
            zIndex: 2,
            filter: "drop-shadow(0 2px 4px rgba(0,0,0,0.2))",
          }}
        >
          {medal}
        </span>
      )}

      <div
        style={{
          fontWeight: 700,
          fontSize: "clamp(2.3rem, 3.06vw, 3.06rem)",
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
          gap: 8,
        }}
      >
        <span
          style={{
            fontVariantNumeric: "tabular-nums",
            fontWeight: 800,
            color: zc,
            fontSize: "clamp(2.8rem, 4vw, 4.4rem)",
            lineHeight: 1,
            letterSpacing: "-0.02em",
          }}
        >
          {data.heart_rate}
        </span>
        <span style={{ color: zc, fontSize: "clamp(1.2rem, 1.6vw, 2rem)" }}>
          ♥
        </span>
        <span
          style={{
            fontVariantNumeric: "tabular-nums",
            fontWeight: 700,
            color: "var(--mon-text-dim)",
            fontSize: "clamp(1.6rem, 2vw, 2.4rem)",
            whiteSpace: "nowrap",
            lineHeight: 1,
          }}
        >
          <span style={{ color: "#F97316", marginRight: 4, display: "inline-flex", verticalAlign: "middle" }}>
            <svg width="1em" height="1em" viewBox="0 0 24 24" fill="currentColor">
              <path d="M13.5.67s.74 2.65.74 4.8c0 2.06-1.35 3.73-3.41 3.73-2.07 0-3.63-1.67-3.63-3.73l.03-.36C5.21 7.51 4 10.62 4 14c0 4.42 3.58 8 8 8s8-3.58 8-8C20 8.61 17.41 3.8 13.5.67zM11.71 19c-1.78 0-3.22-1.4-3.22-3.14 0-1.62 1.05-2.76 2.81-3.12 1.77-.36 3.6-1.21 4.62-2.58.39 1.29.59 2.65.59 4.04 0 2.65-2.15 4.8-4.8 4.8z"/>
            </svg>
          </span>
          {Math.round(data.calories)}
        </span>
      </div>

      {spark}
    </div>
  );
}

function buildSparkline(history: HrPoint[], stroke: string, fill: string) {
  if (history.length < 2) return null;

  const W = 120;
  const H = 20;
  const hrs = history.map((p) => p.hr);
  const min = Math.min(...hrs);
  const max = Math.max(...hrs);
  const range = max - min || 1;

  const pts = history.map((p, i) => {
    const x = (i / (history.length - 1)) * W;
    const y = H - ((p.hr - min) / range) * (H - 2) - 1;
    return `${x.toFixed(1)},${y.toFixed(1)}`;
  });

  const fillPts = `0,${H} ${pts.join(" ")} ${W},${H}`;

  return (
    <svg
      width="100%"
      height={20}
      viewBox={`0 0 ${W} ${H}`}
      preserveAspectRatio="none"
      style={{ display: "block", marginTop: 2 }}
    >
      <polyline points={fillPts} fill={fill} stroke="none" />
      <polyline
        points={pts.join(" ")}
        fill="none"
        stroke={stroke}
        strokeWidth={1.5}
        strokeLinejoin="round"
        strokeLinecap="round"
      />
    </svg>
  );
}
