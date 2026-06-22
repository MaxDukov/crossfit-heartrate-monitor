"""Pydantic-схемы для валидации запросов и ответов API."""

from datetime import datetime
from pydantic import BaseModel, Field


# ── Athletes ──────────────────────────────────────────────

class AthleteCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    max_hr: int = Field(default=190, ge=60, le=250)


class AthleteUpdate(BaseModel):
    name: str | None = Field(None, min_length=1, max_length=100)
    max_hr: int | None = Field(None, ge=60, le=250)


class AthleteOut(BaseModel):
    id: str
    name: str
    max_hr: int
    created_at: datetime
    updated_at: datetime

    model_config = {"from_attributes": True}


# ── Sensors ───────────────────────────────────────────────

class SensorAssign(BaseModel):
    athlete_id: str


class SensorOut(BaseModel):
    device_id: int
    athlete_id: str | None = None
    athlete_name: str | None = None
    last_hr: int | None = None
    last_seen_at: datetime | None = None
    battery_level: int | None = None


# ── Sessions ──────────────────────────────────────────────

class SessionCreate(BaseModel):
    name: str | None = None


class SessionOut(BaseModel):
    id: str
    name: str | None
    started_at: datetime
    ended_at: datetime | None
    athlete_count: int = 0

    model_config = {"from_attributes": True}


class SessionAthleteAdd(BaseModel):
    athlete_id: str


# ── HR data ───────────────────────────────────────────────

class HrReadingOut(BaseModel):
    heart_rate: int
    zone: int
    timestamp: datetime


class HrUpdate(BaseModel):
    """Формат WebSocket-сообщения клиенту."""
    type: str = "hr_update"
    device_id: int
    athlete_id: str | None
    athlete_name: str | None
    heart_rate: int
    zone: int
    zone_percent: float
    max_hr: int | None


class NewSensorEvent(BaseModel):
    """WebSocket-событие: обнаружен новый датчик."""
    type: str = "new_sensor"
    device_id: int


# ── Analytics ─────────────────────────────────────────────

class ZoneDistribution(BaseModel):
    zone_1_seconds: int = 0
    zone_2_seconds: int = 0
    zone_3_seconds: int = 0
    zone_4_seconds: int = 0


class SessionStats(BaseModel):
    session_id: str
    session_name: str | None
    avg_hr: float
    max_hr: int
    min_hr: int
    duration_seconds: int
    zones: ZoneDistribution


class AthleteStats(BaseModel):
    total_sessions: int
    total_duration_seconds: int
    avg_hr: float
    max_hr_ever: int
