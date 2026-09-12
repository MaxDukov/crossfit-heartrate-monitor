# ChangeLog

Все заметные изменения проекта CF-Monitor документируются в этом файле.

Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/),
версионирование следует [Semantic Versioning](https://semver.org/lang/ru/).

---

## [Неопубликовано] — v0.4.0

### Добавлено
- **Go-бэкенд** (`backend-go/`): полная замена FastAPI в проде — один статический
  бинарь, pure-Go SQLite (`modernc.org/sqlite`, WAL), ANT+ через
  [openant-go](https://github.com/MaxDukov/openant-go), 55 REST-эндпоинтов,
  WebSocket, watchdog БД, atomic-переключение mock/ANT (`/api/system/mode`);
  живые `hr_readings` на каждое HR-событие
- **Планирование тренировок** (`docs/design/trainings_workflow/draft1.MD`, 7 экранов):
  макроциклы → календарь слотов → рекомендации генератора (темы/паттерны/инвентарь)
  → multi-wod дни с бюджетом 60/65 мин → проведение слота → результаты
  (RPE, scaling) → PR и 1ПМ (`personal_records`, `athlete_1rm`) → аналитика цикла
  (план/факт, темы, динамика 1ПМ)
- **Экран `/wod`**: вкладки Быстрый выбор / Планирование / Конструктор /
  Библиотека / Движения / Инвентарь
- **Библиотека шаблонов**: карточки в стиле быстрого выбора, редактирование
  (PUT + `GET ?raw=1` с нескалированными движениями), удаление с проверками —
  409 со ссылками на циклы при включении в план, архивация выполнявшихся
  (скрыты из генерации/назначения), восстановление из архива
- **Защита пользовательских данных**: `is_custom` у правленых шаблонов,
  tombstones `wod_templates_deleted` / `movements_deleted` против восстановления
  сидом после рестарта
- **История тренировок** (HistoryPage) и страница цикла с аналитикой
- **CI**: Go (vet + test) и фронтенд (tsc + eslint + build) в GitHub Actions
- **Тесты**: workflows циклов, результаты/PR, генератор WoD

### Изменено
- **Инвентарь**: переименование в единственное число («Диски», «Гриф для
  штанги», «Гантеля»…), убраны иконки, количество вводится с клавиатуры
  (стираемое значение, выделение при фокусе)
- README переписан под Go-стек и текущую функциональность

---

## [v0.3.0] — 2026-09-01

### Добавлено
- **Расчёт калорий**: формула Keyt `(0.6309 × HR + 0.1988 × вес + 0.2017 × возраст − 55.0963) / 4.184`, упрощённая `HR × 0.014` без данных
- **Поля спортсмена**: `weight_kg` и `age` в модели Athlete (+ авто-миграция БД)
- **Calories в WebSocket**: накопительное значение ккал за сессию, передаётся в каждом `hr_update`
- **Калории на карточке**: отображение `N ккал` рядом с зоной на AthleteCard
- **Поля вес/возраст в AthletesPage**: при создании и редактировании спортсмена
- **Mock dev-режим** (`CF_DEV_MODE=1`): 8 виртуальных датчиков с реалистичными ЧСС-данными (60-220 bpm), без ANT+ стика
- **Автосоздание спортсменов**: mock-коллектор автоматически создаёт 8 тестовых спортсменов (Анна, Борис, Вера...) и привязывает их к датчикам, с тестовыми весом/возрастом
- **Экранная клавиатура** (`OnScreenKeyboard`): RU/EN раскладки, открывается при фокусе на input в AthletesPage
- **Игнорирование датчиков**: rest-API `ignore/unignore`, suppressed в алертах и WebSocket, серая плашка "Чужой"
- **Дорожная карта функций** (`docs/design/feature-roadmap.md`): 10 фич, 3 приоритетных тира
- **Анализ конкурентов** (`docs/design/competitor-analysis.md`): Myzone, Orangetheory, WHOOP, SugarWOD, BTWB, Polar
- **UI-предложения** (`docs/design/ui-proposals.md`)
- Данный ChangeLog

### Изменено
- **SensorsPage**: разделён на вкладки «Обнаруженные» (непривязанные) и «Привязанные» — назначенные датчики уходят из обнаружения
- **AthleteCard**: цветной фон по зоне пульса (rgba градиент, 18-22% light / 30% dark, transition 0.8s)
- **AthleteCard**: адаптивные размеры через container queries (`cqw`): имя 16cqw, ЧСС 28cqw, график 16cqw
- **Dashboard**: адаптивная сетка 1-8 спортсменов (1x1 → 4x2), убраны пустые div-хаки
- Цвета зон: зелёный `#0a8a06`, жёлтый `#f5e505`, красный `#DC2626`
- Миграция БД: авто-`ALTER TABLE` для колонки `ignored` в `_run_migrations()`

### Удалено
- **Спидометр (gauge)**: убран из AthleteCard — ЧСС число теперь единственный центральный элемент

### Инфраструктура
- Установлен `xserver-xorg-input-libinput` на сервере, убран флаг `-nocursor` — клавиатура/мышь работают
- `.gitignore`: `backend/.venv/`, `*.db`, `__pycache__/`

---

## [v0.2.0] — 2026-07-23

### Добавлено
- REST API: athletes CRUD, sensors assign/unassign, sessions, analytics, equipment, wods
- WebSocket: real-time трансляция ЧСС и события `new_sensor`
- ANT+ collector (`AntCollector`) с `_is_sensor_assigned()` / `_is_sensor_ignored()` проверками
- SQLite с auto-seed: 17 equipment, 80 movements, 54 wod templates
- WoD generator с фильтрацией по оборудованию, 3 варианта
- Frontend: 7 страниц (Dashboard, Athletes, Sensors, Sessions, Analytics, Equipment, WodBuilder)
- Zustand store с 15-мин историей ЧСС, WebSocket с авто-реконнектом
- Light/dark тема, навигация снизу, WorkoutTimer, WodPanel, NewSensorAlert
- Systemd-сервисы, 512M swap, Openbox autostart с Chromium kiosk + watchdog

---

## [v0.1.0] — 2026-07-20

### Добавлено
- Базовая структура проекта: FastAPI backend + React/TypeScript frontend
- MVP Dashboard с AthleteCard
- ANT+ интеграция через `openant`
- Первый деплой на Raspberry Pi
