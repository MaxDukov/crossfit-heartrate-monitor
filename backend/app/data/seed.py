"""Инициализация БД сид-данными (инвентарь, движения, шаблоны тренировок)."""

import logging

from ..database import SessionLocal
from ..models import Equipment, Movement, WodTemplate, WodTemplateMovement
from .equipment import EQUIPMENT_SEED
from .movements import MOVEMENTS_SEED
from .wod_templates import WOD_TEMPLATES_SEED

_logger = logging.getLogger(__name__)


def seed_db():
    """Заполняет БД сид-данными, если таблицы пустые."""
    db = SessionLocal()
    try:
        # ── Инвентарь ──
        if db.query(Equipment).count() == 0:
            for item in EQUIPMENT_SEED:
                db.add(Equipment(**item))
            db.commit()
            _logger.info("Seeded %s equipment items", len(EQUIPMENT_SEED))

        # ── Движения ──
        if db.query(Movement).count() == 0:
            for mv in MOVEMENTS_SEED:
                mv_dict = {**mv}
                mv_dict["themes"] = ",".join(mv_dict.get("themes", []))
                mv_dict["equipment_keys"] = ",".join(mv_dict.get("equipment_keys", []))
                db.add(Movement(**mv_dict))
            db.commit()
            _logger.info("Seeded %s movements", len(MOVEMENTS_SEED))

        # ── Шаблоны тренировок ──
        if db.query(WodTemplate).count() == 0:
            for tpl in WOD_TEMPLATES_SEED:
                movements_data = tpl.pop("movements", [])
                template = WodTemplate(**tpl)
                db.add(template)
                db.flush()
                for m in movements_data:
                    db.add(WodTemplateMovement(template_id=template.id, **m))
            db.commit()
            _logger.info("Seeded %s wod templates", len(WOD_TEMPLATES_SEED))

    except Exception as e:
        db.rollback()
        _logger.error("Seed error: %s", e)
    finally:
        db.close()
