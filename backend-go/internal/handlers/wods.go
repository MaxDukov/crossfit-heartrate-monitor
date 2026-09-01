package handlers

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/maxdukov/cf/backend-go/internal/services"
)

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

// GetActiveWod: GET /api/wods/active — текущий активный WoD или null.
func (a *App) GetActiveWod(w http.ResponseWriter, r *http.Request) {
	var wodID string
	err := a.DB.QueryRow(`SELECT id FROM wods WHERE is_active = 1 LIMIT 1`).Scan(&wodID)
	if err == sql.ErrNoRows {
		writeJSON(w, 200, nil)
		return
	}
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	out := a.wodToJSON(w, wodID)
	if out == nil {
		httpError(w, 500, "active wod query failed")
		return
	}
	writeJSON(w, 200, out)
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

// ListWodTemplates: GET /api/wods/templates?theme=&search=&limit=50.
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

	type tplItem struct {
		ID          string
		Name        string
		Format      string
		DurationMin int
		Intensity   string
		Theme       string
		IsBenchmark bool
		MovCount    int
	}
	var items []tplItem
	{
		q := `SELECT t.id, t.name, t.format, t.duration_min, t.intensity, t.theme, COALESCE(t.is_benchmark,0),
		      (SELECT COUNT(*) FROM wod_template_movements m WHERE m.template_id = t.id)
		      FROM wod_templates t WHERE ` + where + ` ORDER BY COALESCE(t.is_benchmark,0) DESC, t.name LIMIT ?`
		args = append(args, limit)
		rows, err := a.DB.Query(q, args...)
		if err != nil {
			httpError(w, 500, err.Error())
			return
		}
		for rows.Next() {
			var it tplItem
			if rows.Scan(&it.ID, &it.Name, &it.Format, &it.DurationMin, &it.Intensity, &it.Theme, &it.IsBenchmark, &it.MovCount) == nil {
				items = append(items, it)
			}
		}
		_ = rows.Close()
	}

	out := []map[string]any{}
	for _, it := range items {
		out = append(out, map[string]any{
			"template_id":     it.ID,
			"name":            it.Name,
			"format":          it.Format,
			"duration_min":    it.DurationMin,
			"intensity":       it.Intensity,
			"theme":           it.Theme,
			"is_benchmark":    it.IsBenchmark,
			"movements_count": it.MovCount,
		})
	}
	writeJSON(w, 200, out)
}

// GetWodTemplate: GET /api/wods/templates/{template_id} — полная карточка.
func (a *App) GetWodTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "template_id")
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
	search := r.URL.Query().Get("search")
	q := `SELECT "key", name, modality, muscle_group, themes, equipment_keys, difficulty, scaling_beginner, scaling_intermediate
	      FROM movements`
	args := []any{}
	if search != "" {
		q += ` WHERE name LIKE ? OR "key" LIKE ?`
		args = append(args, "%"+search+"%", "%"+search+"%")
	}
	q += ` ORDER BY name LIMIT 200`

	rows, err := a.DB.Query(q, args...)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	defer func() { _ = rows.Close() }()

	out := []map[string]any{}
	for rows.Next() {
		var key, name, modality, muscle, themes, eq, difficulty string
		var sb, si sql.NullString
		if rows.Scan(&key, &name, &modality, &muscle, &themes, &eq, &difficulty, &sb, &si) == nil {
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
