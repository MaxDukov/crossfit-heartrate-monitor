package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/maxdukov/cf/backend-go/internal/db"
)

// ── Sessions: жизненный цикл тренировочных сессий ─────────────

type sessionCreateReq struct {
	Name *string `json:"name"`
}

type sessionAthleteAddReq struct {
	AthleteID string `json:"athlete_id"`
}

type sessionRow struct {
	ID         string         `db:"id"`
	Name       sql.NullString `db:"name"`
	StartedAt  sql.NullString `db:"started_at"`
	EndedAt    sql.NullString `db:"ended_at"`
	AthleteCnt int            `db:"athlete_count"`
}

const sessionSelect = `
	SELECT s.id, s.name, s.started_at, s.ended_at,
	       (SELECT COUNT(*) FROM session_athletes sa WHERE sa.session_id = s.id AND sa.left_at IS NULL) AS athlete_count
	FROM sessions s`

func (s sessionRow) toJSON() map[string]any {
	return map[string]any{
		"id":            s.ID,
		"name":          nullStr(s.Name),
		"started_at":    isoOrNull(s.StartedAt),
		"ended_at":      isoOrNull(s.EndedAt),
		"athlete_count": s.AthleteCnt,
	}
}

// ListSessions: GET /api/sessions?limit=50
func (a *App) ListSessions(w http.ResponseWriter, r *http.Request) {
	limit := intQueryParam(r, "limit", 50)
	rows, err := a.DB.Query(sessionSelect+` ORDER BY s.started_at DESC LIMIT ?`, limit)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	defer func() { _ = rows.Close() }()

	out := []map[string]any{}
	for rows.Next() {
		var row sessionRow
		if err := rows.Scan(&row.ID, &row.Name, &row.StartedAt, &row.EndedAt, &row.AthleteCnt); err != nil {
			httpError(w, 500, err.Error())
			return
		}
		out = append(out, row.toJSON())
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

// CreateSession: POST /api/sessions (201; 400, если уже есть активная)
func (a *App) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req sessionCreateReq
	if !decodeJSON(w, r, &req) {
		return
	}

	var activeID string
	err := a.DB.QueryRow(`SELECT id FROM sessions WHERE ended_at IS NULL LIMIT 1`).Scan(&activeID)
	if err == nil {
		httpError(w, 400, "Уже есть активная сессия — сначала завершите её")
		return
	}
	if err != sql.ErrNoRows {
		httpError(w, 500, err.Error())
		return
	}

	id := db.NewUUID()
	var name any
	if req.Name != nil {
		name = *req.Name
	}
	now := db.NowDB()
	if _, err := a.DB.Exec(`INSERT INTO sessions (id, name, started_at) VALUES (?, ?, ?)`, id, name, now); err != nil {
		httpError(w, 500, err.Error())
		return
	}

	var row sessionRow
	if err := a.DB.QueryRow(sessionSelect+` WHERE s.id = ?`, id).
		Scan(&row.ID, &row.Name, &row.StartedAt, &row.EndedAt, &row.AthleteCnt); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, row.toJSON())
}

// GetActiveSession: GET /api/sessions/active (null, если нет активной)
func (a *App) GetActiveSession(w http.ResponseWriter, r *http.Request) {
	var row sessionRow
	err := a.DB.QueryRow(sessionSelect+` WHERE s.ended_at IS NULL LIMIT 1`).
		Scan(&row.ID, &row.Name, &row.StartedAt, &row.EndedAt, &row.AthleteCnt)
	if err == sql.ErrNoRows {
		writeJSON(w, 200, nil)
		return
	}
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, row.toJSON())
}

// EndSession: POST /api/sessions/{session_id}/end
func (a *App) EndSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "session_id")

	var endedAt sql.NullString
	err := a.DB.QueryRow(`SELECT ended_at FROM sessions WHERE id = ?`, id).Scan(&endedAt)
	if err == sql.ErrNoRows {
		httpError(w, 404, "Сессия не найдена")
		return
	}
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if endedAt.Valid {
		httpError(w, 400, "Сессия уже завершена")
		return
	}

	now := db.NowDB()
	tx, err := a.DB.Begin()
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`UPDATE sessions SET ended_at = ? WHERE id = ?`, now, id); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if _, err := tx.Exec(`UPDATE session_athletes SET left_at = ? WHERE session_id = ? AND left_at IS NULL`, now, id); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		httpError(w, 500, err.Error())
		return
	}

	var row sessionRow
	if err := a.DB.QueryRow(sessionSelect+` WHERE s.id = ?`, id).
		Scan(&row.ID, &row.Name, &row.StartedAt, &row.EndedAt, &row.AthleteCnt); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, row.toJSON())
}

// AddAthleteToSession: POST /api/sessions/{session_id}/athletes (201)
func (a *App) AddAthleteToSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "session_id")

	var req sessionAthleteAddReq
	if !decodeJSON(w, r, &req) {
		return
	}

	var endedAt sql.NullString
	err := a.DB.QueryRow(`SELECT ended_at FROM sessions WHERE id = ?`, sessionID).Scan(&endedAt)
	if err == sql.ErrNoRows {
		httpError(w, 404, "Сессия не найдена")
		return
	}
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if endedAt.Valid {
		httpError(w, 400, "Сессия уже завершена")
		return
	}

	var athleteID string
	err = a.DB.QueryRow(`SELECT id FROM athletes WHERE id = ?`, req.AthleteID).Scan(&athleteID)
	if err == sql.ErrNoRows {
		httpError(w, 404, "Спортсмен не найден")
		return
	}
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}

	var exists string
	err = a.DB.QueryRow(
		`SELECT id FROM session_athletes WHERE session_id = ? AND athlete_id = ? AND left_at IS NULL`,
		sessionID, athleteID).Scan(&exists)
	if err == nil {
		httpError(w, 400, "Спортсмен уже в сессии")
		return
	}
	if err != sql.ErrNoRows {
		httpError(w, 500, err.Error())
		return
	}

	if _, err := a.DB.Exec(
		`INSERT INTO session_athletes (id, session_id, athlete_id, joined_at) VALUES (?, ?, ?, ?)`,
		db.NewUUID(), sessionID, athleteID, db.NowDB(),
	); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, map[string]string{"status": "added"})
}

// RemoveAthleteFromSession: DELETE /api/sessions/{session_id}/athletes/{athlete_id} (204)
func (a *App) RemoveAthleteFromSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "session_id")
	athleteID := chi.URLParam(r, "athlete_id")

	res, err := a.DB.Exec(
		`UPDATE session_athletes SET left_at = ? WHERE session_id = ? AND athlete_id = ? AND left_at IS NULL`,
		db.NowDB(), sessionID, athleteID,
	)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httpError(w, 404, "Спортсмен не найден в сессии")
		return
	}
	w.WriteHeader(204)
}

// ── helpers (общие) ───────────────────────────────────────────

func deviceIDParam(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(chi.URLParam(r, "device_id"))
	if err != nil {
		httpError(w, 422, "device_id должен быть целым числом")
		return 0, false
	}
	return id, true
}

func intQueryParam(r *http.Request, name string, def int) int {
	if v := r.URL.Query().Get(name); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
