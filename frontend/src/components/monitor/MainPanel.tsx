import type { HrUpdate } from "../../types";

interface Props {
  variant: "tri" | "split";
  entries: HrUpdate[];
}

export default function MainPanel({ variant, entries }: Props) {
  const avgHr =
    entries.length > 0
      ? Math.round(entries.reduce((s, e) => s + e.heart_rate, 0) / entries.length)
      : 0;
  const totalKcal = entries.reduce((s, e) => s + Math.round(e.calories), 0);

  const triFont = variant === "tri";

  return (
    <div
      style={{
        background: "var(--mon-center-bg)",
        border: "1px solid var(--mon-border)",
        borderRadius: 16,
        padding: "24px 5%",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "space-evenly",
        textAlign: "center",
        height: "100%",
        minHeight: 0,
        boxShadow: variant === "split" ? "var(--mon-shadow)" : "none",
      }}
    >
      <h1
        style={{
          display: "flex",
          alignItems: "baseline",
          justifyContent: "center",
          gap: 16,
          fontSize: triFont
            ? "clamp(1.92rem, 3.83vw, 4.6rem)"
            : "clamp(1.53rem, 3.07vw, 3.83rem)",
          fontWeight: 900,
          letterSpacing: "-0.03em",
          margin: 0,
          lineHeight: 1,
          textTransform: "uppercase",
          width: "90%",
          color: "var(--mon-text)",
        }}
      >
        <span style={{ color: "var(--mon-text-muted)" }}>WOD:</span> AMRAP 20:00
      </h1>

      <span
        style={{
          fontSize: triFont
            ? "clamp(0.86rem, 1.15vw, 1.27rem)"
            : "0.86rem",
          fontWeight: 700,
          letterSpacing: "0.2em",
          textTransform: "uppercase",
          color: "#16A34A",
          background: "rgba(22,163,74,0.12)",
          border: "1px solid rgba(22,163,74,0.3)",
          padding: "6px 18px",
          borderRadius: 999,
        }}
      >
        Work Phase
      </span>

      <div
        style={{
          fontVariantNumeric: "tabular-nums",
          fontSize: triFont
            ? "clamp(3.91rem, 10.75vw, 13.69rem)"
            : "clamp(3.45rem, 9.2vw, 11.5rem)",
          fontWeight: 800,
          color: "#16A34A",
          lineHeight: 0.9,
          letterSpacing: "-0.04em",
          textShadow: "0 0 60px rgba(22,163,74,0.2)",
          width: "90%",
        }}
      >
        14:32
      </div>

      <table
        style={{
          width: triFont ? "80%" : "95%",
          maxWidth: triFont ? 580 : 1040,
          borderCollapse: "collapse",
          fontSize: triFont
            ? "clamp(1.4rem, 2.28vw, 2.53rem)"
            : "clamp(1.27rem, 2.07vw, 2.3rem)",
          fontWeight: 600,
          color: "var(--mon-text)",
        }}
      >
        <thead>
          <tr>
            <th style={thStyle(triFont, 1)}>Раунды</th>
            <th style={thStyle(triFont, 2)}>Повторы</th>
            <th style={thStyle(triFont, 3)}>Упражнение</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td style={tdRounds} rowSpan={3}>5×</td>
            <td style={tdReps}>10</td>
            <td style={tdExercise}>Thrusters (40/25 кг)</td>
          </tr>
          <tr>
            <td style={tdReps}>15</td>
            <td style={tdExercise}>Pull-ups</td>
          </tr>
          <tr>
            <td style={tdReps}>20</td>
            <td style={tdExercise}>Box Jumps (60/50 см)</td>
          </tr>
        </tbody>
      </table>

      <div
        style={{
          display: "flex",
          gap: 32,
          alignItems: "center",
          fontVariantNumeric: "tabular-nums",
          fontSize: "clamp(1.09rem, 1.5vw, 1.61rem)",
          color: "var(--mon-text-muted)",
          fontWeight: 600,
        }}
      >
        <span>
          <b style={{ color: "var(--mon-text)", fontWeight: 800, marginRight: 6 }}>
            Средний HR:
          </b>
          {avgHr}
        </span>
        <span style={{ color: "var(--mon-border)" }}>|</span>
        <span>
          <b style={{ color: "var(--mon-text)", fontWeight: 800, marginRight: 6 }}>
            Всего ккал:
          </b>
          {totalKcal}
        </span>
        <span style={{ color: "var(--mon-border)" }}>|</span>
        <span>
          <b style={{ color: "var(--mon-text)", fontWeight: 800, marginRight: 6 }}>
            Активны:
          </b>
          {entries.length}
        </span>
      </div>
    </div>
  );
}

function thStyle(tri: boolean, col: number): React.CSSProperties {
  return {
    textAlign: "center",
    width: col === 1 ? "18%" : col === 2 ? "18%" : undefined,
    fontSize: tri ? "clamp(0.8rem, 0.98vw, 1.04rem)" : "clamp(0.8rem, 0.98vw, 1.04rem)",
    fontWeight: 700,
    letterSpacing: "0.18em",
    textTransform: "uppercase",
    color: "var(--mon-text-muted)",
    padding: "8px 16px",
    borderBottom: "2px solid var(--mon-border)",
  };
}

const tdRounds: React.CSSProperties = {
  fontWeight: 900,
  fontVariantNumeric: "tabular-nums",
  color: "#16A34A",
  fontSize: "1.3em",
  textAlign: "center",
  verticalAlign: "middle",
  padding: "10px 16px",
  borderBottom: "1px solid var(--mon-border)",
};

const tdReps: React.CSSProperties = {
  fontWeight: 800,
  fontVariantNumeric: "tabular-nums",
  color: "var(--mon-text)" ,
  padding: "10px 16px",
  borderBottom: "1px solid var(--mon-border)",
};

const tdExercise: React.CSSProperties = {
  color: "var(--mon-text-dim)",
  padding: "10px 16px",
  borderBottom: "1px solid var(--mon-border)",
};
