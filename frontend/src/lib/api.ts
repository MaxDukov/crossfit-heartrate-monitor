import type { Athlete, Sensor, Session, SessionStats, AthleteStats, Equipment, GymInventoryItem, Wod, WodVariant, CycleSummary, CycleDetail, Recommendation, SlotDetail, SaveResultResponse, Movement, WodTemplateItem, TemplateForEdit, CycleAnalytics } from "../types";

// Пayload редактора движений.
export interface MovementPayload {
  key?: string;
  name: string;
  modality: string;
  muscle_group: string;
  themes: string[];
  equipment_keys: string[];
  difficulty: string;
  scaling_beginner?: string;
  scaling_intermediate?: string;
}

const BASE = "/api";

async function request<T>(url: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${url}`, {
    headers: { "Content-Type": "application/json" },
    ...opts,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    const e = new Error(err.detail || res.statusText) as Error & {
      status?: number;
      body?: Record<string, unknown>;
    };
    e.status = res.status;
    e.body = err;
    throw e;
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
    active: (date?: string) =>
      request<Wod | null>(`/wods/active${date ? `?date=${date}` : ""}`),
    endActive: () => request<void>("/wods/active/end", { method: "POST" }),
    history: (limit = 20) => request<Wod[]>(`/wods/history?limit=${limit}`),
    templates: (params?: { theme?: string; search?: string; limit?: number; include_archived?: boolean }) => {
      const q = new URLSearchParams();
      if (params?.theme) q.set("theme", params.theme);
      if (params?.search) q.set("search", params.search);
      if (params?.include_archived) q.set("include_archived", "1");
      q.set("limit", String(params?.limit ?? 100));
      return request<WodTemplateItem[]>(`/wods/templates?${q}`);
    },
    template: (id: string, groupLevel = "intermediate") =>
      request<WodVariant>(`/wods/templates/${id}?group_level=${groupLevel}`),
    templateRaw: (id: string) =>
      request<TemplateForEdit>(`/wods/templates/${id}?raw=1`),
    updateTemplate: (
      id: string,
      data: {
        name?: string;
        format: string;
        duration_min: number;
        intensity: string;
        theme: string;
        description?: string;
        movements: { movement_key: string; reps?: number | null; weight_male?: number | null; weight_female?: number | null; rounds_note?: string }[];
      }
    ) =>
      request<{ template_id: string; warnings: string[] | null }>(`/wods/templates/${id}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }),
    deleteTemplate: (id: string) =>
      request<{ deleted?: boolean; archived?: boolean }>(`/wods/templates/${id}`, {
        method: "DELETE",
      }),
    restoreTemplate: (id: string) =>
      request<{ restored: boolean }>(`/wods/templates/${id}/restore`, { method: "POST" }),
    custom: (data: {
      name?: string;
      format: string;
      duration_min: number;
      intensity: string;
      theme: string;
      description?: string;
      movements: { movement_key: string; reps?: number | null; weight_male?: number | null; weight_female?: number | null; rounds_note?: string }[];
    }) =>
      request<{ template_id: string; warnings: string[] | null }>("/wods/custom", {
        method: "POST",
        body: JSON.stringify(data),
      }),
  },
  cycles: {
    list: () => request<CycleSummary[]>("/cycles"),
    create: (data: {
      name: string;
      goal?: string;
      weeks: number;
      start_date: string;
      modality_priority?: string;
      groups: { name: string; weekdays: number[]; third_day_off_cycle?: boolean }[];
    }) =>
      request<{ id: string; warnings: string[] | null }>("/cycles", {
        method: "POST",
        body: JSON.stringify(data),
      }),
    get: (id: string) => request<CycleDetail>(`/cycles/${id}`),
    setStatus: (id: string, status: string) =>
      request<{ id: string; status: string }>(`/cycles/${id}/status`, {
        method: "PUT",
        body: JSON.stringify({ status }),
      }),
    delete: (id: string) => request<void>(`/cycles/${id}`, { method: "DELETE" }),
    analytics: (id: string) => request<CycleAnalytics>(`/cycles/${id}/analytics`),
    setThirdDayOff: (cycleId: string, groupId: string, enabled: boolean) =>
      request<{ applied: boolean }>(`/cycles/${cycleId}/groups/${groupId}/third-day-off`, {
        method: "PUT",
        body: JSON.stringify({ enabled }),
      }),
    replanThirdDayOff: (cycleId: string, groupId: string, mode: "rollback" | "auto") =>
      request<{ cleared?: number; reassigned?: number; dropped?: number }>(
        `/cycles/${cycleId}/groups/${groupId}/third-day-off/replan`,
        { method: "POST", body: JSON.stringify({ mode }) }
      ),
  },
  slots: {
    get: (id: string) => request<SlotDetail>(`/slots/${id}`),
    recommendations: (id: string, groupLevel = "intermediate") =>
      request<Recommendation[]>(`/slots/${id}/recommendations?group_level=${groupLevel}`),
    assign: (id: string, templateId: string, groupLevel = "intermediate") =>
      request<{ slot_id: string; wod_id: string; warnings: string[] | null }>(`/slots/${id}/assign`, {
        method: "POST",
        body: JSON.stringify({ template_id: templateId, group_level: groupLevel }),
      }),
    unassign: (id: string) => request<void>(`/slots/${id}/assign`, { method: "DELETE" }),
    unassignWod: (id: string, wodId: string) =>
      request<void>(`/slots/${id}/assign/${wodId}`, { method: "DELETE" }),
    updateMovements: (id: string, movements: { movement_key: string; reps: number | null; weight_male: number | null; weight_female: number | null }[]) =>
      request<Wod>(`/slots/${id}/movements`, {
        method: "PUT",
        body: JSON.stringify({ movements }),
      }),
    start: (id: string) => request<{ slot_id: string; session_id: string }>(`/slots/${id}/start`, { method: "POST" }),
    complete: (id: string) => request<void>(`/slots/${id}/complete`, { method: "POST" }),
    saveResult: (id: string, data: {
      athlete_id: string;
      time_seconds?: number | null;
      rounds?: number | null;
      reps?: number | null;
      weight_kg?: number | null;
      scaled_version?: string;
      rpe?: number | null;
      notes?: string;
      movements?: { movement_key: string; weight_kg?: number | null; reps?: number | null }[];
    }) => request<SaveResultResponse>(`/slots/${id}/results`, { method: "POST", body: JSON.stringify(data) }),
  },
  movements: {
    list: (search?: string) =>
      request<Movement[]>(`/movements${search ? `?search=${encodeURIComponent(search)}` : ""}`),
    create: (m: MovementPayload) =>
      request<{ key: string }>("/movements", { method: "POST", body: JSON.stringify(m) }),
    update: (key: string, m: MovementPayload) =>
      request<{ key: string }>(`/movements/${key}`, { method: "PUT", body: JSON.stringify(m) }),
    delete: (key: string) =>
      request<void>(`/movements/${key}`, { method: "DELETE" }),
  },
  system: {
    getMode: () => request<{ mode: string }>("/system/mode"),
    setMode: (mode: string) =>
      request<{ mode: string }>("/system/mode", {
        method: "POST",
        body: JSON.stringify({ mode }),
      }),
  },
};
