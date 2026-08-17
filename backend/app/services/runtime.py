"""Runtime-состояние приложения: режим коллектора, инстанс, калории.

Вынесено из main.py, чтобы routers/system.py не создавал
циклический импорт (main → routers → main).
"""

import asyncio
import logging
import os
import time
from typing import Optional, Protocol

from ..database import SessionLocal
from ..hr_zones import calc_zone, calc_percent, calc_calories_per_min
from ..models import Athlete, Sensor
from .ws_manager import manager

_logger = logging.getLogger(__name__)


class Collector(Protocol):
    """Интерфейс коллектора (реализован AntCollector и MockCollector)."""

    def start(self):
        """Запускает коллектор."""

    def stop(self):
        """Останавливает коллектор."""


class CollectorRuntime:
    """Управляет жизненным циклом коллектора и аккумулятором калорий."""

    def __init__(self):
        dev_mode = os.environ.get("CF_DEV_MODE", "0") == "1"
        self.collector: Optional[Collector] = None
        self.mode: str = "mock" if dev_mode else "ant"
        self._main_loop: Optional[asyncio.AbstractEventLoop] = None
        self._calories_accum: dict[int, float] = {}
        self._last_hr_time: dict[int, float] = {}

    def set_loop(self, loop: asyncio.AbstractEventLoop):
        """Сохраняет главный event loop для broadcast из фоновых потоков."""
        self._main_loop = loop

    def on_hr_data(self, device_id: int, hr: int, battery: int):
        """Callback коллектора: считает калории и рассылает HR через WS."""
        db = SessionLocal()
        try:
            sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
            athlete_id = sensor.athlete_id if sensor else None
            max_hr = 190
            athlete_name = None
            weight_kg = None
            age = None

            if athlete_id:
                athlete = db.query(Athlete).filter(Athlete.id == athlete_id).first()
                if athlete:
                    max_hr = athlete.max_hr
                    athlete_name = athlete.name
                    weight_kg = athlete.weight_kg
                    age = athlete.age

            zone = calc_zone(hr, max_hr)
            pct = calc_percent(hr, max_hr)

            now = time.monotonic()
            prev_t = self._last_hr_time.get(device_id)
            if prev_t is not None:
                dt_min = (now - prev_t) / 60.0
                kcal_rate = calc_calories_per_min(hr, weight_kg, age)
                self._calories_accum[device_id] = (
                    self._calories_accum.get(device_id, 0.0) + kcal_rate * dt_min
                )
            self._last_hr_time[device_id] = now

            calories = round(self._calories_accum.get(device_id, 0.0), 1)

            payload = {
                "type": "hr_update",
                "device_id": device_id,
                "athlete_id": athlete_id,
                "athlete_name": athlete_name,
                "heart_rate": hr,
                "zone": zone,
                "zone_percent": pct,
                "max_hr": max_hr,
                "calories": calories,
            }

            if self._main_loop and self._main_loop.is_running():
                asyncio.run_coroutine_threadsafe(
                    manager.broadcast(payload), self._main_loop
                )
        except Exception as e:
            _logger.error("HR data callback error: %s", e)
        finally:
            db.close()

    def on_new_sensor(self, device_id: int):
        """Callback коллектора: новый датчик — уведомляет фронтенд."""
        if self._main_loop and self._main_loop.is_running():
            asyncio.run_coroutine_threadsafe(
                manager.broadcast({"type": "new_sensor", "device_id": device_id}),
                self._main_loop,
            )

    def _create_collector(self, mode: str) -> Collector:
        """Создаёт коллектор нужного типа по режиму."""
        # Ленивые импорты: mock-режим не тянет openant (железозависимый пакет)
        if mode == "mock":
            from .mock_collector import MockCollector  # pylint: disable=import-outside-toplevel
            return MockCollector(
                on_hr_data=self.on_hr_data, on_new_sensor=self.on_new_sensor
            )
        from .ant_collector import AntCollector  # pylint: disable=import-outside-toplevel
        return AntCollector(
            max_sensors=8,
            on_hr_data=self.on_hr_data,
            on_new_sensor=self.on_new_sensor,
        )

    def start(self, mode: Optional[str] = None):
        """Создаёт и запускает коллектор указанного (или текущего) режима."""
        if mode is not None:
            self.mode = mode
        self.collector = self._create_collector(self.mode)
        self.collector.start()
        _logger.info("Collector started (%s)", self.mode)

    def stop(self):
        """Останавливает текущий коллектор."""
        if self.collector:
            self.collector.stop()
        self.collector = None

    def switch(self, mode: str):
        """Останавливает текущий коллектор и запускает новый режим."""
        self.stop()
        self._calories_accum.clear()
        self._last_hr_time.clear()
        self.start(mode)


runtime = CollectorRuntime()
