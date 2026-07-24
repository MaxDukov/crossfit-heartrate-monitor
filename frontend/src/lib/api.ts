import type { Athlete, Sensor, Session, SessionStats, AthleteStats, Equipment, GymInventoryItem, Wod, WodVariant } from "../types";

const BASE = "/api";

async function request<T>(url: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${url}`, {
    headers: { "Content-Type": "application/json" },
    ...opts,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.detail || res.statusText);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  athletes: {
    list: () => request<Athlete[]>("/athletes"),
    create: (data: { name: string; max_hr: number; weight_kg?: number; age?: number }) =>
      request<Athlete>("/athletes", { method: "POST", body: JSON.stringify(data) }),
    update: (id: string, data: Partial<{ name: string; max_hr: number; weight_kg?: number; age?: number }>) =>
      request<Athlete>(`/athletes/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    delete: (id: string) =>
      request<void>(`/athletes/${id}`, { method: "DELETE" }),
  },
  sensors: {
    list: () => request<Sensor[]>("/sensors"),
    assign: (deviceId: number, athleteId: string) =>
      request<Sensor>(`/sensors/${deviceId}/assign`, {
        method: "POST",
        body: JSON.stringify({ athlete_id: athleteId }),
      }),
    unassign: (deviceId: number) =>
      request<Sensor>(`/sensors/${deviceId}/assign`, { method: "DELETE" }),
    ignore: (deviceId: number) =>
      request<Sensor>(`/sensors/${deviceId}/ignore`, { method: "POST" }),
    unignore: (deviceId: number) =>
      request<Sensor>(`/sensors/${deviceId}/unignore`, { method: "POST" }),
  },
  sessions: {
    list: () => request<Session[]>("/sessions"),
    active: () => request<Session | null>("/sessions/active"),
    create: (name?: string) =>
      request<Session>("/sessions", { method: "POST", body: JSON.stringify({ name }) }),
    end: (id: string) =>
      request<Session>(`/sessions/${id}/end`, { method: "POST" }),
    addAthlete: (sessionId: string, athleteId: string) =>
      request<void>(`/sessions/${sessionId}/athletes`, {
        method: "POST",
        body: JSON.stringify({ athlete_id: athleteId }),
      }),
    removeAthlete: (sessionId: string, athleteId: string) =>
      request<void>(`/sessions/${sessionId}/athletes/${athleteId}`, { method: "DELETE" }),
  },
  analytics: {
    athleteStats: (id: string) => request<AthleteStats>(`/analytics/athletes/${id}/stats`),
    athleteHistory: (id: string) => request<SessionStats[]>(`/analytics/athletes/${id}/history`),
  },
  equipment: {
    list: () => request<Equipment[]>("/equipment"),
    inventory: () => request<GymInventoryItem[]>("/equipment/inventory"),
    updateInventory: (items: GymInventoryItem[]) =>
      request<GymInventoryItem[]>("/equipment/inventory", {
        method: "PUT",
        body: JSON.stringify({ items }),
      }),
  },
  wods: {
    generate: (theme: string, groupLevel: string) =>
      request<WodVariant[]>("/wods/generate", {
        method: "POST",
        body: JSON.stringify({ theme, group_level: groupLevel }),
      }),
    select: (templateId: string, groupLevel: string) =>
      request<Wod>("/wods/select", {
        method: "POST",
        body: JSON.stringify({ template_id: templateId, group_level: groupLevel }),
      }),
    active: () => request<Wod | null>("/wods/active"),
    endActive: () => request<void>("/wods/active/end", { method: "POST" }),
    history: (limit = 20) => request<Wod[]>(`/wods/history?limit=${limit}`),
  },
};
