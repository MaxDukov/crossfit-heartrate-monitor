package handlers

import (
	"database/sql"
	"net/http"
)

// ── Sensors: список, привязка/отвязка, игнорирование ──────────

type sensorAssignReq struct {
	AthleteID string `json:"athlete_id"`
}

type sensorRow struct {
	DeviceID     int            `db:"device_id"`
	AthleteID    sql.NullString `db:"athlete_id"`
	AthleteName  sql.NullString `db:"athlete_name"`
	LastHr       sql.NullInt64  `db:"last_hr"`
	LastSeenAt   sql.NullString `db:"last_seen_at"`
	BatteryLevel sql.NullInt64  `db:"battery_level"`
	Ignored      bool           `db:"ignored"`
}

const sensorSelect = `
	SELECT s.device_id, s.athlete_id, a.name AS athlete_name,
	       s.last_hr, s.last_seen_at, s.battery_level, COALESCE(s.ignored, 0) AS ignored
	FROM sensors s
	LEFT JOIN athletes a ON a.id = s.athlete_id`

func (s sensorRow) toJSON(a *App) map[string]any {
	return map[string]any{
		"device_id":     s.DeviceID,
		"athlete_id":    nullStr(s.AthleteID),
		"athlete_name":  nullStr(s.AthleteName),
		"last_hr":       nullInt64(s.LastHr),
		"last_seen_at":  isoOrNull(s.LastSeenAt),
		"battery_level": nullInt64(s.BatteryLevel),
		"ignored":       s.Ignored,
	}
}

func (a *App) getSensorRow(w http.ResponseWriter, deviceID int, notFoundMsg string) *sensorRow {
	var row sensorRow
	err := a.DB.QueryRow(sensorSelect+` WHERE s.device_id = ?`, deviceID).
		Scan(&row.DeviceID, &row.AthleteID, &row.AthleteName, &row.LastHr, &row.LastSeenAt, &row.BatteryLevel, &row.Ignored)
	if err == sql.ErrNoRows {
		httpError(w, 404, notFoundMsg)
		return nil
	}
	if err != nil {
		httpError(w, 500, err.Error())
		return nil
	}
	return &row
}

// ListSensors: GET /api/sensors
func (a *App) ListSensors(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.Query(sensorSelect + ` ORDER BY s.last_seen_at DESC`)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	defer func() { _ = rows.Close() }()

	out := []map[string]any{}
	for rows.Next() {
		var row sensorRow
		if err := rows.Scan(&row.DeviceID, &row.AthleteID, &row.AthleteName, &row.LastHr, &row.LastSeenAt, &row.BatteryLevel, &row.Ignored); err != nil {
			httpError(w, 500, err.Error())
			return
		}
		out = append(out, row.toJSON(a))
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

// AssignSensor: POST /api/sensors/{device_id}/assign
func (a *App) AssignSensor(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := deviceIDParam(w, r)
	if !ok {
		return
	}

	row := a.getSensorRow(w, deviceID, "Датчик не найден")
	if row == nil {
		return
	}
	if row.Ignored {
		httpError(w, 400, "Датчик проигнорирован — верните его в активные")
		return
	}

	var req sensorAssignReq
	if !decodeJSON(w, r, &req) {
		return
	}

	var athleteID string
	err := a.DB.QueryRow(`SELECT id FROM athletes WHERE id = ?`, req.AthleteID).Scan(&athleteID)
	if err == sql.ErrNoRows {
		httpError(w, 404, "Спортсмен не найден")
		return
	}
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}

	tx, err := a.DB.Begin()
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Взаимно-исключающая привязка: старый датчик спортсмена отвязываем.
	if _, err := tx.Exec(`UPDATE sensors SET athlete_id = NULL WHERE athlete_id = ? AND device_id != ?`, athleteID, deviceID); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if _, err := tx.Exec(`UPDATE sensors SET athlete_id = ? WHERE device_id = ?`, athleteID, deviceID); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if err := tx.Commit(); err != nil {
		httpError(w, 500, err.Error())
		return
	}

	row = a.getSensorRowOr500(w, deviceID)
	if row == nil {
		return
	}
	writeJSON(w, 200, row.toJSON(a))
}

// UnassignSensor: DELETE /api/sensors/{device_id}/assign
func (a *App) UnassignSensor(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := deviceIDParam(w, r)
	if !ok {
		return
	}
	row := a.getSensorRow(w, deviceID, "Датчик не найден")
	if row == nil {
		return
	}
	if _, err := a.DB.Exec(`UPDATE sensors SET athlete_id = NULL WHERE device_id = ?`, deviceID); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	row = a.getSensorRowOr500(w, deviceID)
	if row == nil {
		return
	}
	writeJSON(w, 200, row.toJSON(a))
}

// IgnoreSensor: POST /api/sensors/{device_id}/ignore
func (a *App) IgnoreSensor(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := deviceIDParam(w, r)
	if !ok {
		return
	}
	row := a.getSensorRow(w, deviceID, "Датчик не найден")
	if row == nil {
		return
	}
	if _, err := a.DB.Exec(`UPDATE sensors SET ignored = 1, athlete_id = NULL WHERE device_id = ?`, deviceID); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	row = a.getSensorRowOr500(w, deviceID)
	if row == nil {
		return
	}
	writeJSON(w, 200, row.toJSON(a))
}

// UnignoreSensor: POST /api/sensors/{device_id}/unignore
func (a *App) UnignoreSensor(w http.ResponseWriter, r *http.Request) {
	deviceID, ok := deviceIDParam(w, r)
	if !ok {
		return
	}
	row := a.getSensorRow(w, deviceID, "Датчик не найден")
	if row == nil {
		return
	}
	if _, err := a.DB.Exec(`UPDATE sensors SET ignored = 0 WHERE device_id = ?`, deviceID); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	row = a.getSensorRowOr500(w, deviceID)
	if row == nil {
		return
	}
	writeJSON(w, 200, row.toJSON(a))
}
