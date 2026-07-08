"""Каталог шаблонов тренировок для сид-данных.

Структура записи:
  name           — название
  format         — amrap | for_time | emom | tabata | chipper | ladder | death_by | strength
  duration_min   — плановая длительность (мин)
  intensity      — low | medium | high
  theme          — legs | arms_shoulders | clean_jerk | snatch | cardio_metcon | gymnastics | core | full_body
  is_benchmark   — True для классических бенчмарков CrossFit
  description    — описание на русском
  movements      — список движений:
    movement_key, movement_name, reps, weight_male, weight_female, sort_order, rounds_note
"""

WOD_TEMPLATES_SEED = [
    # ═══════════════════════════════════════════════════════════
    # КЛАССИЧЕСКИЕ БЕНЧМАРКИ (The Girls + Hero WODs)
    # ═══════════════════════════════════════════════════════════
    {
        "name": "Fran", "format": "for_time", "duration_min": 5, "intensity": "high",
        "theme": "full_body", "is_benchmark": True,
        "description": "21-15-9: трастер (43/30 кг) + подтягивания",
        "movements": [
            {"movement_key": "thruster",  "movement_name": "Трастер",       "reps": None, "weight_male": 43, "weight_female": 30, "sort_order": 0, "rounds_note": "21-15-9"},
            {"movement_key": "pull_up",   "movement_name": "Подтягивания",   "reps": None, "weight_male": None, "weight_female": None, "sort_order": 1, "rounds_note": "21-15-9"},
        ],
    },
    {
        "name": "Helen", "format": "for_time", "duration_min": 12, "intensity": "high",
        "theme": "cardio_metcon", "is_benchmark": True,
        "description": "3 раунда: бег 400 м + 21 мах гирей (24/16 кг) + 12 подтягиваний",
        "movements": [
            {"movement_key": "run",              "movement_name": "Бег 400 м",          "reps": 400,  "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "kettlebell_swing", "movement_name": "Махи гирей",          "reps": 21,   "weight_male": 24,  "weight_female": 16,  "sort_order": 1},
            {"movement_key": "pull_up",          "movement_name": "Подтягивания",        "reps": 12,   "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },
    {
        "name": "Cindy", "format": "amrap", "duration_min": 20, "intensity": "medium",
        "theme": "gymnastics", "is_benchmark": True,
        "description": "20 мин AMRAP: 5 подтягиваний + 10 отжиманий + 15 приседаний",
        "movements": [
            {"movement_key": "pull_up",    "movement_name": "Подтягивания",   "reps": 5,  "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "push_up",    "movement_name": "Отжимания",      "reps": 10, "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "air_squat",  "movement_name": "Приседания",     "reps": 15, "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },
    {
        "name": "Grace", "format": "for_time", "duration_min": 5, "intensity": "high",
        "theme": "clean_jerk", "is_benchmark": True,
        "description": "30 толчков (61/43 кг) на время",
        "movements": [
            {"movement_key": "clean_and_jerk", "movement_name": "Толчок (C&J)", "reps": 30, "weight_male": 61, "weight_female": 43, "sort_order": 0},
        ],
    },
    {
        "name": "Isabel", "format": "for_time", "duration_min": 4, "intensity": "high",
        "theme": "snatch", "is_benchmark": True,
        "description": "30 рывков (61/43 кг) на время",
        "movements": [
            {"movement_key": "snatch", "movement_name": "Рывок", "reps": 30, "weight_male": 61, "weight_female": 43, "sort_order": 0},
        ],
    },
    {
        "name": "Karen", "format": "for_time", "duration_min": 12, "intensity": "high",
        "theme": "legs", "is_benchmark": True,
        "description": "150 бросков медбола (9/6 кг) на время",
        "movements": [
            {"movement_key": "wall_ball", "movement_name": "Wall ball", "reps": 150, "weight_male": 9, "weight_female": 6, "sort_order": 0},
        ],
    },
    {
        "name": "Annie", "format": "for_time", "duration_min": 10, "intensity": "medium",
        "theme": "cardio_metcon", "is_benchmark": True,
        "description": "50-40-30-20-10: дабл-андеры + ситапы",
        "movements": [
            {"movement_key": "double_under", "movement_name": "Double-under", "reps": None, "weight_male": None, "weight_female": None, "sort_order": 0, "rounds_note": "50-40-30-20-10"},
            {"movement_key": "sit_up",        "movement_name": "Ситапы",       "reps": None, "weight_male": None, "weight_female": None, "sort_order": 1, "rounds_note": "50-40-30-20-10"},
        ],
    },
    {
        "name": "Murph", "format": "chipper", "duration_min": 45, "intensity": "high",
        "theme": "full_body", "is_benchmark": True,
        "description": "Бег 1 миля + 100 PU + 200 PU + 300 приседаний + бег 1 миля",
        "movements": [
            {"movement_key": "run",        "movement_name": "Бег 1 миля (1600 м)", "reps": 1600, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "pull_up",    "movement_name": "Подтягивания",         "reps": 100,  "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "push_up",    "movement_name": "Отжимания",            "reps": 200,  "weight_male": None, "weight_female": None, "sort_order": 2},
            {"movement_key": "air_squat",  "movement_name": "Приседания",           "reps": 300,  "weight_male": None, "weight_female": None, "sort_order": 3},
            {"movement_key": "run",        "movement_name": "Бег 1 миля (1600 м)", "reps": 1600, "weight_male": None, "weight_female": None, "sort_order": 4},
        ],
    },
    {
        "name": "DT", "format": "for_time", "duration_min": 12, "intensity": "high",
        "theme": "clean_jerk", "is_benchmark": True,
        "description": "5 раундов: 12 становая (70/45 кг) + 9 hang power clean + 6 push jerk",
        "movements": [
            {"movement_key": "deadlift",         "movement_name": "Становая тяга",     "reps": 12, "weight_male": 70, "weight_female": 45, "sort_order": 0},
            {"movement_key": "hang_power_clean", "movement_name": "Hang power clean",  "reps": 9,  "weight_male": 70, "weight_female": 45, "sort_order": 1},
            {"movement_key": "push_jerk",        "movement_name": "Push jerk",         "reps": 6,  "weight_male": 70, "weight_female": 45, "sort_order": 2},
        ],
    },
    {
        "name": "Elizabeth", "format": "for_time", "duration_min": 10, "intensity": "high",
        "theme": "clean_jerk", "is_benchmark": True,
        "description": "21-15-9: cleans (61/43 кг) + отжимания на кольцах",
        "movements": [
            {"movement_key": "squat_clean", "movement_name": "Squat clean",        "reps": None, "weight_male": 61, "weight_female": 43, "sort_order": 0, "rounds_note": "21-15-9"},
            {"movement_key": "ring_dip",    "movement_name": "Отжимания на кольцах","reps": None, "weight_male": None,"weight_female": None,"sort_order": 1, "rounds_note": "21-15-9"},
        ],
    },
    {
        "name": "Nancy", "format": "for_time", "duration_min": 15, "intensity": "medium",
        "theme": "legs", "is_benchmark": True,
        "description": "5 раундов: бег 400 м + 15 оверхед-приседаний (43/30 кг)",
        "movements": [
            {"movement_key": "run",             "movement_name": "Бег 400 м",       "reps": 400, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "overhead_squat",  "movement_name": "Оверхед-приседания","reps": 15,"weight_male": 43,  "weight_female": 30,  "sort_order": 1},
        ],
    },
    {
        "name": "Angie", "format": "for_time", "duration_min": 20, "intensity": "high",
        "theme": "gymnastics", "is_benchmark": True,
        "description": "100 PU + 100 PU + 100 ситапов + 100 приседаний",
        "movements": [
            {"movement_key": "pull_up",    "movement_name": "Подтягивания",   "reps": 100, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "push_up",    "movement_name": "Отжимания",      "reps": 100, "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "sit_up",     "movement_name": "Ситапы",         "reps": 100, "weight_male": None, "weight_female": None, "sort_order": 2},
            {"movement_key": "air_squat",  "movement_name": "Приседания",     "reps": 100, "weight_male": None, "weight_female": None, "sort_order": 3},
        ],
    },
    {
        "name": "Chelsea", "format": "emom", "duration_min": 30, "intensity": "medium",
        "theme": "gymnastics", "is_benchmark": True,
        "description": "30 мин EMOM: 5 PU + 10 отжиманий + 15 приседаний",
        "movements": [
            {"movement_key": "pull_up",    "movement_name": "Подтягивания",   "reps": 5,  "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "push_up",    "movement_name": "Отжимания",      "reps": 10, "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "air_squat",  "movement_name": "Приседания",     "reps": 15, "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },
    {
        "name": "Mary", "format": "amrap", "duration_min": 20, "intensity": "medium",
        "theme": "gymnastics", "is_benchmark": True,
        "description": "20 мин AMRAP: 5 HSPU + 10 пистолетов + 15 подтягиваний",
        "movements": [
            {"movement_key": "handstand_push_up", "movement_name": "HSPU",            "reps": 5,  "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "pistol_squat",      "movement_name": "Писталеты",       "reps": 10, "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "pull_up",           "movement_name": "Подтягивания",     "reps": 15, "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },
    {
        "name": "Eva", "format": "for_time", "duration_min": 45, "intensity": "high",
        "theme": "cardio_metcon", "is_benchmark": True,
        "description": "5 раундов: бег 800 м + 30 KB swing (24/16 кг) + 30 подтягиваний",
        "movements": [
            {"movement_key": "run",              "movement_name": "Бег 800 м",   "reps": 800, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "kettlebell_swing", "movement_name": "Махи гирей",   "reps": 30,  "weight_male": 24,  "weight_female": 16,  "sort_order": 1},
            {"movement_key": "pull_up",          "movement_name": "Подтягивания", "reps": 30,  "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },

    # ═══════════════════════════════════════════════════════════
    # НОГИ
    # ═══════════════════════════════════════════════════════════
    {
        "name": "Leg Burner", "format": "amrap", "duration_min": 12, "intensity": "high",
        "theme": "legs", "is_benchmark": False,
        "description": "12 мин AMRAP: 20 wall ball + 20 запрыгиваний на бокс + 20 махи гирей",
        "movements": [
            {"movement_key": "wall_ball",         "movement_name": "Wall ball",      "reps": 20, "weight_male": 9, "weight_female": 6, "sort_order": 0},
            {"movement_key": "box_jump",          "movement_name": "Запрыгивания",   "reps": 20, "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "kettlebell_swing",  "movement_name": "Махи гирей",     "reps": 20, "weight_male": 24, "weight_female": 16, "sort_order": 2},
        ],
    },
    {
        "name": "Squat Ladder", "format": "ladder", "duration_min": 15, "intensity": "medium",
        "theme": "legs", "is_benchmark": False,
        "description": "EMOM 10: чётная мин — 10 фронтальных приседаний, нечётная — 10 выпадов",
        "movements": [
            {"movement_key": "front_squat",   "movement_name": "Фронтальные приседания", "reps": 10, "weight_male": 43, "weight_female": 30, "sort_order": 0, "rounds_note": "Чётные минуты"},
            {"movement_key": "walking_lunge", "movement_name": "Выпады с шагом",          "reps": 10, "weight_male": 43, "weight_female": 30, "sort_order": 1, "rounds_note": "Нечётные минуты"},
        ],
    },
    {
        "name": "Wall Ball Hell", "format": "for_time", "duration_min": 10, "intensity": "high",
        "theme": "legs", "is_benchmark": False,
        "description": "100-80-60-40-20: wall ball + воздушные приседания",
        "movements": [
            {"movement_key": "wall_ball",   "movement_name": "Wall ball",   "reps": None, "weight_male": 9, "weight_female": 6, "sort_order": 0, "rounds_note": "100-80-60-40-20"},
            {"movement_key": "air_squat",   "movement_name": "Приседания",   "reps": None, "weight_male": None, "weight_female": None, "sort_order": 1, "rounds_note": "100-80-60-40-20"},
        ],
    },
    {
        "name": "Pistols & Box", "format": "for_time", "duration_min": 15, "intensity": "medium",
        "theme": "legs", "is_benchmark": False,
        "description": "5 раундов: 10 пистолетов + 15 запрыгиваний на бокс + бег 400 м",
        "movements": [
            {"movement_key": "pistol_squat", "movement_name": "Писталеты",       "reps": 10,  "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "box_jump",     "movement_name": "Запрыгивания",     "reps": 15,  "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "run",          "movement_name": "Бег 400 м",       "reps": 400, "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },
    {
        "name": "Deadlift Metcon", "format": "amrap", "duration_min": 10, "intensity": "high",
        "theme": "legs", "is_benchmark": False,
        "description": "10 мин AMRAP: 10 становая (70/50) + 20 бёрпи + 30 воздушных приседаний",
        "movements": [
            {"movement_key": "deadlift",    "movement_name": "Становая тяга",  "reps": 10, "weight_male": 70, "weight_female": 50, "sort_order": 0},
            {"movement_key": "burpee",      "movement_name": "Бёрпи",           "reps": 20, "weight_male": None,"weight_female": None,"sort_order": 1},
            {"movement_key": "air_squat",   "movement_name": "Приседания",      "reps": 30, "weight_male": None,"weight_female": None,"sort_order": 2},
        ],
    },

    # ═══════════════════════════════════════════════════════════
    # РУКИ / ПЛЕЧИ
    # ═══════════════════════════════════════════════════════════
    {
        "name": "Push Press Burner", "format": "amrap", "duration_min": 10, "intensity": "high",
        "theme": "arms_shoulders", "is_benchmark": False,
        "description": "10 мин AMRAP: 10 жимовой толчок (43/30) + 15 отжиманий + 200 м бег",
        "movements": [
            {"movement_key": "push_press", "movement_name": "Жимовой толчок", "reps": 10,  "weight_male": 43, "weight_female": 30, "sort_order": 0},
            {"movement_key": "push_up",    "movement_name": "Отжимания",      "reps": 15,  "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "run",        "movement_name": "Бег 200 м",     "reps": 200, "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },
    {
        "name": "HSPU Hell", "format": "emom", "duration_min": 12, "intensity": "high",
        "theme": "arms_shoulders", "is_benchmark": False,
        "description": "12 мин EMOM: 8 HSPU + 10 отжиманий на кольцах",
        "movements": [
            {"movement_key": "handstand_push_up", "movement_name": "HSPU",                 "reps": 8,  "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "ring_dip",          "movement_name": "Отжимания на кольцах", "reps": 10, "weight_male": None, "weight_female": None, "sort_order": 1},
        ],
    },
    {
        "name": "Shoulder Destroyer", "format": "for_time", "duration_min": 12, "intensity": "high",
        "theme": "arms_shoulders", "is_benchmark": False,
        "description": "5 раундов: 12 армейский жим (43/30) + 15 отжиманий + 12 жим гантелей",
        "movements": [
            {"movement_key": "shoulder_press", "movement_name": "Армейский жим",    "reps": 12, "weight_male": 43, "weight_female": 30, "sort_order": 0},
            {"movement_key": "push_up",         "movement_name": "Отжимания",        "reps": 15, "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "dumbbell_press",  "movement_name": "Жим гантелей",     "reps": 12, "weight_male": 15, "weight_female": 10, "sort_order": 2},
        ],
    },
    {
        "name": "Devil Press Sprint", "format": "for_time", "duration_min": 10, "intensity": "high",
        "theme": "arms_shoulders", "is_benchmark": False,
        "description": "5 раундов: 10 devil press (2×15/10 кг) + 15 бёрпи",
        "movements": [
            {"movement_key": "devil_press", "movement_name": "Devil press", "reps": 10, "weight_male": 15, "weight_female": 10, "sort_order": 0},
            {"movement_key": "burpee",      "movement_name": "Бёрпи",       "reps": 15, "weight_male": None, "weight_female": None, "sort_order": 1},
        ],
    },

    # ═══════════════════════════════════════════════════════════
    # ТОЛЧОК (CLEAN & JERK)
    # ═══════════════════════════════════════════════════════════
    {
        "name": "Clean Complex", "format": "emom", "duration_min": 12, "intensity": "medium",
        "theme": "clean_jerk", "is_benchmark": False,
        "description": "12 мин EMOM: 3 power clean + 1 push jerk (60/40 кг)",
        "movements": [
            {"movement_key": "power_clean", "movement_name": "Power clean",  "reps": 3, "weight_male": 60, "weight_female": 40, "sort_order": 0},
            {"movement_key": "push_jerk",   "movement_name": "Push jerk",    "reps": 1, "weight_male": 60, "weight_female": 40, "sort_order": 1},
        ],
    },
    {
        "name": "Clean Ladder", "format": "ladder", "duration_min": 15, "intensity": "high",
        "theme": "clean_jerk", "is_benchmark": False,
        "description": "Каждую минуту: 1 squat clean, добавляя вес. До отказа.",
        "movements": [
            {"movement_key": "squat_clean", "movement_name": "Squat clean", "reps": 1, "weight_male": None, "weight_female": None, "sort_order": 0, "rounds_note": "Рост веса каждую минуту"},
        ],
    },
    {
        "name": "Hang Power Metcon", "format": "amrap", "duration_min": 10, "intensity": "high",
        "theme": "clean_jerk", "is_benchmark": False,
        "description": "10 мин AMRAP: 10 hang power clean (50/35) + 10 толчков гантелей + 10 бёрпи",
        "movements": [
            {"movement_key": "hang_power_clean", "movement_name": "Hang power clean",   "reps": 10, "weight_male": 50, "weight_female": 35, "sort_order": 0},
            {"movement_key": "db_clean",          "movement_name": "Взятие гантелей",     "reps": 10, "weight_male": 15, "weight_female": 10, "sort_order": 1},
            {"movement_key": "burpee",            "movement_name": "Бёрпи",               "reps": 10, "weight_male": None,"weight_female": None,"sort_order": 2},
        ],
    },
    {
        "name": "Clean & Box", "format": "for_time", "duration_min": 12, "intensity": "medium",
        "theme": "clean_jerk", "is_benchmark": False,
        "description": "5 раундов: 10 power clean (50/35) + 15 запрыгиваний на бокс",
        "movements": [
            {"movement_key": "power_clean", "movement_name": "Power clean",    "reps": 10, "weight_male": 50, "weight_female": 35, "sort_order": 0},
            {"movement_key": "box_jump",    "movement_name": "Запрыгивания",    "reps": 15, "weight_male": None, "weight_female": None, "sort_order": 1},
        ],
    },

    # ═══════════════════════════════════════════════════════════
    # РЫВОК (SNATCH)
    # ═══════════════════════════════════════════════════════════
    {
        "name": "Snatch Skill", "format": "emom", "duration_min": 12, "intensity": "low",
        "theme": "snatch", "is_benchmark": False,
        "description": "12 мин EMOM: 2 power snatch + 2 OHS (40/25 кг)",
        "movements": [
            {"movement_key": "power_snatch",    "movement_name": "Power snatch",    "reps": 2, "weight_male": 40, "weight_female": 25, "sort_order": 0},
            {"movement_key": "overhead_squat",  "movement_name": "OHS",             "reps": 2, "weight_male": 40, "weight_female": 25, "sort_order": 1},
        ],
    },
    {
        "name": "Snatch Burner", "format": "amrap", "duration_min": 8, "intensity": "high",
        "theme": "snatch", "is_benchmark": False,
        "description": "8 мин AMRAP: 10 рывок гантели + 10 бёрпи + 200 м бег",
        "movements": [
            {"movement_key": "alt_db_snatch", "movement_name": "Рывок гантели", "reps": 10,  "weight_male": 20, "weight_female": 15, "sort_order": 0},
            {"movement_key": "burpee",        "movement_name": "Бёрпи",         "reps": 10,  "weight_male": None,"weight_female": None,"sort_order": 1},
            {"movement_key": "run",           "movement_name": "Бег 200 м",    "reps": 200, "weight_male": None,"weight_female": None,"sort_order": 2},
        ],
    },
    {
        "name": "Snatch Ladder", "format": "ladder", "duration_min": 15, "intensity": "medium",
        "theme": "snatch", "is_benchmark": False,
        "description": "Каждую минуту: 1 snatch balance + 1 snatch. Рост веса.",
        "movements": [
            {"movement_key": "snatch_balance", "movement_name": "Snatch balance", "reps": 1, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "snatch",         "movement_name": "Рывок",          "reps": 1, "weight_male": None, "weight_female": None, "sort_order": 1},
        ],
    },
    {
        "name": "Isabel Lite", "format": "for_time", "duration_min": 6, "intensity": "high",
        "theme": "snatch", "is_benchmark": False,
        "description": "21 рывок (43/30 кг) + 21 отжимания",
        "movements": [
            {"movement_key": "snatch",  "movement_name": "Рывок",    "reps": 21, "weight_male": 43, "weight_female": 30, "sort_order": 0},
            {"movement_key": "push_up", "movement_name": "Отжимания", "reps": 21, "weight_male": None,"weight_female": None,"sort_order": 1},
        ],
    },

    # ═══════════════════════════════════════════════════════════
    # КАРДИО / METCON
    # ═══════════════════════════════════════════════════════════
    {
        "name": "Cardio Blast", "format": "amrap", "duration_min": 15, "intensity": "high",
        "theme": "cardio_metcon", "is_benchmark": False,
        "description": "15 мин AMRAP: 500 м гребля + 15 бёрпи + 30 дабл-андеров",
        "movements": [
            {"movement_key": "cal_row",       "movement_name": "Гребля 500 м",      "reps": 500, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "burpee",        "movement_name": "Бёрпи",              "reps": 15,  "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "double_under",  "movement_name": "Double-under",       "reps": 30,  "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },
    {
        "name": "Rowing Hell", "format": "for_time", "duration_min": 20, "intensity": "high",
        "theme": "cardio_metcon", "is_benchmark": False,
        "description": "5 раундов: 500 м гребля + 25 бёрпи + 200 м бег",
        "movements": [
            {"movement_key": "cal_row",  "movement_name": "Гребля 500 м", "reps": 500, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "burpee",   "movement_name": "Бёрпи",         "reps": 25,  "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "run",      "movement_name": "Бег 200 м",    "reps": 200, "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },
    {
        "name": "Bike & Burpee", "format": "emom", "duration_min": 12, "intensity": "high",
        "theme": "cardio_metcon", "is_benchmark": False,
        "description": "12 мин EMOM: 10 кал AirBike + 10 бёрпи",
        "movements": [
            {"movement_key": "cal_bike", "movement_name": "AirBike 10 кал", "reps": 10, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "burpee",   "movement_name": "Бёрпи",           "reps": 10, "weight_male": None, "weight_female": None, "sort_order": 1},
        ],
    },
    {
        "name": "Death by Burpee", "format": "death_by", "duration_min": 20, "intensity": "high",
        "theme": "cardio_metcon", "is_benchmark": False,
        "description": "1-я минута: 1 бёрпи. 2-я: 2. И т.д. до отказа.",
        "movements": [
            {"movement_key": "burpee", "movement_name": "Бёрпи", "reps": 1, "weight_male": None, "weight_female": None, "sort_order": 0, "rounds_note": "+1 каждую минуту"},
        ],
    },
    {
        "name": "Tabata Mash", "format": "tabata", "duration_min": 16, "intensity": "high",
        "theme": "cardio_metcon", "is_benchmark": False,
        "description": "8 раундов (20с/10с): отжимания → приседания → бёрпи → махи гирей. 4 цикла.",
        "movements": [
            {"movement_key": "push_up",           "movement_name": "Отжимания",      "reps": None, "weight_male": None, "weight_female": None, "sort_order": 0, "rounds_note": "20с/10с × 8"},
            {"movement_key": "air_squat",         "movement_name": "Приседания",      "reps": None, "weight_male": None, "weight_female": None, "sort_order": 1, "rounds_note": "20с/10с × 8"},
            {"movement_key": "burpee",            "movement_name": "Бёрпи",           "reps": None, "weight_male": None, "weight_female": None, "sort_order": 2, "rounds_note": "20с/10с × 8"},
            {"movement_key": "kettlebell_swing",  "movement_name": "Махи гирей",      "reps": None, "weight_male": 16,  "weight_female": 12,  "sort_order": 3, "rounds_note": "20с/10с × 8"},
        ],
    },
    {
        "name": "Double Under Burner", "format": "for_time", "duration_min": 8, "intensity": "medium",
        "theme": "cardio_metcon", "is_benchmark": False,
        "description": "5 раундов: 50 дабл-андеров + 20 ситапов + 10 отжиманий",
        "movements": [
            {"movement_key": "double_under", "movement_name": "Double-under", "reps": 50, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "sit_up",       "movement_name": "Ситапы",        "reps": 20, "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "push_up",      "movement_name": "Отжимания",      "reps": 10, "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },

    # ═══════════════════════════════════════════════════════════
    # ГИМНАСТИКА
    # ═══════════════════════════════════════════════════════════
    {
        "name": "Pull-up Pyramid", "format": "ladder", "duration_min": 12, "intensity": "high",
        "theme": "gymnastics", "is_benchmark": False,
        "description": "1-2-3-4-5-6-7-8-9-10: подтягивания + отжимания. И обратно.",
        "movements": [
            {"movement_key": "pull_up", "movement_name": "Подтягивания", "reps": None, "weight_male": None, "weight_female": None, "sort_order": 0, "rounds_note": "1-2-3-4-5-6-7-8-9-10"},
            {"movement_key": "push_up", "movement_name": "Отжимания",     "reps": None, "weight_male": None, "weight_female": None, "sort_order": 1, "rounds_note": "1-2-3-4-5-6-7-8-9-10"},
        ],
    },
    {
        "name": "Muscle-up Skill", "format": "emom", "duration_min": 15, "intensity": "medium",
        "theme": "gymnastics", "is_benchmark": False,
        "description": "15 мин EMOM: 3 chest-to-bar + 3 отжимания на кольцах",
        "movements": [
            {"movement_key": "chest_to_bar", "movement_name": "Chest-to-bar",      "reps": 3, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "ring_dip",     "movement_name": "Отжимания на кольцах","reps": 3, "weight_male": None, "weight_female": None, "sort_order": 1},
        ],
    },
    {
        "name": "T2B Burner", "format": "amrap", "duration_min": 10, "intensity": "high",
        "theme": "gymnastics", "is_benchmark": False,
        "description": "10 мин AMRAP: 10 toes-to-bar + 15 отжиманий + 20 воздушных приседаний",
        "movements": [
            {"movement_key": "toes_to_bar", "movement_name": "Toes-to-bar",   "reps": 10, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "push_up",      "movement_name": "Отжимания",     "reps": 15, "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "air_squat",    "movement_name": "Приседания",    "reps": 20, "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },
    {
        "name": "Handstand Work", "format": "emom", "duration_min": 10, "intensity": "low",
        "theme": "gymnastics", "is_benchmark": False,
        "description": "10 мин EMOM: 30с стойка на руках у стены + 5 HSPU",
        "movements": [
            {"movement_key": "handstand_walk",     "movement_name": "Стойка на руках",     "reps": 1, "weight_male": None, "weight_female": None, "sort_order": 0, "rounds_note": "30 секунд"},
            {"movement_key": "handstand_push_up",  "movement_name": "HSPU",                "reps": 5, "weight_male": None, "weight_female": None, "sort_order": 1},
        ],
    },
    {
        "name": "Rope Climb Metcon", "format": "for_time", "duration_min": 12, "intensity": "medium",
        "theme": "gymnastics", "is_benchmark": False,
        "description": "5 раундов: 3 лазания по канату + 15 подтягиваний + 400 м бег",
        "movements": [
            {"movement_key": "rope_climb", "movement_name": "Лазание по канату", "reps": 3,   "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "pull_up",    "movement_name": "Подтягивания",       "reps": 15,  "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "run",        "movement_name": "Бег 400 м",         "reps": 400, "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },

    # ═══════════════════════════════════════════════════════════
    # КОР / ПРЕСС
    # ═══════════════════════════════════════════════════════════
    {
        "name": "Core Crusher", "format": "amrap", "duration_min": 10, "intensity": "medium",
        "theme": "core", "is_benchmark": False,
        "description": "10 мин AMRAP: 20 ситапов + 15 русские скручивания + 10 toes-to-bar + 30с планка",
        "movements": [
            {"movement_key": "sit_up",          "movement_name": "Ситапы",            "reps": 20, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "russian_twist",   "movement_name": "Русские скручивания","reps": 15, "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "toes_to_bar",     "movement_name": "Toes-to-bar",       "reps": 10, "weight_male": None, "weight_female": None, "sort_order": 2},
            {"movement_key": "plank",           "movement_name": "Планка 30с",         "reps": 1,  "weight_male": None, "weight_female": None, "sort_order": 3, "rounds_note": "30 секунд"},
        ],
    },
    {
        "name": "GHD Hell", "format": "for_time", "duration_min": 10, "intensity": "high",
        "theme": "core", "is_benchmark": False,
        "description": "5 раундов: 30 GHD ситапов + 15 разгибаний на GHD + 30с hollow hold",
        "movements": [
            {"movement_key": "ghd_sit_up",      "movement_name": "GHD ситапы",       "reps": 30, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "back_extension",  "movement_name": "Разгибания на GHD","reps": 15, "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "hollow_hold",     "movement_name": "Hollow hold",       "reps": 1,  "weight_male": None, "weight_female": None, "sort_order": 2, "rounds_note": "30 секунд"},
        ],
    },
    {
        "name": "Core Tabata", "format": "tabata", "duration_min": 8, "intensity": "medium",
        "theme": "core", "is_benchmark": False,
        "description": "8 раундов (20с/10с): ситапы → hollow hold → русские скручивания → superman",
        "movements": [
            {"movement_key": "sit_up",         "movement_name": "Ситапы",            "reps": None, "weight_male": None, "weight_female": None, "sort_order": 0, "rounds_note": "20с/10с × 8"},
            {"movement_key": "hollow_hold",    "movement_name": "Hollow hold",        "reps": None, "weight_male": None, "weight_female": None, "sort_order": 1, "rounds_note": "20с/10с × 8"},
            {"movement_key": "russian_twist",  "movement_name": "Русские скручивания","reps": None, "weight_male": None, "weight_female": None, "sort_order": 2, "rounds_note": "20с/10с × 8"},
            {"movement_key": "superman_hold",  "movement_name": "Superman",           "reps": None, "weight_male": None, "weight_female": None, "sort_order": 3, "rounds_note": "20с/10с × 8"},
        ],
    },
    {
        "name": "V-Up Burner", "format": "emom", "duration_min": 10, "intensity": "medium",
        "theme": "core", "is_benchmark": False,
        "description": "10 мин EMOM: 15 V-up + 10 подъём коленей в висе",
        "movements": [
            {"movement_key": "v_up",                  "movement_name": "V-up",                   "reps": 15, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "hanging_knee_raise",    "movement_name": "Подъём коленей в висе", "reps": 10, "weight_male": None, "weight_female": None, "sort_order": 1},
        ],
    },

    # ═══════════════════════════════════════════════════════════
    # ПОЛНОЕ ТЕЛО
    # ═══════════════════════════════════════════════════════════
    {
        "name": "Full Body Blitz", "format": "amrap", "duration_min": 12, "intensity": "high",
        "theme": "full_body", "is_benchmark": False,
        "description": "12 мин AMRAP: 10 трастер (43/30) + 10 бёрпи + 10 подтягиваний + 200 м бег",
        "movements": [
            {"movement_key": "thruster", "movement_name": "Трастер",       "reps": 10,  "weight_male": 43, "weight_female": 30, "sort_order": 0},
            {"movement_key": "burpee",   "movement_name": "Бёрпи",          "reps": 10,  "weight_male": None,"weight_female": None,"sort_order": 1},
            {"movement_key": "pull_up",  "movement_name": "Подтягивания",    "reps": 10,  "weight_male": None,"weight_female": None,"sort_order": 2},
            {"movement_key": "run",      "movement_name": "Бег 200 м",      "reps": 200, "weight_male": None,"weight_female": None,"sort_order": 3},
        ],
    },
    {
        "name": "Thruster & Pull", "format": "for_time", "duration_min": 15, "intensity": "high",
        "theme": "full_body", "is_benchmark": False,
        "description": "21-18-15-12-9-6-3: трастер (43/30) + chest-to-bar",
        "movements": [
            {"movement_key": "thruster",    "movement_name": "Трастер",       "reps": None, "weight_male": 43, "weight_female": 30, "sort_order": 0, "rounds_note": "21-18-15-12-9-6-3"},
            {"movement_key": "chest_to_bar","movement_name": "Chest-to-bar",   "reps": None, "weight_male": None,"weight_female": None,"sort_order": 1, "rounds_note": "21-18-15-12-9-6-3"},
        ],
    },
    {
        "name": "Dumbbell Complex", "format": "amrap", "duration_min": 10, "intensity": "high",
        "theme": "full_body", "is_benchmark": False,
        "description": "10 мин AMRAP: 50 дабл-андеров + 10 man maker (2×15/10 кг)",
        "movements": [
            {"movement_key": "double_under", "movement_name": "Double-under", "reps": 50, "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "man_maker",    "movement_name": "Man maker",     "reps": 10, "weight_male": 15,  "weight_female": 10,  "sort_order": 1},
        ],
    },
    {
        "name": "Chipper Madness", "format": "chipper", "duration_min": 25, "intensity": "high",
        "theme": "full_body", "is_benchmark": False,
        "description": "50 wall ball + 40 бёрпи + 30 трастер (43/30) + 20 pull-up + 10 clean (61/43)",
        "movements": [
            {"movement_key": "wall_ball",       "movement_name": "Wall ball",     "reps": 50, "weight_male": 9,  "weight_female": 6,  "sort_order": 0},
            {"movement_key": "burpee",          "movement_name": "Бёрпи",          "reps": 40, "weight_male": None,"weight_female": None,"sort_order": 1},
            {"movement_key": "thruster",        "movement_name": "Трастер",        "reps": 30, "weight_male": 43, "weight_female": 30, "sort_order": 2},
            {"movement_key": "pull_up",         "movement_name": "Подтягивания",    "reps": 20, "weight_male": None,"weight_female": None,"sort_order": 3},
            {"movement_key": "clean_and_jerk",  "movement_name": "Толчок (C&J)",   "reps": 10, "weight_male": 61, "weight_female": 43, "sort_order": 4},
        ],
    },
    {
        "name": "Fight Gone Bad", "format": "amrap", "duration_min": 17, "intensity": "high",
        "theme": "full_body", "is_benchmark": True,
        "description": "3 раунда по 5 мин: станция 1 мин. wall ball + SDHP + прыжки на бокс + push press + row. 1 мин отдых между раундами.",
        "movements": [
            {"movement_key": "wall_ball",        "movement_name": "Wall ball",          "reps": None, "weight_male": 9,  "weight_female": 6,  "sort_order": 0, "rounds_note": "1 мин на станцию"},
            {"movement_key": "sumo_deadlift",    "movement_name": "SDHP (становая сумо)","reps": None,"weight_male": 34, "weight_female": 25, "sort_order": 1, "rounds_note": "1 мин на станцию"},
            {"movement_key": "box_jump",         "movement_name": "Запрыгивания на бокс","reps": None,"weight_male": None,"weight_female": None,"sort_order": 2, "rounds_note": "1 мин на станцию"},
            {"movement_key": "push_press",       "movement_name": "Push press",         "reps": None, "weight_male": 34, "weight_female": 25, "sort_order": 3, "rounds_note": "1 мин на станцию"},
            {"movement_key": "cal_row",          "movement_name": "Гребля 14 кал",      "reps": 14,  "weight_male": None,"weight_female": None,"sort_order": 4, "rounds_note": "1 мин на станцию"},
        ],
    },
    {
        "name": "Sandbag Special", "format": "for_time", "duration_min": 15, "intensity": "medium",
        "theme": "full_body", "is_benchmark": False,
        "description": "5 раундов: 15 взятие мешка + 200 м бег с мешком + 15 бёрпи",
        "movements": [
            {"movement_key": "sandbag_clean", "movement_name": "Взятие мешка", "reps": 15,  "weight_male": None, "weight_female": None, "sort_order": 0},
            {"movement_key": "run",           "movement_name": "Бег с мешком 200 м", "reps": 200, "weight_male": None, "weight_female": None, "sort_order": 1},
            {"movement_key": "burpee",        "movement_name": "Бёрпи",          "reps": 15,  "weight_male": None, "weight_female": None, "sort_order": 2},
        ],
    },
    {
        "name": "KB Domination", "format": "amrap", "duration_min": 12, "intensity": "high",
        "theme": "full_body", "is_benchmark": False,
        "description": "12 мин AMRAP: 15 KB clean + 15 KB swing + 15 goblet squat",
        "movements": [
            {"movement_key": "kb_clean",        "movement_name": "Взятие гири",     "reps": 15, "weight_male": 24, "weight_female": 16, "sort_order": 0},
            {"movement_key": "kettlebell_swing","movement_name": "Махи гирей",       "reps": 15, "weight_male": 24, "weight_female": 16, "sort_order": 1},
            {"movement_key": "goblet_squat",   "movement_name": "Goblet squat",     "reps": 15, "weight_male": 24, "weight_female": 16, "sort_order": 2},
        ],
    },
]
