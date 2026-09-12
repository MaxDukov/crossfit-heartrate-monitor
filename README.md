# CF-Monitor — мониторинг тренировок CrossFit

Real-time мониторинг пульса зала + полный контур планирования тренировок:
ANT+ датчики транслируют ЧСС на live-экран для ТВ, а тренер планирует
макроциклы, назначает WoD на дни, проводит тренировки и собирает
результаты (RPE, scaling, PR, 1ПМ) с аналитикой цикла.

**Продакшн**: Docker Compose на сервере зала — Go-бэкенд + nginx-фронтенд
(порты `:80` и `:8000`), SQLite в `./data`.

## Возможности

**Мониторинг (ТВ-экран)**
- До 8 ANT+ поясов одновременно (8 wildcard-каналов одного стика,
  автодетект новых датчиков)
- Адаптивная сетка 1→8 карточек, 3 стиля монитора (Basic / TriColumn / SplitScreen)
- 4 цветовые зоны от % максимального ЧСС, подсветка карточки, калории (формула Keyt)
- График истории ЧСС за 15 минут, медали 🥇🥈🥉 по текущей ЧСС
- Сессии: старт/финиш, состав группы, персистентные `hr_readings` для аналитики
- Активный WoD с таймером прямо на мониторе (AMRAP/EMOM/Tabata/For time)

**Планирование WoD (экран `/wod`)**
- **Планирование**: макроциклы → календарь слотов по дням → рекомендации
  генератора с учётом тем/паттернов/инвентаря → multi-wod дни (бюджет 60–65 мин)
- **Быстрый выбор**: генерация 3 вариантов тренировки под тему/группу
- **Конструктор**: ручная сборка WoD с проверкой паттернов и предупреждениями
- **Библиотека**: 54 шаблона (бенчмарки + собственные), редактирование,
  архивация выполнявшихся, защита от удаления включённых в план
- **Движения**: каталог 87 движений со scaling (М/Ж веса, проценты 1ПМ)
- **Инвентарь**: учёт оборудования зала, влияет на генерацию
- **Результаты**: ввод результатов слота → `personal_records`, `athlete_1rm`,
  RPE; аналитика цикла: план/факт, темы, динамика 1ПМ

## Архитектура

```
ANT+ пояс → USB-стик → collectors/ant (8 wildcard-каналов)
                     → HRProcessor (зоны, калории, dedup)
                     → WebSocket /ws  → React (ТВ-монитор)
                     → SQLite hr_readings (сессии, аналитика)

Тренер → React /wod → REST /api/* (циклы, слоты, шаблоны, результаты)
                    → SQLite (22 таблицы: HR, WoD, циклы, PR)
```

| Слой | Технологии |
|------|-----------|
| Бэкенд (актуальный) | Go 1.25, chi, modernc.org/sqlite (pure Go, WAL), coder/websocket, openant-go |
| Фронтенд | React 19, TypeScript, Vite 8, Tailwind CSS 4, Zustand 5, Recharts 3 |
| БД | SQLite, авто-миграции + seed (инвентарь/движения/шаблоны) с защитой пользовательских записей |
| Деплой | Docker Compose (backend-go + nginx), kiosk-браузер на хосте |

Подробности бэкенда — [backend-go/README.md](backend-go/README.md).

## Структура проекта

```
cf/
├── backend-go/            # ★ актуальный бэкенд (Go) — см. его README
│   ├── cmd/server/        # точка входа (chi, graceful shutdown)
│   └── internal/
│       ├── handlers/      # 55 REST-эндпоинтов, WebSocket hub
│       ├── services/      # HR-процессор, генератор WoD, циклы, результаты
│       ├── db/            # SQLite: миграции, seed, tombstones
│       ├── collectors/    # ant (ANT+ USB) | mock (CF_DEV_MODE=1)
│       └── data/          # справочники (генерируются tools/gen_seed.py)
├── frontend/              # React SPA: монитор, /wod, история, аналитика, настройки
├── backend/               # FastAPI-версия (архив, отстаёт от Go: 27/55 эндпоинтов)
├── ant_hr_monitor.py      # CLI-прототип (reference)
├── hr.py, web_app.py      # Flask-прототип (не деплоится)
├── docker-compose.yml     # продакшн: backend-go + nginx-фронт
└── docs/design/           # roadmap, анализ конкурентов, UI-макеты (docs/design/UI/README.md)
```

## REST API (кратко)

| Группа | Эндпоинты | Назначение |
|--------|-----------|------------|
| athletes | 4 | CRUD спортсменов (max HR, вес, возраст) |
| sensors | 5 | привязка/игнорирование ANT+ датчиков |
| sessions | 6 | старт/финиш сессии, состав, стрик-аналитика |
| analytics | 2 | статистика и история по атлету |
| wods | 11 | генерация, шаблоны (CRUD/архив/restore), активный WoD |
| cycles | 8 | макроциклы и группы |
| slots | 10 | календарь, рекомендации, проведение, результаты, PR |
| movements / equipment | 7 | справочники и инвентарь |
| system | 2 | health, dbstats, переключение mock/ANT |

WebSocket `/ws`: `hr_update`, `new_sensor`. Полный список —
`backend-go/internal/handlers/routes.go`.

## Запуск для разработки

```bash
# Бэкенд в mock-режиме (без ANT+ стика), :8001
cd backend-go
go build -o cf-server ./cmd/server
CF_DEV_MODE=1 CF_DB_PATH=/tmp/cf_monitor.db ./cf-server

# Фронтенд, :5173 (проксирует /api и /ws на :8001)
cd frontend
npm install
npm run dev
```

Тесты и проверки:

```bash
cd backend-go && go vet ./... && go test ./...
cd frontend && npx tsc -b --noEmit && npm run lint && npm run build
```

## Продакшн-деплой (Docker Compose)

```bash
docker compose up -d --build
# фронт: :80 и :8000 (kiosk-браузер хоста), API проксируется nginx'ом
```

- Бэкенд-контейнер запускается с доступом к `/dev/bus/usb` (ANT+ стик)
- БД: `./data/cf_monitor.db` (volume)
- `GET /api/health` — health-check; `GET /api/debug/dbstats` — состояние пула SQLite
- Watchdog: 4 подряд неуспешных пинга БД → выход процесса (перезапускается контейнером)

### Kiosk (ТВ)

Хост-браузер в kiosk-режиме открывает `http://localhost:8000/`
(исторический порт Python-бэкенда сохранён). Обновление фронта —
пересборка образа/копирование `dist` и перезагрузка страницы.

## Конфигурация

| Переменная | По умолчанию | Назначение |
|------------|--------------|------------|
| `CF_DB_PATH` | `/tmp/cf_monitor.db` | путь к SQLite |
| `CF_FRONTEND_DIR` | — | каталог собранного SPA (раздача статикой бэкенда) |
| `CF_DEV_MODE` | `0` | `1` — mock-коллектор + тестовые атлеты |
| `CF_PORT` | `8001` | порт HTTP (`8000` в Docker) |

## Железо

- ANT+ USB-стик: Dynastream ANTUSB2 (`0fcf:1008`) или ANTUSB-m (`0fcf:1009`)
- Пояса: любые ANT+ HR-датчики (Garmin, Polar, …) — до 8 одновременно.
  Лимит — 8 каналов одного стика; работа с несколькими стиками (16+ поясов)
  сейчас не поддерживается, коллектор открывает только первый найденный
- Дисплей 55"+ через HDMI, хост — любой Linux (ARM64/x86_64)

udev-правило для доступа к стику без root:

```bash
# /etc/udev/rules.d/99-ant-usb.rules
SUBSYSTEM=="usb", ATTRS{idVendor}=="0fcf", ATTRS{idProduct}=="1008", MODE="0666", GROUP="plugdev"
SUBSYSTEM=="usb", ATTRS{idVendor}=="0fcf", ATTRS{idProduct}=="1009", MODE="0666", GROUP="plugdev"
```

## История версий

См. [CHANGELOG.md](CHANGELOG.md). UI-макеты и их каталог —
[docs/design/UI/README.md](docs/design/UI/README.md), дорожная карта —
[docs/design/feature-roadmap.md](docs/design/feature-roadmap.md).
