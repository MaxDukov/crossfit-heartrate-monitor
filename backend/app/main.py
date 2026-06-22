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
from fastapi.responses import FileResponse

from .database import init_db
from .models import Sensor, Athlete
from .database import SessionLocal
from .hr_zones import calc_zone, calc_percent
from .services.ws_manager import manager
from .services.ant_collector import AntCollector

_logger = logging.getLogger(__name__)

collector: AntCollector | None = None
_main_loop: asyncio.AbstractEventLoop | None = None


def _on_hr_data(device_id: int, hr: int, battery: int):
    """Callback из ANT+ collector: рассылает HR данные через WebSocket."""
    db = SessionLocal()
    try:
        sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
        athlete_id = sensor.athlete_id if sensor else None
        max_hr = 190
        athlete_name = None

        if athlete_id:
            athlete = db.query(Athlete).filter(Athlete.id == athlete_id).first()
            if athlete:
                max_hr = athlete.max_hr
                athlete_name = athlete.name

        zone = calc_zone(hr, max_hr)
        pct = calc_percent(hr, max_hr)

        payload = {
            "type": "hr_update",
            "device_id": device_id,
            "athlete_id": athlete_id,
            "athlete_name": athlete_name,
            "heart_rate": hr,
            "zone": zone,
            "zone_percent": pct,
            "max_hr": max_hr,
        }

        if _main_loop and _main_loop.is_running():
            asyncio.run_coroutine_threadsafe(manager.broadcast(payload), _main_loop)
    except Exception as e:
        _logger.error(f"HR data callback error: {e}")
    finally:
        db.close()


def _on_new_sensor(device_id: int):
    """Callback: новый датчик обнаружен — уведомляет фронтенд."""
    if _main_loop and _main_loop.is_running():
        asyncio.run_coroutine_threadsafe(
            manager.broadcast({"type": "new_sensor", "device_id": device_id}),
            _main_loop,
        )


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Управление жизненным циклом: startup/shutdown."""
    global collector, _main_loop
    _main_loop = asyncio.get_running_loop()
    init_db()
    _logger.info("Database initialized")

    collector = AntCollector(
        max_sensors=8,
        on_hr_data=_on_hr_data,
        on_new_sensor=_on_new_sensor,
    )
    collector.start()
    _logger.info("ANT+ collector started")

    yield

    if collector:
        collector.stop()
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

from .routers import athletes, sensors, sessions, analytics

app.include_router(athletes.router)
app.include_router(sensors.router)
app.include_router(sessions.router)
app.include_router(analytics.router)


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
print(f"[CF-MONITOR] CF_FRONTEND_DIR={FRONTEND_DIR}, isdir={os.path.isdir(FRONTEND_DIR) if FRONTEND_DIR else 'N/A'}")

if FRONTEND_DIR and os.path.isdir(FRONTEND_DIR):
    print(f"[CF-MONITOR] Mounting static files from {FRONTEND_DIR}")
    app.mount("/", StaticFiles(directory=FRONTEND_DIR, html=True), name="static")
