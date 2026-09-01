# CF-Monitor Backend (Go)

Go-переписывание Python-бэкенда (`backend/app`, FastAPI). Полная
совместимость API-контракта с Python-версией: те же эндпоинты,
форматы JSON, коды ответов и WebSocket-события — фронтенд работает
с любой из версий без изменений.

## Отличия от Python-версии

- Один статический бинарь (~15 МБ), pure Go (без CGo) —
  кросс-компиляция под ARM (Raspberry Pi) без тулчейнов.
- SQLite через `modernc.org/sqlite` (pure Go), WAL-режим,
  один writer-коннект — нет конкурентных записей.
- **Исправленные дефекты:**
  - `hr_readings` теперь записывается при каждом HR-событии —
    аналитика (`/api/analytics/*`) живая;
  - переключение коллектора (`/api/system/mode`) атомарно;
  - дедлоки исключены архитектурно (едиственное соединение +
    чтения до транзакций).

## Запуск

```bash
go build -o cf-server ./cmd/server

CF_DEV_MODE=1 \
CF_DB_PATH=/tmp/cf_monitor.db \
CF_FRONTEND_DIR=../frontend/dist \
CF_PORT=8001 \
./cf-server
```

| Переменная       | По умолчанию       | Назначение                        |
|------------------|--------------------|-----------------------------------|
| `CF_DB_PATH`     | `/tmp/cf_monitor.db` | путь к SQLite                   |
| `CF_FRONTEND_DIR`| —                  | каталог собранного SPA           |
| `CF_DEV_MODE`    | `0`                | `1` — старт в mock-режиме        |
| `CF_PORT`        | `8001`             | порт HTTP (`8000` в Docker)      |

## Структура

```
cmd/server/          — точка входа (роутер chi, graceful shutdown)
internal/config/     — env-конфигурация
internal/db/         — SQLite, миграции, сиды
internal/data/       — справочники (инвентарь, движения, WoD-шаблоны)
internal/ws/         — WebSocket hub (broadcast)
internal/hrzones/    — пульсовые зоны, калории (формула Keyt)
internal/collectors/ — интерфейс Collector + mock-генератор
internal/handlers/   — REST API (athletes/sensors/sessions/...)
internal/services/   — HR-процессор, генератор WoD
tools/gen_seed.py   — конвертер сидов из Python (регенерация)
migrations docs      — см. internal/db/db.go (Migrate)
```

## Тесты / линт

```bash
go test ./...
go vet ./...
golangci-lint run   # если установлен
```

## ANT+

Коллектор реализован на базе [openant-go](https://github.com/MaxDukov/openant-go)
(Go-порт Python-openant): профиль HeartRate, 8 wildcard-каналов
(`device_id=0`), авто-рестарт через 5 с при ошибке стика.

- USB-стик: ANTUSB2 (0fcf:1008) / ANTUSB-m (0fcf:1009) через libusb;
  также поддержаны serial/CDC-стики.
- Сборка требует CGo + libusb (`brew install libusb` / `apt install
  libusb-1.0-0-dev`); в Docker это учтено.
- Юнит-тесты коллектора работают на встроенном симуляторе стика
  (`anttest`) — железо для CI не нужно: покрывают wildcard-поиск,
  реаттач канала, дедупликацию пульса и дубли `OnNewSensor`.

## Совместимость БД

Формат дат и схема идентичны SQLAlchemy-версии: обе версии могут
работать с одной и той же БД (параллельная эксплуатация при
переезде).
