package data

// EquipmentItem — запись каталога инвентаря.
type EquipmentItem struct {
	Key      string
	Name     string
	Category string
	Icon     string
}

// MovementItem — запись каталога движений.
type MovementItem struct {
	Key                 string
	Name                string
	Modality            string
	MuscleGroup         string
	Themes              []string
	EquipmentKeys       []string
	Difficulty          string
	ScalingBeginner     string
	ScalingIntermediate string
}

// WodTemplateItem — шаблон тренировки.
type WodTemplateItem struct {
	Name        string
	Format      string
	DurationMin int
	Intensity   string
	Theme       string
	IsBenchmark bool
	Description string
	Movements   []WodTemplateMovementItem
}

// WodTemplateMovementItem — движение в шаблоне тренировки.
type WodTemplateMovementItem struct {
	MovementKey  string
	MovementName string
	Reps         *int
	WeightMale   *int
	WeightFemale *int
	SortOrder    int
	RoundsNote   string
}

func intPtr(v int) *int { return &v }

// SPtr возвращает указатель на строку (для nullable-полей в БД).
func SPtr(s string) *string { return &s }
