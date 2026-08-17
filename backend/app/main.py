"""CF-Monitor Backend — FastAPI приложение.

Запускает ANT+ collector в фоновом потоке, обслуживает REST API
и WebSocket для real-time трансляции ЧСС на фронтенд.
"""

import asyncio
import logging
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI, WebSocket, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles

from .database import init_db
from .data.seed import seed_db
from .routers import athletes, sensors, sessions, analytics, equipment, wods, system
from .services.runtime import runtime
from .services.ws_manager import manager

_logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(_app: FastAPI):
    """Управление жизненным циклом: startup/shutdown."""
    runtime.set_loop(asyncio.get_running_loop())
    init_db()
    seed_db()
    _logger.info("Database initialized and seeded")

    runtime.start()

    yield

    runtime.stop()
    _logger.info("Shutdown complete")


app = FastAPI(
    title="CF-Monitor",
    description="CrossFit HR monitoring system",
    version="0.1.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(athletes.router)
app.include_router(sensors.router)
app.include_router(sessions.router)
app.include_router(analytics.router)
app.include_router(equipment.router)
app.include_router(wods.router)
app.include_router(system.router)


@app.get("/api/health")
def health():
    """Проверка работоспособности бэкенда."""
    return {"status": "ok"}


@app.websocket("/ws")
async def websocket_endpoint(ws: WebSocket):
    """WebSocket endpoint для real-time ЧСС данных."""
    await manager.connect(ws)
    try:
        while True:
            await ws.receive_text()
    except WebSocketDisconnect:
        manager.disconnect(ws)


FRONTEND_DIR = os.environ.get("CF_FRONTEND_DIR", "")
_logger.info("CF_FRONTEND_DIR=%s, isdir=%s",
             FRONTEND_DIR,
             os.path.isdir(FRONTEND_DIR) if FRONTEND_DIR else "N/A")

if FRONTEND_DIR and os.path.isdir(FRONTEND_DIR):
    _logger.info("Mounting static files from %s", FRONTEND_DIR)
    app.mount("/", StaticFiles(directory=FRONTEND_DIR, html=True), name="static")
