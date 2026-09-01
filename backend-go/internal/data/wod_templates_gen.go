package data

// WodTemplatesSeed — шаблоны тренировок.
//
// Принципы наполнения (расписание зала: 10 мин разминка + 5 мин инструктаж +
// 30–40 мин основная часть + 5–10 мин заминка):
//   - «длинная» основная часть: AMRAP/EMOM/лесенка 30 мин, chipper/for_time 20–25 мин;
//   - «пара» по 15 мин: короткие метконы со штангой и спринты;
//   - бенчмарки — классические, не меняются;
//   - дистанции (бег/гребля/велосипед) хранятся в названии движения,
//     reps всегда NULL — иначе множители уровня портят метраж (500 м × 0.75 = 375).
var WodTemplatesSeed = []WodTemplateItem{
	{
		Name: "Fran", Format: "for_time", DurationMin: 5, Intensity: "high", Theme: "full_body", IsBenchmark: true,
		Description: "21-15-9: трастер (43/30 кг) + подтягивания",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "thruster", MovementName: "Трастер", WeightMale: intPtr(43), WeightFemale: intPtr(30), SortOrder: 0, RoundsNote: "21-15-9"},
			{MovementKey: "pull_up", MovementName: "Подтягивания", SortOrder: 1, RoundsNote: "21-15-9"},
		},
	},
	{
		Name: "Helen", Format: "for_time", DurationMin: 12, Intensity: "high", Theme: "cardio_metcon", IsBenchmark: true,
		Description: "3 раунда: бег 400 м + 21 мах гирей (24/16 кг) + 12 подтягиваний",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "run", MovementName: "Бег 400 м", SortOrder: 0},
			{MovementKey: "kettlebell_swing", MovementName: "Махи гирей", Reps: intPtr(21), WeightMale: intPtr(24), WeightFemale: intPtr(16), SortOrder: 1},
			{MovementKey: "pull_up", MovementName: "Подтягивания", Reps: intPtr(12), SortOrder: 2},
		},
	},
	{
		Name: "Cindy", Format: "amrap", DurationMin: 20, Intensity: "medium", Theme: "gymnastics", IsBenchmark: true,
		Description: "20 мин AMRAP: 5 подтягиваний + 10 отжиманий + 15 приседаний",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "pull_up", MovementName: "Подтягивания", Reps: intPtr(5), SortOrder: 0},
			{MovementKey: "push_up", MovementName: "Отжимания", Reps: intPtr(10), SortOrder: 1},
			{MovementKey: "air_squat", MovementName: "Приседания", Reps: intPtr(15), SortOrder: 2},
		},
	},
	{
		Name: "Grace", Format: "for_time", DurationMin: 5, Intensity: "high", Theme: "clean_jerk", IsBenchmark: true,
		Description: "30 толчков (61/43 кг) на время",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "clean_and_jerk", MovementName: "Толчок (C&J)", Reps: intPtr(30), WeightMale: intPtr(61), WeightFemale: intPtr(43), SortOrder: 0},
		},
	},
	{
		Name: "Isabel", Format: "for_time", DurationMin: 4, Intensity: "high", Theme: "snatch", IsBenchmark: true,
		Description: "30 рывков (61/43 кг) на время",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "snatch", MovementName: "Рывок", Reps: intPtr(30), WeightMale: intPtr(61), WeightFemale: intPtr(43), SortOrder: 0},
		},
	},
	{
		Name: "Karen", Format: "for_time", DurationMin: 12, Intensity: "high", Theme: "legs", IsBenchmark: true,
		Description: "150 бросков медбола (9/6 кг) на время",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "wall_ball", MovementName: "Wall ball", Reps: intPtr(150), WeightMale: intPtr(9), WeightFemale: intPtr(6), SortOrder: 0},
		},
	},
	{
		Name: "Annie", Format: "for_time", DurationMin: 10, Intensity: "medium", Theme: "cardio_metcon", IsBenchmark: true,
		Description: "50-40-30-20-10: дабл-андеры + ситапы",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "double_under", MovementName: "Double-under", SortOrder: 0, RoundsNote: "50-40-30-20-10"},
			{MovementKey: "sit_up", MovementName: "Ситапы", SortOrder: 1, RoundsNote: "50-40-30-20-10"},
		},
	},
	{
		Name: "Murph", Format: "chipper", DurationMin: 45, Intensity: "high", Theme: "full_body", IsBenchmark: true,
		Description: "Бег 1 миля + 100 PU + 200 PU + 300 приседаний + бег 1 миля",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "run", MovementName: "Бег 1 миля (1600 м)", SortOrder: 0},
			{MovementKey: "pull_up", MovementName: "Подтягивания", Reps: intPtr(100), SortOrder: 1},
			{MovementKey: "push_up", MovementName: "Отжимания", Reps: intPtr(200), SortOrder: 2},
			{MovementKey: "air_squat", MovementName: "Приседания", Reps: intPtr(300), SortOrder: 3},
			{MovementKey: "run", MovementName: "Бег 1 миля (1600 м)", SortOrder: 4},
		},
	},
	{
		Name: "DT", Format: "for_time", DurationMin: 12, Intensity: "high", Theme: "clean_jerk", IsBenchmark: true,
		Description: "5 раундов: 12 становая (70/45 кг) + 9 hang power clean + 6 push jerk",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "deadlift", MovementName: "Становая тяга", Reps: intPtr(12), WeightMale: intPtr(70), WeightFemale: intPtr(45), SortOrder: 0},
			{MovementKey: "hang_power_clean", MovementName: "Hang power clean", Reps: intPtr(9), WeightMale: intPtr(70), WeightFemale: intPtr(45), SortOrder: 1},
			{MovementKey: "push_jerk", MovementName: "Push jerk", Reps: intPtr(6), WeightMale: intPtr(70), WeightFemale: intPtr(45), SortOrder: 2},
		},
	},
	{
		Name: "Elizabeth", Format: "for_time", DurationMin: 10, Intensity: "high", Theme: "clean_jerk", IsBenchmark: true,
		Description: "21-15-9: cleans (61/43 кг) + отжимания на кольцах",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "squat_clean", MovementName: "Squat clean", WeightMale: intPtr(61), WeightFemale: intPtr(43), SortOrder: 0, RoundsNote: "21-15-9"},
			{MovementKey: "ring_dip", MovementName: "Отжимания на кольцах", SortOrder: 1, RoundsNote: "21-15-9"},
		},
	},
	{
		Name: "Nancy", Format: "for_time", DurationMin: 15, Intensity: "medium", Theme: "legs", IsBenchmark: true,
		Description: "5 раундов: бег 400 м + 15 оверхед-приседаний (43/30 кг)",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "run", MovementName: "Бег 400 м", SortOrder: 0},
			{MovementKey: "overhead_squat", MovementName: "Оверхед-приседания", Reps: intPtr(15), WeightMale: intPtr(43), WeightFemale: intPtr(30), SortOrder: 1},
		},
	},
	{
		Name: "Angie", Format: "for_time", DurationMin: 20, Intensity: "high", Theme: "gymnastics", IsBenchmark: true,
		Description: "100 PU + 100 PU + 100 ситапов + 100 приседаний",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "pull_up", MovementName: "Подтягивания", Reps: intPtr(100), SortOrder: 0},
			{MovementKey: "push_up", MovementName: "Отжимания", Reps: intPtr(100), SortOrder: 1},
			{MovementKey: "sit_up", MovementName: "Ситапы", Reps: intPtr(100), SortOrder: 2},
			{MovementKey: "air_squat", MovementName: "Приседания", Reps: intPtr(100), SortOrder: 3},
		},
	},
	{
		Name: "Chelsea", Format: "emom", DurationMin: 30, Intensity: "medium", Theme: "gymnastics", IsBenchmark: true,
		Description: "30 мин EMOM: 5 PU + 10 отжиманий + 15 приседаний",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "pull_up", MovementName: "Подтягивания", Reps: intPtr(5), SortOrder: 0},
			{MovementKey: "push_up", MovementName: "Отжимания", Reps: intPtr(10), SortOrder: 1},
			{MovementKey: "air_squat", MovementName: "Приседания", Reps: intPtr(15), SortOrder: 2},
		},
	},
	{
		Name: "Mary", Format: "amrap", DurationMin: 20, Intensity: "medium", Theme: "gymnastics", IsBenchmark: true,
		Description: "20 мин AMRAP: 5 HSPU + 10 пистолетов + 15 подтягиваний",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "handstand_push_up", MovementName: "HSPU", Reps: intPtr(5), SortOrder: 0},
			{MovementKey: "pistol_squat", MovementName: "Писталеты", Reps: intPtr(10), SortOrder: 1},
			{MovementKey: "pull_up", MovementName: "Подтягивания", Reps: intPtr(15), SortOrder: 2},
		},
	},
	{
		Name: "Eva", Format: "for_time", DurationMin: 45, Intensity: "high", Theme: "cardio_metcon", IsBenchmark: true,
		Description: "5 раундов: бег 800 м + 30 KB swing (24/16 кг) + 30 подтягиваний",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "run", MovementName: "Бег 800 м", SortOrder: 0},
			{MovementKey: "kettlebell_swing", MovementName: "Махи гирей", Reps: intPtr(30), WeightMale: intPtr(24), WeightFemale: intPtr(16), SortOrder: 1},
			{MovementKey: "pull_up", MovementName: "Подтягивания", Reps: intPtr(30), SortOrder: 2},
		},
	},
	// ── Ноги ──
	{
		Name: "Leg Burner", Format: "amrap", DurationMin: 30, Intensity: "high", Theme: "legs", IsBenchmark: false,
		Description: "30 мин AMRAP: 20 wall ball (9/6) + 20 запрыгиваний на бокс + 20 махи гирей (24/16)",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "wall_ball", MovementName: "Wall ball", Reps: intPtr(20), WeightMale: intPtr(9), WeightFemale: intPtr(6), SortOrder: 0},
			{MovementKey: "box_jump", MovementName: "Запрыгивания", Reps: intPtr(20), SortOrder: 1},
			{MovementKey: "kettlebell_swing", MovementName: "Махи гирей", Reps: intPtr(20), WeightMale: intPtr(24), WeightFemale: intPtr(16), SortOrder: 2},
		},
	},
	{
		Name: "Squat Ladder", Format: "ladder", DurationMin: 30, Intensity: "medium", Theme: "legs", IsBenchmark: false,
		Description: "EMOM 30: чётная минута — 10 фронтальных приседаний (35/25), нечётная — 10 выпадов с шагом (35/25)",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "front_squat", MovementName: "Фронтальные приседания", Reps: intPtr(10), WeightMale: intPtr(35), WeightFemale: intPtr(25), SortOrder: 0, RoundsNote: "Чётные минуты"},
			{MovementKey: "walking_lunge", MovementName: "Выпады с шагом", Reps: intPtr(10), WeightMale: intPtr(35), WeightFemale: intPtr(25), SortOrder: 1, RoundsNote: "Нечётные минуты"},
		},
	},
	{
		Name: "Wall Ball Hell", Format: "for_time", DurationMin: 25, Intensity: "high", Theme: "legs", IsBenchmark: false,
		Description: "100-80-60-40-20: wall ball (9/6) + воздушные приседания",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "wall_ball", MovementName: "Wall ball", WeightMale: intPtr(9), WeightFemale: intPtr(6), SortOrder: 0, RoundsNote: "100-80-60-40-20"},
			{MovementKey: "air_squat", MovementName: "Приседания", SortOrder: 1, RoundsNote: "100-80-60-40-20"},
		},
	},
	{
		Name: "Pistols & Box", Format: "for_time", DurationMin: 20, Intensity: "medium", Theme: "legs", IsBenchmark: false,
		Description: "5 раундов: 10 пистолетов + 15 запрыгиваний на бокс + бег 400 м",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "pistol_squat", MovementName: "Писталеты", Reps: intPtr(10), SortOrder: 0},
			{MovementKey: "box_jump", MovementName: "Запрыгивания", Reps: intPtr(15), SortOrder: 1},
			{MovementKey: "run", MovementName: "Бег 400 м", SortOrder: 2},
		},
	},
	{
		Name: "Deadlift Metcon", Format: "amrap", DurationMin: 15, Intensity: "high", Theme: "legs", IsBenchmark: false,
		Description: "15 мин AMRAP: 8 становая (70/50) + 16 бёрпи + 24 воздушных приседаний",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "deadlift", MovementName: "Становая тяга", Reps: intPtr(8), WeightMale: intPtr(70), WeightFemale: intPtr(50), SortOrder: 0},
			{MovementKey: "burpee", MovementName: "Бёрпи", Reps: intPtr(16), SortOrder: 1},
			{MovementKey: "air_squat", MovementName: "Приседания", Reps: intPtr(24), SortOrder: 2},
		},
	},
	// ── Руки / Плечи ──
	{
		Name: "Push Press Burner", Format: "amrap", DurationMin: 15, Intensity: "high", Theme: "arms_shoulders", IsBenchmark: false,
		Description: "15 мин AMRAP: 10 жимовой толчок (43/30) + 15 отжиманий + 200 м бег",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "push_press", MovementName: "Жимовой толчок", Reps: intPtr(10), WeightMale: intPtr(43), WeightFemale: intPtr(30), SortOrder: 0},
			{MovementKey: "push_up", MovementName: "Отжимания", Reps: intPtr(15), SortOrder: 1},
			{MovementKey: "run", MovementName: "Бег 200 м", SortOrder: 2},
		},
	},
	{
		Name: "HSPU Hell", Format: "emom", DurationMin: 30, Intensity: "high", Theme: "arms_shoulders", IsBenchmark: false,
		Description: "EMOM 30, 3 блока: 6 HSPU / 8 отжиманий на кольцах / 45с планка",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "handstand_push_up", MovementName: "HSPU", Reps: intPtr(6), SortOrder: 0, RoundsNote: "Начиная с 1-й, каждая 3-я минута"},
			{MovementKey: "ring_dip", MovementName: "Отжимания на кольцах", Reps: intPtr(8), SortOrder: 1, RoundsNote: "Начиная со 2-й, каждая 3-я минута"},
			{MovementKey: "plank", MovementName: "Планка 45с", Reps: intPtr(1), SortOrder: 2, RoundsNote: "Начиная с 3-й, каждая 3-я минута"},
		},
	},
	{
		Name: "Shoulder Destroyer", Format: "for_time", DurationMin: 20, Intensity: "high", Theme: "arms_shoulders", IsBenchmark: false,
		Description: "5 раундов: 12 жимовой толчок (43/30) + 15 отжиманий + 12 жим гантелей (15/10)",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "push_press", MovementName: "Жимовой толчок", Reps: intPtr(12), WeightMale: intPtr(43), WeightFemale: intPtr(30), SortOrder: 0},
			{MovementKey: "push_up", MovementName: "Отжимания", Reps: intPtr(15), SortOrder: 1},
			{MovementKey: "dumbbell_press", MovementName: "Жим гантелей", Reps: intPtr(12), WeightMale: intPtr(15), WeightFemale: intPtr(10), SortOrder: 2},
		},
	},
	{
		Name: "Devil Press Sprint", Format: "for_time", DurationMin: 20, Intensity: "high", Theme: "arms_shoulders", IsBenchmark: false,
		Description: "5 раундов: 10 devil press (2×15/10 кг) + 15 бёрпи",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "devil_press", MovementName: "Devil press", Reps: intPtr(10), WeightMale: intPtr(15), WeightFemale: intPtr(10), SortOrder: 0},
			{MovementKey: "burpee", MovementName: "Бёрпи", Reps: intPtr(15), SortOrder: 1},
		},
	},
	// ── Толчок (C&J) ──
	{
		Name: "Clean Complex", Format: "emom", DurationMin: 12, Intensity: "medium", Theme: "clean_jerk", IsBenchmark: false,
		Description: "12 мин EMOM: 3 power clean + 1 push jerk (60/40 кг), комплекс без скидывания",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "power_clean", MovementName: "Power clean", Reps: intPtr(3), WeightMale: intPtr(60), WeightFemale: intPtr(40), SortOrder: 0, RoundsNote: "Комплекс без скидывания"},
			{MovementKey: "push_jerk", MovementName: "Push jerk", Reps: intPtr(1), WeightMale: intPtr(60), WeightFemale: intPtr(40), SortOrder: 1, RoundsNote: "Комплекс без скидывания"},
		},
	},
	{
		Name: "Clean Ladder", Format: "ladder", DurationMin: 15, Intensity: "high", Theme: "clean_jerk", IsBenchmark: false,
		Description: "Каждую минуту: 1 squat clean, добавляя вес. До отказа.",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "squat_clean", MovementName: "Squat clean", Reps: intPtr(1), SortOrder: 0, RoundsNote: "Рост веса каждую минуту"},
		},
	},
	{
		Name: "Hang Power Metcon", Format: "amrap", DurationMin: 15, Intensity: "high", Theme: "clean_jerk", IsBenchmark: false,
		Description: "15 мин AMRAP: 8 hang power clean (50/35) + 8 взятие гантелей (15/10) + 8 бёрпи",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "hang_power_clean", MovementName: "Hang power clean", Reps: intPtr(8), WeightMale: intPtr(50), WeightFemale: intPtr(35), SortOrder: 0},
			{MovementKey: "db_clean", MovementName: "Взятие гантелей", Reps: intPtr(8), WeightMale: intPtr(15), WeightFemale: intPtr(10), SortOrder: 1},
			{MovementKey: "burpee", MovementName: "Бёрпи", Reps: intPtr(8), SortOrder: 2},
		},
	},
	{
		Name: "Clean & Box", Format: "for_time", DurationMin: 25, Intensity: "medium", Theme: "clean_jerk", IsBenchmark: false,
		Description: "6 раундов: 8 power clean (50/35) + 12 запрыгиваний на бокс",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "power_clean", MovementName: "Power clean", Reps: intPtr(8), WeightMale: intPtr(50), WeightFemale: intPtr(35), SortOrder: 0},
			{MovementKey: "box_jump", MovementName: "Запрыгивания", Reps: intPtr(12), SortOrder: 1},
		},
	},
	// ── Рывок ──
	{
		Name: "Snatch Skill", Format: "emom", DurationMin: 15, Intensity: "low", Theme: "snatch", IsBenchmark: false,
		Description: "EMOM 15, 3 блока: 3 power snatch (40/25) / 3 OHS (40/25) / 10 good morning (20/15)",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "power_snatch", MovementName: "Power snatch", Reps: intPtr(3), WeightMale: intPtr(40), WeightFemale: intPtr(25), SortOrder: 0, RoundsNote: "Начиная с 1-й, каждая 3-я минута"},
			{MovementKey: "overhead_squat", MovementName: "OHS", Reps: intPtr(3), WeightMale: intPtr(40), WeightFemale: intPtr(25), SortOrder: 1, RoundsNote: "Начиная со 2-й, каждая 3-я минута"},
			{MovementKey: "good_morning", MovementName: "Good morning", Reps: intPtr(10), WeightMale: intPtr(20), WeightFemale: intPtr(15), SortOrder: 2, RoundsNote: "Начиная с 3-й, каждая 3-я минута"},
		},
	},
	{
		Name: "Snatch Burner", Format: "amrap", DurationMin: 20, Intensity: "high", Theme: "snatch", IsBenchmark: false,
		Description: "20 мин AMRAP: 10 рывок гантели (20/15) + 10 бёрпи + 200 м бег",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "alt_db_snatch", MovementName: "Рывок гантели", Reps: intPtr(10), WeightMale: intPtr(20), WeightFemale: intPtr(15), SortOrder: 0},
			{MovementKey: "burpee", MovementName: "Бёрпи", Reps: intPtr(10), SortOrder: 1},
			{MovementKey: "run", MovementName: "Бег 200 м", SortOrder: 2},
		},
	},
	{
		Name: "Snatch Ladder", Format: "ladder", DurationMin: 15, Intensity: "medium", Theme: "snatch", IsBenchmark: false,
		Description: "Каждую минуту: 1 snatch balance + 1 snatch. Рост веса.",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "snatch_balance", MovementName: "Snatch balance", Reps: intPtr(1), SortOrder: 0},
			{MovementKey: "snatch", MovementName: "Рывок", Reps: intPtr(1), SortOrder: 1},
		},
	},
	{
		Name: "Isabel Lite", Format: "for_time", DurationMin: 6, Intensity: "high", Theme: "snatch", IsBenchmark: false,
		Description: "21 рывок (43/30 кг) + 21 отжимания",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "snatch", MovementName: "Рывок", Reps: intPtr(21), WeightMale: intPtr(43), WeightFemale: intPtr(30), SortOrder: 0},
			{MovementKey: "push_up", MovementName: "Отжимания", Reps: intPtr(21), SortOrder: 1},
		},
	},
	// ── Кардио / Metcon ──
	{
		Name: "Cardio Blast", Format: "amrap", DurationMin: 30, Intensity: "high", Theme: "cardio_metcon", IsBenchmark: false,
		Description: "30 мин AMRAP: 500 м гребля + 15 бёрпи + 30 дабл-андеров (раунд ≈ 4 мин, ориентир 7–8 раундов)",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "cal_row", MovementName: "Гребля 500 м", SortOrder: 0},
			{MovementKey: "burpee", MovementName: "Бёрпи", Reps: intPtr(15), SortOrder: 1},
			{MovementKey: "double_under", MovementName: "Double-under", Reps: intPtr(30), SortOrder: 2},
		},
	},
	{
		Name: "Rowing Hell", Format: "for_time", DurationMin: 25, Intensity: "high", Theme: "cardio_metcon", IsBenchmark: false,
		Description: "5 раундов: 500 м гребля + 25 бёрпи + 200 м бег",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "cal_row", MovementName: "Гребля 500 м", SortOrder: 0},
			{MovementKey: "burpee", MovementName: "Бёрпи", Reps: intPtr(25), SortOrder: 1},
			{MovementKey: "run", MovementName: "Бег 200 м", SortOrder: 2},
		},
	},
	{
		Name: "Bike & Burpee", Format: "emom", DurationMin: 30, Intensity: "high", Theme: "cardio_metcon", IsBenchmark: false,
		Description: "EMOM 30, 3 блока: 10 кал AirBike / 6 бёрпи / 24 дабл-андера",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "cal_bike", MovementName: "AirBike 10 кал", SortOrder: 0, RoundsNote: "Начиная с 1-й, каждая 3-я минута"},
			{MovementKey: "burpee", MovementName: "Бёрпи", Reps: intPtr(6), SortOrder: 1, RoundsNote: "Начиная со 2-й, каждая 3-я минута"},
			{MovementKey: "double_under", MovementName: "Double-under", Reps: intPtr(24), SortOrder: 2, RoundsNote: "Начиная с 3-й, каждая 3-я минута"},
		},
	},
	{
		Name: "Death by Burpee", Format: "death_by", DurationMin: 20, Intensity: "high", Theme: "cardio_metcon", IsBenchmark: false,
		Description: "1-я минута: 1 бёрпи. 2-я: 2. И т.д. до отказа.",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "burpee", MovementName: "Бёрпи", Reps: intPtr(1), SortOrder: 0, RoundsNote: "+1 каждую минуту"},
		},
	},
	{
		Name: "Tabata Mash", Format: "tabata", DurationMin: 16, Intensity: "high", Theme: "cardio_metcon", IsBenchmark: false,
		Description: "8 раундов (20с/10с): отжимания → приседания → бёрпи → махи гирей. 4 цикла.",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "push_up", MovementName: "Отжимания", SortOrder: 0, RoundsNote: "20с/10с × 8"},
			{MovementKey: "air_squat", MovementName: "Приседания", SortOrder: 1, RoundsNote: "20с/10с × 8"},
			{MovementKey: "burpee", MovementName: "Бёрпи", SortOrder: 2, RoundsNote: "20с/10с × 8"},
			{MovementKey: "kettlebell_swing", MovementName: "Махи гирей", WeightMale: intPtr(16), WeightFemale: intPtr(12), SortOrder: 3, RoundsNote: "20с/10с × 8"},
		},
	},
	{
		Name: "Double Under Burner", Format: "for_time", DurationMin: 15, Intensity: "medium", Theme: "cardio_metcon", IsBenchmark: false,
		Description: "5 раундов: 50 дабл-андеров + 20 ситапов + 10 отжиманий",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "double_under", MovementName: "Double-under", Reps: intPtr(50), SortOrder: 0},
			{MovementKey: "sit_up", MovementName: "Ситапы", Reps: intPtr(20), SortOrder: 1},
			{MovementKey: "push_up", MovementName: "Отжимания", Reps: intPtr(10), SortOrder: 2},
		},
	},
	// ── Гимнастика ──
	{
		Name: "Pull-up Pyramid", Format: "ladder", DurationMin: 20, Intensity: "high", Theme: "gymnastics", IsBenchmark: false,
		Description: "1-2-3-4-5-6-7-8-9-10: подтягивания + отжимания. И обратно.",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "pull_up", MovementName: "Подтягивания", SortOrder: 0, RoundsNote: "1-2-3-4-5-6-7-8-9-10"},
			{MovementKey: "push_up", MovementName: "Отжимания", SortOrder: 1, RoundsNote: "1-2-3-4-5-6-7-8-9-10"},
		},
	},
	{
		Name: "Muscle-up Skill", Format: "emom", DurationMin: 15, Intensity: "medium", Theme: "gymnastics", IsBenchmark: false,
		Description: "EMOM 15, 3 блока: 5 chest-to-bar / 8 отжиманий на кольцах / 30с hollow hold",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "chest_to_bar", MovementName: "Chest-to-bar", Reps: intPtr(5), SortOrder: 0, RoundsNote: "Начиная с 1-й, каждая 3-я минута"},
			{MovementKey: "ring_dip", MovementName: "Отжимания на кольцах", Reps: intPtr(8), SortOrder: 1, RoundsNote: "Начиная со 2-й, каждая 3-я минута"},
			{MovementKey: "hollow_hold", MovementName: "Hollow hold 30с", Reps: intPtr(1), SortOrder: 2, RoundsNote: "Начиная с 3-й, каждая 3-я минута"},
		},
	},
	{
		Name: "T2B Burner", Format: "amrap", DurationMin: 20, Intensity: "high", Theme: "gymnastics", IsBenchmark: false,
		Description: "20 мин AMRAP: 10 toes-to-bar + 15 отжиманий + 20 воздушных приседаний",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "toes_to_bar", MovementName: "Toes-to-bar", Reps: intPtr(10), SortOrder: 0},
			{MovementKey: "push_up", MovementName: "Отжимания", Reps: intPtr(15), SortOrder: 1},
			{MovementKey: "air_squat", MovementName: "Приседания", Reps: intPtr(20), SortOrder: 2},
		},
	},
	{
		Name: "Handstand Work", Format: "emom", DurationMin: 15, Intensity: "low", Theme: "gymnastics", IsBenchmark: false,
		Description: "EMOM 15, 3 блока: 30с стойка на руках у стены / 5 HSPU / 10 отжиманий",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "handstand_walk", MovementName: "Стойка на руках 30с", SortOrder: 0, RoundsNote: "Начиная с 1-й, каждая 3-я минута"},
			{MovementKey: "handstand_push_up", MovementName: "HSPU", Reps: intPtr(5), SortOrder: 1, RoundsNote: "Начиная со 2-й, каждая 3-я минута"},
			{MovementKey: "push_up", MovementName: "Отжимания", Reps: intPtr(10), SortOrder: 2, RoundsNote: "Начиная с 3-й, каждая 3-я минута"},
		},
	},
	{
		Name: "Rope Climb Metcon", Format: "for_time", DurationMin: 25, Intensity: "medium", Theme: "gymnastics", IsBenchmark: false,
		Description: "5 раундов: 3 лазания по канату + 15 подтягиваний + 400 м бег",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "rope_climb", MovementName: "Лазание по канату", Reps: intPtr(3), SortOrder: 0},
			{MovementKey: "pull_up", MovementName: "Подтягивания", Reps: intPtr(15), SortOrder: 1},
			{MovementKey: "run", MovementName: "Бег 400 м", SortOrder: 2},
		},
	},
	// ── Кор / Пресс ──
	{
		Name: "Core Crusher", Format: "amrap", DurationMin: 20, Intensity: "medium", Theme: "core", IsBenchmark: false,
		Description: "20 мин AMRAP: 20 ситапов + 15 русские скручивания + 10 toes-to-bar + 30с планка",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "sit_up", MovementName: "Ситапы", Reps: intPtr(20), SortOrder: 0},
			{MovementKey: "russian_twist", MovementName: "Русские скручивания", Reps: intPtr(15), SortOrder: 1},
			{MovementKey: "toes_to_bar", MovementName: "Toes-to-bar", Reps: intPtr(10), SortOrder: 2},
			{MovementKey: "plank", MovementName: "Планка 30с", Reps: intPtr(1), SortOrder: 3, RoundsNote: "30 секунд"},
		},
	},
	{
		Name: "GHD Hell", Format: "for_time", DurationMin: 20, Intensity: "high", Theme: "core", IsBenchmark: false,
		Description: "5 раундов: 20 GHD ситапов + 15 разгибаний на GHD + 30с hollow hold",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "ghd_sit_up", MovementName: "GHD ситапы", Reps: intPtr(20), SortOrder: 0},
			{MovementKey: "back_extension", MovementName: "Разгибания на GHD", Reps: intPtr(15), SortOrder: 1},
			{MovementKey: "hollow_hold", MovementName: "Hollow hold", Reps: intPtr(1), SortOrder: 2, RoundsNote: "30 секунд"},
		},
	},
	{
		Name: "Core Tabata", Format: "tabata", DurationMin: 8, Intensity: "medium", Theme: "core", IsBenchmark: false,
		Description: "8 раундов (20с/10с): ситапы → hollow hold → русские скручивания → superman",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "sit_up", MovementName: "Ситапы", SortOrder: 0, RoundsNote: "20с/10с × 8"},
			{MovementKey: "hollow_hold", MovementName: "Hollow hold", SortOrder: 1, RoundsNote: "20с/10с × 8"},
			{MovementKey: "russian_twist", MovementName: "Русские скручивания", SortOrder: 2, RoundsNote: "20с/10с × 8"},
			{MovementKey: "superman_hold", MovementName: "Superman", SortOrder: 3, RoundsNote: "20с/10с × 8"},
		},
	},
	{
		Name: "V-Up Burner", Format: "emom", DurationMin: 15, Intensity: "medium", Theme: "core", IsBenchmark: false,
		Description: "EMOM 15, 3 блока: 15 V-up / 10 подъём коленей в висе / 20 русские скручивания",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "v_up", MovementName: "V-up", Reps: intPtr(15), SortOrder: 0, RoundsNote: "Начиная с 1-й, каждая 3-я минута"},
			{MovementKey: "hanging_knee_raise", MovementName: "Подъём коленей в висе", Reps: intPtr(10), SortOrder: 1, RoundsNote: "Начиная со 2-й, каждая 3-я минута"},
			{MovementKey: "russian_twist", MovementName: "Русские скручивания", Reps: intPtr(20), SortOrder: 2, RoundsNote: "Начиная с 3-й, каждая 3-я минута"},
		},
	},
	// ── Полное тело ──
	{
		Name: "Full Body Blitz", Format: "amrap", DurationMin: 30, Intensity: "high", Theme: "full_body", IsBenchmark: false,
		Description: "30 мин AMRAP: 10 wall ball (9/6) + 10 отжиманий + 10 подтягиваний + 200 м бег",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "wall_ball", MovementName: "Wall ball", Reps: intPtr(10), WeightMale: intPtr(9), WeightFemale: intPtr(6), SortOrder: 0},
			{MovementKey: "push_up", MovementName: "Отжимания", Reps: intPtr(10), SortOrder: 1},
			{MovementKey: "pull_up", MovementName: "Подтягивания", Reps: intPtr(10), SortOrder: 2},
			{MovementKey: "run", MovementName: "Бег 200 м", SortOrder: 3},
		},
	},
	{
		Name: "Thruster & Pull", Format: "for_time", DurationMin: 25, Intensity: "high", Theme: "full_body", IsBenchmark: false,
		Description: "21-18-15-12-9-6-3: трастер (43/30) + chest-to-bar",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "thruster", MovementName: "Трастер", WeightMale: intPtr(43), WeightFemale: intPtr(30), SortOrder: 0, RoundsNote: "21-18-15-12-9-6-3"},
			{MovementKey: "chest_to_bar", MovementName: "Chest-to-bar", SortOrder: 1, RoundsNote: "21-18-15-12-9-6-3"},
		},
	},
	{
		Name: "Dumbbell Complex", Format: "amrap", DurationMin: 15, Intensity: "high", Theme: "full_body", IsBenchmark: false,
		Description: "15 мин AMRAP: 40 дабл-андеров + 8 man maker (2×15/10 кг)",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "double_under", MovementName: "Double-under", Reps: intPtr(40), SortOrder: 0},
			{MovementKey: "man_maker", MovementName: "Man maker", Reps: intPtr(8), WeightMale: intPtr(15), WeightFemale: intPtr(10), SortOrder: 1},
		},
	},
	{
		Name: "Chipper Madness", Format: "chipper", DurationMin: 25, Intensity: "high", Theme: "full_body", IsBenchmark: false,
		Description: "50 wall ball + 40 бёрпи + 30 трастер (43/30) + 20 pull-up + 10 clean (61/43)",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "wall_ball", MovementName: "Wall ball", Reps: intPtr(50), WeightMale: intPtr(9), WeightFemale: intPtr(6), SortOrder: 0},
			{MovementKey: "burpee", MovementName: "Бёрпи", Reps: intPtr(40), SortOrder: 1},
			{MovementKey: "thruster", MovementName: "Трастер", Reps: intPtr(30), WeightMale: intPtr(43), WeightFemale: intPtr(30), SortOrder: 2},
			{MovementKey: "pull_up", MovementName: "Подтягивания", Reps: intPtr(20), SortOrder: 3},
			{MovementKey: "clean_and_jerk", MovementName: "Толчок (C&J)", Reps: intPtr(10), WeightMale: intPtr(61), WeightFemale: intPtr(43), SortOrder: 4},
		},
	},
	{
		Name: "Fight Gone Bad", Format: "amrap", DurationMin: 17, Intensity: "high", Theme: "full_body", IsBenchmark: true,
		Description: "3 раунда по 5 мин: станция 1 мин. wall ball + SDHP + прыжки на бокс + push press + row. 1 мин отдых между раундами.",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "wall_ball", MovementName: "Wall ball", WeightMale: intPtr(9), WeightFemale: intPtr(6), SortOrder: 0, RoundsNote: "1 мин на станцию"},
			{MovementKey: "sumo_deadlift", MovementName: "SDHP (становая сумо)", WeightMale: intPtr(34), WeightFemale: intPtr(25), SortOrder: 1, RoundsNote: "1 мин на станцию"},
			{MovementKey: "box_jump", MovementName: "Запрыгивания на бокс", SortOrder: 2, RoundsNote: "1 мин на станцию"},
			{MovementKey: "push_press", MovementName: "Push press", WeightMale: intPtr(34), WeightFemale: intPtr(25), SortOrder: 3, RoundsNote: "1 мин на станцию"},
			{MovementKey: "cal_row", MovementName: "Гребля 14 кал", SortOrder: 4, RoundsNote: "1 мин на станцию"},
		},
	},
	{
		Name: "Sandbag Special", Format: "for_time", DurationMin: 25, Intensity: "medium", Theme: "full_body", IsBenchmark: false,
		Description: "5 раундов: 15 взятие мешка + 200 м бег с мешком + 15 бёрпи",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "sandbag_clean", MovementName: "Взятие мешка", Reps: intPtr(15), SortOrder: 0},
			{MovementKey: "run", MovementName: "Бег с мешком 200 м", SortOrder: 1},
			{MovementKey: "burpee", MovementName: "Бёрпи", Reps: intPtr(15), SortOrder: 2},
		},
	},
	{
		Name: "KB Domination", Format: "amrap", DurationMin: 20, Intensity: "high", Theme: "full_body", IsBenchmark: false,
		Description: "20 мин AMRAP: 12 KB clean (24/16) + 12 KB swing + 12 goblet squat",
		Movements: []WodTemplateMovementItem{
			{MovementKey: "kb_clean", MovementName: "Взятие гири", Reps: intPtr(12), WeightMale: intPtr(24), WeightFemale: intPtr(16), SortOrder: 0},
			{MovementKey: "kettlebell_swing", MovementName: "Махи гирей", Reps: intPtr(12), WeightMale: intPtr(24), WeightFemale: intPtr(16), SortOrder: 1},
			{MovementKey: "goblet_squat", MovementName: "Goblet squat", Reps: intPtr(12), WeightMale: intPtr(24), WeightFemale: intPtr(16), SortOrder: 2},
		},
	},
}
