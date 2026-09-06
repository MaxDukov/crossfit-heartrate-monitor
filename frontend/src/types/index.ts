export interface Athlete {
  id: string;
  name: string;
  max_hr: number;
  weight_kg: number | null;
  age: number | null;
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
  ignored: boolean;
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
  calories: number;
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
  source?: "active" | "slot";
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

// ── Планирование тренировок (draft1.MD) ─────────────────────

export interface CycleSummary {
  id: string;
  name: string;
  goal: string | null;
  weeks: number;
  start_date: string;
  status: "planned" | "active" | "completed";
  modality_priority: string | null;
  created_at: string | null;
  groups_count: number;
  slots_total: number;
  slots_filled: number;
  slots_completed: number;
  fill_percent: number;
}

export interface CycleGroup {
  id: string;
  name: string;
  weekdays: number[];
  weekdays_names: string[];
}

export type SlotStatus = "empty" | "planned" | "in_progress" | "completed" | "skipped";

export interface SlotView {
  id: string;
  group_id: string;
  group_name: string;
  slot_date: string;
  day_number: number;
  status: SlotStatus;
  kind: "regular" | "off_cycle";
  wod_id: string | null;
  template_id: string | null;
  wod_name: string | null;
  wod_format: string | null;
  intensity: string | null;
  theme: string | null;
  notes: string | null;
}

export interface CycleDetail extends CycleSummary {
  groups: CycleGroup[];
  slots: SlotView[];
  warnings: string[] | null;
}

export interface Recommendation extends WodVariant {
  reason: string;
}

export interface SlotDetail {
  id: string;
  cycle_id: string;
  group_id: string;
  group_name: string;
  slot_date: string;
  day_number: number;
  status: SlotStatus;
  kind: "regular" | "off_cycle";
  template_id: string | null;
  wod_id: string | null;
  notes: string | null;
  session_id: string | null;
  wod: Wod | null;
  participants: { id: string; name: string; max_hr: number }[];
  results: WorkoutResult[];
}

export interface WorkoutResult {
  id: string;
  athlete_id: string;
  athlete_name: string;
  time_seconds: number | null;
  rounds: number | null;
  reps: number | null;
  weight_kg: number | null;
  scaled_version: string | null;
  rpe: number | null;
  notes: string | null;
  created_at: string;
}

export interface PRInfo {
  is_pr: boolean;
  type?: string;
  description?: string;
  previous_best?: number | null;
  new_value?: number | null;
}

export interface SaveResultResponse {
  result_id: string;
  pr: PRInfo;
  previous_attempts: { date: string; value: number; kind: string }[];
}

export interface Movement {
  key: string;
  name: string;
  modality: string;
  muscle_group: string;
  themes: string;
  equipment_keys: string;
  difficulty: string;
  scaling_beginner: string | null;
  scaling_intermediate: string | null;
}

export interface WodTemplateItem {
  template_id: string;
  name: string;
  format: string;
  duration_min: number;
  intensity: string;
  theme: string;
  is_benchmark: boolean;
  movements_count: number;
}

export interface GroupCycleStats {
  group_id: string;
  name: string;
  slots_total: number;
  slots_filled: number;
  slots_completed: number;
  slots_skipped: number;
}

export interface ThemeStat {
  theme: string;
  theme_name: string;
  planned: number;
  done: number;
}

export interface AthleteCycleStats {
  athlete_id: string;
  name: string;
  workouts_done: number;
  avg_rpe: number | null;
  total_volume_kg: number;
  pr_count: number;
}

export interface RMProgress {
  athlete_name: string;
  movement_key: string;
  movement_name: string;
  date: string;
  value: number;
}

export interface CycleAnalytics {
  slots_total: number;
  slots_filled: number;
  slots_completed: number;
  slots_skipped: number;
  fill_percent: number;
  completion_percent: number;
  groups: GroupCycleStats[];
  themes: ThemeStat[];
  athletes: AthleteCycleStats[];
  rm_progress: RMProgress[];
  advice: string[];
}

export const STATUS_LABELS: Record<SlotStatus, string> = {
  empty: "Свободен",
  planned: "Запланирован",
  in_progress: "Идёт",
  completed: "Проведён",
  skipped: "Пропущен",
};

export const MODALITY_LABELS: Record<string, string> = {
  strength: "Сила",
  gymnastics: "Гимнастика",
  cardio: "Кардио",
};

export const WEEKDAY_LABELS = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];
