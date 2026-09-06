// Package services — циклы планирования тренировок (draft1.MD).
package services

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/maxdukov/cf/backend-go/internal/db"
)

const slotDateLayout = "2006-01-02"

var themeNames = map[string]string{
	"legs": "Ноги", "arms_shoulders": "Руки и плечи", "clean_jerk": "Толчок",
	"snatch": "Рывок", "cardio_metcon": "Кардио / меткон", "gymnastics": "Гимнастика",
	"core": "Кор", "full_body": "Всё тело",
}

var modalityThemes = map[string][]string{
	"strength":   {"clean_jerk", "snatch", "legs", "arms_shoulders"},
	"gymnastics": {"gymnastics", "core"},
	"cardio":     {"cardio_metcon", "full_body"},
}

// BuildTemplatePreview — карточка шаблона с отскалированными движениями.
func BuildTemplatePreview(d *sql.DB, templateID, groupLevel string) *WodVariant {
	var tpl wodTemplateRow
	err := d.QueryRow(`
		SELECT id, name, format, duration_min, intensity, theme, COALESCE(is_benchmark, 0), description
		FROM wod_templates WHERE id = ?`, templateID,
	).Scan(&tpl.ID, &tpl.Name, &tpl.Format, &tpl.DurationMin, &tpl.Intensity, &tpl.Theme, &tpl.IsBenchmark, &tpl.Description)
	if err != nil {
		return nil
	}
	v := buildWodFromTemplate(d, tpl, groupLevel)
	return &v
}

// ThemeName — русское название темы.
func ThemeName(theme string) string {
	if n, ok := themeNames[theme]; ok {
		return n
	}
	return theme
}

// ── Создание цикла (Экран 1) ──────────────────────────────────

// CycleGroupInput — группа цикла с днями недели (ISO: 1=Пн … 7=Вс).
type CycleGroupInput struct {
	Name             string `json:"name"`
	Weekdays         []int  `json:"weekdays"`
	ThirdDayOffCycle bool   `json:"third_day_off_cycle"`
}

// Подсказки тренеру для дней вне цикла (чередуются).
const (
	noteOffCycleTechnique = "Вне цикла · день техники: отработайте движение из текущего цикла (слабое звено) — 4×5 с паузами, лёгкий вес"
	noteOffCycleTesting   = "Вне цикла · тестирование: 1ПМ по основному движению цикла — thorough разминка, 3 попытки, зафиксируйте результат"
)

// CreateCycleInput — запрос POST /api/cycles.
type CreateCycleInput struct {
	Name             string            `json:"name"`
	Goal             string            `json:"goal"`
	Weeks            int               `json:"weeks"`
	StartDate        string            `json:"start_date"` // YYYY-MM-DD
	ModalityPriority string            `json:"modality_priority"`
	Groups           []CycleGroupInput `json:"groups"`
}

// CreateCycle создаёт цикл, группы и автогенерирует слоты.
// Возвращает ID цикла и предупреждения планирования (Экран 1).
func CreateCycle(d *sql.DB, in CreateCycleInput) (string, []string, error) {
	if strings.TrimSpace(in.Name) == "" {
		return "", nil, fmt.Errorf("укажите название цикла")
	}
	if in.Weeks <= 0 || in.Weeks > 52 {
		in.Weeks = 8
	}
	start, err := time.Parse(slotDateLayout, in.StartDate)
	if err != nil {
		return "", nil, fmt.Errorf("start_date должен быть в формате YYYY-MM-DD")
	}
	if len(in.Groups) == 0 {
		return "", nil, fmt.Errorf("добавьте хотя бы одну группу")
	}

	var warnings []string
	seen := map[string]bool{}
	groups := make([]CycleGroupInput, 0, len(in.Groups))
	for _, g := range in.Groups {
		if strings.TrimSpace(g.Name) == "" {
			return "", nil, fmt.Errorf("у каждой группы должно быть название")
		}
		if len(g.Weekdays) == 0 {
			return "", nil, fmt.Errorf("группа «%s»: укажите дни недели", g.Name)
		}
		wds := make([]int, 0, len(g.Weekdays))
		for _, wd := range g.Weekdays {
			if wd < 1 || wd > 7 {
				return "", nil, fmt.Errorf("группа «%s»: дни недели 1–7 (Пн–Вс)", g.Name)
			}
			wds = append(wds, wd)
		}
		sort.Ints(wds)
		g.Weekdays = wds
		groups = append(groups, g)
		if seen[g.Name] {
			warnings = append(warnings, fmt.Sprintf("Группы с именем «%s» совпадают", g.Name))
		}
		seen[g.Name] = true
		// Правило 48 часов: тренировки в соседние дни.
		for i := 1; i < len(wds); i++ {
			wrapCase := wds[i] == 7 && len(wds) > 2 && wds[0] == 1
			if wds[i]-wds[i-1] == 1 && !wrapCase {
				warnings = append(warnings, fmt.Sprintf(
					"Группа «%s»: между %s и %s меньше 48 часов восстановления",
					g.Name, weekdayName(wds[i-1]), weekdayName(wds[i])))
				break
			}
		}
	}
	if len(warnings) == 0 {
		warnings = nil
	}

	// Слоты считаем заранее (до транзакции).
	type slotInsert struct {
		groupID, date, kind, notes string
		dayNumber                  int
	}
	var allSlots []slotInsert

	tx, err := d.Begin()
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = tx.Rollback() }()

	cycleID := db.NewUUID()
	var goal any
	if strings.TrimSpace(in.Goal) != "" {
		goal = in.Goal
	}
	var modality any
	if strings.TrimSpace(in.ModalityPriority) != "" {
		modality = in.ModalityPriority
	}
	if _, err := tx.Exec(`
		INSERT INTO training_cycles (id, name, goal, weeks, start_date, status, modality_priority, created_at)
		VALUES (?, ?, ?, ?, ?, 'planned', ?, ?)`,
		cycleID, in.Name, goal, in.Weeks, in.StartDate, modality, db.NowDB(),
	); err != nil {
		return "", nil, err
	}

	end := start.AddDate(0, 0, in.Weeks*7)
	for _, g := range groups {
		groupID := db.NewUUID()
		wdSet := map[int]bool{}
		for _, wd := range g.Weekdays {
			wdSet[wd] = true
		}
		if _, err := tx.Exec(`
			INSERT INTO cycle_groups (id, cycle_id, name, weekdays, third_day_off) VALUES (?, ?, ?, ?, ?)`,
			groupID, cycleID, g.Name, joinInts(g.Weekdays), g.ThirdDayOffCycle,
		); err != nil {
			return "", nil, err
		}
		dayNumber := 0
		for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
			if wdSet[isoWeekday(day)] {
				dayNumber++
				kind, notes := "regular", ""
				if g.ThirdDayOffCycle && dayNumber%3 == 0 {
					kind = "off_cycle"
					notes = noteOffCycleTechnique
					if (dayNumber/3)%2 == 0 {
						notes = noteOffCycleTesting
					}
				}
				allSlots = append(allSlots, slotInsert{
					groupID: groupID, date: day.Format(slotDateLayout),
					kind: kind, notes: notes, dayNumber: dayNumber,
				})
			}
		}
	}

	for _, s := range allSlots {
		if _, err := tx.Exec(`
			INSERT INTO cycle_slots (id, cycle_id, group_id, slot_date, day_number, status, kind, notes)
			VALUES (?, ?, ?, ?, ?, 'empty', ?, ?)`,
			db.NewUUID(), cycleID, s.groupID, s.date, s.dayNumber, s.kind, nullIfEmpty(s.notes),
		); err != nil {
			return "", nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", nil, err
	}
	return cycleID, warnings, nil
}

func isoWeekday(t time.Time) int { return (int(t.Weekday())+6)%7 + 1 }

// normDate — SQLite может вернуть DATE как RFC3339; режем до YYYY-MM-DD.
func normDate(s string) string {
	if len(s) > 10 {
		return s[:10]
	}
	return s
}

// parseSlotDate парсит дату слота в любом из форматов БД.
func parseSlotDate(s string) (time.Time, error) {
	s = normDate(s)
	if t, err := time.Parse(slotDateLayout, s); err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339, s)
}

func weekdayName(iso int) string {
	names := []string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"}
	if iso >= 1 && iso <= 7 {
		return names[iso-1]
	}
	return "?"
}

func joinInts(items []int) string {
	parts := make([]string, len(items))
	for i, v := range items {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return strings.Join(parts, ",")
}

// ── Чтение циклов ─────────────────────────────────────────────

// CycleSummary — элемент списка циклов.
type CycleSummary struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Goal             *string `json:"goal"`
	Weeks            int     `json:"weeks"`
	StartDate        string  `json:"start_date"`
	Status           string  `json:"status"`
	ModalityPriority *string `json:"modality_priority"`
	CreatedAt        *string `json:"created_at"`
	GroupsCount      int     `json:"groups_count"`
	SlotsTotal       int     `json:"slots_total"`
	SlotsFilled      int     `json:"slots_filled"`
	SlotsCompleted   int     `json:"slots_completed"`
	FillPercent      float64 `json:"fill_percent"`
}

// ListCycles — все циклы с прогрессом заполнения.
func ListCycles(d *sql.DB) []CycleSummary {
	rows, err := d.Query(`
		SELECT c.id, c.name, c.goal, c.weeks, c.start_date, c.status, c.modality_priority, c.created_at,
		       (SELECT COUNT(*) FROM cycle_groups g WHERE g.cycle_id = c.id),
		       (SELECT COUNT(*) FROM cycle_slots s WHERE s.cycle_id = c.id),
		       (SELECT COUNT(*) FROM cycle_slots s WHERE s.cycle_id = c.id AND s.status != 'empty'),
		       (SELECT COUNT(*) FROM cycle_slots s WHERE s.cycle_id = c.id AND s.status = 'completed')
		FROM training_cycles c ORDER BY c.created_at DESC`)
	if err != nil {
		return nil
	}
	defer func() { _ = rows.Close() }()

	out := []CycleSummary{}
	for rows.Next() {
		var c CycleSummary
		var goal, modality, created sql.NullString
		if err := rows.Scan(&c.ID, &c.Name, &goal, &c.Weeks, &c.StartDate, &c.Status, &modality, &created,
			&c.GroupsCount, &c.SlotsTotal, &c.SlotsFilled, &c.SlotsCompleted); err != nil {
			continue
		}
		c.Goal = nullStrToPtr(goal)
		c.ModalityPriority = nullStrToPtr(modality)
		c.CreatedAt = isoDatePtr(created)
		if c.SlotsTotal > 0 {
			c.FillPercent = float64(c.SlotsFilled) * 100 / float64(c.SlotsTotal)
		}
		out = append(out, c)
	}
	return out
}

// nullIfEmpty — пустую строку пишем в БД как NULL.
func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func isoDatePtr(s sql.NullString) *string {
	if !s.Valid || s.String == "" {
		return nil
	}
	t, err := time.Parse(db.DBTimeLayout, s.String)
	if err != nil {
		if t2, err2 := time.Parse(time.RFC3339Nano, s.String); err2 == nil {
			v := t2.UTC().Format("2006-01-02T15:04:05.000000Z07:00")
			return &v
		}
		return &s.String
	}
	v := t.UTC().Format("2006-01-02T15:04:05.000000Z07:00")
	return &v
}

// CycleGroup — группа цикла.
type CycleGroup struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Weekdays      []int    `json:"weekdays"`
	WeekdaysNames []string `json:"weekdays_names"`
	ThirdDayOff   bool     `json:"third_day_off"`
}

// SlotConflict — запланированный слот третьего дня при включении режима.
type SlotConflict struct {
	SlotID    string  `json:"slot_id"`
	SlotDate  string  `json:"slot_date"`
	DayNumber int     `json:"day_number"`
	WodName   *string `json:"wod_name"`
}

// SlotView — слот календаря (Экран 2).
type SlotView struct {
	ID         string        `json:"id"`
	GroupID    string        `json:"group_id"`
	GroupName  string        `json:"group_name"`
	SlotDate   string        `json:"slot_date"`
	DayNumber  int           `json:"day_number"`
	Status     string        `json:"status"`
	Kind       string        `json:"kind"`
	WodID      *string       `json:"wod_id"`
	TemplateID *string       `json:"template_id"`
	WodName    *string       `json:"wod_name"`
	WodFormat  *string       `json:"wod_format"`
	Intensity  *string       `json:"intensity"`
	Theme      *string       `json:"theme"`
	Notes      *string       `json:"notes"`
	Wods       []SlotWodItem `json:"wods"`
}

// Константы длительности тренировочного дня (Экран 5): план 60 минут
// (разминка 10 + заминка 5), плотный день — до 65 (подтверждение тренера),
// жёсткий потолок для тренировок: 65 − 15 = 50 минут.
const (
	SlotDayMin      = 60
	SlotDayMaxMin   = 65
	SlotWarmupMin   = 10
	SlotCooldownMin = 5
	SlotWodCapMin   = SlotDayMaxMin - SlotWarmupMin - SlotCooldownMin
)

// SlotWodItem — одна тренировка дня (без движений, для календаря и списков).
type SlotWodItem struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Format      string             `json:"format"`
	DurationMin int                `json:"duration_min"`
	Intensity   string             `json:"intensity"`
	Theme       string             `json:"theme"`
	Movements   []SlotWodMovement  `json:"movements"`
}

// SlotWodMovement — строка состава тренировки (для подсказок и панели).
type SlotWodMovement struct {
	MovementName string  `json:"movement_name"`
	Reps         *int    `json:"reps"`
	WeightMale   *int    `json:"weight_male"`
	WeightFemale *int    `json:"weight_female"`
	RoundsNote   *string `json:"rounds_note"`
}

// attachSlotWodMovements дополняет тренировки дня их составом одним запросом.
func attachSlotWodMovements(d *sql.DB, items []SlotWodItem) {
	if len(items) == 0 {
		return
	}
	ids := make([]string, 0, len(items))
	idx := map[string]int{}
	for i, it := range items {
		ids = append(ids, it.ID)
		idx[it.ID] = i
		items[i].Movements = []SlotWodMovement{}
	}
	ph := make([]string, len(ids))
	args := make([]any, 0, len(ids))
	for i, id := range ids {
		ph[i] = "?"
		args = append(args, id)
	}
	q := `
		SELECT wod_id, movement_name, reps, weight_male, weight_female, rounds_note
		FROM wod_movements WHERE wod_id IN (` + strings.Join(ph, ",") + `)
		ORDER BY sort_order`
	rows, err := d.Query(q, args...)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var wodID string
		var m SlotWodMovement
		var reps, wm, wf sql.NullInt64
		var rn sql.NullString
		if err := rows.Scan(&wodID, &m.MovementName, &reps, &wm, &wf, &rn); err != nil {
			continue
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
		if rn.Valid {
			v := rn.String
			m.RoundsNote = &v
		}
		if i, ok := idx[wodID]; ok {
			items[i].Movements = append(items[i].Movements, m)
		}
	}
}

// slotWodsList — тренировки дня в порядке добавления (с составом).
func slotWodsList(d *sql.DB, slotID string) ([]SlotWodItem, error) {
	rows, err := d.Query(`
		SELECT w.id, w.name, w.format, w.duration_min, w.intensity, w.theme
		FROM slot_wods sw JOIN wods w ON w.id = sw.wod_id
		WHERE sw.slot_id = ?
		ORDER BY sw.position, w.created_at`, slotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SlotWodItem
	for rows.Next() {
		var it SlotWodItem
		if err := rows.Scan(&it.ID, &it.Name, &it.Format, &it.DurationMin, &it.Intensity, &it.Theme); err == nil {
			out = append(out, it)
		}
	}
	attachSlotWodMovements(d, out)
	return out, nil
}

// CycleDetail — полный цикл с группами и слотами.
type CycleDetail struct {
	CycleSummary
	Groups   []CycleGroup `json:"groups"`
	Slots    []SlotView   `json:"slots"`
	Warnings []string     `json:"warnings"`
}

// GetCycle — цикл с календарём и подсветкой нарушений (2 тяжёлых подряд).
func GetCycle(d *sql.DB, cycleID string) *CycleDetail {
	var c CycleDetail
	var goal, modality, created sql.NullString
	err := d.QueryRow(`
		SELECT id, name, goal, weeks, start_date, status, modality_priority, created_at
		FROM training_cycles WHERE id = ?`, cycleID,
	).Scan(&c.ID, &c.Name, &goal, &c.Weeks, &c.StartDate, &c.Status, &modality, &created)
	if err != nil {
		return nil
	}
	c.Goal = nullStrToPtr(goal)
	c.ModalityPriority = nullStrToPtr(modality)
	c.CreatedAt = isoDatePtr(created)

	// Группы.
	grows, err := d.Query(`SELECT id, name, weekdays, COALESCE(third_day_off, 0) FROM cycle_groups WHERE cycle_id = ? ORDER BY name`, cycleID)
	if err != nil {
		return nil
	}
	groups := []CycleGroup{}
	for grows.Next() {
		var g CycleGroup
		var wd string
		var thirdOff bool
		if err := grows.Scan(&g.ID, &g.Name, &wd, &thirdOff); err != nil {
			continue
		}
		g.ThirdDayOff = thirdOff
		g.Weekdays = parseInts(wd)
		for _, w := range g.Weekdays {
			g.WeekdaysNames = append(g.WeekdaysNames, weekdayName(w))
		}
		groups = append(groups, g)
	}
	_ = grows.Close()
	c.Groups = groups

	// Слоты.
	srows, err := d.Query(`
		SELECT s.id, s.group_id, g.name, s.slot_date, s.day_number, s.status, COALESCE(s.kind, 'regular'), s.wod_id, s.template_id, s.notes,
		       w.name, w.format, w.intensity, w.theme
		FROM cycle_slots s
		JOIN cycle_groups g ON g.id = s.group_id
		LEFT JOIN wods w ON w.id = s.wod_id
		WHERE s.cycle_id = ?
		ORDER BY s.slot_date, g.name`, cycleID)
	if err != nil {
		return nil
	}
	slots := []SlotView{}
	for srows.Next() {
		var s SlotView
		var wodID, tplID, notes, wname, wformat, wintensity, wtheme sql.NullString
		if err := srows.Scan(&s.ID, &s.GroupID, &s.GroupName, &s.SlotDate, &s.DayNumber, &s.Status, &s.Kind,
			&wodID, &tplID, &notes, &wname, &wformat, &wintensity, &wtheme); err != nil {
			continue
		}
		s.SlotDate = normDate(s.SlotDate)
		s.WodID = nullStrToPtr(wodID)
		s.TemplateID = nullStrToPtr(tplID)
		s.Notes = nullStrToPtr(notes)
		s.WodName = nullStrToPtr(wname)
		s.WodFormat = nullStrToPtr(wformat)
		s.Intensity = nullStrToPtr(wintensity)
		s.Theme = nullStrToPtr(wtheme)
		slots = append(slots, s)
	}
	_ = srows.Close()

	// Тренировки каждого дня (может быть несколько — слот на 60 минут).
	for i := range slots {
		if list, err := slotWodsList(d, slots[i].ID); err == nil && len(list) > 0 {
			slots[i].Wods = list
		}
	}
	c.Slots = slots

	c.GroupsCount = len(groups)
	c.SlotsTotal = len(slots)
	for _, s := range slots {
		if s.Status != "empty" {
			c.SlotsFilled++
		}
		if s.Status == "completed" {
			c.SlotsCompleted++
		}
	}
	if c.SlotsTotal > 0 {
		c.FillPercent = float64(c.SlotsFilled) * 100 / float64(c.SlotsTotal)
	}

	// Подсветка нарушений: два высокоинтенсивных подряд у группы (Экран 2).
	c.Warnings = slotSequenceWarnings(slots)
	return &c
}

func slotSequenceWarnings(slots []SlotView) []string {
	var out []string
	byGroup := map[string][]SlotView{}
	for _, s := range slots {
		byGroup[s.GroupID] = append(byGroup[s.GroupID], s)
	}
	for gid, list := range byGroup {
		var prev SlotView
		for i, s := range list {
			if i > 0 && s.Intensity != nil && prev.Intensity != nil &&
				*s.Intensity == "high" && *prev.Intensity == "high" {
				out = append(out, fmt.Sprintf(
					"«%s»: две высокоинтенсивные тренировки подряд (%s и %s) — рекомендуем снизить нагрузку",
					list[0].GroupName, prev.SlotDate, s.SlotDate))
			}
			if s.Status != "skipped" {
				prev = s
			}
			_ = gid
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func parseInts(csv string) []int {
	var out []int
	for _, p := range strings.Split(csv, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var v int
		if _, err := fmt.Sscanf(p, "%d", &v); err == nil {
			out = append(out, v)
		}
	}
	return out
}

// UpdateCycleStatus — смена статуса (planned/active/completed).
// thirdDayOffNote возвращает чередующуюся подсказку для 3/6/9… дня.
func thirdDayOffNote(dayNumber int) string {
	if (dayNumber/3)%2 == 0 {
		return noteOffCycleTechnique
	}
	return noteOffCycleTesting
}

// markThirdDaySlots проставляет kind='off_cycle' свободным слотам 3/6/9… дня группы.
func markThirdDaySlots(d *sql.DB, groupID string) error {
	rows, err := d.Query(`
		SELECT id, day_number FROM cycle_slots
		WHERE group_id = ? AND day_number % 3 = 0 AND wod_id IS NULL`, groupID)
	if err != nil {
		return err
	}
	type idn struct {
		id    string
		dayNo int
	}
	var targets []idn
	for rows.Next() {
		var t idn
		if err := rows.Scan(&t.id, &t.dayNo); err == nil {
			targets = append(targets, t)
		}
	}
	_ = rows.Close()
	for _, t := range targets {
		if _, err := d.Exec(`
			UPDATE cycle_slots SET kind = 'off_cycle', notes = ? WHERE id = ?`,
			thirdDayOffNote(t.dayNo), t.id,
		); err != nil {
			return err
		}
	}
	return nil
}

// SetThirdDayOff включает/выключает режим «третий день вне цикла» у группы.
// При включении, если на третьи дни уже назначены тренировки (будущие planned),
// изменения НЕ применяются — возвращается список конфликтов для диалога.
func SetThirdDayOff(d *sql.DB, groupID string, enabled bool) (applied bool, conflicts []SlotConflict, err error) {
	var cycleID string
	var current bool
	err = d.QueryRow(`
		SELECT cycle_id, COALESCE(third_day_off, 0) FROM cycle_groups WHERE id = ?`, groupID,
	).Scan(&cycleID, &current)
	if err == sql.ErrNoRows {
		return false, nil, fmt.Errorf("группа не найдена")
	}
	if err != nil {
		return false, nil, err
	}
	if enabled == current {
		return true, nil, nil
	}

	if !enabled {
		// Выключение: свободные третьи дни возвращаем в цикл, назначенные не трогаем.
		if _, err := d.Exec(`UPDATE cycle_groups SET third_day_off = 0 WHERE id = ?`, groupID); err != nil {
			return false, nil, err
		}
		if _, err := d.Exec(`
			UPDATE cycle_slots SET kind = 'regular', notes = NULL
			WHERE group_id = ? AND day_number % 3 = 0 AND kind = 'off_cycle' AND wod_id IS NULL`, groupID,
		); err != nil {
			return false, nil, err
		}
		return true, nil, nil
	}

	// Включение: ищем конфликты — назначенные тренировки на будущих третьих днях.
	today := time.Now().Format(slotDateLayout)
	rows, err := d.Query(`
		SELECT s.id, s.slot_date, s.day_number, w.name
		FROM cycle_slots s LEFT JOIN wods w ON w.id = s.wod_id
		WHERE s.group_id = ? AND s.day_number % 3 = 0
		  AND s.wod_id IS NOT NULL AND s.status = 'planned' AND s.slot_date >= ?`,
		groupID, today)
	if err != nil {
		return false, nil, err
	}
	defer func() { _ = rows.Close() }()
	conflicts = []SlotConflict{}
	for rows.Next() {
		var c SlotConflict
		var date, wodName sql.NullString
		if err := rows.Scan(&c.SlotID, &date, &c.DayNumber, &wodName); err == nil {
			c.SlotDate = normDate(date.String)
			c.WodName = nullStrToPtr(wodName)
			conflicts = append(conflicts, c)
		}
	}
	if len(conflicts) > 0 {
		return false, conflicts, nil
	}
	if err := markThirdDaySlots(d, groupID); err != nil {
		return false, nil, err
	}
	if _, err := d.Exec(`UPDATE cycle_groups SET third_day_off = 1 WHERE id = ?`, groupID); err != nil {
		return false, nil, err
	}
	return true, nil, nil
}

// ReplanThirdDayOff перепланирует группу после включения режима:
//   - rollback: снять все запланированные тренировки в будущем (полный откат);
//   - auto: сохранить последовательность тренировок, распределив их только
//     по регулярным (первые два дня недели) слотам; лишние («хвост») удалить.
func ReplanThirdDayOff(d *sql.DB, groupID, mode string) (map[string]int, error) {
	if mode != "rollback" && mode != "auto" {
		return nil, fmt.Errorf("mode должен быть rollback или auto")
	}
	today := time.Now().Format(slotDateLayout)

	// Будущие запланированные слоты группы (по порядку).
	type slotRef struct {
		id    string
		wodID string
		kind  string
		dayNo int
	}
	rows, err := d.Query(`
		SELECT s.id, COALESCE(s.wod_id, ''), COALESCE(s.kind, 'regular'), s.day_number
		FROM cycle_slots s
		WHERE s.group_id = ? AND s.status = 'planned' AND s.slot_date >= ? AND s.wod_id IS NOT NULL
		ORDER BY s.slot_date, s.day_number`, groupID, today)
	if err != nil {
		return nil, err
	}
	var planned []slotRef
	for rows.Next() {
		var s slotRef
		if err := rows.Scan(&s.id, &s.wodID, &s.kind, &s.dayNo); err == nil {
			planned = append(planned, s)
		}
	}
	_ = rows.Close()

	// Полный откат: снять всё будущее запланированное.
	dropped := 0
	if mode == "rollback" {
		for _, s := range planned {
			if _, err := d.Exec(`DELETE FROM slot_wods WHERE slot_id = ?`, s.id); err != nil {
				return nil, err
			}
			if _, err := d.Exec(`
				UPDATE cycle_slots SET status = 'empty', wod_id = NULL, template_id = NULL, notes = NULL
				WHERE id = ?`, s.id); err != nil {
				return nil, err
			}
			if _, err := d.Exec(`DELETE FROM wods WHERE id = ?`, s.wodID); err != nil {
				return nil, err
			}
		}
		if err := markThirdDaySlots(d, groupID); err != nil {
			return nil, err
		}
		if _, err := d.Exec(`UPDATE cycle_groups SET third_day_off = 1 WHERE id = ?`, groupID); err != nil {
			return nil, err
		}
		return map[string]int{"cleared": len(planned)}, nil
	}

	// Автоперепланирование: снять всё, затем назначить последовательность
	// только на регулярные слоты (первые два дня недельного цикла).
	for _, s := range planned {
		if _, err := d.Exec(`DELETE FROM slot_wods WHERE slot_id = ?`, s.id); err != nil {
			return nil, err
		}
		if _, err := d.Exec(`
			UPDATE cycle_slots SET status = 'empty', wod_id = NULL, template_id = NULL, notes = NULL
			WHERE id = ?`, s.id); err != nil {
			return nil, err
		}
	}
	if err := markThirdDaySlots(d, groupID); err != nil {
		return nil, err
	}
	if _, err := d.Exec(`UPDATE cycle_groups SET third_day_off = 1 WHERE id = ?`, groupID); err != nil {
		return nil, err
	}

	// Цели: будущие регулярные свободные слоты по порядку.
	trows, err := d.Query(`
		SELECT id FROM cycle_slots
		WHERE group_id = ? AND slot_date >= ? AND COALESCE(kind, 'regular') = 'regular' AND status = 'empty'
		ORDER BY slot_date, day_number`, groupID, today)
	if err != nil {
		return nil, err
	}
	var targets []string
	for trows.Next() {
		var id string
		if err := trows.Scan(&id); err == nil {
			targets = append(targets, id)
		}
	}
	_ = trows.Close()

	reassigned := 0
	for i, s := range planned {
		if i >= len(targets) {
			dropped++ // «хвост» — тренировок больше, чем регулярных слотов
			if _, err := d.Exec(`DELETE FROM wods WHERE id = ?`, s.wodID); err != nil {
				return nil, err
			}
			continue
		}
		if _, err := d.Exec(`DELETE FROM slot_wods WHERE wod_id = ?`, s.wodID); err != nil {
			return nil, err
		}
		if _, err := d.Exec(`
			INSERT INTO slot_wods (slot_id, wod_id, position) VALUES (?, ?, 0)`, targets[i], s.wodID,
		); err != nil {
			return nil, err
		}
		if _, err := d.Exec(`
			UPDATE cycle_slots SET status = 'planned', wod_id = ? WHERE id = ?`, s.wodID, targets[i],
		); err != nil {
			return nil, err
		}
		reassigned++
	}
	return map[string]int{"reassigned": reassigned, "dropped": dropped}, nil
}

func UpdateCycleStatus(d *sql.DB, cycleID, status string) error {
	switch status {
	case "planned", "active", "completed":
	default:
		return fmt.Errorf("status должен быть planned, active или completed")
	}
	res, err := d.Exec(`UPDATE training_cycles SET status = ? WHERE id = ?`, status, cycleID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ── Рекомендации для слота (Экран 3) ──────────────────────────

// Recommendation — рекомендованная тренировка с обоснованием.
type Recommendation struct {
	WodVariant
	Reason string `json:"reason"`
}

// SlotRecommendations подбирает 3–5 вариантов для слота с обоснованием.
func SlotRecommendations(d *sql.DB, slotID, groupLevel string) ([]Recommendation, error) {
	if groupLevel == "" {
		groupLevel = "intermediate"
	}

	var cycleModality, slotDate, groupID, kind string
	err := d.QueryRow(`
		SELECT COALESCE(c.modality_priority, ''), s.slot_date, s.group_id, COALESCE(s.kind, 'regular')
		FROM cycle_slots s JOIN training_cycles c ON c.id = s.cycle_id
		WHERE s.id = ?`, slotID,
	).Scan(&cycleModality, &slotDate, &groupID, &kind)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("слот не найден")
	}
	if err != nil {
		return nil, err
	}

	// День вне цикла (техника/тесты) — штатные шаблоны не предлагаем,
	// тренер видит подсказку в notes слота и назначает вручную.
	if kind == "off_cycle" {
		return []Recommendation{}, nil
	}

	// Темы предыдущих 2–3 заполненных слотов группы (не повторять стимул).
	recentThemes := map[string]bool{}
	{
		rows, err := d.Query(`
			SELECT w.theme FROM cycle_slots s JOIN wods w ON w.id = s.wod_id
			WHERE s.group_id = ? AND s.status IN ('planned','completed') AND s.id != ?
			ORDER BY s.slot_date DESC LIMIT 3`, groupID, slotID)
		if err == nil {
			for rows.Next() {
				var t string
				if rows.Scan(&t) == nil {
					recentThemes[t] = true
				}
			}
			_ = rows.Close()
		}
	}

	// Кандидаты: темы из приоритета цикла (или все).
	themes, ok := modalityThemes[strings.TrimSpace(cycleModality)]
	if !ok || len(themes) == 0 {
		themes = []string{"legs", "arms_shoulders", "clean_jerk", "snatch", "cardio_metcon", "gymnastics", "core", "full_body"}
	}

	available := map[string]bool{}
	{
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
	}

	var candidates []wodTemplateRow
	for _, theme := range themes {
		rows, err := d.Query(`
			SELECT id, name, format, duration_min, intensity, theme, COALESCE(is_benchmark, 0), description
			FROM wod_templates WHERE theme = ?`, theme)
		if err != nil {
			continue
		}
		for rows.Next() {
			var t wodTemplateRow
			if rows.Scan(&t.ID, &t.Name, &t.Format, &t.DurationMin, &t.Intensity, &t.Theme, &t.IsBenchmark, &t.Description) == nil {
				if len(available) == 0 || templateHasEquipment(d, t, available) {
					candidates = append(candidates, t)
				}
			}
		}
		_ = rows.Close()
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	// Скоринг: свежесть темы, бенчмарк, разнообразие форматов.
	type scored struct {
		tpl    wodTemplateRow
		score  int
		reason string
	}
	var pool []scored
	usedFormats := map[string]bool{}
	usedNames := map[string]bool{}
	for _, t := range candidates {
		if usedNames[t.Name] {
			continue
		}
		score := 0
		reasonParts := []string{fmt.Sprintf("Развивает %s", strings.ToLower(ThemeName(t.Theme)))}

		if recentThemes[t.Theme] {
			score -= 10
			reasonParts = append(reasonParts, "⚠ тема повторяется с последними тренировками группы")
		} else if len(recentThemes) > 0 {
			score += 5
			reasonParts = append(reasonParts, "не повторяет последние тренировки группы")
		}
		if t.IsBenchmark {
			score += 3
			reasonParts = append(reasonParts, "бенчмарк — прогресс измерим")
		}
		if _, ok := modalityThemes[strings.TrimSpace(cycleModality)]; ok && containsTheme(modalityThemes[cycleModality], t.Theme) {
			score += 4
			reasonParts = append(reasonParts, fmt.Sprintf("соответствует приоритету цикла (%s)", cycleModality))
		}
		if !usedFormats[t.Format] {
			score += 2
		}
		pool = append(pool, scored{tpl: t, score: score, reason: strings.Join(reasonParts, "; ")})
	}

	sort.SliceStable(pool, func(i, j int) bool {
		if pool[i].score != pool[j].score {
			return pool[i].score > pool[j].score
		}
		return pool[i].tpl.Name < pool[j].tpl.Name
	})

	// Рандомизируем верхнюю часть, сохраняя порядок качества.
	top := pool
	if len(top) > 12 {
		top = top[:12]
	}
	rand.Shuffle(len(top), func(i, j int) { top[i], top[j] = top[j], top[i] })

	out := []Recommendation{}
	formatsUsed := map[string]bool{}
	for _, p := range top {
		if len(out) >= 5 {
			break
		}
		if formatsUsed[p.tpl.Format] && len(out) < 4 {
			continue // разнообразие форматов среди первых 4
		}
		out = append(out, Recommendation{
			WodVariant: buildWodFromTemplate(d, p.tpl, groupLevel),
			Reason:     p.reason,
		})
		usedNames[p.tpl.Name] = true
		formatsUsed[p.tpl.Format] = true
	}
	return out, nil
}

func containsTheme(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// ── Назначение на слот (Экран 5) ──────────────────────────────

// AssignTemplate материализует шаблон в слот (is_active=0).
// Возвращает предупреждения (правило 48 часов и др.).
func AssignTemplate(d *sql.DB, slotID, templateID, groupLevel string) (string, []string, error) {
	if groupLevel == "" {
		groupLevel = "intermediate"
	}

	var status, groupID, slotDate string
	err := d.QueryRow(`
		SELECT status, group_id, slot_date FROM cycle_slots WHERE id = ?`, slotID,
	).Scan(&status, &groupID, &slotDate)
	if err == sql.ErrNoRows {
		return "", nil, fmt.Errorf("слот не найден")
	}
	if err != nil {
		return "", nil, err
	}
	if status == "in_progress" || status == "completed" {
		return "", nil, fmt.Errorf("нельзя изменить уже идущую или завершённую тренировку")
	}

	// Проверка длительности: разминка 10 + сумма тренировок + заминка 5 ≤ 60.
	var newDur int
	if err := d.QueryRow(`SELECT duration_min FROM wod_templates WHERE id = ?`, templateID).Scan(&newDur); err != nil {
		return "", nil, fmt.Errorf("Template %s not found", templateID)
	}
	existing, err := slotWodsList(d, slotID)
	if err != nil {
		return "", nil, err
	}
	sum := 0
	for _, w := range existing {
		sum += w.DurationMin
	}
	if sum+newDur > SlotWodCapMin {
		return "", nil, fmt.Errorf(
			"Выходит за временной лимит: тренировки займут %d из %d доступных минут — день с разминкой и заминкой превысит %d",
			sum+newDur, SlotWodCapMin, SlotDayMaxMin)
	}

	wodID, err := CreateWodForSlot(d, templateID, groupLevel)
	if err != nil {
		return "", nil, err
	}

	if _, err := d.Exec(`
		INSERT INTO slot_wods (slot_id, wod_id, position) VALUES (?, ?, ?)`,
		slotID, wodID, len(existing)); err != nil {
		return "", nil, err
	}
	// wod_id слота всегда указывает на первую тренировку дня (для монитора и сессий).
	if _, err := d.Exec(`
		UPDATE cycle_slots SET status = 'planned', template_id = ?, wod_id = COALESCE(wod_id, ?)
		WHERE id = ?`, templateID, wodID, slotID); err != nil {
		return "", nil, err
	}

	warnings := assignWarnings(d, groupID, slotID, slotDate, templateID)
	return wodID, warnings, nil
}

// SlotWods — экспортированный список тренировок дня (для хендлеров).
func SlotWods(d *sql.DB, slotID string) ([]SlotWodItem, error) {
	return slotWodsList(d, slotID)
}

// UnassignSlotWod снимает одну тренировку дня. Если снимали первую —
// первой становится следующая по порядку; когда день опустел, слот сбрасывается.
func UnassignSlotWod(d *sql.DB, slotID, wodID string) error {
	var status string
	var slotWodID sql.NullString
	err := d.QueryRow(`SELECT status, wod_id FROM cycle_slots WHERE id = ?`, slotID).Scan(&status, &slotWodID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("слот не найден")
	}
	if err != nil {
		return err
	}
	if status == "in_progress" || status == "completed" {
		return fmt.Errorf("нельзя снять тренировку с идущего или завершённого дня")
	}

	res, err := d.Exec(`DELETE FROM slot_wods WHERE slot_id = ? AND wod_id = ?`, slotID, wodID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("тренировка не найдена в этом дне")
	}
	if _, err := d.Exec(`DELETE FROM wods WHERE id = ?`, wodID); err != nil {
		return err
	}

	remaining, err := slotWodsList(d, slotID)
	if err != nil {
		return err
	}
	if len(remaining) == 0 {
		_, err = d.Exec(`UPDATE cycle_slots SET status = 'empty', wod_id = NULL, template_id = NULL WHERE id = ?`, slotID)
		return err
	}
	if slotWodID.Valid && slotWodID.String == wodID {
		_, err = d.Exec(`UPDATE cycle_slots SET wod_id = ? WHERE id = ?`, remaining[0].ID, slotID)
	}
	return err
}

// assignWarnings — проверки 48 часов и повтора нагрузки (Экран 5).
func assignWarnings(d *sql.DB, groupID, slotID, slotDate, templateID string) []string {
	var out []string
	slotT, err := parseSlotDate(slotDate)
	if err != nil {
		return nil
	}

	var tplIntensity, tplTheme string
	_ = d.QueryRow(`SELECT intensity, theme FROM wod_templates WHERE id = ?`, templateID).
		Scan(&tplIntensity, &tplTheme)

	rows, err := d.Query(`
		SELECT s.slot_date, s.template_id, w.intensity, w.theme, w.name
		FROM cycle_slots s JOIN wods w ON w.id = s.wod_id
		WHERE s.group_id = ? AND s.id != ? AND s.status IN ('planned','in_progress','completed')`,
		groupID, slotID)
	if err != nil {
		return nil
	}
	type near struct {
		date          string
		sameTemplate  bool
		highIntensity bool
		sameTheme     bool
		name          string
	}
	var nears []near
	for rows.Next() {
		var n near
		var date string
		var sTplID, intensity, theme, name sql.NullString
		if rows.Scan(&date, &sTplID, &intensity, &theme, &name) == nil {
			n.date = date
			n.sameTemplate = sTplID.Valid && sTplID.String == templateID
			n.highIntensity = intensity.Valid && intensity.String == "high"
			n.sameTheme = theme.Valid && theme.String == tplTheme
			n.name = name.String
			nears = append(nears, n)
		}
	}
	_ = rows.Close()

	for _, n := range nears {
		t, err := parseSlotDate(n.date)
		if err != nil {
			continue
		}
		diff := t.Sub(slotT)
		if diff < 0 {
			diff = -diff
		}
		if diff.Hours() >= 48 {
			continue
		}
		switch {
		case n.sameTemplate:
			out = append(out, fmt.Sprintf(
				"«%s» уже назначена этой группе %s (менее 48 часов назад)", n.name, n.date))
		case n.highIntensity && tplIntensity == "high":
			out = append(out, fmt.Sprintf(
				"Вчера/завтра у группы высокоинтенсивная «%s» — две тяжёлые подряд не рекомендуются", n.name))
		case n.sameTheme:
			out = append(out, fmt.Sprintf(
				"Похожая нагрузка (%s) была %s — рекомендуем сменить стимул", ThemeName(tplTheme), n.date))
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// UnassignSlot снимает назначение (слот → empty, экземпляр WoD удаляем).
func UnassignSlot(d *sql.DB, slotID string) error {
	var status string
	var wodID sql.NullString
	err := d.QueryRow(`SELECT status, wod_id FROM cycle_slots WHERE id = ?`, slotID).Scan(&status, &wodID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("слот не найден")
	}
	if err != nil {
		return err
	}
	if status == "in_progress" || status == "completed" {
		return fmt.Errorf("нельзя снять назначение с идущей или завершённой тренировки")
	}
	dayWods, err := slotWodsList(d, slotID)
	if err != nil {
		return err
	}
	if _, err := d.Exec(`DELETE FROM slot_wods WHERE slot_id = ?`, slotID); err != nil {
		return err
	}
	if _, err := d.Exec(`UPDATE cycle_slots SET status = 'empty', wod_id = NULL, template_id = NULL WHERE id = ?`, slotID); err != nil {
		return err
	}
	for _, w := range dayWods {
		if _, err := d.Exec(`DELETE FROM wods WHERE id = ?`, w.ID); err != nil {
			return err
		}
	}
	// Легаси: wod_id мог не попасть в slot_wods (до миграции) — подчистим.
	if wodID.Valid {
		if _, err := d.Exec(`DELETE FROM wods WHERE id = ?`, wodID.String); err != nil {
			return err
		}
	}
	return nil
}

// UpdateSlotMovements — корректировка весов/повторов для группы (Экран 5).
func UpdateSlotMovements(d *sql.DB, slotID string, movements []WodVariantMovement) error {
	var wodID sql.NullString
	err := d.QueryRow(`SELECT wod_id FROM cycle_slots WHERE id = ?`, slotID).Scan(&wodID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("слот не найден")
	}
	if err != nil {
		return err
	}
	if !wodID.Valid {
		return fmt.Errorf("слот ещё не заполнен")
	}

	// Читаем текущие строки до транзакции.
	type row struct {
		id  string
		key string
	}
	var rowsData []row
	{
		rows, err := d.Query(`SELECT id, movement_key FROM wod_movements WHERE wod_id = ? ORDER BY sort_order`, wodID.String)
		if err != nil {
			return err
		}
		for rows.Next() {
			var r row
			if rows.Scan(&r.id, &r.key) == nil {
				rowsData = append(rowsData, r)
			}
		}
		_ = rows.Close()
	}

	byKey := map[string]WodVariantMovement{}
	for _, m := range movements {
		byKey[m.MovementKey] = m
	}

	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, r := range rowsData {
		m, ok := byKey[r.key]
		if !ok {
			continue
		}
		var reps, wm, wf any
		if m.Reps != nil {
			reps = *m.Reps
		}
		if m.WeightMale != nil {
			wm = *m.WeightMale
		}
		if m.WeightFemale != nil {
			wf = *m.WeightFemale
		}
		if _, err := tx.Exec(`
			UPDATE wod_movements SET reps = ?, weight_male = ?, weight_female = ? WHERE id = ?`,
			reps, wm, wf, r.id); err != nil {
			return err
		}
	}
	return tx.Commit()
}
