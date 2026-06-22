# CF-Monitor — CrossFit Heart Rate Monitoring System

Real-time heart rate monitoring for CrossFit gyms. ANT+ USB sensors feed live HR data to a FastAPI backend, which broadcasts to a React dashboard optimized for large TV displays.

## Features

- **Live HR monitoring** — up to 8 ANT+ heart rate sensors simultaneously
- **Adaptive grid layout** — 1 sensor (full screen), 2–3 (horizontal row), 4 (2×2), 5–8 (3×3 without center cell)
- **HR zones** — 4 color-coded zones (blue/green/amber/red) based on percentage of athlete's max HR
- **Speedometer gauge** — semicircle dial with zone segments and needle for each athlete
- **15-minute HR history chart** — real-time area chart per sensor
- **Athlete management** — create, edit, delete athletes with custom max HR
- **Sensor assignment** — pair ANT+ sensors to athletes; auto-detection of new sensors
- **Training sessions** — start/end sessions, add/remove athletes, track duration
- **Analytics** — per-athlete stats (avg HR, max HR, total sessions, zone distribution)
- **Kiosk mode** — auto-launches Chromium full-screen on boot (no keyboard/mouse needed)

## Architecture

```
ANT+ Sensor → USB Dongle → AntCollector (thread) → WebSocket → React Dashboard
                                      ↓
                               SQLite (HR readings, sessions)
```

### Backend (FastAPI + SQLite)

- **ANT+ Collector** — background thread running `openant` Node with up to 8 HR channels (auto-discovery via wildcard `device_id=0`); dispatches data to the asyncio event loop via `run_coroutine_threadsafe`
- **REST API** — 15 endpoints for athletes, sensors, sessions, and analytics
- **WebSocket** — broadcasts `hr_update` and `new_sensor` events to all connected clients
- **HR Zones** — Zone 1 (≤60%, blue/Recovery), Zone 2 (61–80%, green/Moderate), Zone 3 (81–100%, amber/High), Zone 4 (>100%, red/Critical)

### Frontend (React + TypeScript + Vite + Tailwind)

- **Zustand store** — manages WebSocket connection (auto-reconnect), live HR data, 15-minute history buffer, and sensor list
- **Recharts** — area charts for HR history
- **Responsive grid** — CSS Grid with dynamic columns/rows based on active sensor count
- **Dark theme** — optimized for 55"+ TV displays with large fonts

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | FastAPI, Uvicorn, SQLAlchemy, SQLite |
| ANT+ | openant 1.3.4, ANTUSB2 stick |
| Frontend | React 19, TypeScript, Vite 8, Tailwind CSS 4 |
| State | Zustand 5 |
| Charts | Recharts 3 |
| Deployment | systemd (production), Docker Compose (optional) |

## Project Structure

```
cf/
├── ant_hr_monitor.py              # Standalone CLI monitor (reference prototype)
├── docker-compose.yml             # Optional Docker deployment
├── backend/
│   ├── Dockerfile
│   ├── requirements.txt
│   └── app/
│       ├── main.py                # FastAPI app, lifespan, WebSocket, callbacks
│       ├── database.py            # SQLAlchemy engine & session (SQLite)
│       ├── models.py              # Athlete, Sensor, Session, SessionAthlete, HrReading
│       ├── schemas.py             # Pydantic request/response models
│       ├── hr_zones.py            # Zone & percentage calculation
│       ├── routers/
│       │   ├── athletes.py        # GET/POST/PUT/DELETE /api/athletes
│       │   ├── sensors.py         # GET /api/sensors, POST/DELETE assign
│       │   ├── sessions.py        # Session lifecycle management
│       │   └── analytics.py       # Per-athlete stats & history
│       └── services/
│           ├── ant_collector.py   # Background ANT+ collector (thread)
│           └── ws_manager.py      # WebSocket connection manager
└── frontend/
    ├── Dockerfile
    ├── nginx.conf                 # Reverse proxy for Docker deployment
    └── src/
        ├── App.tsx                # Router (Dashboard, Athletes, Sensors, Sessions, Analytics)
        ├── components/
        │   ├── AthleteCard.tsx     # HR card with gauge, BPM, history chart
        │   ├── Layout.tsx          # Nav bar, WebSocket init
        │   └── NewSensorAlert.tsx  # Sensor auto-detection notification
        ├── lib/
        │   ├── api.ts             # REST API client
        │   ├── store.ts           # Zustand store (WebSocket + HR data + history)
        │   └── zones.ts           # Zone colors, names, gradients
        ├── pages/
        │   ├── Dashboard.tsx      # Live HR grid monitor
        │   ├── AthletesPage.tsx   # Athlete CRUD
        │   ├── SensorsPage.tsx    # Sensor list & assignment
        │   ├── SessionsPage.tsx   # Training session management
        │   └── AnalyticsPage.tsx  # Per-athlete analytics
        └── types/index.ts         # TypeScript interfaces
```

## API Reference

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/athletes` | List all athletes |
| POST | `/api/athletes` | Create athlete `{name, max_hr}` |
| PUT | `/api/athletes/{id}` | Update athlete |
| DELETE | `/api/athletes/{id}` | Delete athlete |
| GET | `/api/sensors` | List all detected sensors |
| POST | `/api/sensors/{device_id}/assign` | Assign sensor to athlete `{athlete_id}` |
| DELETE | `/api/sensors/{device_id}/assign` | Unassign sensor |
| GET | `/api/sessions` | List sessions |
| POST | `/api/sessions` | Create session `{name?}` |
| GET | `/api/sessions/active` | Get active session |
| POST | `/api/sessions/{id}/end` | End session |
| POST | `/api/sessions/{id}/athletes` | Add athlete to session |
| DELETE | `/api/sessions/{id}/athletes/{athlete_id}` | Remove athlete from session |
| GET | `/api/analytics/athletes/{id}/stats` | Athlete aggregate stats |
| GET | `/api/analytics/athletes/{id}/history` | Athlete session history |
| WS | `/ws` | Real-time HR updates & sensor events |

## Getting Started

### Prerequisites

- Python 3.11+
- Node.js 18+
- ANT+ USB stick (e.g., ANTUSB2)
- ANT+ heart rate chest straps
- Linux host with USB access (for systemd deployment)

### Backend

```bash
cd backend
pip install -r requirements.txt

# Set environment variables
export CF_DB_PATH=./cf_monitor.db         # SQLite database path
export CF_FRONTEND_DIR=../frontend/dist    # Optional: serve frontend from backend

# Run
python -m uvicorn app.main:app --host 0.0.0.0 --port 8000
```

### Frontend

```bash
cd frontend
npm install
npm run dev          # Development (proxies API to localhost:8000)
npm run build        # Production build → dist/
```

### Docker (Optional)

```bash
docker-compose up -d
# Backend on :8000, Frontend on :80
```

## Production Deployment (systemd + Kiosk)

Designed for a Raspberry Pi or similar Linux device connected to a TV display.

### 1. Backend service (`/etc/systemd/system/cf-monitor.service`)

```ini
[Unit]
Description=CF-Monitor Backend
After=network.target

[Service]
Type=simple
User=ma.dukov
Environment=CF_FRONTEND_DIR=/home/ma.dukov/cf-frontend
Environment=CF_DB_PATH=/home/ma.dukov/cf_monitor.db
Environment=PYTHONPATH=/home/ma.dukov
ExecStart=/usr/bin/python3 -u -m uvicorn cf_backend.app.main:app --host 0.0.0.0 --port 8000
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### 2. Kiosk service (`/etc/systemd/system/cf-kiosk.service`)

Launches Xorg + Openbox + Chromium in kiosk mode on tty1.

```ini
[Unit]
Description=CF-Monitor Kiosk Browser
After=cf-monitor.service
Requires=cf-monitor.service
Conflicts=getty@tty1.service

[Service]
Type=simple
User=ma.dukov
Environment=DISPLAY=:0
Environment=HOME=/home/ma.dukov
TTYPath=/dev/tty1
TTYReset=yes
TTYVHangup=yes
StandardInput=tty
ExecStartPre=/bin/sh -c "while ! curl -s http://localhost:8000/api/health >/dev/null 2>&1; do sleep 2; done"
ExecStart=/usr/bin/xinit /usr/bin/openbox-session -- /usr/bin/Xorg :0 vt1 -nocursor -nolisten tcp
Restart=always
RestartSec=5

[Install]
WantedBy=graphical.target
```

### 3. Openbox autostart (`~/.config/openbox/autostart`)

```bash
xset s off
xset -dpms
xset s noblank

( while ! curl -s http://localhost:8000/api/health >/dev/null 2>&1; do
  sleep 2
done
chromium --kiosk --noerrdialogs --disable-translate --no-first-run --incognito \
  --disable-features=TranslateUI http://localhost:8000/ ) &
```

### 4. Enable & start

```bash
sudo systemctl daemon-reload
sudo systemctl enable cf-monitor cf-kiosk
sudo systemctl start cf-monitor cf-kiosk
```

## Hardware

- **ANT+ USB Stick** — Dynastream ANTUSB2 (`0fcf:1008`) or ANTUSB-m (`0fcf:1009`)
- **Sensors** — Any ANT+ heart rate chest strap (Garmin, Polar, etc.)
- **Display** — 55"+ TV via HDMI
- **Host** — Raspberry Pi 4/5 (ARM64) or any Linux x86_64

### udev rules

Ensure the USB stick is accessible without root:

```bash
# /etc/udev/rules.d/99-ant-usb.rules
SUBSYSTEM=="usb", ATTRS{idVendor}=="0fcf", ATTRS{idProduct}=="1008", MODE="0666", GROUP="plugdev"
SUBSYSTEM=="usb", ATTRS{idVendor}=="0fcf", ATTRS{idProduct}=="1009", MODE="0666", GROUP="plugdev"
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CF_DB_PATH` | `/tmp/cf_monitor.db` | SQLite database file path |
| `CF_FRONTEND_DIR` | *(empty)* | If set, backend serves static frontend from this directory |

## Updating Frontend

Since the kiosk has no keyboard, redeploy and restart the browser remotely:

```bash
cd frontend && npm run build
scp -r dist/* user@host:~/cf-frontend/
ssh user@host "sudo systemctl restart cf-kiosk"
```
