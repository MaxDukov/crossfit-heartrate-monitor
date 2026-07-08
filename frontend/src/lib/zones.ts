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
  1: "linear-gradient(135deg, #F0FDF4, #ECFDF5)",
  2: "linear-gradient(135deg, #F0FDF4, #ECFDF5)",
  3: "linear-gradient(135deg, #FFFBEB, #FEFCE8)",
  4: "linear-gradient(135deg, #FEF2F2, #FFF7ED)",
};

const ZONE_GRADIENT_DARK: Record<number, string> = {
  1: "linear-gradient(135deg, #022C22, #065F46)",
  2: "linear-gradient(135deg, #022C22, #065F46)",
  3: "linear-gradient(135deg, #451A03, #92400E)",
  4: "linear-gradient(135deg, #450A0A, #991B1B)",
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
