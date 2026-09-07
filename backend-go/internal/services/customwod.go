// Package services — конструктор тренировок с нуля (Экран 4 draft1.MD).
package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/maxdukov/cf/backend-go/internal/db"
)

// CustomMovementInput — строка упражнения конструктора.
type CustomMovementInput struct {
	MovementKey  string `json:"movement_key"`
	Reps         *int   `json:"reps"`
	WeightMale   *int   `json:"weight_male"`
	WeightFemale *int   `json:"weight_female"`
	RoundsNote   string `json:"rounds_note"`
}

// CustomWodInput — запрос POST /api/wods/custom.
type CustomWodInput struct {
	Name        string                `json:"name"`
	Format      string                `json:"format"`
	DurationMin int                   `json:"duration_min"`
	Intensity   string                `json:"intensity"`
	Theme       string                `json:"theme"`
	Description string                `json:"description"`
	Movements   []CustomMovementInput `json:"movements"`
}

var validFormats = map[string]bool{
	"amrap": true, "for_time": true, "emom": true, "tabata": true,
	"chipper": true, "ladder": true, "death_by": true, "strength": true,
}

var validIntensities = map[string]bool{"low": true, "medium": true, "high": true}

// CreateCustomWod сохраняет тренировку из конструктора как шаблон
// (wod_templates, is_benchmark=0). Возвращает template_id + предупреждения.
func CreateCustomWod(d *sql.DB, in CustomWodInput) (string, []string, error) {
	warnings, err := validateCustomWod(d, &in)
	if err != nil {
		return "", nil, err
	}

	// Автоимя: протокол + фокус + дата.
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = fmt.Sprintf("%s · %s · %s",
			formatNames[in.Format], ThemeName(in.Theme), time.Now().Format("02.01"))
	}

	templateID := db.NewUUID()
	tx, err := d.Begin()
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var desc any
	if strings.TrimSpace(in.Description) != "" {
		desc = in.Description
	}
	if _, err := tx.Exec(`
		INSERT INTO wod_templates (id, name, format, duration_min, intensity, theme, is_benchmark, description, is_custom)
		VALUES (?, ?, ?, ?, ?, ?, 0, ?, 1)`,
		templateID, name, in.Format, in.DurationMin, in.Intensity, in.Theme, desc,
	); err != nil {
		return "", nil, err
	}

	for i, m := range in.Movements {
		var mname string
		if err := tx.QueryRow(`SELECT name FROM movements WHERE "key" = ?`, m.MovementKey).Scan(&mname); err != nil {
			return "", nil, fmt.Errorf("движение «%s» не найдено", m.MovementKey)
		}
		var reps, wm, wf, rn any
		if m.Reps != nil {
			reps = *m.Reps
		}
		if m.WeightMale != nil {
			wm = *m.WeightMale
		}
		if m.WeightFemale != nil {
			wf = *m.WeightFemale
		}
		if m.RoundsNote != "" {
			rn = m.RoundsNote
		}
		if _, err := tx.Exec(`
			INSERT INTO wod_template_movements (id, template_id, movement_key, movement_name, reps, weight_male, weight_female, sort_order, rounds_note)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			db.NewUUID(), templateID, m.MovementKey, mname, reps, wm, wf, i+1, rn,
		); err != nil {
			return "", nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", nil, err
	}
	if len(warnings) == 0 {
		warnings = nil
	}
	return templateID, warnings, nil
}

// validateCustomWod — общие проверки конструктора (создание и правка).
func validateCustomWod(d *sql.DB, in *CustomWodInput) ([]string, error) {
	if !validFormats[in.Format] {
		return nil, fmt.Errorf("неизвестный формат: %s", in.Format)
	}
	if !validIntensities[in.Intensity] {
		in.Intensity = "medium"
	}
	if _, ok := themeNames[in.Theme]; !ok {
		return nil, fmt.Errorf("неизвестная тема: %s", in.Theme)
	}
	if in.DurationMin <= 0 || in.DurationMin > 120 {
		return nil, fmt.Errorf("длительность — от 1 до 120 минут")
	}
	if len(in.Movements) == 0 {
		return nil, fmt.Errorf("добавьте хотя бы одно упражнение")
	}
	if len(in.Movements) > 15 {
		return nil, fmt.Errorf("слишком много упражнений (максимум 15)")
	}

	// Проверка движений + сбор предупреждений (Экран 4).
	var warnings []string
	eqCount := map[string]int{} // требуемое количество снарядов по типам
	patterns := map[string]int{}
	for i := range in.Movements {
		m := &in.Movements[i]
		var name, modality, muscle, eqKeys sql.NullString
		err := d.QueryRow(`
			SELECT name, modality, muscle_group, equipment_keys FROM movements WHERE "key" = ?`, m.MovementKey,
		).Scan(&name, &modality, &muscle, &eqKeys)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("движение «%s» не найдено в каталоге", m.MovementKey)
		}
		if err != nil {
			return nil, err
		}
		if eqKeys.Valid && strings.TrimSpace(eqKeys.String) != "" {
			for _, k := range strings.Split(eqKeys.String, ",") {
				k = strings.TrimSpace(k)
				if k != "" {
					eqCount[k]++
				}
			}
		}
		if muscle.Valid {
			p := patternOf(muscle.String)
			if p != "" {
				patterns[p]++
			}
		}
	}

	// Проверка оборудования: наличие в инвентаре + достаточность количества.
	{
		rows, err := d.Query(`SELECT equipment_key, quantity FROM gym_inventory`)
		if err == nil {
			inv := map[string]int{}
			for rows.Next() {
				var k string
				var q int
				if rows.Scan(&k, &q) == nil {
					inv[k] = q
				}
			}
			_ = rows.Close()
			if len(inv) > 0 {
				for k, need := range eqCount {
					if have, ok := inv[k]; !ok {
						warnings = append(warnings, "Снаряд «"+k+"» отсутствует в инвентаре зала")
					} else if need > have {
						warnings = append(warnings, fmt.Sprintf(
							"Снарядов «%s» нужно %d, в зале %d", k, need, have))
					}
				}
			}
		}
	}

	// Баланс паттернов: тяга/толчок/ноги/кор.
	total := 0
	for _, c := range patterns {
		total += c
	}
	if total > 0 {
		for _, p := range []string{"ноги", "тяга", "толчок", "кор"} {
			if patterns[p] == 0 {
				warnings = append(warnings, "Паттерн «"+p+"» не задействован")
			} else if patterns[p]*100/total > 60 {
				warnings = append(warnings, "Паттерн «"+p+"» перегружен ("+itoaLocal(patterns[p]*100/total)+"% движений)")
			}
		}
	}

	// Подсказка по стимулу.
	if in.Format == "strength" && in.Intensity == "high" {
		warnings = append(warnings, "Силовая работа с высокой интенсивностью — проверьте план недели")
	}
	if in.DurationMin >= 25 && in.Intensity == "high" {
		warnings = append(warnings, "Высокоинтенсивная работа дольше 25 минут — это гликолитический стимул, не аэробный")
	}
	// EMOM-ротация: 3 блока упражнений (по минуте на блок).
	if in.Format == "emom" && len(in.Movements) != 3 {
		warnings = append(warnings, fmt.Sprintf(
			"EMOM: рекомендуется 3 блока упражнений (ротация по минутам), сейчас блоков %d", len(in.Movements)))
	}
	// Основная часть занятия: 30–40 мин (длинная тренировка) или пара по ~15 мин.
	if in.Format == "amrap" || in.Format == "for_time" || in.Format == "chipper" || in.Format == "ladder" {
		if in.DurationMin >= 18 && in.DurationMin < 28 {
			warnings = append(warnings, "Основная часть занятия — 30–40 мин: удлините тренировку либо запланируйте пару по ~15 мин")
		}
	}
	return warnings, nil
}

// UpdateCustomWod — правка существующего шаблона из библиотеки.
// Шаблон помечается is_custom = 1: seed-refresh больше не перетирает правки.
func UpdateCustomWod(d *sql.DB, templateID string, in CustomWodInput) ([]string, error) {
	var id string
	if err := d.QueryRow(`SELECT id FROM wod_templates WHERE id = ?`, templateID).Scan(&id); err == sql.ErrNoRows {
		return nil, fmt.Errorf("шаблон не найден")
	} else if err != nil {
		return nil, err
	}
	warnings, err := validateCustomWod(d, &in)
	if err != nil {
		return nil, err
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = fmt.Sprintf("%s · %s · %s",
			formatNames[in.Format], ThemeName(in.Theme), time.Now().Format("02.01"))
	}

	tx, err := d.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var desc any
	if strings.TrimSpace(in.Description) != "" {
		desc = in.Description
	}
	if _, err := tx.Exec(`
		UPDATE wod_templates SET name = ?, format = ?, duration_min = ?, intensity = ?, theme = ?, description = ?, is_custom = 1
		WHERE id = ?`,
		name, in.Format, in.DurationMin, in.Intensity, in.Theme, desc, templateID,
	); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`DELETE FROM wod_template_movements WHERE template_id = ?`, templateID); err != nil {
		return nil, err
	}
	for i, m := range in.Movements {
		var mname string
		if err := tx.QueryRow(`SELECT name FROM movements WHERE "key" = ?`, m.MovementKey).Scan(&mname); err != nil {
			return nil, fmt.Errorf("движение «%s» не найдено", m.MovementKey)
		}
		var reps, wm, wf, rn any
		if m.Reps != nil {
			reps = *m.Reps
		}
		if m.WeightMale != nil {
			wm = *m.WeightMale
		}
		if m.WeightFemale != nil {
			wf = *m.WeightFemale
		}
		if m.RoundsNote != "" {
			rn = m.RoundsNote
		}
		if _, err := tx.Exec(`
			INSERT INTO wod_template_movements (id, template_id, movement_key, movement_name, reps, weight_male, weight_female, sort_order, rounds_note)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			db.NewUUID(), templateID, m.MovementKey, mname, reps, wm, wf, i+1, rn,
		); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	if len(warnings) == 0 {
		warnings = nil
	}
	return warnings, nil
}

// TemplateForEdit — шаблон с исходными (нескалированными) движениями
// для предзаполнения конструктора в режиме правки.
type TemplateForEdit struct {
	TemplateID  string                `json:"template_id"`
	Name        string                `json:"name"`
	Format      string                `json:"format"`
	DurationMin int                   `json:"duration_min"`
	Intensity   string                `json:"intensity"`
	Theme       string                `json:"theme"`
	Description *string               `json:"description"`
	IsBenchmark bool                  `json:"is_benchmark"`
	Archived    bool                  `json:"archived"`
	Movements   []CustomMovementInput `json:"movements"`
}

// GetTemplateForEdit возвращает шаблон как есть — без скалирования.
func GetTemplateForEdit(d *sql.DB, templateID string) *TemplateForEdit {
	var t TemplateForEdit
	var desc sql.NullString
	var archived int
	err := d.QueryRow(`
		SELECT id, name, format, duration_min, intensity, theme, description, COALESCE(is_benchmark, 0), COALESCE(archived, 0)
		FROM wod_templates WHERE id = ?`, templateID,
	).Scan(&t.TemplateID, &t.Name, &t.Format, &t.DurationMin, &t.Intensity, &t.Theme, &desc, &t.IsBenchmark, &archived)
	if err != nil {
		return nil
	}
	t.Archived = archived == 1
	if desc.Valid {
		t.Description = &desc.String
	}
	rows, err := d.Query(`
		SELECT movement_key, reps, weight_male, weight_female, COALESCE(rounds_note, '')
		FROM wod_template_movements WHERE template_id = ? ORDER BY sort_order`, templateID)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()
	t.Movements = []CustomMovementInput{}
	for rows.Next() {
		var m CustomMovementInput
		var reps, wm, wf sql.NullInt64
		if err := rows.Scan(&m.MovementKey, &reps, &wm, &wf, &m.RoundsNote); err != nil {
			return nil
		}
		if reps.Valid {
			v := int(reps.Int64)
			m.Reps = &v
		}
		if wm.Valid {
			v := int(wm.Int64)
			m.WeightMale = &v
		}
		if wf.Valid {
			v := int(wf.Int64)
			m.WeightFemale = &v
		}
		t.Movements = append(t.Movements, m)
	}
	return &t
}

// patternOf — маппинг muscle_group на паттерн баланса.
func patternOf(muscleGroup string) string {
	switch strings.TrimSpace(muscleGroup) {
	case "legs", "quads", "hamstrings", "glutes":
		return "ноги"
	case "back", "lats":
		return "тяга"
	case "chest", "shoulders", "triceps", "arms":
		return "толчок"
	case "core", "abs":
		return "кор"
	}
	return ""
}
