import { create } from "zustand";
import type { HrUpdate, Sensor } from "../types";
import { api } from "../lib/api";

interface HrPoint {
  t: number;
  hr: number;
}

interface HrState {
  hrData: Record<number, HrUpdate>;
  hrHistory: Record<number, HrPoint[]>;
  sensors: Sensor[];
  newSensors: number[];
  ws: WebSocket | null;
  connectWs: () => void;
  disconnectWs: () => void;
  fetchSensors: () => Promise<void>;
  dismissNewSensor: (deviceId: number) => void;
}

const FIFTEEN_MIN = 15 * 60 * 1000;

export const useHrStore = create<HrState>((set, get) => ({
  hrData: {},
  hrHistory: {},
  sensors: [],
  newSensors: [],
  ws: null,

  connectWs: () => {
    const existing = get().ws;
    if (existing && existing.readyState === WebSocket.OPEN) return;

    const protocol = location.protocol === "https:" ? "wss:" : "ws:";
    const ws = new WebSocket(`${protocol}//${location.host}/ws`);

    ws.onmessage = (e) => {
      const msg = JSON.parse(e.data);
      if (msg.type === "hr_update") {
        const now = Date.now();
        const point = { t: now, hr: msg.heart_rate };
        set((s) => {
          const prev = s.hrHistory[msg.device_id] || [];
          const cutoff = now - FIFTEEN_MIN;
          const next = [...prev, point].filter((p) => p.t >= cutoff);
          return {
            hrData: { ...s.hrData, [msg.device_id]: msg },
            hrHistory: { ...s.hrHistory, [msg.device_id]: next },
          };
        });
      } else if (msg.type === "new_sensor") {
        set((s) => ({
          newSensors: s.newSensors.includes(msg.device_id)
            ? s.newSensors
            : [...s.newSensors, msg.device_id],
        }));
        get().fetchSensors();
      }
    };

    ws.onclose = () => {
      setTimeout(() => get().connectWs(), 3000);
    };

    set({ ws });
  },

  disconnectWs: () => {
    const ws = get().ws;
    if (ws) ws.close();
    set({ ws: null });
  },

  fetchSensors: async () => {
    const sensors = await api.sensors.list();
    set({ sensors });
  },

  dismissNewSensor: (deviceId: number) => {
    set((s) => ({
      newSensors: s.newSensors.filter((id) => id !== deviceId),
    }));
  },
}));
