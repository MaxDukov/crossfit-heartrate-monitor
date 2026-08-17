"""Pydantic-схемы для валидации запросов и ответов API."""

from datetime import datetime
from pydantic import BaseModel, Field, field_validator


# ── Athletes ──────────────────────────────────────────────

class AthleteCreate(BaseModel):
    """Данные создания спортсмена."""
    name: str = Field(..., min_length=1, max_length=100)
    max_hr: int = Field(default=190, ge=60, le=250)
    weight_kg: float | None = Field(None, ge=30, le=250)
    age: int | None = Field(None, ge=10, le=100)


class AthleteUpdate(BaseModel):
    """Данные частичного обновления спортсмена."""
    name: str | None = Field(None, min_length=1, max_length=100)
    max_hr: int | None = Field(None, ge=60, le=250)
    weight_kg: float | None = Field(None, ge=30, le=250)
    age: int | None = Field(None, ge=10, le=100)


class AthleteOut(BaseModel):
    """DTO спортсмена для ответов API."""
    id: str
    name: str
    max_hr: int
    weight_kg: float | None = None
    age: int | None = None
    created_at: datetime
    updated_at: datetime

    model_config = {"from_attributes": True}


# ── Sensors ───────────────────────────────────────────────

class SensorAssign(BaseModel):
    """Тело запроса привязки датчика к спортсмену."""
    athlete_id: str


class SensorOut(BaseModel):
    """DTO датчика с именем привязанного спортсмена."""
    device_id: int
    athlete_id: str | None = None
    athlete_name: str | None = None
    last_hr: int | None = None
    last_seen_at: datetime | None = None
    battery_level: int | None = None
    ignored: bool = False


# ── Sessions ──────────────────────────────────────────────

class SessionCreate(BaseModel):
    """Тело запроса создания тренировочной сессии."""
    name: str | None = None


class SessionOut(BaseModel):
    """DTO тренировочной сессии."""
    id: str
    name: str | None
    started_at: datetime
    ended_at: datetime | None
    athlete_count: int = 0

    model_config = {"from_attributes": True}


class SessionAthleteAdd(BaseModel):
    """Тело запроса добавления спортсмена в сессию."""
    athlete_id: str


# ── HR data ───────────────────────────────────────────────

class HrReadingOut(BaseModel):
    """DTO одиночного измерения ЧСС."""
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
    calories: float = 0.0


class NewSensorEvent(BaseModel):
    """WebSocket-событие: обнаружен новый датчик."""
    type: str = "new_sensor"
    device_id: int


# ── Analytics ─────────────────────────────────────────────

class ZoneDistribution(BaseModel):
    """Распределение времени по пульсовым зонам (сек)."""
    zone_1_seconds: int = 0
    zone_2_seconds: int = 0
    zone_3_seconds: int = 0
    zone_4_seconds: int = 0


class SessionStats(BaseModel):
    """Агрегированная статистика сессии."""
    session_id: str
    session_name: str | None
    avg_hr: float
    max_hr: int
    min_hr: int
    duration_seconds: int
    zones: ZoneDistribution


class AthleteStats(BaseModel):
    """Агрегированная статистика спортсмена за всё время."""
    total_sessions: int
    total_duration_seconds: int
    avg_hr: float
    max_hr_ever: int


# ── Equipment / Инвентарь ───────────────────────────────────

class EquipmentOut(BaseModel):
    """DTO единицы инвентаря."""
    key: str
    name: str
    category: str
    icon: str

    model_config = {"from_attributes": True}


class GymInventoryOut(BaseModel):
    """DTO позиции инвентаря зала."""
    equipment_key: str
    quantity: int

    model_config = {"from_attributes": True}


class GymInventoryUpdate(BaseModel):
    """Тело запроса обновления инвентаря зала."""
    items: list[dict]  # [{"equipment_key": "barbell", "quantity": 4}, ...]


# ── Movement ────────────────────────────────────────────────

class MovementOut(BaseModel):
    """DTO движения с распарсенными списками тем/инвентаря."""
    key: str
    name: str
    modality: str
    muscle_group: str
    themes: list[str]
    equipment_keys: list[str]
    difficulty: str
    scaling_beginner: str | None = None
    scaling_intermediate: str | None = None

    @field_validator("themes", "equipment_keys", mode="before")
    @classmethod
    def split_comma(cls, v):
        """Разбивает строку с запятыми в список."""
        if isinstance(v, str):
            return [s.strip() for s in v.split(",") if s.strip()]
        return v

    model_config = {"from_attributes": True}


# ── WodTemplate ─────────────────────────────────────────────

class WodTemplateMovementOut(BaseModel):
    """DTO движения в шаблоне WoD."""
    movement_key: str
    movement_name: str
    reps: int | None = None
    weight_male: int | None = None
    weight_female: int | None = None
    sort_order: int = 0
    rounds_note: str | None = None

    model_config = {"from_attributes": True}


class WodTemplateOut(BaseModel):
    """DTO шаблона WoD."""
    id: str
    name: str
    format: str
    duration_min: int
    intensity: str
    theme: str
    is_benchmark: bool
    description: str | None = None
    movements: list[WodTemplateMovementOut] = []

    model_config = {"from_attributes": True}


# ── Wod ─────────────────────────────────────────────────────

class WodMovementOut(BaseModel):
    """DTO движения в созданном WoD."""
    movement_key: str
    movement_name: str
    reps: int | None = None
    weight_male: int | None = None
    weight_female: int | None = None
    sort_order: int = 0
    scaling_note: str | None = None
    rounds_note: str | None = None

    model_config = {"from_attributes": True}


class WodOut(BaseModel):
    """DTO созданного WoD."""
    id: str
    name: str
    format: str
    duration_min: int
    intensity: str
    theme: str
    group_level: str
    description: str | None = None
    is_active: bool
    created_at: datetime
    movements: list[WodMovementOut] = []

    model_config = {"from_attributes": True}


class WodGenerateRequest(BaseModel):
    """Тело запроса генерации вариантов WoD."""
    theme: str
    group_level: str = "intermediate"


class WodSelectRequest(BaseModel):
    """Тело запроса выбора шаблона WoD."""
    template_id: str
    group_level: str = "intermediate"
