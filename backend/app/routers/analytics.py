"""API для аналитики: история тренировок, статистика по зонам."""

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy import func
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Athlete, HrReading, SessionAthlete, Session as TrainingSession
from ..schemas import AthleteStats, SessionStats, ZoneDistribution

router = APIRouter(prefix="/api/analytics", tags=["analytics"])


@router.get("/athletes/{athlete_id}/stats", response_model=AthleteStats)
def athlete_stats(athlete_id: str, db: Session = Depends(get_db)):
    """Возвращает агрегированную статистику спортсмена."""
    athlete = db.query(Athlete).filter(Athlete.id == athlete_id).first()
    if not athlete:
        raise HTTPException(404, "Спортсмен не найден")

    total_sessions = (
        db.query(func.count(func.distinct(HrReading.session_id)))
        .filter(HrReading.athlete_id == athlete_id, HrReading.session_id.isnot(None))
        .scalar() or 0
    )

    agg = (
        db.query(
            func.count(HrReading.id).label("count"),
            func.avg(HrReading.heart_rate).label("avg"),
            func.max(HrReading.heart_rate).label("max_hr"),
        )
        .filter(HrReading.athlete_id == athlete_id)
        .first()
    )

    duration = (
        db.query(func.sum(
            func.strftime("%s", SessionAthlete.left_at) - func.strftime("%s", SessionAthlete.joined_at)
        ))
        .filter(SessionAthlete.athlete_id == athlete_id, SessionAthlete.left_at.isnot(None))
        .scalar() or 0
    )

    return AthleteStats(
        total_sessions=total_sessions,
        total_duration_seconds=int(duration),
        avg_hr=round(agg.avg, 1) if agg and agg.avg else 0,
        max_hr_ever=agg.max_hr if agg and agg.max_hr else 0,
    )


@router.get("/athletes/{athlete_id}/history", response_model=list[SessionStats])
def athlete_history(athlete_id: str, limit: int = 20, db: Session = Depends(get_db)):
    """Возвращает историю тренировок спортсмена со статистикой по зонам."""
    athlete = db.query(Athlete).filter(Athlete.id == athlete_id).first()
    if not athlete:
        raise HTTPException(404, "Спортсмен не найден")

    readings = (
        db.query(
            HrReading.session_id,
            func.avg(HrReading.heart_rate).label("avg_hr"),
            func.max(HrReading.heart_rate).label("max_hr"),
            func.min(HrReading.heart_rate).label("min_hr"),
            func.count(HrReading.id).label("count"),
            func.sum(func.case((HrReading.zone == 1, 1), else_=0)).label("z1"),
            func.sum(func.case((HrReading.zone == 2, 1), else_=0)).label("z2"),
            func.sum(func.case((HrReading.zone == 3, 1), else_=0)).label("z3"),
            func.sum(func.case((HrReading.zone == 4, 1), else_=0)).label("z4"),
        )
        .filter(HrReading.athlete_id == athlete_id, HrReading.session_id.isnot(None))
        .group_by(HrReading.session_id)
        .order_by(HrReading.session_id.desc())
        .limit(limit)
        .all()
    )

    result = []
    for r in readings:
        session = db.query(TrainingSession).filter(TrainingSession.id == r.session_id).first()
        if not session:
            continue
        duration = 0
        if session.ended_at and session.started_at:
            duration = int((session.ended_at - session.started_at).total_seconds())
        result.append(SessionStats(
            session_id=r.session_id,
            session_name=session.name,
            avg_hr=round(r.avg_hr, 1) if r.avg_hr else 0,
            max_hr=r.max_hr or 0,
            min_hr=r.min_hr or 0,
            duration_seconds=duration,
            zones=ZoneDistribution(
                zone_1_seconds=int(r.z1 or 0),
                zone_2_seconds=int(r.z2 or 0),
                zone_3_seconds=int(r.z3 or 0),
                zone_4_seconds=int(r.z4 or 0),
            ),
        ))
    return result
