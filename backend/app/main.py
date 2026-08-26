"""CF-Monitor Backend — FastAPI приложение.

Запускает ANT+ collector в фоновом потоке, обслуживает REST API
и WebSocket для real-time трансляции ЧСС на фронтенд.
"""

import time
import asyncio
import logging
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI, WebSocket, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from starlette.exceptions import HTTPException

from .database import init_db
from .data.seed import seed_db
from .models import Sensor, Athlete
from .database import SessionLocal
from .hr_zones import calc_zone, calc_percent, calc_calories_per_min
from .services.ws_manager import manager
from .services.mock_collector import MockCollector
from .routers import athletes, sensors, sessions, analytics, equipment, wods, system

_logger = logging.getLogger(__name__)

DEV_MODE = os.environ.get("CF_DEV_MODE", "0") == "1"

collector: "MockCollector | None" = None
current_mode: str = "mock" if DEV_MODE else "ant"
_main_loop: asyncio.AbstractEventLoop | None = None
_calories_accum: dict[int, float] = {}
_last_hr_time: dict[int, float] = {}


def _on_hr_data(device_id: int, hr: int, _battery: int):
    """Callback из ANT+ collector: рассылает HR данные через WebSocket."""
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

        now = time.monotonic()
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
        _logger.error("HR data callback error: %s", e)
    finally:
        db.close()


def _on_new_sensor(device_id: int):
    """Callback: новый датчик обнаружен — уведомляет фронтенд."""
    if _main_loop and _main_loop.is_running():
        asyncio.run_coroutine_threadsafe(
            manager.broadcast({"type": "new_sensor", "device_id": device_id}),
            _main_loop,
        )


def _start_collector(mode: str):
    """Создаёт и запускает коллектор нужного типа."""
    global collector, current_mode
    if mode == "mock":
        collector = MockCollector(on_hr_data=_on_hr_data, on_new_sensor=_on_new_sensor)
    else:
        from .services.ant_collector import AntCollector
        collector = AntCollector(
            max_sensors=8,
            on_hr_data=_on_hr_data,
            on_new_sensor=_on_new_sensor,
        )
    collector.start()
    current_mode = mode
    _logger.info("Collector started (%s)", mode)


def switch_collector(mode: str):
    """Останавливает текущий коллектор и запускает новый."""
    if collector:
        collector.stop()
    _calories_accum.clear()
    _last_hr_time.clear()
    _start_collector(mode)


@asynccontextmanager
async def lifespan(_app: FastAPI):
    """Управление жизненным циклом: startup/shutdown."""
    global _main_loop
    _main_loop = asyncio.get_running_loop()
    init_db()
    seed_db()
    _logger.info("Database initialized and seeded")

    _start_collector(current_mode)

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
print(
    f"[CF-MONITOR] CF_FRONTEND_DIR={FRONTEND_DIR}, "
    f"isdir={os.path.isdir(FRONTEND_DIR) if FRONTEND_DIR else 'N/A'}"
)

if FRONTEND_DIR and os.path.isdir(FRONTEND_DIR):
    print(f"[CF-MONITOR] Mounting static files from {FRONTEND_DIR}")

    class SPAStaticFiles(StaticFiles):
        """StaticFiles с SPA-fallback: неизвестные пути отдают index.html,

        чтобы клиентские роуты (/athletes, /sensors) работали при F5/прямом входе.
        """

        async def get_response(self, path: str, scope):  # type: ignore[override]
            """Отдать файл; для неизвестных путей — index.html (SPA-роутинг)."""
            try:
                return await super().get_response(path, scope)
            except HTTPException as exc:
                if exc.status_code == 404 and not path.startswith("api/"):
                    return await super().get_response("index.html", scope)
                raise

    app.mount("/", SPAStaticFiles(directory=FRONTEND_DIR, html=True), name="static")
