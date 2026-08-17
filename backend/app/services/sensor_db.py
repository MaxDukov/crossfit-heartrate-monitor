"""Общие DB-хелперы для коллекторов (ANT+ / mock).

Устраняет дублирование кода между ant_collector и mock_collector.
"""

import logging
from datetime import datetime, timezone

from ..database import SessionLocal
from ..models import Sensor

_logger = logging.getLogger(__name__)


def upsert_sensor(device_id: int):
    """Создаёт запись датчика в БД, если её ещё нет."""
    db = SessionLocal()
    try:
        sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
        if not sensor:
            sensor = Sensor(device_id=device_id)
            db.add(sensor)
            db.commit()
    except Exception as e:
        _logger.error("DB upsert sensor error: %s", e)
        db.rollback()
    finally:
        db.close()


def update_sensor_hr(device_id: int, hr: int, battery: int):
    """Обновляет последние показания датчика (ЧСС, батарея, время)."""
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
        _logger.error("DB update sensor HR error: %s", e)
        db.rollback()
    finally:
        db.close()


def is_sensor_ignored(device_id: int) -> bool:
    """Проверяет, помечен ли датчик как проигнорированный."""
    db = SessionLocal()
    try:
        sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
        return sensor is not None and sensor.ignored
    except Exception as e:
        _logger.error("DB check ignored error: %s", e)
        return False
    finally:
        db.close()


def is_sensor_assigned(device_id: int) -> bool:
    """Проверяет, привязан ли датчик к спортсмену."""
    db = SessionLocal()
    try:
        sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
        return sensor is not None and sensor.athlete_id is not None
    except Exception as e:
        _logger.error("DB check assigned error: %s", e)
        return False
    finally:
        db.close()
