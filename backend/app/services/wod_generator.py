"""Генератор тренировок — подбирает 3 варианта WoD по теме и инвентарю.

Алгоритм:
1. Получить доступный инвентарь зала
2. Отфильтровать шаблоны по теме
3. Проверить, что все движения шаблона доступны с имеющимся инвентарём
4. Выбрать 3 варианта: бенчмарк / сбалансированный / высокоинтенсивный
5. Отскалировать повторения по уровню группы
"""

import random
import logging

from sqlalchemy.orm import Session

from ..models import (
    WodTemplate, WodTemplateMovement,
    Movement, GymInventory, Wod, WodMovement,
)

_logger = logging.getLogger(__name__)

# Множители повторений по уровню группы
LEVEL_MULTIPLIERS = {
    "beginner":     0.5,
    "intermediate": 0.75,
    "advanced":     1.0,
    "elite":        1.25,
}

# Целевые пульсовые зоны по формату тренировки
FORMAT_TARGET_ZONES = {
    "amrap":      [3, 4],   # Высокая-критическая
    "for_time":   [3, 4],
    "emom":       [2, 3],   # Умеренная-высокая
    "tabata":     [3, 4],
    "chipper":    [2, 3],   # Длинная — умеренная-высокая
    "ladder":     [3, 4],
    "death_by":   [3, 4],
    "strength":   [2],       # Умеренная
}

FORMAT_NAMES = {
    "amrap":     "AMRAP",
    "for_time":  "На время",
    "emom":      "EMOM",
    "tabata":    "Tabata",
    "chipper":   "Chipper",
    "ladder":    "Лесенка",
    "death_by":  "Death By",
    "strength":  "Силовая",
}

INTENSITY_NAMES = {
    "low":    "Низкая",
    "medium": "Средняя",
    "high":   "Высокая",
}

LEVEL_NAMES = {
    "beginner":     "Начальный",
    "intermediate": "Средний",
    "advanced":     "Продвинутый (RX)",
    "elite":        "Элитный",
}


def _get_available_equipment_keys(db: Session) -> set[str]:
    """Возвращает множество ключей инвентаря, доступного в зале."""
    items = db.query(GymInventory).all()
    return {item.equipment_key for item in items}


def _template_has_equipment(
    db: Session, template: WodTemplate, available: set[str]
) -> bool:
    """Проверяет, что все движения шаблона доступны с имеющимся инвентарём."""
    movements = db.query(WodTemplateMovement).filter(
        WodTemplateMovement.template_id == template.id
    ).all()

    for tm in movements:
        mv = db.query(Movement).filter(Movement.key == tm.movement_key).first()
        if mv and mv.equipment_keys:
            needed = {k.strip() for k in mv.equipment_keys.split(",") if k.strip()}
            if not needed.issubset(available):
                return False
    return True


def _scale_reps(reps: int | None, level: str) -> int | None:
    """Масштабирует количество повторений по уровню группы."""
    if reps is None:
        return None
    mult = LEVEL_MULTIPLIERS.get(level, 1.0)
    scaled = int(round(reps * mult))
    return max(1, scaled)


def _scale_weight(weight: int | None, level: str) -> int | None:
    """Масштабирует вес по уровню группы."""
    if weight is None:
        return None
    mult = LEVEL_MULTIPLIERS.get(level, 1.0)
    scaled = int(round(weight * mult))
    # Округляем до ближайших 2.5 кг
    rounded = round(scaled / 2.5) * 2.5
    return int(rounded) if rounded == int(rounded) else int(rounded)


def _get_scaling_note(movement_key: str, level: str, db: Session) -> str | None:
    """Возвращает текстовое описание скалирования для движения."""
    mv = db.query(Movement).filter(Movement.key == movement_key).first()
    if not mv:
        return None
    if level == "beginner" and mv.scaling_beginner:
        return mv.scaling_beginner
    if level == "intermediate" and mv.scaling_intermediate:
        return mv.scaling_intermediate
    return None


def _build_wod_from_template(
    template: WodTemplate,
    level: str,
    db: Session,
) -> dict:
    """Создаёт словарь WoD из шаблона с учётом скалирования."""
    movements = db.query(WodTemplateMovement).filter(
        WodTemplateMovement.template_id == template.id
    ).order_by(WodTemplateMovement.sort_order).all()

    wod_movements = []
    for tm in movements:
        wod_movements.append({
            "movement_key": tm.movement_key,
            "movement_name": tm.movement_name,
            "reps": _scale_reps(tm.reps, level),
            "weight_male": _scale_weight(tm.weight_male, level),
            "weight_female": _scale_weight(tm.weight_female, level),
            "sort_order": tm.sort_order,
            "scaling_note": _get_scaling_note(tm.movement_key, level, db),
            "rounds_note": tm.rounds_note,
        })

    return {
        "template_id": template.id,
        "name": template.name,
        "format": template.format,
        "format_name": FORMAT_NAMES.get(template.format, template.format),
        "duration_min": template.duration_min,
        "intensity": template.intensity,
        "intensity_name": INTENSITY_NAMES.get(template.intensity, template.intensity),
        "theme": template.theme,
        "is_benchmark": template.is_benchmark,
        "description": template.description,
        "target_zones": FORMAT_TARGET_ZONES.get(template.format, [2, 3]),
        "movements": wod_movements,
    }


def generate_wods(db: Session, theme: str, group_level: str = "intermediate") -> list[dict]:
    """Генерирует 3 варианта WoD по теме и уровню.

    Возвращает список словарей с полной информацией о тренировке,
    включая отскалированные движения.
    """
    available = _get_available_equipment_keys(db)

    # Все шаблоны по теме
    templates = db.query(WodTemplate).filter(
        WodTemplate.theme == theme
    ).all()

    # Фильтр по доступному инвентарю
    suitable = [
        t for t in templates
        if _template_has_equipment(db, t, available)
    ]

    # Если инвентарь не настроен — показываем всё
    if not available:
        suitable = templates

    if not suitable:
        _logger.warning(f"No suitable templates for theme={theme}")
        return []

    # Разделяем на категории
    benchmarks = [t for t in suitable if t.is_benchmark]
    medium = [t for t in suitable if t.intensity == "medium" and not t.is_benchmark]
    high = [t for t in suitable if t.intensity == "high" and not t.is_benchmark]
    low = [t for t in suitable if t.intensity == "low" and not t.is_benchmark]

    result = []

    # Вариант A: Классический бенчмарк
    if benchmarks:
        result.append(_build_wod_from_template(
            random.choice(benchmarks), group_level, db
        ))

    # Вариант B: Сбалансированный
    pool_b = medium + low + (benchmarks if not result else [])
    if pool_b:
        choice = random.choice(pool_b)
        wod = _build_wod_from_template(choice, group_level, db)
        if not result or wod["name"] != result[0]["name"]:
            result.append(wod)

    # Вариант C: Высокоинтенсивный
    pool_c = high + medium + (benchmarks if len(result) < 2 else [])
    if pool_c:
        choice = random.choice(pool_c)
        wod = _build_wod_from_template(choice, group_level, db)
        if not any(w["name"] == wod["name"] for w in result):
            result.append(wod)

    # Добиваем до 3 вариантов
    remaining = [t for t in suitable if not any(r["name"] == t.name for r in result)]
    while len(result) < 3 and remaining:
        choice = random.choice(remaining)
        remaining.remove(choice)
        result.append(_build_wod_from_template(choice, group_level, db))

    return result[:3]


def create_wod_from_template(
    db: Session,
    template_id: str,
    group_level: str = "intermediate",
    session_id: str | None = None,
) -> Wod:
    """Создаёт активный WoD из выбранного шаблона."""
    template = db.query(WodTemplate).filter(WodTemplate.id == template_id).first()
    if not template:
        raise ValueError(f"Template {template_id} not found")

    # Деактивируем предыдущие активные WoD
    db.query(Wod).filter(Wod.is_active == True).update({"is_active": False})

    wod = Wod(
        name=template.name,
        format=template.format,
        duration_min=template.duration_min,
        intensity=template.intensity,
        theme=template.theme,
        group_level=group_level,
        description=template.description,
        is_active=True,
        session_id=session_id,
    )
    db.add(wod)
    db.flush()

    movements = db.query(WodTemplateMovement).filter(
        WodTemplateMovement.template_id == template_id
    ).order_by(WodTemplateMovement.sort_order).all()

    for tm in movements:
        db.add(WodMovement(
            wod_id=wod.id,
            movement_key=tm.movement_key,
            movement_name=tm.movement_name,
            reps=_scale_reps(tm.reps, group_level),
            weight_male=_scale_weight(tm.weight_male, group_level),
            weight_female=_scale_weight(tm.weight_female, group_level),
            sort_order=tm.sort_order,
            scaling_note=_get_scaling_note(tm.movement_key, group_level, db),
            rounds_note=tm.rounds_note,
        ))

    db.commit()
    db.refresh(wod)
    return wod
