# UI Design Catalog — CF-Monitor

Статичные HTML-макеты и исследование дизайна для дашборда мониторинга пульса.

## Структура

### Исследование
- **[research.md](research.md)** — исследование UI/UX: кросс-индустриальные паттерны (медицина, Ф1, Bloomberg, NOC, киберспорт), теория цвета для TV, типографика для 10-foot UI, WCAG

### Стилевые макеты (8 спортсменов, 4×2)
Каждый файл — автономный HTML, открывается прямо в браузере.

| Файл | Стиль | Палитра | Шрифт | Применение |
|---|---|---|---|---|
| [01-cartoon.html](01-cartoon.html) | Мультяшный | Яркие пастельные `#4ADE80` `#FBBF24` `#F87171` | Fredoka | Детский спорт |
| [02-neon.html](02-neon.html) | Неоновый | Чёрный + неон `#00FF9F` `#00B8FF` `#FF003C` | Orbitron / Rajdhani | Киберспорт, ночной режим |
| [03-strict.html](03-strict.html) | Строгий | Приглушённые `#2D6A4F` `#A16207` `#991B1B` | IBM Plex Sans / Mono | Медицина, профи |
| [04-techno.html](04-techno.html) | Техно | HUD `#76FF03` `#00E5FF` `#FFAB00` `#FF1744` | JetBrains Mono / Roboto | Ф1 телеметрия, гаджеты |
| [05-apple.html](05-apple.html) | Apple | Мягкие `#30D158` `#FFD60A` `#FF453A` на чёрном | SF Pro (system) | Премиум, минимализм |
| [05-apple-16.html](05-apple-16.html) | Apple (4×4) | Тот же стиль, 16 спортсменов | SF Pro (system) | Премиум, больше атлетов |
| [06-bloomberg.html](06-bloomberg.html) | Терминал | Чёрный + `#00FF65` `#FFD000` `#FF3B30` | Helvetica Narrow / Mono | Максимальная плотность данных |
| [07-retro.html](07-retro.html) | Ретро 8-bit | NES `#00AA00` `#AAAA00` `#AA0000` | Press Start 2P / VT323 | Геймификация, фан |

### Layout-макеты

#### Сетки (Grid)

| Файл | Раскладка | Кол-во | Описание |
|---|---|---|---|
| [grid-4.html](grid-4.html) | 2×2 | 4 | Широкие карточки, крупный BPM |
| [grid-8.html](grid-8.html) | 4×2 | 8 | Стандартная сетка |
| [grid-16.html](grid-16.html) | 4×4 | 16 | Мини-карточки: имя + BPM + зон-цвет |

#### Split-screen (карточки + центральный блок)

| Файл | Тема | BPM-цвет | Особенности |
|---|---|---|---|
| [split-screen.html](split-screen.html) | тёмная | по зоне | Базовый сплит |
| [split-screen-5x3.html](split-screen-5x3.html) | тёмная | по зоне | 5 столбцов × 3 строки = 15 атлетов |
| [split-screen-4x4.html](split-screen-4x4.html) | тёмная | по зоне | 4×4 = 16 атлетов |
| [split-screen-4x4-v2.html](split-screen-4x4-v2.html) | тёмная | по зоне | Увеличенные шрифты |
| [split-screen-4x4-light-v2.html](split-screen-4x4-light-v2.html) | светлая | по зоне | Увеличенные шрифты |

#### Tri-column 20-60-20 (карточки слева/справа, WoD в центре)

Левая и правая колонки — по 4 карточки спортсменов. Центральный блок — название тренировки, таймер, таблица упражнений, статы группы.

| Файл | Тема | BPM-цвет | Версия | Особенности |
|---|---|---|---|---|
| [tri-column-4x4.html](tri-column-4x4.html) | тёмная | по зоне | v1 | Базовая |
| [tri-column-4x4-v2.html](tri-column-4x4-v2.html) | тёмная | по зоне | v2 | 4× имя, 2× BPM, 4× ккал |
| [tri-column-4x4-light-v2.html](tri-column-4x4-light-v2.html) | светлая | по зоне | v2 | 4× имя, 2× BPM, 4× ккал |
| [tri-column-4x4-neutral-v2.html](tri-column-4x4-neutral-v2.html) | тёмная | нейтральный | v2 | 4× имя, 2× BPM, 4× ккал |
| [tri-column-4x4-light-neutral-v2.html](tri-column-4x4-light-neutral-v2.html) | светлая | нейтральный | v2 | 4× имя, 2× BPM, 4× ккал |
| [tri-column-4x4-neutral-v3.html](tri-column-4x4-neutral-v3.html) | тёмная | нейтральный | v3 | 9px зон-линия, −10% имя/BPM, WOD inline, таблица |
| [tri-column-4x4-light-neutral-v3.html](tri-column-4x4-light-neutral-v3.html) | светлая | нейтральный | v3 | 9px зон-линия, −10% имя/BPM, WOD inline, таблица |
| [tri-column-4x4-neutral-v4.html](tri-column-4x4-neutral-v4.html) | тёмная | нейтральный | v4 | Яркие цвета зон `#15FD07` `#FDE507` `#FD0707` |
| [tri-column-4x4-light-neutral-v4.html](tri-column-4x4-light-neutral-v4.html) | светлая | нейтральный | v4 | Яркие цвета зон `#15FD07` `#FDE507` `#FD0707` |
| [tri-column-4x4-neutral-v5.html](tri-column-4x4-neutral-v5.html) | тёмная | нейтральный | v5 | WOD: двоеточие, 3-кол таблица (Раунды\|Повторы\|Упражнение) |
| [tri-column-4x4-light-neutral-v5.html](tri-column-4x4-light-neutral-v5.html) | светлая | нейтральный | v5 | WOD: двоеточие, 3-кол таблица (Раунды\|Повторы\|Упражнение) |
| [tri-column-4x4-neutral-v6.html](tri-column-4x4-neutral-v6.html) | тёмная | нейтральный | v6 | Лидерборд убран, медали 🥇🥈🥉 в карточках (×2), +15% шрифты центра |
| [tri-column-4x4-light-neutral-v6.html](tri-column-4x4-light-neutral-v6.html) | светлая | нейтральный | v6 | Лидерборд убран, медали 🥇🥈🥉 в карточках (×2), +15% шрифты центра |

## Как смотреть

Откройте любой `.html` файл в браузере (Chrome/Chromium):
```bash
open docs/design/UI/01-cartoon.html
```

Для просмотра на 55" TV — откройте в Chromium kiosk mode:
```bash
chromium-browser --kiosk docs/design/UI/tri-column-4x4-light-neutral-v6.html
```

## Зонные цвета

### Оригинальная палитра

| Зона | % max HR | Цвет (dark) | Цвет (light) |
|---|---|---|---|
| 1 — Восстановление | ≤60% | `#0a8a06` | `#0a8a06` |
| 2 — Умеренная | 61-80% | `#0a8a06` | `#0a8a06` |
| 3 — Высокая | 81-100% | `#EAB308` | `#CA8A04` |
| 4 — Критическая | >100% | `#EF4444` | `#DC2626` |

### Яркая палитра (v4+)

| Зона | % max HR | Цвет |
|---|---|---|
| 1 — Восстановление | ≤60% | `#15FD07` |
| 2 — Умеренная | 61-80% | `#15FD07` |
| 3 — Высокая | 81-100% | `#FDE507` |
| 4 — Критическая | >100% | `#FD0707` |

## Рекомендации

Подробные рекомендации по выбору стиля, типографики и цвета — в [research.md](research.md).
