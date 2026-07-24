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
from .data.seed import seed_db
from .models import Sensor, Athlete
from .database import SessionLocal
from .hr_zones import calc_zone, calc_percent, calc_calories_per_min
from .services.ws_manager import manager
from .services.mock_collector import MockCollector

_logger = logging.getLogger(__name__)

DEV_MODE = os.environ.get("CF_DEV_MODE", "0") == "1"

if not DEV_MODE:
    from .services.ant_collector import AntCollector

collector: "AntCollector | MockCollector | None" = None
_main_loop: asyncio.AbstractEventLoop | None = None
_calories_accum: dict[int, float] = {}
_last_hr_time: dict[int, float] = {}


def _on_hr_data(device_id: int, hr: int, battery: int):
    """Callback из ANT+ collector: рассылает HR данные через WebSocket."""
    import time as _time
    db = SessionLocal()
    try:
        sensor = db.query(Sensor).filter(Sensor.device_id == device_id).first()
        athlete_id = sensor.athlete_id if sensor else None
        max_hr = 190
        athlete_name = None
        weight_kg = None
        age = None

        if athlete_id:
            athlete = db.query(Athlete).filter(Athlete.id == athlete_id).first()
            if athlete:
                max_hr = athlete.max_hr
                athlete_name = athlete.name
                weight_kg = athlete.weight_kg
                age = athlete.age

        zone = calc_zone(hr, max_hr)
        pct = calc_percent(hr, max_hr)

        now = _time.monotonic()
        prev_t = _last_hr_time.get(device_id)
        if prev_t is not None:
            dt_min = (now - prev_t) / 60.0
            kcal_rate = calc_calories_per_min(hr, weight_kg, age)
            _calories_accum[device_id] = _calories_accum.get(device_id, 0.0) + kcal_rate * dt_min
        _last_hr_time[device_id] = now

        calories = round(_calories_accum.get(device_id, 0.0), 1)

        payload = {
            "type": "hr_update",
            "device_id": device_id,
            "athlete_id": athlete_id,
            "athlete_name": athlete_name,
            "heart_rate": hr,
            "zone": zone,
            "zone_percent": pct,
            "max_hr": max_hr,
            "calories": calories,
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
    seed_db()
    _logger.info("Database initialized and seeded")

    if DEV_MODE:
        _logger.info("=== CF DEV MODE — mock collector (8 virtual sensors) ===")
        collector = MockCollector(
            on_hr_data=_on_hr_data,
            on_new_sensor=_on_new_sensor,
        )
    else:
        collector = AntCollector(
            max_sensors=8,
            on_hr_data=_on_hr_data,
            on_new_sensor=_on_new_sensor,
        )
    collector.start()
    _logger.info(f"Collector started ({'mock' if DEV_MODE else 'ANT+'})")

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

from .routers import athletes, sensors, sessions, analytics, equipment, wods

app.include_router(athletes.router)
app.include_router(sensors.router)
app.include_router(sessions.router)
app.include_router(analytics.router)
app.include_router(equipment.router)
app.include_router(wods.router)


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
