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
	if !validFormats[in.Format] {
		return "", nil, fmt.Errorf("неизвестный формат: %s", in.Format)
	}
	if !validIntensities[in.Intensity] {
		in.Intensity = "medium"
	}
	if _, ok := themeNames[in.Theme]; !ok {
		return "", nil, fmt.Errorf("неизвестная тема: %s", in.Theme)
	}
	if in.DurationMin <= 0 || in.DurationMin > 120 {
		return "", nil, fmt.Errorf("длительность — от 1 до 120 минут")
	}
	if len(in.Movements) == 0 {
		return "", nil, fmt.Errorf("добавьте хотя бы одно упражнение")
	}
	if len(in.Movements) > 15 {
		return "", nil, fmt.Errorf("слишком много упражнений (максимум 15)")
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
			return "", nil, fmt.Errorf("движение «%s» не найдено в каталоге", m.MovementKey)
		}
		if err != nil {
			return "", nil, err
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
		INSERT INTO wod_templates (id, name, format, duration_min, intensity, theme, is_benchmark, description)
		VALUES (?, ?, ?, ?, ?, ?, 0, ?)`,
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
