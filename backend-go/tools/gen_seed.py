#!/usr/bin/env python3
"""Конвертация Python-сидов (equipment/movements/wod_templates) в Go-слайсы.

Запуск из корня репозитория:
    python3 backend-go/tools/gen_seed.py
Выходные файлы перезаписываются в backend-go/internal/data/.
"""

import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "..", "..", "backend"))

from app.data.equipment import EQUIPMENT_SEED
from app.data.movements import MOVEMENTS_SEED
from app.data.wod_templates import WOD_TEMPLATES_SEED


def go_str(s):
    if s is None:
        return ""
    return '"' + str(s).replace('\\', '\\\\').replace('"', '\\"') + '"'


def gen_equipment():
    lines = []
    lines.append("package data")
    lines.append("")
    lines.append("// EquipmentSeed — каталог инвентаря (сгенерировано из app/data/equipment.py).")
    lines.append("var EquipmentSeed = []EquipmentItem{")
    for item in EQUIPMENT_SEED:
        lines.append(
            "\t{Key: %s, Name: %s, Category: %s, Icon: %s},"
            % (go_str(item["key"]), go_str(item["name"]), go_str(item["category"]), go_str(item["icon"]))
        )
    lines.append("}")
    return "\n".join(lines) + "\n"


def gen_movements():
    lines = []
    lines.append("package data")
    lines.append("")
    lines.append("// MovementsSeed — каталог движений (сгенерировано из app/data/movements.py).")
    lines.append("var MovementsSeed = []MovementItem{")
    for mv in MOVEMENTS_SEED:
        themes = ", ".join(go_str(t) for t in mv.get("themes", []))
        equip = ", ".join(go_str(k) for k in mv.get("equipment_keys", []))
        lines.append(
            "\t{Key: %s, Name: %s, Modality: %s, MuscleGroup: %s, Themes: []string{%s}, EquipmentKeys: []string{%s}, Difficulty: %s, ScalingBeginner: %s, ScalingIntermediate: %s},"
            % (
                go_str(mv["key"]), go_str(mv["name"]), go_str(mv["modality"]),
                go_str(mv["muscle_group"]), themes, equip, go_str(mv["difficulty"]),
                go_str(mv.get("scaling_beginner", "")), go_str(mv.get("scaling_intermediate", "")),
            )
        )
    lines.append("}")
    return "\n".join(lines) + "\n"


def gen_wod_templates():
    lines = []
    lines.append("package data")
    lines.append("")
    lines.append("// WodTemplatesSeed — шаблоны тренировок (сгенерировано из app/data/wod_templates.py).")
    lines.append("var WodTemplatesSeed = []WodTemplateItem{")
    for tpl in WOD_TEMPLATES_SEED:
        lines.append("\t{")
        lines.append("\t\tName: %s, Format: %s, DurationMin: %d, Intensity: %s, Theme: %s, IsBenchmark: %s," % (
            go_str(tpl["name"]), go_str(tpl["format"]), tpl["duration_min"],
            go_str(tpl["intensity"]), go_str(tpl["theme"]),
            "true" if tpl["is_benchmark"] else "false",
        ))
        desc = tpl.get("description")
        if desc:
            lines.append("\t\tDescription: %s," % go_str(desc))
        lines.append("\t\tMovements: []WodTemplateMovementItem{")
        for m in tpl.get("movements", []):
            parts = ["\t\t\t{MovementKey: %s, MovementName: %s" % (go_str(m["movement_key"]), go_str(m["movement_name"]))]
            if m.get("reps") is not None:
                parts.append("Reps: intPtr(%d)" % m["reps"])
            if m.get("weight_male") is not None:
                parts.append("WeightMale: intPtr(%d)" % m["weight_male"])
            if m.get("weight_female") is not None:
                parts.append("WeightFemale: intPtr(%d)" % m["weight_female"])
            parts.append("SortOrder: %d" % m.get("sort_order", 0))
            if m.get("rounds_note"):
                parts.append("RoundsNote: %s" % go_str(m["rounds_note"]))
            lines.append("\t\t\t" + ", ".join(parts) + "},")
        lines.append("\t\t},")
        lines.append("\t},")
    lines.append("}")
    return "\n".join(lines) + "\n"


def main():
    out_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "internal", "data")
    os.makedirs(out_dir, exist_ok=True)

    files = {
        "equipment_gen.go": gen_equipment(),
        "movements_gen.go": gen_movements(),
        "wod_templates_gen.go": gen_wod_templates(),
    }
    for name, content in files.items():
        path = os.path.join(out_dir, name)
        with open(path, "w", encoding="utf-8") as f:
            f.write(content)
        print(f"wrote {path} ({len(content.splitlines())} lines)")


if __name__ == "__main__":
    main()
