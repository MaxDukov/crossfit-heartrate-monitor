import { create } from "zustand";

export type DashboardStyle = "basic" | "tri-column" | "split-screen";

const STORAGE_KEY = "cf-dashboard-style";

function getInitial(): DashboardStyle {
  if (typeof window === "undefined") return "basic";
  return (localStorage.getItem(STORAGE_KEY) as DashboardStyle) || "basic";
}

interface StyleState {
  style: DashboardStyle;
  setStyle: (s: DashboardStyle) => void;
}

export const useStyle = create<StyleState>((set) => ({
  style: getInitial(),
  setStyle: (s) => {
    localStorage.setItem(STORAGE_KEY, s);
    set({ style: s });
  },
}));
