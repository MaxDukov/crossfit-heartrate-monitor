"""API для управления ANT+ датчиками: список, привязка/отвязка, игнорирование."""

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Sensor
from ..schemas import SensorAssign, SensorOut

router = APIRouter(prefix="/api/sensors", tags=["sensors"])


def _sensor_to_out(s: Sensor) -> SensorOut:
    """Преобразует ORM-модель в DTO с именем спортсмена."""
    return SensorOut(
        device_id=s.device_id,
        athlete_id=s.athlete_id,
        athlete_name=s.athlete.name if s.athlete else None,
        last_hr=s.last_hr,
        last_seen_at=s.last_seen_at,
        battery_level=s.battery_level,
        ignored=s.ignored,
    )


@router.get("", response_model=list[SensorOut])
def list_sensors(db: Session = Depends(get_db)):
    """Возвращает все обнаруженные ANT+ датчики с текущим статусом."""
    rows = db.query(Sensor).order_by(Sensor.last_seen_at.desc()).all()
    return [_sensor_to_out(s) for s in rows]


@router.post("/{device_id}/assign", response_model=SensorOut)
def assign_sensor(device_id: int, data: SensorAssign, db: Session = Depends(get_db)):
    """Привязывает датчик к спортсмену."""
    sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
    if not sensor:
        raise HTTPException(404, "Датчик не найден")
    if sensor.ignored:
        raise HTTPException(400, "Датчик проигнорирован — верните его в активные")

    from ..models import Athlete
    athlete = db.query(Athlete).filter(Athlete.id == data.athlete_id).first()
    if not athlete:
        raise HTTPException(404, "Спортсмен не найден")

    existing = db.query(Sensor).filter(Sensor.athlete_id == data.athlete_id).first()
    if existing and existing.device_id != device_id:
        existing.athlete_id = None

    sensor.athlete_id = data.athlete_id
    db.commit()
    db.refresh(sensor)
    return _sensor_to_out(sensor)


@router.delete("/{device_id}/assign", response_model=SensorOut)
def unassign_sensor(device_id: int, db: Session = Depends(get_db)):
    """Отвязывает датчик от спортсмена."""
    sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
    if not sensor:
        raise HTTPException(404, "Датчик не найден")
    sensor.athlete_id = None
    db.commit()
    db.refresh(sensor)
    return _sensor_to_out(sensor)


@router.post("/{device_id}/ignore", response_model=SensorOut)
def ignore_sensor(device_id: int, db: Session = Depends(get_db)):
    """Помечает датчик как проигнорированный (чужой)."""
    sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
    if not sensor:
        raise HTTPException(404, "Датчик не найден")
    sensor.ignored = True
    sensor.athlete_id = None
    db.commit()
    db.refresh(sensor)
    return _sensor_to_out(sensor)


@router.post("/{device_id}/unignore", response_model=SensorOut)
def unignore_sensor(device_id: int, db: Session = Depends(get_db)):
    """Возвращает датчик из проигнорированных в активные."""
    sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
    if not sensor:
        raise HTTPException(404, "Датчик не найден")
    sensor.ignored = False
    db.commit()
    db.refresh(sensor)
    return _sensor_to_out(sensor)
