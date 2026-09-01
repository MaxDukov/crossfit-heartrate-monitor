package handlers

import (
	"database/sql"
	"net/http"

	"github.com/maxdukov/cf/backend-go/internal/db"
)

// ── Equipment: каталог инвентаря и инвентарь зала ─────────────

// ListEquipment: GET /api/equipment
func (a *App) ListEquipment(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.Query(`SELECT "key", name, category, icon FROM equipment ORDER BY category, name`)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	defer func() { _ = rows.Close() }()

	out := []map[string]any{}
	for rows.Next() {
		var key, name, category, icon string
		if err := rows.Scan(&key, &name, &category, &icon); err != nil {
			httpError(w, 500, err.Error())
			return
		}
		out = append(out, map[string]any{
			"key": key, "name": name, "category": category, "icon": icon,
		})
	}
	if out == nil {
		out = []map[string]any{}
	}
	writeJSON(w, 200, out)
}

// ListInventory: GET /api/equipment/inventory
func (a *App) ListInventory(w http.ResponseWriter, r *http.Request) {
	out, ok := a.queryInventory(w)
	if !ok {
		return
	}
	writeJSON(w, 200, out)
}

type inventoryItemReq struct {
	EquipmentKey string `json:"equipment_key"`
	Quantity     *int   `json:"quantity"`
}

type inventoryUpdateReq struct {
	Items []inventoryItemReq `json:"items"`
}

// UpdateInventory: PUT /api/equipment/inventory — полная перезапись.
func (a *App) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	var req inventoryUpdateReq
	if !decodeJSON(w, r, &req) {
		return
	}

	tx, err := a.DB.Begin()
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`DELETE FROM gym_inventory`); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	for _, item := range req.Items {
		qty := 1
		if item.Quantity != nil {
			qty = *item.Quantity
		}
		if _, err := tx.Exec(
			`INSERT INTO gym_inventory (id, equipment_key, quantity) VALUES (?, ?, ?)`,
			db.NewUUID(), item.EquipmentKey, qty,
		); err != nil {
			httpError(w, 500, err.Error())
			return
		}
	}
	if err := tx.Commit(); err != nil {
		httpError(w, 500, err.Error())
		return
	}

	out, ok := a.queryInventory(w)
	if !ok {
		return
	}
	writeJSON(w, 200, out)
}

func (a *App) queryInventory(w http.ResponseWriter) ([]map[string]any, bool) {
	rows, err := a.DB.Query(`SELECT equipment_key, quantity FROM gym_inventory`)
	if err != nil {
		httpError(w, 500, err.Error())
		return nil, false
	}
	defer func() { _ = rows.Close() }()

	out := []map[string]any{}
	for rows.Next() {
		var key string
		var qty int
		if err := rows.Scan(&key, &qty); err != nil {
			httpError(w, 500, err.Error())
			return nil, false
		}
		out = append(out, map[string]any{"equipment_key": key, "quantity": qty})
	}
	return out, true
}

// null-хелперы, общие для handlers.
func nullStr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func nullInt64(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

func nullInt0(ni sql.NullInt64) int {
	if !ni.Valid {
		return 0
	}
	return int(ni.Int64)
}

func nullFloat(nf sql.NullFloat64) *float64 {
	if !nf.Valid {
		return nil
	}
	return &nf.Float64
}

func nullFloat64(nf sql.NullFloat64) any {
	if !nf.Valid {
		return nil
	}
	return nf.Float64
}

func nullInt64Arg(ni sql.NullInt64) any {
	if !ni.Valid {
		return nil
	}
	return ni.Int64
}

func isoOrNull(ns sql.NullString) any {
	if !ns.Valid {
		return nil
	}
	return db.ISO(ns.String)
}

func scanAthlete(rows *sql.Rows, row *athleteRow) error {
	return rows.Scan(&row.ID, &row.Name, &row.MaxHr, &row.WeightKg, &row.Age, &row.CreatedAt, &row.UpdatedAt)
}

func (a *App) getSensorRowOr500(w http.ResponseWriter, deviceID int) *sensorRow {
	row := a.getSensorRow(w, deviceID, "Датчик не найден")
	if row == nil {
		httpError(w, 500, "датчик исчез после обновления")
		return nil
	}
	return row
}

// round1 округляет float до 1 знака (аналог Python round(v, 1)).
func round1(v float64) float64 {
	sign := 1.0
	if v < 0 {
		sign = -1.0
	}
	return float64(int64(v*10+sign*0.5)) / 10
}
