import type { Athlete, Sensor, Session, SessionStats, AthleteStats } from "../types";

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
    create: (data: { name: string; max_hr: number }) =>
      request<Athlete>("/athletes", { method: "POST", body: JSON.stringify(data) }),
    update: (id: string, data: Partial<{ name: string; max_hr: number }>) =>
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
};
