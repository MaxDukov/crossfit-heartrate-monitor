export interface Athlete {
  id: string;
  name: string;
  max_hr: number;
  created_at: string;
  updated_at: string;
}

export interface Sensor {
  device_id: number;
  athlete_id: string | null;
  athlete_name: string | null;
  last_hr: number | null;
  last_seen_at: string | null;
  battery_level: number | null;
}

export interface Session {
  id: string;
  name: string | null;
  started_at: string;
  ended_at: string | null;
  athlete_count: number;
}

export interface HrUpdate {
  type: "hr_update";
  device_id: number;
  athlete_id: string | null;
  athlete_name: string | null;
  heart_rate: number;
  zone: number;
  zone_percent: number;
  max_hr: number | null;
}

export interface NewSensorEvent {
  type: "new_sensor";
  device_id: number;
}

export type WsMessage = HrUpdate | NewSensorEvent;

export interface SessionStats {
  session_id: string;
  session_name: string | null;
  avg_hr: number;
  max_hr: number;
  min_hr: number;
  duration_seconds: number;
  zones: {
    zone_1_seconds: number;
    zone_2_seconds: number;
    zone_3_seconds: number;
    zone_4_seconds: number;
  };
}

export interface AthleteStats {
  total_sessions: number;
  total_duration_seconds: number;
  avg_hr: number;
  max_hr_ever: number;
}

// ── WoD / Тренировки ───────────────────────────────────────

export interface Equipment {
  key: string;
  name: string;
  category: string;
  icon: string;
}

export interface GymInventoryItem {
  equipment_key: string;
  quantity: number;
}

export interface WodMovement {
  movement_key: string;
  movement_name: string;
  reps: number | null;
  weight_male: number | null;
  weight_female: number | null;
  sort_order: number;
  scaling_note: string | null;
  rounds_note: string | null;
}

export interface Wod {
  id: string;
  name: string;
  format: string;
  duration_min: number;
  intensity: string;
  theme: string;
  group_level: string;
  description: string | null;
  is_active: boolean;
  created_at: string;
  movements: WodMovement[];
}

export interface WodVariant {
  template_id: string;
  name: string;
  format: string;
  format_name: string;
  duration_min: number;
  intensity: string;
  intensity_name: string;
  theme: string;
  is_benchmark: boolean;
  description: string | null;
  target_zones: number[];
  movements: WodMovement[];
}

export type WodTheme = "legs" | "arms_shoulders" | "clean_jerk" | "snatch" | "cardio_metcon" | "gymnastics" | "core" | "full_body";

export const THEME_LABELS: Record<string, string> = {
  legs: "Ноги",
  arms_shoulders: "Руки / Плечи",
  clean_jerk: "Толчок (C&J)",
  snatch: "Рывок",
  cardio_metcon: "Кардио / Metcon",
  gymnastics: "Гимнастика",
  core: "Кор / Пресс",
  full_body: "Полное тело",
};

export const THEME_ICONS: Record<string, string> = {
  legs: "🦵",
  arms_shoulders: "💪",
  clean_jerk: "🏋️",
  snatch: "⚡",
  cardio_metcon: "🏃",
  gymnastics: "🤸",
  core: "🔥",
  full_body: "🎯",
};

export const LEVEL_LABELS: Record<string, string> = {
  beginner: "Начальный",
  intermediate: "Средний",
  advanced: "Продвинутый (RX)",
  elite: "Элитный",
};

export const FORMAT_LABELS: Record<string, string> = {
  amrap: "AMRAP",
  for_time: "На время",
  emom: "EMOM",
  tabata: "Tabata",
  chipper: "Chipper",
  ladder: "Лесенка",
  death_by: "Death By",
  strength: "Силовая",
};
