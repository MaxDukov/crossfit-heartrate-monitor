"""API для генерации и управления WoD (тренировками дня)."""

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Wod, WodMovement
from ..schemas import WodGenerateRequest, WodSelectRequest, WodOut
from ..services.wod_generator import generate_wods, create_wod_from_template

router = APIRouter(prefix="/api/wods", tags=["wods"])


def _wod_to_out(wod: Wod, movements: list[WodMovement]) -> dict:
    return {
        "id": wod.id,
        "name": wod.name,
        "format": wod.format,
        "duration_min": wod.duration_min,
        "intensity": wod.intensity,
        "theme": wod.theme,
        "group_level": wod.group_level,
        "description": wod.description,
        "is_active": wod.is_active,
        "created_at": wod.created_at,
        "movements": [
            {
                "movement_key": m.movement_key,
                "movement_name": m.movement_name,
                "reps": m.reps,
                "weight_male": m.weight_male,
                "weight_female": m.weight_female,
                "sort_order": m.sort_order,
                "scaling_note": m.scaling_note,
                "rounds_note": m.rounds_note,
            }
            for m in movements
        ],
    }


@router.post("/generate")
def generate(req: WodGenerateRequest, db: Session = Depends(get_db)):
    """Генерирует 3 варианта WoD по теме и уровню группы."""
    variants = generate_wods(db, req.theme, req.group_level)
    if not variants:
        raise HTTPException(404, "Нет подходящих шаблонов для данной темы и инвентаря")
    return variants


@router.post("/select", response_model=WodOut)
def select_wod(req: WodSelectRequest, db: Session = Depends(get_db)):
    """Создаёт активный WoD из выбранного шаблона."""
    try:
        wod = create_wod_from_template(db, req.template_id, req.group_level)
    except ValueError as e:
        raise HTTPException(404, str(e))
    movements = db.query(WodMovement).filter(
        WodMovement.wod_id == wod.id
    ).order_by(WodMovement.sort_order).all()
    return _wod_to_out(wod, movements)


@router.get("/active")
def get_active_wod(db: Session = Depends(get_db)):
    """Возвращает текущий активный WoD или null."""
    wod = db.query(Wod).filter(Wod.is_active == True).first()
    if not wod:
        return None
    movements = db.query(WodMovement).filter(
        WodMovement.wod_id == wod.id
    ).order_by(WodMovement.sort_order).all()
    return _wod_to_out(wod, movements)


@router.post("/active/end", status_code=204)
def end_active_wod(db: Session = Depends(get_db)):
    """Деактивирует текущий активный WoD."""
    db.query(Wod).filter(Wod.is_active == True).update({"is_active": False})
    db.commit()


@router.get("/history")
def list_history(limit: int = 20, db: Session = Depends(get_db)):
    """Возвращает историю тренировок."""
    wods = db.query(Wod).order_by(Wod.created_at.desc()).limit(limit).all()
    result = []
    for wod in wods:
        movements = db.query(WodMovement).filter(
            WodMovement.wod_id == wod.id
        ).order_by(WodMovement.sort_order).all()
        result.append(_wod_to_out(wod, movements))
    return result
