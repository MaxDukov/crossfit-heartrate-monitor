"""System router — управление режимом коллектора (mock / ant)."""

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

from ..services.runtime import runtime

router = APIRouter(prefix="/api/system", tags=["system"])


class ModeRequest(BaseModel):
    """Тело запроса переключения режима коллектора."""

    mode: str


@router.get("/mode")
def get_mode():
    """Возвращает текущий режим коллектора."""
    return {"mode": runtime.mode}


@router.post("/mode")
def set_mode(req: ModeRequest):
    """Переключает режим коллектора (mock / ant)."""
    if req.mode not in ("mock", "ant"):
        raise HTTPException(status_code=400, detail="mode must be 'mock' or 'ant'")
    runtime.switch(req.mode)
    return {"mode": req.mode}
