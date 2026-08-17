import type { HrUpdate } from "../../types";

const MEDAL_SYMBOLS = ["🥇", "🥈", "🥉"] as const;

export function computeMedals(
  hrData: Record<number, HrUpdate>
): Map<number, string> {
  const sorted = Object.values(hrData)
    .filter((d) => d.calories > 0)
    .sort((a, b) => b.calories - a.calories);

  const medals = new Map<number, string>();
  for (let i = 0; i < Math.min(3, sorted.length); i++) {
    medals.set(sorted[i].device_id, MEDAL_SYMBOLS[i]);
  }
  return medals;
}
