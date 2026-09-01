package services

import (
	"database/sql"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"strings"

	"github.com/maxdukov/cf/backend-go/internal/db"
)

// ── Генератор тренировок — 3 варианта WoD по теме и инвентарю ──

// Множители повторений по уровню группы.
var levelMultipliers = map[string]float64{
	"beginner": 0.5, "intermediate": 0.75, "advanced": 1.0, "elite": 1.25,
}

// Целевые пульсовые зоны по формату тренировки.
var formatTargetZones = map[string][]int{
	"amrap": {3, 4}, "for_time": {3, 4}, "emom": {2, 3}, "tabata": {3, 4},
	"chipper": {2, 3}, "ladder": {3, 4}, "death_by": {3, 4}, "strength": {2},
}

var formatNames = map[string]string{
	"amrap": "AMRAP", "for_time": "На время", "emom": "EMOM", "tabata": "Tabata",
	"chipper": "Chipper", "ladder": "Лесенка", "death_by": "Death By", "strength": "Силовая",
}

var intensityNames = map[string]string{
	"low": "Низкая", "medium": "Средняя", "high": "Высокая",
}

// WodVariantMovement — отскалированное движение в варианте генерации.
type WodVariantMovement struct {
	MovementKey  string  `json:"movement_key"`
	MovementName string  `json:"movement_name"`
	Reps         *int    `json:"reps"`
	WeightMale   *int    `json:"weight_male"`
	WeightFemale *int    `json:"weight_female"`
	SortOrder    int     `json:"sort_order"`
	ScalingNote  *string `json:"scaling_note"`
	RoundsNote   *string `json:"rounds_note"`
}

// WodVariant — один сгенерированный вариант WoD (формат /api/wods/generate).
type WodVariant struct {
	TemplateID    string               `json:"template_id"`
	Name          string               `json:"name"`
	Format        string               `json:"format"`
	FormatName    string               `json:"format_name"`
	DurationMin   int                  `json:"duration_min"`
	Intensity     string               `json:"intensity"`
	IntensityName string               `json:"intensity_name"`
	Theme         string               `json:"theme"`
	IsBenchmark   bool                 `json:"is_benchmark"`
	Description   *string              `json:"description"`
	TargetZones   []int                `json:"target_zones"`
	Movements     []WodVariantMovement `json:"movements"`
}

// WodMovementRow — строка wod_template_movements / wod_movements.
type WodMovementRow struct {
	MovementKey  string
	MovementName string
	Reps         sql.NullInt64
	WeightMale   sql.NullInt64
	WeightFemale sql.NullInt64
	SortOrder    int
	RoundsNote   sql.NullString
}

type wodTemplateRow struct {
	ID          string
	Name        string
	Format      string
	DurationMin int
	Intensity   string
	Theme       string
	IsBenchmark bool
	Description sql.NullString
}

// pythonRound — округление как в Python (banker's rounding),
// чтобы скалирование совпадало с референсом 1:1.
func pythonRound(v float64) int {
	f := math.Floor(v)
	diff := v - f
	switch {
	case diff > 0.5:
		return int(f + 1)
	case diff < 0.5:
		return int(f)
	default: // ровно .5 — к чётному
		if int(f)%2 == 0 {
			return int(f)
		}
		return int(f + 1)
	}
}

func scaleReps(reps sql.NullInt64, level string) sql.NullInt64 {
	if !reps.Valid {
		return sql.NullInt64{}
	}
	mult := levelMultipliers[level]
	if mult == 0 {
		mult = 1.0
	}
	scaled := pythonRound(float64(reps.Int64) * mult)
	if scaled < 1 {
		scaled = 1
	}
	return sql.NullInt64{Int64: int64(scaled), Valid: true}
}

func scaleWeight(weight sql.NullInt64, level string) sql.NullInt64 {
	if !weight.Valid {
		return sql.NullInt64{}
	}
	mult := levelMultipliers[level]
	if mult == 0 {
		mult = 1.0
	}
	scaled := pythonRound(float64(weight.Int64) * mult)
	// Округление до ближайших 2.5 кг (banker's rounding + усечение, как в Python:
	// round(scaled/2.5)*2.5 → int() усекает .5: 32.5 → 32).
	rounded := pythonRound(float64(scaled)/2.5) * 5 / 2
	return sql.NullInt64{Int64: int64(rounded), Valid: true}
}

func getScalingNote(d *sql.DB, movementKey, level string) *string {
	var sb, si sql.NullString
	err := d.QueryRow(`SELECT scaling_beginner, scaling_intermediate FROM movements WHERE "key" = ?`, movementKey).
		Scan(&sb, &si)
	if err != nil {
		return nil
	}
	if level == "beginner" && sb.Valid && sb.String != "" {
		return &sb.String
	}
	if level == "intermediate" && si.Valid && si.String != "" {
		return &si.String
	}
	return nil
}

func templateMovements(d *sql.DB, templateID string) []WodMovementRow {
	rows, err := d.Query(`
		SELECT movement_key, movement_name, reps, weight_male, weight_female, sort_order, rounds_note
		FROM wod_template_movements WHERE template_id = ? ORDER BY sort_order`, templateID)
	if err != nil {
		slog.Error("template movements", "err", err)
		return nil
	}
	defer func() { _ = rows.Close() }()

	var out []WodMovementRow
	for rows.Next() {
		var m WodMovementRow
		if err := rows.Scan(&m.MovementKey, &m.MovementName, &m.Reps, &m.WeightMale, &m.WeightFemale, &m.SortOrder, &m.RoundsNote); err != nil {
			slog.Error("scan template movement", "err", err)
			continue
		}
		out = append(out, m)
	}
	return out
}

func buildWodFromTemplate(d *sql.DB, tpl wodTemplateRow, level string) WodVariant {
	movements := templateMovements(d, tpl.ID)
	wm := make([]WodVariantMovement, 0, len(movements))
	for _, m := range movements {
		wm = append(wm, WodVariantMovement{
			MovementKey:  m.MovementKey,
			MovementName: m.MovementName,
			Reps:         nullInt64ToPtr(scaleReps(m.Reps, level)),
			WeightMale:   nullInt64ToPtr(scaleWeight(m.WeightMale, level)),
			WeightFemale: nullInt64ToPtr(scaleWeight(m.WeightFemale, level)),
			SortOrder:    m.SortOrder,
			ScalingNote:  getScalingNote(d, m.MovementKey, level),
			RoundsNote:   nullStrToPtr(m.RoundsNote),
		})
	}

	zones, ok := formatTargetZones[tpl.Format]
	if !ok {
		zones = []int{2, 3}
	}
	formatName, ok := formatNames[tpl.Format]
	if !ok {
		formatName = tpl.Format
	}
	intensityName, ok := intensityNames[tpl.Intensity]
	if !ok {
		intensityName = tpl.Intensity
	}

	return WodVariant{
		TemplateID:    tpl.ID,
		Name:          tpl.Name,
		Format:        tpl.Format,
		FormatName:    formatName,
		DurationMin:   tpl.DurationMin,
		Intensity:     tpl.Intensity,
		IntensityName: intensityName,
		Theme:         tpl.Theme,
		IsBenchmark:   tpl.IsBenchmark,
		Description:   nullStrToPtr(tpl.Description),
		TargetZones:   zones,
		Movements:     wm,
	}
}

// GenerateWods генерирует до 3 вариантов WoD по теме и уровню.
func GenerateWods(d *sql.DB, theme, groupLevel string) []WodVariant {
	available := map[string]bool{}
	rows, err := d.Query(`SELECT equipment_key FROM gym_inventory`)
	if err == nil {
		for rows.Next() {
			var k string
			if rows.Scan(&k) == nil {
				available[k] = true
			}
		}
		_ = rows.Close()
	}

	tplRows, err := d.Query(`
		SELECT id, name, format, duration_min, intensity, theme, COALESCE(is_benchmark, 0), description
		FROM wod_templates WHERE theme = ?`, theme)
	if err != nil {
		slog.Error("generate wods: templates query", "err", err)
		return nil
	}
	var templates []wodTemplateRow
	for tplRows.Next() {
		var t wodTemplateRow
		if err := tplRows.Scan(&t.ID, &t.Name, &t.Format, &t.DurationMin, &t.Intensity, &t.Theme, &t.IsBenchmark, &t.Description); err != nil {
			continue
		}
		templates = append(templates, t)
	}
	_ = tplRows.Close()

	// Фильтр по доступному инвентарю.
	suitable := make([]wodTemplateRow, 0, len(templates))
	for _, t := range templates {
		if templateHasEquipment(d, t, available) {
			suitable = append(suitable, t)
		}
	}
	// Если инвентарь не настроен — показываем всё.
	if len(available) == 0 {
		suitable = templates
	}
	if len(suitable) == 0 {
		slog.Warn("no suitable templates", "theme", theme)
		return nil
	}

	partition := func(pred func(t wodTemplateRow) bool) []wodTemplateRow {
		var out []wodTemplateRow
		for _, t := range suitable {
			if pred(t) {
				out = append(out, t)
			}
		}
		return out
	}
	benchmarks := partition(func(t wodTemplateRow) bool { return t.IsBenchmark })
	medium := partition(func(t wodTemplateRow) bool { return t.Intensity == "medium" && !t.IsBenchmark })
	high := partition(func(t wodTemplateRow) bool { return t.Intensity == "high" && !t.IsBenchmark })
	low := partition(func(t wodTemplateRow) bool { return t.Intensity == "low" && !t.IsBenchmark })

	choice := func(pool []wodTemplateRow) wodTemplateRow { return pool[rand.Intn(len(pool))] }

	var result []WodVariant
	nameIn := func(name string) bool {
		for _, w := range result {
			if w.Name == name {
				return true
			}
		}
		return false
	}

	// Вариант A: классический бенчмарк.
	if len(benchmarks) > 0 {
		result = append(result, buildWodFromTemplate(d, choice(benchmarks), groupLevel))
	}

	// Вариант B: сбалансированный.
	poolB := append(append([]wodTemplateRow{}, medium...), low...)
	if len(result) == 0 {
		poolB = append(poolB, benchmarks...)
	}
	if len(poolB) > 0 {
		wod := buildWodFromTemplate(d, choice(poolB), groupLevel)
		if len(result) == 0 || wod.Name != result[0].Name {
			result = append(result, wod)
		}
	}

	// Вариант C: высокоинтенсивный.
	poolC := append(append([]wodTemplateRow{}, high...), medium...)
	if len(result) < 2 {
		poolC = append(poolC, benchmarks...)
	}
	if len(poolC) > 0 {
		wod := buildWodFromTemplate(d, choice(poolC), groupLevel)
		if !nameIn(wod.Name) {
			result = append(result, wod)
		}
	}

	// Добиваем до 3 вариантов.
	remaining := make([]wodTemplateRow, 0)
	for _, t := range suitable {
		if !nameIn(t.Name) {
			remaining = append(remaining, t)
		}
	}
	for len(result) < 3 && len(remaining) > 0 {
		c := choice(remaining)
		remaining = removeTemplate(remaining, c)
		result = append(result, buildWodFromTemplate(d, c, groupLevel))
	}

	if len(result) > 3 {
		result = result[:3]
	}
	return result
}

func removeTemplate(list []wodTemplateRow, t wodTemplateRow) []wodTemplateRow {
	for i, item := range list {
		if item.ID == t.ID {
			return append(list[:i:i], list[i+1:]...)
		}
	}
	return list
}

func templateHasEquipment(d *sql.DB, tpl wodTemplateRow, available map[string]bool) bool {
	for _, tm := range templateMovements(d, tpl.ID) {
		var equipmentKeys sql.NullString
		err := d.QueryRow(`SELECT equipment_keys FROM movements WHERE "key" = ?`, tm.MovementKey).Scan(&equipmentKeys)
		if err != nil || !equipmentKeys.Valid || strings.TrimSpace(equipmentKeys.String) == "" {
			continue
		}
		for _, k := range strings.Split(equipmentKeys.String, ",") {
			k = strings.TrimSpace(k)
			if k != "" && !available[k] {
				return false
			}
		}
	}
	return true
}

// CreateWodFromTemplate создаёт активный WoD из выбранного шаблона.
// Возвращает ID созданного WoD; ошибка — если шаблон не найден.
func CreateWodFromTemplate(d *sql.DB, templateID, groupLevel string) (string, error) {
	return createWodFromTemplate(d, templateID, groupLevel, true)
}

// CreateWodForSlot создаёт НЕактивный WoD (экземпляр слота).
func CreateWodForSlot(d *sql.DB, templateID, groupLevel string) (string, error) {
	return createWodFromTemplate(d, templateID, groupLevel, false)
}

func createWodFromTemplate(d *sql.DB, templateID, groupLevel string, active bool) (string, error) {
	var tpl wodTemplateRow
	err := d.QueryRow(`
		SELECT id, name, format, duration_min, intensity, theme, COALESCE(is_benchmark, 0), description
		FROM wod_templates WHERE id = ?`, templateID,
	).Scan(&tpl.ID, &tpl.Name, &tpl.Format, &tpl.DurationMin, &tpl.Intensity, &tpl.Theme, &tpl.IsBenchmark, &tpl.Description)
	if err == sql.ErrNoRows {
		//nolint:staticcheck // сообщение 1:1 с Python-версией (detail фронтенду)
		return "", fmt.Errorf("Template %s not found", templateID)
	}
	if err != nil {
		return "", err
	}

	// Всё чтение — ДО открытия транзакции (MaxOpenConns(1): вложенные
	// запросы через пул при живом tx дают дедлок).
	movements := templateMovements(d, templateID)
	type movInsert struct {
		key, name    string
		reps, wm, wf sql.NullInt64
		sortOrder    int
		scalingNote  any
		roundsNote   any
	}
	inserts := make([]movInsert, 0, len(movements))
	for _, tm := range movements {
		var scalingNote any
		if sn := getScalingNote(d, tm.MovementKey, groupLevel); sn != nil {
			scalingNote = *sn
		}
		var roundsNote any
		if tm.RoundsNote.Valid {
			roundsNote = tm.RoundsNote.String
		}
		inserts = append(inserts, movInsert{
			key: tm.MovementKey, name: tm.MovementName,
			reps:      scaleReps(tm.Reps, groupLevel),
			wm:        scaleWeight(tm.WeightMale, groupLevel),
			wf:        scaleWeight(tm.WeightFemale, groupLevel),
			sortOrder: tm.SortOrder, scalingNote: scalingNote, roundsNote: roundsNote,
		})
	}

	tx, err := d.Begin()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	if active {
		if _, err := tx.Exec(`UPDATE wods SET is_active = 0 WHERE is_active = 1`); err != nil {
			return "", err
		}
	}

	activeFlag := 0
	if active {
		activeFlag = 1
	}
	wodID := db.NewUUID()
	var desc any
	if tpl.Description.Valid {
		desc = tpl.Description.String
	}
	if _, err := tx.Exec(`
		INSERT INTO wods (id, name, format, duration_min, intensity, theme, group_level, description, is_active, created_at, template_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		wodID, tpl.Name, tpl.Format, tpl.DurationMin, tpl.Intensity, tpl.Theme, groupLevel, desc, activeFlag, db.NowDB(), templateID,
	); err != nil {
		return "", err
	}

	for _, m := range inserts {
		if _, err := tx.Exec(`
			INSERT INTO wod_movements (id, wod_id, movement_key, movement_name, reps, weight_male, weight_female, sort_order, scaling_note, rounds_note)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			db.NewUUID(), wodID, m.key, m.name,
			nullInt64Any(m.reps), nullInt64Any(m.wm), nullInt64Any(m.wf),
			m.sortOrder, m.scalingNote, m.roundsNote,
		); err != nil {
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return wodID, nil
}

func nullInt64ToPtr(ni sql.NullInt64) *int {
	if !ni.Valid {
		return nil
	}
	v := int(ni.Int64)
	return &v
}

func nullStrToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

// nullInt64Any — драйвер-аргумент для nullable INTEGER.
func nullInt64Any(ni sql.NullInt64) any {
	if !ni.Valid {
		return nil
	}
	return ni.Int64
}
