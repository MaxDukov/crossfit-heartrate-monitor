const ZONE_COLORS_LIGHT: Record<number, string> = {
  1: "#0a8a06",
  2: "#0a8a06",
  3: "#f5e505",
  4: "#DC2626",
};

const ZONE_COLORS_DARK: Record<number, string> = {
  1: "#0a8a06",
  2: "#0a8a06",
  3: "#f5e505",
  4: "#EF4444",
};

const ZONE_GRADIENT_LIGHT: Record<number, string> = {
  1: "linear-gradient(160deg, rgba(10,138,6,0.18) 0%, rgba(10,138,6,0.06) 100%)",
  2: "linear-gradient(160deg, rgba(10,138,6,0.18) 0%, rgba(10,138,6,0.06) 100%)",
  3: "linear-gradient(160deg, rgba(245,229,5,0.22) 0%, rgba(245,229,5,0.08) 100%)",
  4: "linear-gradient(160deg, rgba(220,38,38,0.18) 0%, rgba(220,38,38,0.06) 100%)",
};

const ZONE_GRADIENT_DARK: Record<number, string> = {
  1: "linear-gradient(160deg, rgba(10,138,6,0.30) 0%, rgba(10,138,6,0.10) 100%)",
  2: "linear-gradient(160deg, rgba(10,138,6,0.30) 0%, rgba(10,138,6,0.10) 100%)",
  3: "linear-gradient(160deg, rgba(245,229,5,0.30) 0%, rgba(245,229,5,0.10) 100%)",
  4: "linear-gradient(160deg, rgba(239,68,68,0.30) 0%, rgba(239,68,68,0.10) 100%)",
};

export const ZONE_NAMES: Record<number, string> = {
  1: "Восстановление",
  2: "Умеренная",
  3: "Высокая",
  4: "Критическая",
};

export function getZoneColors(theme: string): Record<number, string> {
  return theme === "dark" ? ZONE_COLORS_DARK : ZONE_COLORS_LIGHT;
}

export function getZoneGradient(theme: string): Record<number, string> {
  return theme === "dark" ? ZONE_GRADIENT_DARK : ZONE_GRADIENT_LIGHT;
}
