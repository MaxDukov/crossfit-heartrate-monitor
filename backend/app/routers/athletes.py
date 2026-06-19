"""CRUD API для управления спортсменами."""

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Athlete
from ..schemas import AthleteCreate, AthleteUpdate, AthleteOut

router = APIRouter(prefix="/api/athletes", tags=["athletes"])


@router.get("", response_model=list[AthleteOut])
def list_athletes(db: Session = Depends(get_db)):
    """Возвращает список всех спортсменов."""
    return db.query(Athlete).order_by(Athlete.name).all()


@router.post("", response_model=AthleteOut, status_code=201)
def create_athlete(data: AthleteCreate, db: Session = Depends(get_db)):
    """Создаёт нового спортсмена."""
    athlete = Athlete(name=data.name, max_hr=data.max_hr)
    db.add(athlete)
    db.commit()
    db.refresh(athlete)
    return athlete


@router.put("/{athlete_id}", response_model=AthleteOut)
def update_athlete(athlete_id: str, data: AthleteUpdate, db: Session = Depends(get_db)):
    """Обновляет данные спортсмена (имя, max_hr)."""
    athlete = db.query(Athlete).filter(Athlete.id == athlete_id).first()
    if not athlete:
        raise HTTPException(404, "Спортсмен не найден")
    if data.name is not None:
        athlete.name = data.name
    if data.max_hr is not None:
        athlete.max_hr = data.max_hr
    db.commit()
    db.refresh(athlete)
    return athlete


@router.delete("/{athlete_id}", status_code=204)
def delete_athlete(athlete_id: str, db: Session = Depends(get_db)):
    """Удаляет спортсмена из системы."""
    athlete = db.query(Athlete).filter(Athlete.id == athlete_id).first()
    if not athlete:
        raise HTTPException(404, "Спортсмен не найден")
    db.delete(athlete)
    db.commit()
