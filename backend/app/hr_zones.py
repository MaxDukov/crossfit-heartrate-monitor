"""Расчёт пульсовых зон по % от максимальной ЧСС.

Зоны (по ТЗ):
  Zone 1:  ≤60% max_hr  — Восстановление   (blue)
  Zone 2:  61-80%        — Умеренная         (green)
  Zone 3:  81-100%       — Высокая           (amber)
  Zone 4:  >100%         — Критическая       (red)
"""

ZONE_COLORS = {
    1: "#3B82F6",
    2: "#22C55E",
    3: "#F59E0B",
    4: "#EF4444",
}

ZONE_NAMES = {
    1: "Восстановление",
    2: "Умеренная",
    3: "Высокая",
    4: "Критическая",
}


def calc_zone(hr: int, max_hr: int) -> int:
    """Возвращает номер зоны (1-4) для текущего пульса."""
    pct = hr / max_hr * 100
    if pct <= 60:
        return 1
    if pct <= 80:
        return 2
    if pct <= 100:
        return 3
    return 4


def calc_percent(hr: int, max_hr: int) -> float:
    """Возвращает % от максимальной ЧСС."""
    return round(hr / max_hr * 100, 1)


def calc_calories_per_min(hr: int, weight_kg: float | None = None, age: int | None = None) -> float:
    """Расчёт расхода калорий (ккал/мин).

    Формула Keyt (точная, при наличии веса и возраста):
        (0.6309 * HR + 0.1988 * вес + 0.2017 * возраст - 55.0963) / 4.184

    Упрощённая (без веса/возраста):
        HR * 0.014
    """
    if weight_kg and age:
        kcal = (0.6309 * hr + 0.1988 * weight_kg + 0.2017 * age - 55.0963) / 4.184
        return round(max(0.0, kcal), 2)
    return round(max(0.0, hr * 0.014), 2)
