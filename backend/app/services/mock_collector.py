"""Mock-коллектор для development-режима.

Генерирует случайные ЧСС-данные для 8 виртуальных датчиков без ANT+ стика.
Запускается через env CF_DEV_MODE=1.
"""

import logging
import random
import threading
import time
from datetime import datetime, timezone
from typing import Callable, Optional

from ..database import SessionLocal
from ..models import Sensor, Athlete

_logger = logging.getLogger(__name__)

MOCK_NAMES = [
    "Анна", "Борис", "Вера", "Глеб",
    "Дина", "Егор", "Жанна", "Захар",
]

MOCK_RANGES: dict[int, tuple[int, int]] = {
    1: (60, 90),
    2: (60, 90),
    3: (90, 160),
    4: (90, 160),
    5: (140, 200),
    6: (140, 200),
    7: (160, 220),
    8: (160, 220),
}


class MockCollector:
    """Фоновый генератор mock-данных для разработки."""

    def __init__(
        self,
        on_hr_data: Optional[Callable] = None,
        on_new_sensor: Optional[Callable] = None,
    ):
        self._on_hr_data = on_hr_data
        self._on_new_sensor = on_new_sensor
        self._thread: Optional[threading.Thread] = None
        self._running = False
        self._prev_hr: dict[int, int] = {}

    def start(self):
        if self._running:
            return
        self._running = True
        self._thread = threading.Thread(target=self._run, daemon=True)
        self._thread.start()
        _logger.info("Mock collector started (8 virtual sensors)")

    def stop(self):
        self._running = False
        _logger.info("Mock collector stopped")

    def _run(self):
        self._ensure_mock_athletes()

        for device_id in MOCK_RANGES:
            self._upsert_sensor(device_id)
            if self._on_new_sensor:
                try:
                    self._on_new_sensor(device_id)
                except Exception as e:
                    _logger.error(f"on_new_sensor callback error: {e}")
            time.sleep(0.3)

        while self._running:
            for device_id, (lo, hi) in MOCK_RANGES.items():
                if not self._running:
                    break

                prev = self._prev_hr.get(device_id)
                if prev is not None:
                    drift = random.randint(-5, 5)
                    hr = max(lo, min(hi, prev + drift))
                else:
                    hr = random.randint(lo, hi)

                self._prev_hr[device_id] = hr
                battery = random.randint(70, 100)

                self._update_sensor_hr(device_id, hr, battery)

                if self._on_hr_data:
                    try:
                        self._on_hr_data(device_id, hr, battery)
                    except Exception as e:
                        _logger.error(f"on_hr_data callback error: {e}")

            time.sleep(2.0)

    def _ensure_mock_athletes(self):
        """Создаёт тестовых спортсменов и привязывает к mock-датчикам."""
        db = SessionLocal()
        try:
            for i, device_id in enumerate(MOCK_RANGES):
                sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
                if not sensor:
                    sensor = Sensor(device_id=device_id)
                    db.add(sensor)
                    db.flush()

                if sensor.athlete_id:
                    continue

                name = MOCK_NAMES[i]
                athlete = db.query(Athlete).filter(Athlete.name == name).first()
                if not athlete:
                    athlete = Athlete(name=name, max_hr=190)
                    db.add(athlete)
                    db.flush()

                sensor.athlete_id = athlete.id

            db.commit()
            _logger.info("Mock athletes created and assigned")
        except Exception as e:
            _logger.error(f"Mock athletes setup error: {e}")
            db.rollback()
        finally:
            db.close()

    def _upsert_sensor(self, device_id: int):
        db = SessionLocal()
        try:
            sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
            if not sensor:
                sensor = Sensor(device_id=device_id)
                db.add(sensor)
                db.commit()
        except Exception as e:
            _logger.error(f"DB upsert sensor error: {e}")
            db.rollback()
        finally:
            db.close()

    def _update_sensor_hr(self, device_id: int, hr: int, battery: int):
        db = SessionLocal()
        try:
            sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
            if sensor:
                sensor.last_hr = hr
                sensor.last_seen_at = datetime.now(timezone.utc)
                if battery != 0xFF:
                    sensor.battery_level = battery
                db.commit()
        except Exception as e:
            _logger.error(f"DB update sensor HR error: {e}")
            db.rollback()
        finally:
            db.close()
