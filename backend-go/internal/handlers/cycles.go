package handlers

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/maxdukov/cf/backend-go/internal/services"
)

// ── Циклы планирования (draft1.MD: Экраны 1–2) ────────────────

// CreateCycle: POST /api/cycles — создать цикл со слотами.
func (a *App) CreateCycle(w http.ResponseWriter, r *http.Request) {
	var req services.CreateCycleInput
	if !decodeJSON(w, r, &req) {
		return
	}
	id, warnings, err := services.CreateCycle(a.DB, req)
	if err != nil {
		httpError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"id": id, "warnings": warnings})
}

// ListCycles: GET /api/cycles.
func (a *App) ListCycles(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, services.ListCycles(a.DB))
}

// GetCycle: GET /api/cycles/{cycle_id} — с группами и календарём.
func (a *App) GetCycle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "cycle_id")
	c := services.GetCycle(a.DB, id)
	if c == nil {
		httpError(w, 404, "Цикл не найден")
		return
	}
	writeJSON(w, 200, c)
}

type cycleStatusReq struct {
	Status string `json:"status"`
}

// UpdateCycleStatus: PUT /api/cycles/{cycle_id}/status.
func (a *App) UpdateCycleStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "cycle_id")
	var req cycleStatusReq
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := services.UpdateCycleStatus(a.DB, id, req.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpError(w, 404, "Цикл не найден")
			return
		}
		httpError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"id": id, "status": req.Status})
}

// DeleteCycle: DELETE /api/cycles/{cycle_id} (204).
func (a *App) DeleteCycle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "cycle_id")
	res, err := a.DB.Exec(`DELETE FROM training_cycles WHERE id = ?`, id)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httpError(w, 404, "Цикл не найден")
		return
	}
	w.WriteHeader(204)
}

// GetCycleAnalytics: GET /api/cycles/{cycle_id}/analytics (Экран 7).
func (a *App) GetCycleAnalytics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "cycle_id")
	res, err := services.GetCycleAnalytics(a.DB, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpError(w, 404, "Цикл не найден")
			return
		}
		httpError(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, res)
}

// normSlotDate — SQLite может вернуть DATE с временем; режем до YYYY-MM-DD.
func normSlotDate(s string) string {
	if len(s) > 10 {
		return s[:10]
	}
	return s
}

// ── Слоты (Экраны 3, 5, 6) ────────────────────────────────────

// GetSlot: GET /api/slots/{slot_id} — карточка слота с WoD и результатами.
func (a *App) GetSlot(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "slot_id")
	var (
		cycleID, groupID, groupName, slotDate, status string
		dayNumber                                     int
		templateID, notes, sessionID                  sql.NullString
		wodID                                         sql.NullString
		kind                                          string
	)
	err := a.DB.QueryRow(`
		SELECT s.cycle_id, s.group_id, g.name, s.slot_date, s.day_number, s.status,
		       COALESCE(s.kind, 'regular'), s.template_id, s.notes, s.wod_id, s.session_id
		FROM cycle_slots s JOIN cycle_groups g ON g.id = s.group_id
		WHERE s.id = ?`, id,
	).Scan(&cycleID, &groupID, &groupName, &slotDate, &dayNumber, &status, &kind, &templateID, &notes, &wodID, &sessionID)
	if err == sql.ErrNoRows {
		httpError(w, 404, "Слот не найден")
		return
	}
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	slotDate = normSlotDate(slotDate)

	out := map[string]any{
		"id":          id,
		"cycle_id":    cycleID,
		"group_id":    groupID,
		"group_name":  groupName,
		"slot_date":   slotDate,
		"day_number":  dayNumber,
		"status":      status,
		"kind":        kind,
		"template_id": nullStr(templateID),
		"wod_id":      nullStr(wodID),
		"notes":       nullStr(notes),
		"session_id":  nullStr(sessionID),
	}

	if wodID.Valid {
		out["wod"] = a.wodToJSON(w, wodID.String)
	} else {
		out["wod"] = nil
	}

	// Все тренировки дня (может быть несколько в пределах 60 минут).
	if list, err := services.SlotWods(a.DB, id); err == nil {
		out["wods"] = list
		total := 0
		for _, it := range list {
			total += it.DurationMin
		}
		out["wods_total_min"] = total
	}

	// Участники live-сессии (Экран 6).
	if sessionID.Valid {
		var ids []string
		rows, err := a.DB.Query(`
			SELECT athlete_id FROM session_athletes
			WHERE session_id = ? AND left_at IS NULL`, sessionID.String)
		if err == nil {
			for rows.Next() {
				var aid string
				if rows.Scan(&aid) == nil {
					ids = append(ids, aid)
				}
			}
			_ = rows.Close()
		}
		athletes := []map[string]any{}
		for _, aid := range ids {
			var name string
			var maxHr int
			if a.DB.QueryRow(`SELECT name, max_hr FROM athletes WHERE id = ?`, aid).Scan(&name, &maxHr) == nil {
				athletes = append(athletes, map[string]any{"id": aid, "name": name, "max_hr": maxHr})
			}
		}
		out["participants"] = athletes
	} else {
		out["participants"] = []map[string]any{}
	}

	out["results"] = services.ListSlotResults(a.DB, id)
	writeJSON(w, 200, out)
}

// SlotRecommendations: GET /api/slots/{slot_id}/recommendations (Экран 3).
func (a *App) SlotRecommendations(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "slot_id")
	level := r.URL.Query().Get("group_level")
	recs, err := services.SlotRecommendations(a.DB, id, level)
	if err != nil {
		httpError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, recs)
}

type slotAssignReq struct {
	TemplateID string `json:"template_id"`
	GroupLevel string `json:"group_level"`
}

// AssignSlot: POST /api/slots/{slot_id}/assign — назначить шаблон (Экран 5).
func (a *App) AssignSlot(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "slot_id")
	var req slotAssignReq
	if !decodeJSON(w, r, &req) {
		return
	}
	wodID, warnings, err := services.AssignTemplate(a.DB, id, req.TemplateID, req.GroupLevel)
	if err != nil {
		httpError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{
		"slot_id":  id,
		"wod_id":   wodID,
		"warnings": warnings,
	})
}

// UnassignSlot: DELETE /api/slots/{slot_id}/assign (204).
func (a *App) UnassignSlot(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "slot_id")
	if err := services.UnassignSlot(a.DB, id); err != nil {
		httpError(w, 400, err.Error())
		return
	}
	w.WriteHeader(204)
}

// UnassignSlotWod: DELETE /api/slots/{slot_id}/assign/{wod_id} — снять
// одну тренировку из дня (204).
func (a *App) UnassignSlotWod(w http.ResponseWriter, r *http.Request) {
	slotID := chi.URLParam(r, "slot_id")
	wodID := chi.URLParam(r, "wod_id")
	if err := services.UnassignSlotWod(a.DB, slotID, wodID); err != nil {
		httpError(w, 400, err.Error())
		return
	}
	w.WriteHeader(204)
}

type slotMovementsReq struct {
	Movements []services.WodVariantMovement `json:"movements"`
}

// UpdateSlotMovements: PUT /api/slots/{slot_id}/movements — веса по группе (Экран 5).
func (a *App) UpdateSlotMovements(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "slot_id")
	var req slotMovementsReq
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := services.UpdateSlotMovements(a.DB, id, req.Movements); err != nil {
		httpError(w, 400, err.Error())
		return
	}
	var wodID sql.NullString
	_ = a.DB.QueryRow(`SELECT wod_id FROM cycle_slots WHERE id = ?`, id).Scan(&wodID)
	if wodID.Valid {
		writeJSON(w, 200, a.wodToJSON(w, wodID.String))
		return
	}
	w.WriteHeader(204)
}

// SetGroupThirdDayOff: PUT /api/cycles/{cycle_id}/groups/{group_id}/third-day-off.
// При включении с конфликтами (назначенные третьи дни) — 409 и список конфликтов.
func (a *App) SetGroupThirdDayOff(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "group_id")
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	applied, conflicts, err := services.SetThirdDayOff(a.DB, groupID, req.Enabled)
	if err != nil {
		httpError(w, 400, err.Error())
		return
	}
	if !applied {
		writeJSON(w, 409, map[string]any{
			"detail":    "Третий день был включён в тренировочный цикл. Начать перепланирование?",
			"conflicts": conflicts,
		})
		return
	}
	writeJSON(w, 200, map[string]any{"applied": true})
}

// ReplanGroupThirdDayOff: POST /api/cycles/{cycle_id}/groups/{group_id}/third-day-off/replan.
func (a *App) ReplanGroupThirdDayOff(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "group_id")
	var req struct {
		Mode string `json:"mode"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	res, err := services.ReplanThirdDayOff(a.DB, groupID, req.Mode)
	if err != nil {
		httpError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, res)
}

// StartSlot: POST /api/slots/{slot_id}/start — начать тренировку (Экран 6).
func (a *App) StartSlot(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "slot_id")
	sessionID, err := services.StartSlot(a.DB, id)
	if err != nil {
		httpError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"slot_id": id, "session_id": sessionID})
}

// CompleteSlot: POST /api/slots/{slot_id}/complete (204).
func (a *App) CompleteSlot(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "slot_id")
	if err := services.CompleteSlot(a.DB, id); err != nil {
		httpError(w, 400, err.Error())
		return
	}
	w.WriteHeader(204)
}

// SaveSlotResult: POST /api/slots/{slot_id}/results — логирование (Экран 6).
func (a *App) SaveSlotResult(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "slot_id")
	var req services.SaveResultInput
	if !decodeJSON(w, r, &req) {
		return
	}
	resp, err := services.SaveResult(a.DB, id, req)
	if err != nil {
		httpError(w, 400, err.Error())
		return
	}
	writeJSON(w, 201, resp)
}

// ListSlotResults: GET /api/slots/{slot_id}/results.
func (a *App) ListSlotResults(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "slot_id")
	writeJSON(w, 200, services.ListSlotResults(a.DB, id))
}
