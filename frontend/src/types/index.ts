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
