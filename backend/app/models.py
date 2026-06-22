"""SQLAlchemy ORM-модели приложения."""

import uuid
from datetime import datetime, timezone

from sqlalchemy import String, Integer, ForeignKey, Text, Index
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
