package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/maxdukov/cf/backend-go/internal/db"
)

// ── Athletes: CRUD спортсменов ────────────────────────────────

type athleteCreate struct {
	Name     string   `json:"name"`
	MaxHr    *int     `json:"max_hr"`
	WeightKg *float64 `json:"weight_kg"`
	Age      *int     `json:"age"`
}

type athleteUpdate struct {
	Name     *string  `json:"name"`
	MaxHr    *int     `json:"max_hr"`
	WeightKg *float64 `json:"weight_kg"`
	Age      *int     `json:"age"`
}

// athleteRow — строка таблицы athletes (даты — строки в формате SQLAlchemy).
type athleteRow struct {
	ID        string          `db:"id"`
	Name      string          `db:"name"`
	MaxHr     int             `db:"max_hr"`
	WeightKg  sql.NullFloat64 `db:"weight_kg"`
	Age       sql.NullInt64   `db:"age"`
	CreatedAt sql.NullString  `db:"created_at"`
	UpdatedAt sql.NullString  `db:"updated_at"`
}

func (a athleteRow) toJSON() map[string]any {
	return map[string]any{
		"id":         a.ID,
		"name":       a.Name,
		"max_hr":     a.MaxHr,
		"weight_kg":  nullFloat(a.WeightKg),
		"age":        nullInt64(a.Age),
		"created_at": db.ISO(a.CreatedAt.String),
		"updated_at": db.ISO(a.UpdatedAt.String),
	}
}

func validateAthleteCreate(w http.ResponseWriter, d athleteCreate) bool {
	d.Name = strings.TrimSpace(d.Name)
	if len(d.Name) < 1 || len(d.Name) > 100 {
		httpError(w, 422, "name: длина должна быть от 1 до 100 символов")
		return false
	}
	maxHr := 190
	if d.MaxHr != nil {
		maxHr = *d.MaxHr
	}
	if maxHr < 60 || maxHr > 250 {
		httpError(w, 422, "max_hr: значение должно быть от 60 до 250")
		return false
	}
	if d.WeightKg != nil && (*d.WeightKg < 30 || *d.WeightKg > 250) {
		httpError(w, 422, "weight_kg: значение должно быть от 30 до 250")
		return false
	}
	if d.Age != nil && (*d.Age < 10 || *d.Age > 100) {
		httpError(w, 422, "age: значение должно быть от 10 до 100")
		return false
	}
	return true
}

// ListAthletes: GET /api/athletes
func (a *App) ListAthletes(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.Query(`SELECT id, name, max_hr, weight_kg, age, created_at, updated_at
		FROM athletes ORDER BY name`)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	defer func() { _ = rows.Close() }()

	out := []map[string]any{}
	for rows.Next() {
		var row athleteRow
		if err := scanAthlete(rows, &row); err != nil {
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

// CreateAthlete: POST /api/athletes (201)
func (a *App) CreateAthlete(w http.ResponseWriter, r *http.Request) {
	var req athleteCreate
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if !validateAthleteCreate(w, req) {
		return
	}

	maxHr := 190
	if req.MaxHr != nil {
		maxHr = *req.MaxHr
	}
	now := db.NowDB()
	row := athleteRow{ID: db.NewUUID(), Name: req.Name, MaxHr: maxHr,
		CreatedAt: sql.NullString{String: now, Valid: true},
		UpdatedAt: sql.NullString{String: now, Valid: true}}
	if req.WeightKg != nil {
		row.WeightKg = sql.NullFloat64{Float64: *req.WeightKg, Valid: true}
	}
	if req.Age != nil {
		row.Age = sql.NullInt64{Int64: int64(*req.Age), Valid: true}
	}

	_, err := a.DB.Exec(
		`INSERT INTO athletes (id, name, max_hr, weight_kg, age, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		row.ID, row.Name, row.MaxHr, nullFloat64(row.WeightKg), nullInt64Arg(row.Age), now, now,
	)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, row.toJSON())
}

// UpdateAthlete: PUT /api/athletes/{athlete_id} (patch-семантика: null = не менять)
func (a *App) UpdateAthlete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "athlete_id")

	var row athleteRow
	err := a.DB.QueryRow(`SELECT id, name, max_hr, weight_kg, age, created_at, updated_at
		FROM athletes WHERE id = ?`, id).Scan(
		&row.ID, &row.Name, &row.MaxHr, &row.WeightKg, &row.Age, &row.CreatedAt, &row.UpdatedAt)
	if err == sql.ErrNoRows {
		httpError(w, 404, "Спортсмен не найден")
		return
	}
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}

	var req athleteUpdate
	if !decodeJSON(w, r, &req) {
		return
	}

	if req.Name != nil {
		*req.Name = strings.TrimSpace(*req.Name)
		if len(*req.Name) < 1 || len(*req.Name) > 100 {
			httpError(w, 422, "name: длина должна быть от 1 до 100 символов")
			return
		}
		row.Name = *req.Name
	}
	if req.MaxHr != nil {
		if *req.MaxHr < 60 || *req.MaxHr > 250 {
			httpError(w, 422, "max_hr: значение должно быть от 60 до 250")
			return
		}
		row.MaxHr = *req.MaxHr
	}
	if req.WeightKg != nil {
		if *req.WeightKg < 30 || *req.WeightKg > 250 {
			httpError(w, 422, "weight_kg: значение должно быть от 30 до 250")
			return
		}
		row.WeightKg = sql.NullFloat64{Float64: *req.WeightKg, Valid: true}
	}
	if req.Age != nil {
		if *req.Age < 10 || *req.Age > 100 {
			httpError(w, 422, "age: значение должно быть от 10 до 100")
			return
		}
		row.Age = sql.NullInt64{Int64: int64(*req.Age), Valid: true}
	}

	_, err = a.DB.Exec(
		`UPDATE athletes SET name = ?, max_hr = ?, weight_kg = ?, age = ?, updated_at = ? WHERE id = ?`,
		row.Name, row.MaxHr, nullFloat64(row.WeightKg), nullInt64Arg(row.Age), db.NowDB(), id,
	)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	row.UpdatedAt = sql.NullString{String: db.NowDB(), Valid: true}
	writeJSON(w, 200, row.toJSON())
}

// DeleteAthlete: DELETE /api/athletes/{athlete_id} (204)
func (a *App) DeleteAthlete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "athlete_id")
	res, err := a.DB.Exec(`DELETE FROM athletes WHERE id = ?`, id)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httpError(w, 404, "Спортсмен не найден")
		return
	}
	w.WriteHeader(204)
}

// getAthleteRow достаёт строку спортсмена или пишет 404.
func (a *App) getAthleteRow(w http.ResponseWriter, id string) *athleteRow {
	var row athleteRow
	err := a.DB.QueryRow(`SELECT id, name, max_hr, weight_kg, age, created_at, updated_at
		FROM athletes WHERE id = ?`, id).Scan(
		&row.ID, &row.Name, &row.MaxHr, &row.WeightKg, &row.Age, &row.CreatedAt, &row.UpdatedAt)
	if err == sql.ErrNoRows {
		httpError(w, 404, "Спортсмен не найден")
		return nil
	}
	if err != nil {
		httpError(w, 500, err.Error())
		return nil
	}
	return &row
}
