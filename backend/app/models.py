"""SQLAlchemy ORM-модели приложения."""

import uuid
from datetime import datetime, timezone

from sqlalchemy import String, Integer, Boolean, ForeignKey, Text, Index
from sqlalchemy.orm import Mapped, mapped_column, relationship

from .database import Base


def _uuid() -> str:
    return str(uuid.uuid4())


def _now() -> datetime:
    return datetime.now(timezone.utc)


class Athlete(Base):
    """Спортсмен — имя и максимальная ЧСС для расчёта зон."""
    __tablename__ = "athletes"

    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=_uuid)
    name: Mapped[str] = mapped_column(String(100), nullable=False)
    max_hr: Mapped[int] = mapped_column(Integer, nullable=False, default=190)
    created_at: Mapped[datetime] = mapped_column(default=_now)
    updated_at: Mapped[datetime] = mapped_column(default=_now, onupdate=_now)

    sensor: Mapped["Sensor | None"] = relationship(back_populates="athlete")
    readings: Mapped[list["HrReading"]] = relationship(back_populates="athlete")
    session_links: Mapped[list["SessionAthlete"]] = relationship(back_populates="athlete")


class Sensor(Base):
    """Обнаруженный ANT+ датчик. Привязывается к спортсмену."""
    __tablename__ = "sensors"

    device_id: Mapped[int] = mapped_column(Integer, primary_key=True)
    athlete_id: Mapped[str | None] = mapped_column(
        String(36), ForeignKey("athletes.id", ondelete="SET NULL"), nullable=True
    )
    last_hr: Mapped[int | None] = mapped_column(Integer, nullable=True)
    last_seen_at: Mapped[datetime | None] = mapped_column(nullable=True)
    battery_level: Mapped[int | None] = mapped_column(Integer, nullable=True)
    ignored: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)

    athlete: Mapped["Athlete | None"] = relationship(back_populates="sensor")


class Session(Base):
    """Тренировочная сессия."""
    __tablename__ = "sessions"

    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=_uuid)
    name: Mapped[str | None] = mapped_column(String(200), nullable=True)
    started_at: Mapped[datetime] = mapped_column(default=_now)
    ended_at: Mapped[datetime | None] = mapped_column(nullable=True)

    athletes: Mapped[list["SessionAthlete"]] = relationship(back_populates="session")
    readings: Mapped[list["HrReading"]] = relationship(back_populates="session")


class SessionAthlete(Base):
    """Связь сессия↔спортсмен (кто участвовал, когда вошёл/вышел)."""
    __tablename__ = "session_athletes"
    __table_args__ = (
        Index("ix_sa_session_athlete", "session_id", "athlete_id"),
    )

    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=_uuid)
    session_id: Mapped[str] = mapped_column(
        String(36), ForeignKey("sessions.id", ondelete="CASCADE"), nullable=False
    )
    athlete_id: Mapped[str] = mapped_column(
        String(36), ForeignKey("athletes.id", ondelete="CASCADE"), nullable=False
    )
    joined_at: Mapped[datetime] = mapped_column(default=_now)
    left_at: Mapped[datetime | None] = mapped_column(nullable=True)

    session: Mapped["Session"] = relationship(back_populates="athletes")
    athlete: Mapped["Athlete"] = relationship(back_populates="session_links")


class HrReading(Base):
    """Запись ЧСС — одна точка time-series данных."""
    __tablename__ = "hr_readings"
    __table_args__ = (
        Index("ix_hr_athlete_ts", "athlete_id", "timestamp"),
    )

    id: Mapped[int] = mapped_column(Integer, primary_key=True, autoincrement=True)
    athlete_id: Mapped[str] = mapped_column(
        String(36), ForeignKey("athletes.id", ondelete="CASCADE"), nullable=False
    )
    session_id: Mapped[str | None] = mapped_column(
        String(36), ForeignKey("sessions.id", ondelete="SET NULL"), nullable=True
    )
    heart_rate: Mapped[int] = mapped_column(Integer, nullable=False)
    zone: Mapped[int] = mapped_column(Integer, nullable=False)
    timestamp: Mapped[datetime] = mapped_column(default=_now)

    athlete: Mapped["Athlete"] = relationship(back_populates="readings")
    session: Mapped["Session | None"] = relationship(back_populates="readings")


# ── WoD / Тренировки ────────────────────────────────────────

class Equipment(Base):
    """Каталог инвентаря."""
    __tablename__ = "equipment"

    key: Mapped[str] = mapped_column(String(50), primary_key=True)
    name: Mapped[str] = mapped_column(String(100), nullable=False)
    category: Mapped[str] = mapped_column(String(50), nullable=False)
    icon: Mapped[str] = mapped_column(String(10), nullable=False, default="📦")


class GymInventory(Base):
    """Инвентарь, доступный в зале."""
    __tablename__ = "gym_inventory"

    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=_uuid)
    equipment_key: Mapped[str] = mapped_column(
        String(50), ForeignKey("equipment.key", ondelete="CASCADE"), nullable=False
    )
    quantity: Mapped[int] = mapped_column(Integer, nullable=False, default=1)


class Movement(Base):
    """Каталог движений."""
    __tablename__ = "movements"

    key: Mapped[str] = mapped_column(String(80), primary_key=True)
    name: Mapped[str] = mapped_column(String(100), nullable=False)
    modality: Mapped[str] = mapped_column(String(30), nullable=False)
    muscle_group: Mapped[str] = mapped_column(String(30), nullable=False)
    themes: Mapped[str] = mapped_column(Text, nullable=False, default="")
    equipment_keys: Mapped[str] = mapped_column(Text, nullable=False, default="")
    difficulty: Mapped[str] = mapped_column(String(20), nullable=False, default="intermediate")
    scaling_beginner: Mapped[str | None] = mapped_column(Text, nullable=True)
    scaling_intermediate: Mapped[str | None] = mapped_column(Text, nullable=True)


class WodTemplate(Base):
    """Шаблон тренировки."""
    __tablename__ = "wod_templates"

    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=_uuid)
    name: Mapped[str] = mapped_column(String(200), nullable=False)
    format: Mapped[str] = mapped_column(String(30), nullable=False)
    duration_min: Mapped[int] = mapped_column(Integer, nullable=False)
    intensity: Mapped[str] = mapped_column(String(20), nullable=False, default="medium")
    theme: Mapped[str] = mapped_column(String(30), nullable=False)
    is_benchmark: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    description: Mapped[str | None] = mapped_column(Text, nullable=True)

    movements: Mapped[list["WodTemplateMovement"]] = relationship(
        back_populates="template", cascade="all, delete-orphan",
        order_by="WodTemplateMovement.sort_order",
    )


class WodTemplateMovement(Base):
    """Движение в шаблоне тренировки."""
    __tablename__ = "wod_template_movements"

    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=_uuid)
    template_id: Mapped[str] = mapped_column(
        String(36), ForeignKey("wod_templates.id", ondelete="CASCADE"), nullable=False
    )
    movement_key: Mapped[str] = mapped_column(String(80), nullable=False)
    movement_name: Mapped[str] = mapped_column(String(100), nullable=False)
    reps: Mapped[int | None] = mapped_column(Integer, nullable=True)
    weight_male: Mapped[int | None] = mapped_column(Integer, nullable=True)
    weight_female: Mapped[int | None] = mapped_column(Integer, nullable=True)
    sort_order: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    rounds_note: Mapped[str | None] = mapped_column(String(100), nullable=True)

    template: Mapped["WodTemplate"] = relationship(back_populates="movements")


class Wod(Base):
    """Выбранная тренировка дня."""
    __tablename__ = "wods"

    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=_uuid)
    name: Mapped[str] = mapped_column(String(200), nullable=False)
    format: Mapped[str] = mapped_column(String(30), nullable=False)
    duration_min: Mapped[int] = mapped_column(Integer, nullable=False)
    intensity: Mapped[str] = mapped_column(String(20), nullable=False, default="medium")
    theme: Mapped[str] = mapped_column(String(30), nullable=False)
    group_level: Mapped[str] = mapped_column(String(20), nullable=False, default="intermediate")
    description: Mapped[str | None] = mapped_column(Text, nullable=True)
    is_active: Mapped[bool] = mapped_column(Boolean, nullable=False, default=False)
    created_at: Mapped[datetime] = mapped_column(default=_now)
    session_id: Mapped[str | None] = mapped_column(
        String(36), ForeignKey("sessions.id", ondelete="SET NULL"), nullable=True
    )

    movements: Mapped[list["WodMovement"]] = relationship(
        back_populates="wod", cascade="all, delete-orphan",
        order_by="WodMovement.sort_order",
    )


class WodMovement(Base):
    """Движение в выбранной тренировке."""
    __tablename__ = "wod_movements"

    id: Mapped[str] = mapped_column(String(36), primary_key=True, default=_uuid)
    wod_id: Mapped[str] = mapped_column(
        String(36), ForeignKey("wods.id", ondelete="CASCADE"), nullable=False
    )
    movement_key: Mapped[str] = mapped_column(String(80), nullable=False)
    movement_name: Mapped[str] = mapped_column(String(100), nullable=False)
    reps: Mapped[int | None] = mapped_column(Integer, nullable=True)
    weight_male: Mapped[int | None] = mapped_column(Integer, nullable=True)
    weight_female: Mapped[int | None] = mapped_column(Integer, nullable=True)
    sort_order: Mapped[int] = mapped_column(Integer, nullable=False, default=0)
    scaling_note: Mapped[str | None] = mapped_column(Text, nullable=True)
    rounds_note: Mapped[str | None] = mapped_column(String(100), nullable=True)

    wod: Mapped["Wod"] = relationship(back_populates="movements")
