package data

// EquipmentSeed — каталог инвентаря (сгенерировано из app/data/equipment.py).
// Иконки не используются в интерфейсе — поле оставлено для совместимости схемы.
var EquipmentSeed = []EquipmentItem{
	// Штанги
	{Key: "barbell", Name: "Гриф для штанги", Category: "barbells", Icon: ""},
	{Key: "plate", Name: "Диски", Category: "barbells", Icon: ""},
	{Key: "bench", Name: "Скамья", Category: "barbells", Icon: ""},
	// Турники
	{Key: "pull_up_bar", Name: "Турник", Category: "rig", Icon: ""},
	{Key: "rings", Name: "Кольцо гимнастическое", Category: "rig", Icon: ""},
	{Key: "dip_bars", Name: "Брусья", Category: "rig", Icon: ""},
	{Key: "monkey_bars", Name: "Рукоход", Category: "rig", Icon: ""},
	{Key: "ropes", Name: "Канат (горизонтальный)", Category: "rig", Icon: ""},
	{Key: "rope_climb", Name: "Канат для лазания", Category: "rig", Icon: ""},
	// Свободные веса
	{Key: "dumbbell", Name: "Гантеля", Category: "freeweights", Icon: ""},
	{Key: "kettlebell", Name: "Гиря", Category: "freeweights", Icon: ""},
	{Key: "medball", Name: "Мяч (Медбол)", Category: "freeweights", Icon: ""},
	{Key: "slam_ball", Name: "Слэмбол", Category: "freeweights", Icon: ""},
	{Key: "sandbag", Name: "Мешок (sandbag)", Category: "freeweights", Icon: ""},
	// Кардио
	{Key: "rower", Name: "Гребной тренажёр (C2)", Category: "cardio", Icon: ""},
	{Key: "air_bike", Name: "Велоэргометр (AirBike)", Category: "cardio", Icon: ""},
	{Key: "ski_erg", Name: "Ски-эрг (SkiErg)", Category: "cardio", Icon: ""},
	// Прочее
	{Key: "box", Name: "Коробка", Category: "misc", Icon: ""},
	{Key: "jump_rope", Name: "Скакалка", Category: "misc", Icon: ""},
	{Key: "ghd", Name: "GHD (Glute-Ham Developer)", Category: "misc", Icon: ""},
	{Key: "ab_mat", Name: "AbMat", Category: "misc", Icon: ""},
	{Key: "sled", Name: "Сани", Category: "misc", Icon: ""},
	{Key: "sledgehammer", Name: "Кувалда", Category: "misc", Icon: ""},
	{Key: "bosu", Name: "Полусфера Bosu", Category: "misc", Icon: ""},
	{Key: "parallettes", Name: "Паралетсы", Category: "misc", Icon: ""},
}
