"""System router — управление режимом коллектора (mock / ant)."""

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel

router = APIRouter(prefix="/api/system", tags=["system"])


class ModeRequest(BaseModel):
    mode: str


@router.get("/mode")
def get_mode():
    from ..main import current_mode
    return {"mode": current_mode}


@router.post("/mode")
def set_mode(req: ModeRequest):
    from ..main import switch_collector
    if req.mode not in ("mock", "ant"):
        raise HTTPException(status_code=400, detail="mode must be 'mock' or 'ant'")
    switch_collector(req.mode)
    return {"mode": req.mode}
