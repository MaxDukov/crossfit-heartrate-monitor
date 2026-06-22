"""API для управления тренировочными сессиями."""

from datetime import datetime, timezone

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from sqlalchemy import func

from ..database import get_db
from ..models import Session as TrainingSession, SessionAthlete, HrReading, Athlete
from ..schemas import SessionCreate, SessionOut, SessionAthleteAdd

router = APIRouter(prefix="/api/sessions", tags=["sessions"])


def _session_to_out(s: TrainingSession) -> SessionOut:
    """Преобразует ORM-модель сессии в DTO."""
    return SessionOut(
        id=s.id,
        name=s.name,
        started_at=s.started_at,
        ended_at=s.ended_at,
        athlete_count=len(s.athletes),
    )


@router.get("", response_model=list[SessionOut])
def list_sessions(limit: int = 50, db: Session = Depends(get_db)):
    """Возвращает историю тренировочных сессий."""
    rows = (
        db.query(TrainingSession)
        .order_by(TrainingSession.started_at.desc())
        .limit(limit)
        .all()
    )
    return [_session_to_out(s) for s in rows]


@router.post("", response_model=SessionOut, status_code=201)
def create_session(data: SessionCreate, db: Session = Depends(get_db)):
    """Начинает новую тренировочную сессию."""
    active = (
        db.query(TrainingSession)
        .filter(TrainingSession.ended_at.is_(None))
        .first()
    )
    if active:
        raise HTTPException(400, "Уже есть активная сессия — сначала завершите её")

    session = TrainingSession(name=data.name)
    db.add(session)
    db.commit()
    db.refresh(session)
    return _session_to_out(session)


@router.get("/active", response_model=SessionOut | None)
def get_active_session(db: Session = Depends(get_db)):
    """Возвращает текущую активную сессию (если есть)."""
    s = (
        db.query(TrainingSession)
        .filter(TrainingSession.ended_at.is_(None))
        .first()
    )
    return _session_to_out(s) if s else None


@router.post("/{session_id}/end", response_model=SessionOut)
def end_session(session_id: str, db: Session = Depends(get_db)):
    """Завершает тренировочную сессию."""
    session = db.query(TrainingSession).filter(TrainingSession.id == session_id).first()
    if not session:
        raise HTTPException(404, "Сессия не найдена")
    if session.ended_at:
        raise HTTPException(400, "Сессия уже завершена")
    session.ended_at = datetime.now(timezone.utc)

    for link in session.athletes:
        if not link.left_at:
            link.left_at = session.ended_at

    db.commit()
    db.refresh(session)
    return _session_to_out(session)


@router.post("/{session_id}/athletes", status_code=201)
def add_athlete_to_session(session_id: str, data: SessionAthleteAdd, db: Session = Depends(get_db)):
    """Добавляет спортсмена в активную сессию."""
    session = db.query(TrainingSession).filter(TrainingSession.id == session_id).first()
    if not session:
        raise HTTPException(404, "Сессия не найдена")
    if session.ended_at:
        raise HTTPException(400, "Сессия уже завершена")

    athlete = db.query(Athlete).filter(Athlete.id == data.athlete_id).first()
    if not athlete:
        raise HTTPException(404, "Спортсмен не найден")

    exists = (
        db.query(SessionAthlete)
        .filter(
            SessionAthlete.session_id == session_id,
            SessionAthlete.athlete_id == data.athlete_id,
            SessionAthlete.left_at.is_(None),
        )
        .first()
    )
    if exists:
        raise HTTPException(400, "Спортсмен уже в сессии")

    link = SessionAthlete(session_id=session_id, athlete_id=data.athlete_id)
    db.add(link)
    db.commit()
    return {"status": "added"}


@router.delete("/{session_id}/athletes/{athlete_id}", status_code=204)
def remove_athlete_from_session(session_id: str, athlete_id: str, db: Session = Depends(get_db)):
    """Удаляет спортсмена из активной сессии."""
    link = (
        db.query(SessionAthlete)
        .filter(
            SessionAthlete.session_id == session_id,
            SessionAthlete.athlete_id == athlete_id,
            SessionAthlete.left_at.is_(None),
        )
        .first()
    )
    if not link:
        raise HTTPException(404, "Спортсмен не найден в сессии")
    link.left_at = datetime.now(timezone.utc)
    db.commit()
