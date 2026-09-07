package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/maxdukov/cf/backend-go/internal/services"
)

// normalizeYo приводит «ё» к «е» — поиск «берпи» находит «Бёрпи».
func normalizeYo(s string) string {
	return strings.ReplaceAll(s, "ё", "е")
}

// ── Wods: генерация и управление тренировками дня ─────────────

type wodGenerateReq struct {
	Theme      string `json:"theme"`
	GroupLevel string `json:"group_level"`
}

type wodSelectReq struct {
	TemplateID string `json:"template_id"`
	GroupLevel string `json:"group_level"`
}

// GenerateWods: POST /api/wods/generate — 3 варианта по теме и уровню.
func (a *App) GenerateWods(w http.ResponseWriter, r *http.Request) {
	var req wodGenerateReq
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.GroupLevel == "" {
		req.GroupLevel = "intermediate"
	}
	variants := services.GenerateWods(a.DB, req.Theme, req.GroupLevel)
	if len(variants) == 0 {
		httpError(w, 404, "Нет подходящих шаблонов для данной темы и инвентаря")
		return
	}
	writeJSON(w, 200, variants)
}

// SelectWod: POST /api/wods/select — создать активный WoD из шаблона.
func (a *App) SelectWod(w http.ResponseWriter, r *http.Request) {
	var req wodSelectReq
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.GroupLevel == "" {
		req.GroupLevel = "intermediate"
	}

	wodID, err := services.CreateWodFromTemplate(a.DB, req.TemplateID, req.GroupLevel)
	if err != nil {
		httpError(w, 404, err.Error())
		return
	}
	out := a.wodToJSON(w, wodID)
	if out == nil {
		httpError(w, 500, "wod disappeared after insert")
		return
	}
	writeJSON(w, 200, out)
}

// GetActiveWod: GET /api/wods/active — текущий активный WoD; если активного
// нет — тренировка, запланированная слотом на дату ?date=YYYY-MM-DD
// (дата берётся из часов клиента), иначе null.
// Ответ содержит source: "active" | "slot".
func (a *App) GetActiveWod(w http.ResponseWriter, r *http.Request) {
	var wodID string
	source := "active"
	err := a.DB.QueryRow(`SELECT id FROM wods WHERE is_active = 1 LIMIT 1`).Scan(&wodID)
	if err == sql.ErrNoRows {
		slotWod, ok := a.scheduledWodForDate(r.URL.Query().Get("date"))
		if !ok {
			writeJSON(w, 200, nil)
			return
		}
		wodID, source = slotWod, "slot"
	} else if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	out := a.wodToJSON(w, wodID)
	if out == nil {
		httpError(w, 500, "active wod query failed")
		return
	}
	out["source"] = source
	writeJSON(w, 200, out)
}

// scheduledWodForDate возвращает WoD запланированного (status='planned')
// слота на указанную дату; при нескольких — из активного цикла, затем по дате.
func (a *App) scheduledWodForDate(date string) (string, bool) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return "", false
	}
	var wodID sql.NullString
	err := a.DB.QueryRow(`
		SELECT s.wod_id FROM cycle_slots s
		JOIN training_cycles c ON c.id = s.cycle_id
		WHERE DATE(s.slot_date) = ? AND s.status = 'planned' AND s.wod_id IS NOT NULL
		ORDER BY CASE c.status WHEN 'active' THEN 0 ELSE 1 END, s.slot_date
		LIMIT 1`, date).Scan(&wodID)
	if err != nil || !wodID.Valid {
		return "", false
	}
	return wodID.String, true
}

// EndActiveWod: POST /api/wods/active/end (204)
func (a *App) EndActiveWod(w http.ResponseWriter, r *http.Request) {
	if _, err := a.DB.Exec(`UPDATE wods SET is_active = 0 WHERE is_active = 1`); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	w.WriteHeader(204)
}

// ListWodHistory: GET /api/wods/history?limit=20
func (a *App) ListWodHistory(w http.ResponseWriter, r *http.Request) {
	limit := intQueryParam(r, "limit", 20)
	var ids []string
	{
		rows, err := a.DB.Query(`SELECT id FROM wods ORDER BY created_at DESC LIMIT ?`, limit)
		if err != nil {
			httpError(w, 500, err.Error())
			return
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err == nil {
				ids = append(ids, id)
			}
		}
		rows.Close() //nolint:errcheck // до вложенных запросов (MaxOpenConns(1))
	}

	out := []map[string]any{}
	for _, id := range ids {
		if wod := a.wodToJSON(w, id); wod != nil {
			out = append(out, wod)
		}
	}
	writeJSON(w, 200, out)
}

// wodToJSON собирает полный JSON WoD с движениями.
func (a *App) wodToJSON(w http.ResponseWriter, wodID string) map[string]any {
	var (
		id          string
		name        string
		format      string
		durationMin int
		intensity   string
		theme       string
		groupLevel  string
		description sql.NullString
		isActive    bool
		createdAt   sql.NullString
	)
	err := a.DB.QueryRow(`
		SELECT id, name, format, duration_min, intensity, theme, group_level, description, COALESCE(is_active,0), created_at
		FROM wods WHERE id = ?`, wodID,
	).Scan(&id, &name, &format, &durationMin, &intensity, &theme, &groupLevel, &description, &isActive, &createdAt)
	if err != nil {
		return nil
	}

	mrows, err := a.DB.Query(`
		SELECT movement_key, movement_name, reps, weight_male, weight_female, sort_order, scaling_note, rounds_note
		FROM wod_movements WHERE wod_id = ? ORDER BY sort_order`, wodID)
	if err != nil {
		return nil
	}
	defer func() { _ = mrows.Close() }()

	movements := []map[string]any{}
	for mrows.Next() {
		var (
			mkey, mname string
			reps        sql.NullInt64
			wm, wf      sql.NullInt64
			sortOrder   int
			scalingNote sql.NullString
			roundsNote  sql.NullString
		)
		if err := mrows.Scan(&mkey, &mname, &reps, &wm, &wf, &sortOrder, &scalingNote, &roundsNote); err != nil {
			return nil
		}
		movements = append(movements, map[string]any{
			"movement_key":  mkey,
			"movement_name": mname,
			"reps":          nullInt64(reps),
			"weight_male":   nullInt64(wm),
			"weight_female": nullInt64(wf),
			"sort_order":    sortOrder,
			"scaling_note":  nullStr(scalingNote),
			"rounds_note":   nullStr(roundsNote),
		})
	}

	var desc any
	if description.Valid {
		desc = description.String
	}
	return map[string]any{
		"id":           id,
		"name":         name,
		"format":       format,
		"duration_min": durationMin,
		"intensity":    intensity,
		"theme":        theme,
		"group_level":  groupLevel,
		"description":  desc,
		"is_active":    isActive,
		"created_at":   isoOrNull(createdAt),
		"movements":    movements,
	}
}

// ── Библиотека шаблонов (Экраны 3–5 draft1.MD) ────────────────

// ListWodTemplates: GET /api/wods/templates?theme=&search=&limit=50&include_archived=1.
// По умолчанию архивные шаблоны скрыты.
func (a *App) ListWodTemplates(w http.ResponseWriter, r *http.Request) {
	theme := r.URL.Query().Get("theme")
	search := r.URL.Query().Get("search")
	limit := intQueryParam(r, "limit", 50)
	if limit <= 0 || limit > 500 {
		limit = 50
	}

	where := "1=1"
	args := []any{}
	if theme != "" {
		where += " AND theme = ?"
		args = append(args, theme)
	}
	if search != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+search+"%")
	}
	if r.URL.Query().Get("include_archived") != "1" {
		where += " AND COALESCE(t.archived, 0) = 0"
	}

	type tplItem struct {
		ID          string
		Name        string
		Format      string
		DurationMin int
		Intensity   string
		Theme       string
		IsBenchmark bool
		MovCount    int
		Archived    bool
		Description sql.NullString
	}
	var items []tplItem
	{
		q := `SELECT t.id, t.name, t.format, t.duration_min, t.intensity, t.theme, COALESCE(t.is_benchmark,0),
		      (SELECT COUNT(*) FROM wod_template_movements m WHERE m.template_id = t.id),
		      COALESCE(t.archived, 0), t.description
		      FROM wod_templates t WHERE ` + where + ` ORDER BY COALESCE(t.is_benchmark,0) DESC, t.name LIMIT ?`
		args = append(args, limit)
		rows, err := a.DB.Query(q, args...)
		if err != nil {
			httpError(w, 500, err.Error())
			return
		}
		for rows.Next() {
			var it tplItem
			var archived int
			if rows.Scan(&it.ID, &it.Name, &it.Format, &it.DurationMin, &it.Intensity, &it.Theme, &it.IsBenchmark, &it.MovCount, &archived, &it.Description) == nil {
				it.Archived = archived == 1
				items = append(items, it)
			}
		}
		_ = rows.Close()
	}

	// Движения шаблонов одним запросом — для карточек как в быстром выборе.
	movs := map[string][]map[string]any{}
	if len(items) > 0 {
		ids := make([]string, len(items))
		for i, it := range items {
			ids[i] = it.ID
		}
		q := `SELECT template_id, movement_key, movement_name, reps, weight_male, weight_female, sort_order, COALESCE(rounds_note,'')
		      FROM wod_template_movements WHERE template_id IN (` + placeholders(len(ids)) + `) ORDER BY sort_order`
		if rows, err := a.DB.Query(q, strSliceToAny(ids)...); err == nil {
			for rows.Next() {
				var tplID, mkey, mname, roundsNote string
				var reps, wm, wf sql.NullInt64
				var sortOrder int
				if rows.Scan(&tplID, &mkey, &mname, &reps, &wm, &wf, &sortOrder, &roundsNote) == nil {
					movs[tplID] = append(movs[tplID], map[string]any{
						"movement_key":  mkey,
						"movement_name": mname,
						"reps":          nullInt64(reps),
						"weight_male":   nullInt64(wm),
						"weight_female": nullInt64(wf),
						"sort_order":    sortOrder,
						"rounds_note":   roundsNote,
					})
				}
			}
			_ = rows.Close()
		}
	}

	out := []map[string]any{}
	for _, it := range items {
		movList := movs[it.ID]
		if movList == nil {
			movList = []map[string]any{}
		}
		out = append(out, map[string]any{
			"template_id":     it.ID,
			"name":            it.Name,
			"format":          it.Format,
			"duration_min":    it.DurationMin,
			"intensity":       it.Intensity,
			"theme":           it.Theme,
			"is_benchmark":    it.IsBenchmark,
			"movements_count": it.MovCount,
			"archived":        it.Archived,
			"description":     nullStr(it.Description),
			"movements":       movList,
		})
	}
	writeJSON(w, 200, out)
}

// placeholders — «?, ?, ?» для IN (…).
func placeholders(n int) string {
	return strings.TrimSuffix(strings.Repeat("?,", n), ",")
}

// strSliceToAny — []string → []any для sql-аргументов.
func strSliceToAny(xs []string) []any {
	out := make([]any, len(xs))
	for i, x := range xs {
		out[i] = x
	}
	return out
}

// GetWodTemplate: GET /api/wods/templates/{template_id} — полная карточка.
// ?raw=1 — исходный шаблон без скалирования (режим правки в конструкторе).
func (a *App) GetWodTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "template_id")
	if r.URL.Query().Get("raw") == "1" {
		tpl := services.GetTemplateForEdit(a.DB, id)
		if tpl == nil {
			httpError(w, 404, "Шаблон не найден")
			return
		}
		writeJSON(w, 200, tpl)
		return
	}
	level := r.URL.Query().Get("group_level")
	if level == "" {
		level = "intermediate"
	}
	variants := services.BuildTemplatePreview(a.DB, id, level)
	if variants == nil {
		httpError(w, 404, "Шаблон не найден")
		return
	}
	writeJSON(w, 200, variants)
}

// UpdateWodTemplate: PUT /api/wods/templates/{template_id} — правка из библиотеки.
func (a *App) UpdateWodTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "template_id")
	var req services.CustomWodInput
	if !decodeJSON(w, r, &req) {
		return
	}
	warnings, err := services.UpdateCustomWod(a.DB, id, req)
	if err != nil {
		httpError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"template_id": id, "warnings": warnings})
}

// DeleteWodTemplate: DELETE /api/wods/templates/{template_id}.
// Шаблон в запланированных слотах — 409 со списком циклов для правки плана.
// Выполнявшийся в истории шаблон не удаляется, а помечается архивным.
// Иначе — удаляется безвозвратно (надгробие защищает от сид-refresh).
func (a *App) DeleteWodTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "template_id")

	var name string
	err := a.DB.QueryRow(`SELECT name FROM wod_templates WHERE id = ?`, id).Scan(&name)
	if err == sql.ErrNoRows {
		httpError(w, 404, "Шаблон не найден")
		return
	}
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}

	// Включён в запланированные слоты (основная тренировка дня или доп.).
	type cycleRef struct {
		ID       string `json:"cycle_id"`
		Name     string `json:"cycle_name"`
		SlotID   string `json:"slot_id"`
		SlotDate string `json:"slot_date"`
	}
	planRows, err := a.DB.Query(`
		SELECT c.id, c.name, s.id, s.slot_date FROM cycle_slots s
		JOIN wods w ON w.id = s.wod_id
		JOIN training_cycles c ON c.id = s.cycle_id
		WHERE w.template_id = ? AND s.status = 'planned'
		UNION
		SELECT c.id, c.name, s.id, s.slot_date FROM slot_wods sw
		JOIN wods w ON w.id = sw.wod_id
		JOIN cycle_slots s ON s.id = sw.slot_id
		JOIN training_cycles c ON c.id = s.cycle_id
		WHERE w.template_id = ? AND s.status = 'planned'`, id, id)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	cycles := []cycleRef{}
	for planRows.Next() {
		var c cycleRef
		if planRows.Scan(&c.ID, &c.Name, &c.SlotID, &c.SlotDate) == nil {
			cycles = append(cycles, c)
		}
	}
	_ = planRows.Close()
	if len(cycles) > 0 {
		writeJSON(w, 409, map[string]any{
			"detail": "тренировка включена в план цикла — сначала уберите её из запланированных дней",
			"cycles": cycles,
		})
		return
	}

	// Выполнялся ли когда-либо: проведённая сессия или завершённый слот.
	performed := false
	if err := a.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM wods WHERE template_id = ? AND session_id IS NOT NULL
		) OR EXISTS(
			SELECT 1 FROM cycle_slots s JOIN wods w ON w.id = s.wod_id
			WHERE w.template_id = ? AND s.status = 'completed'
		) OR EXISTS(
			SELECT 1 FROM slot_wods sw JOIN wods w ON w.id = sw.wod_id
			JOIN cycle_slots s ON s.id = sw.slot_id
			WHERE w.template_id = ? AND s.status = 'completed'
		)`, id, id, id).Scan(&performed); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if performed {
		if _, err := a.DB.Exec(`UPDATE wod_templates SET archived = 1 WHERE id = ?`, id); err != nil {
			httpError(w, 500, err.Error())
			return
		}
		writeJSON(w, 200, map[string]any{"archived": true})
		return
	}

	if _, err := a.DB.Exec(`DELETE FROM wod_templates WHERE id = ?`, id); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	// Надгробие: seed-refresh не воскресит шаблон с таким именем.
	if _, err := a.DB.Exec(`INSERT OR IGNORE INTO wod_templates_deleted (name) VALUES (?)`, name); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"deleted": true})
}

// RestoreWodTemplate: POST /api/wods/templates/{template_id}/restore — из архива.
func (a *App) RestoreWodTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "template_id")
	res, err := a.DB.Exec(`UPDATE wod_templates SET archived = 0 WHERE id = ?`, id)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httpError(w, 404, "Шаблон не найден")
		return
	}
	writeJSON(w, 200, map[string]any{"restored": true})
}

// CreateCustomWod: POST /api/wods/custom — конструктор (Экран 4).
func (a *App) CreateCustomWod(w http.ResponseWriter, r *http.Request) {
	var req services.CustomWodInput
	if !decodeJSON(w, r, &req) {
		return
	}
	tplID, warnings, err := services.CreateCustomWod(a.DB, req)
	if err != nil {
		httpError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{
		"template_id": tplID,
		"warnings":    warnings,
	})
}

// ListMovements: GET /api/movements?search= — каталог для конструктора.
func (a *App) ListMovements(w http.ResponseWriter, r *http.Request) {
	search := normalizeYo(strings.ToLower(r.URL.Query().Get("search")))
	q := `SELECT "key", name, modality, muscle_group, themes, equipment_keys, difficulty, scaling_beginner, scaling_intermediate, COALESCE(is_custom, 0)
	      FROM movements ORDER BY name LIMIT 200`

	rows, err := a.DB.Query(q)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	defer func() { _ = rows.Close() }()

	out := []map[string]any{}
	for rows.Next() {
		var key, name, modality, muscle, themes, eq, difficulty string
		var sb, si sql.NullString
		var isCustom int
		if rows.Scan(&key, &name, &modality, &muscle, &themes, &eq, &difficulty, &sb, &si, &isCustom) == nil {
			// Поиск без учёта «ё» и регистра: «берпи» находит «Бёрпи».
			if search != "" &&
				!strings.Contains(normalizeYo(strings.ToLower(name)), search) &&
				!strings.Contains(normalizeYo(strings.ToLower(key)), search) {
				continue
			}
			out = append(out, map[string]any{
				"key":                  key,
				"name":                 name,
				"modality":             modality,
				"muscle_group":         muscle,
				"themes":               themes,
				"equipment_keys":       eq,
				"difficulty":           difficulty,
				"scaling_beginner":     nullStr(sb),
				"scaling_intermediate": nullStr(si),
				"is_custom":            isCustom == 1,
			})
		}
	}
	writeJSON(w, 200, out)
}

// ── System: режим коллектора (mock / ant) ─────────────────────

type modeReq struct {
	Mode string `json:"mode"`
}

// GetMode: GET /api/system/mode
func (a *App) GetMode(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"mode": a.Mode()})
}

// SetMode: POST /api/system/mode — переключить mock/ant.
func (a *App) SetMode(w http.ResponseWriter, r *http.Request) {
	var req modeReq
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Mode != "mock" && req.Mode != "ant" {
		httpError(w, 400, "mode must be 'mock' or 'ant'")
		return
	}
	a.SwitchCollector(req.Mode)
	writeJSON(w, 200, map[string]string{"mode": req.Mode})
}

// ── Health ────────────────────────────────────────────────────

// Health: GET /api/health — проверка бэкенда и БД.
func (a *App) Health(w http.ResponseWriter, r *http.Request) {
	if err := a.DB.Ping(); err != nil {
		writeJSON(w, 503, map[string]string{"status": "error", "detail": "database unavailable"})
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

// DBStats exposes sql.DB pool counters; used to detect connection leaks.
func (a *App) DBStats(w http.ResponseWriter, r *http.Request) {
	s := a.DB.Stats()
	writeJSON(w, 200, map[string]any{
		"in_use": s.InUse, "open": s.OpenConnections, "idle": s.Idle,
		"wait_count": s.WaitCount, "wait_duration_ms": s.WaitDuration.Milliseconds(),
	})
}
