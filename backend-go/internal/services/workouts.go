// Package services — проведение тренировки и логирование результатов (Экран 6 draft1.MD).
package services

import (
	"database/sql"
	"fmt"
	"math"

	"github.com/maxdukov/cf/backend-go/internal/db"
)

// StartSlot начинает тренировку по слоту: создаёт live-сессию,
// привязывает её к WoD (заполняет wods.session_id) и включает
// WoD на мониторе (is_active=1).
func StartSlot(d *sql.DB, slotID string) (string, error) {
	var status, slotDate string
	var wodID sql.NullString
	var cycleName, groupName string
	err := d.QueryRow(`
		SELECT s.status, s.slot_date, s.wod_id, c.name, g.name
		FROM cycle_slots s
		JOIN training_cycles c ON c.id = s.cycle_id
		JOIN cycle_groups g ON g.id = s.group_id
		WHERE s.id = ?`, slotID,
	).Scan(&status, &slotDate, &wodID, &cycleName, &groupName)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("слот не найден")
	}
	if err != nil {
		return "", err
	}
	if status != "planned" {
		return "", fmt.Errorf("начать можно только запланированный слот (статус: %s)", status)
	}
	if !wodID.Valid {
		return "", fmt.Errorf("слот не заполнен — назначьте тренировку")
	}

	// Завершить активную сессию, если была (singleton-контракт как в POST /api/sessions).
	if _, err := d.Exec(`UPDATE sessions SET ended_at = ? WHERE ended_at IS NULL`, db.NowDB()); err != nil {
		return "", err
	}

	sessionID := db.NewUUID()
	sessionName := fmt.Sprintf("%s · %s · %s", cycleName, groupName, ruDate(slotDate))
	if _, err := d.Exec(`
		INSERT INTO sessions (id, name, started_at) VALUES (?, ?, ?)`,
		sessionID, sessionName, db.NowDB(),
	); err != nil {
		return "", err
	}

	tx, err := d.Begin()
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`UPDATE wods SET is_active = 0 WHERE is_active = 1`); err != nil {
		return "", err
	}
	if _, err := tx.Exec(`UPDATE wods SET is_active = 1, session_id = ? WHERE id = ?`, sessionID, wodID.String); err != nil {
		return "", err
	}
	if _, err := tx.Exec(`UPDATE cycle_slots SET status = 'in_progress', session_id = ? WHERE id = ?`, sessionID, slotID); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return sessionID, nil
}

// CompleteSlot завершает тренировку: закрывает сессию, слот — completed.
func CompleteSlot(d *sql.DB, slotID string) error {
	var status string
	var sessionID sql.NullString
	err := d.QueryRow(`SELECT status, session_id FROM cycle_slots WHERE id = ?`, slotID).Scan(&status, &sessionID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("слот не найден")
	}
	if err != nil {
		return err
	}
	if status != "in_progress" {
		return fmt.Errorf("завершить можно только идущую тренировку (статус: %s)", status)
	}

	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if sessionID.Valid {
		if _, err := tx.Exec(`UPDATE sessions SET ended_at = ? WHERE id = ? AND ended_at IS NULL`, db.NowDB(), sessionID.String); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(`UPDATE cycle_slots SET status = 'completed' WHERE id = ?`, slotID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE wods SET is_active = 0 WHERE session_id = ?`, sessionID.String); err != nil {
		return err
	}
	return tx.Commit()
}

func ruDate(iso string) string {
	t, err := parseSlotDate(iso)
	if err != nil {
		return normDate(iso)
	}
	months := []string{"янв", "фев", "мар", "апр", "мая", "июн", "июл", "авг", "сен", "окт", "ноя", "дек"}
	return fmt.Sprintf("%02d %s", t.Day(), months[int(t.Month())-1])
}

// ── Результаты (Экран 6) ──────────────────────────────────────

// ResultMovementInput — фактически использованный вес/повторы движения.
type ResultMovementInput struct {
	MovementKey string   `json:"movement_key"`
	WeightKg    *float64 `json:"weight_kg"`
	Reps        *int     `json:"reps"`
}

// SaveResultInput — запрос POST /api/slots/{id}/results.
type SaveResultInput struct {
	AthleteID     string                `json:"athlete_id"`
	TimeSeconds   *int                  `json:"time_seconds"`
	Rounds        *int                  `json:"rounds"`
	Reps          *int                  `json:"reps"`
	WeightKg      *float64              `json:"weight_kg"`
	ScaledVersion string                `json:"scaled_version"`
	RPE           *int                  `json:"rpe"`
	Notes         string                `json:"notes"`
	Movements     []ResultMovementInput `json:"movements"`
}

// PRInfo — информация о персональном рекорде в ответе.
type PRInfo struct {
	IsPR        bool     `json:"is_pr"`
	Type        string   `json:"type,omitempty"`
	Context     string   `json:"context,omitempty"`
	Previous    *float64 `json:"previous_best,omitempty"`
	New         *float64 `json:"new_value,omitempty"`
	Description string   `json:"description,omitempty"`
}

// SaveResultResponse — ответ на логирование результата.
type SaveResultResponse struct {
	ResultID string      `json:"result_id"`
	PR       PRInfo      `json:"pr"`
	PrevPRs  []PRHistory `json:"previous_attempts"`
}

// PRHistory — предыдущая попытка этого же WoD/движения.
type PRHistory struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
	Kind  string  `json:"kind"`
}

// SaveResult записывает результат, находит PR, обновляет 1ПМ (Epley).
func SaveResult(d *sql.DB, slotID string, in SaveResultInput) (*SaveResultResponse, error) {
	if in.AthleteID == "" {
		return nil, fmt.Errorf("укажите athlete_id")
	}

	var status string
	var wodID sql.NullString
	var tplID sql.NullString
	err := d.QueryRow(`
		SELECT s.status, s.wod_id, COALESCE(w.template_id, '')
		FROM cycle_slots s LEFT JOIN wods w ON w.id = s.wod_id
		WHERE s.id = ?`, slotID,
	).Scan(&status, &wodID, &tplID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("слот не найден")
	}
	if err != nil {
		return nil, err
	}
	if status != "in_progress" && status != "completed" {
		return nil, fmt.Errorf("результаты принимаются только для идущей или завершённой тренировки")
	}

	// История попыток этого шаблона (до записи новой).
	prevAttempts := wTemplateAttempts(d, in.AthleteID, tplID.String, wodID)

	resultID := db.NewUUID()
	var t, rounds, reps any
	if in.TimeSeconds != nil {
		t = *in.TimeSeconds
	}
	if in.Rounds != nil {
		rounds = *in.Rounds
	}
	if in.Reps != nil {
		reps = *in.Reps
	}
	var weight any
	if in.WeightKg != nil {
		weight = *in.WeightKg
	}
	var notes any
	if in.Notes != "" {
		notes = in.Notes
	}
	var sv any
	if in.ScaledVersion != "" {
		sv = in.ScaledVersion
	}
	var wRef any
	if wodID.Valid {
		wRef = wodID.String
	}

	if _, err := d.Exec(`
		INSERT INTO workout_results (id, slot_id, athlete_id, wod_id, time_seconds, rounds, reps, weight_kg, scaled_version, rpe, notes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		resultID, slotID, in.AthleteID, wRef, t, rounds, reps, weight, sv, intPtrAny(in.RPE), notes, db.NowDB(),
	); err != nil {
		return nil, err
	}

	// Per-movement веса (для 1ПМ).
	for _, m := range in.Movements {
		var w, r any
		if m.WeightKg != nil {
			w = *m.WeightKg
		}
		if m.Reps != nil {
			r = *m.Reps
		}
		if _, err := d.Exec(`
			INSERT INTO result_movements (id, result_id, movement_key, weight_kg, reps)
			VALUES (?, ?, ?, ?, ?)`, db.NewUUID(), resultID, m.MovementKey, w, r); err != nil {
			return nil, err
		}
	}

	pr := PRInfo{}

	// PR по времени (fastest_time, ниже — лучше) для того же шаблона.
	if in.TimeSeconds != nil && tplID.String != "" && in.ScaledVersion != "beginner" {
		context := "template:" + tplID.String
		var best sql.NullFloat64
		_ = d.QueryRow(`SELECT value FROM personal_records WHERE athlete_id = ? AND record_type = 'fastest_time' AND context = ?`,
			in.AthleteID, context).Scan(&best)
		cur := float64(*in.TimeSeconds)
		if !best.Valid || cur < best.Float64 {
			if best.Valid {
				prev := best.Float64
				pr.Previous = &prev
			}
			pr.IsPR = true
			pr.Type = "fastest_time"
			pr.Context = context
			pr.New = &cur
			if _, err := d.Exec(`
				INSERT INTO personal_records (id, athlete_id, record_type, context, value, achieved_at, slot_id)
				VALUES (?, ?, 'fastest_time', ?, ?, ?, ?)`,
				db.NewUUID(), in.AthleteID, context, cur, db.NowDB(), slotID); err != nil {
				return nil, err
			}
		}
	}

	// PR по раундам AMRAP (most_rounds, выше — лучше).
	if in.Rounds != nil && tplID.String != "" {
		context := "template:" + tplID.String
		var best sql.NullFloat64
		_ = d.QueryRow(`SELECT value FROM personal_records WHERE athlete_id = ? AND record_type = 'most_rounds' AND context = ?`,
			in.AthleteID, context).Scan(&best)
		cur := float64(*in.Rounds)
		if !best.Valid || cur > best.Float64 {
			if _, err := d.Exec(`
				INSERT INTO personal_records (id, athlete_id, record_type, context, value, achieved_at, slot_id)
				VALUES (?, ?, 'most_rounds', ?, ?, ?, ?)`,
				db.NewUUID(), in.AthleteID, context, cur, db.NowDB(), slotID); err != nil {
				return nil, err
			}
		}
	}

	// 1ПМ по движениям (Epley: 1RM = w × (1 + reps/30)).
	movementPRs := []PRInfo{}
	for _, m := range in.Movements {
		if m.WeightKg == nil || *m.WeightKg <= 0 || m.Reps == nil || *m.Reps <= 0 {
			continue
		}
		est := *m.WeightKg * (1 + float64(*m.Reps)/30)
		est = math.Round(est*2) / 2 // до 0.5 кг

		var prev sql.NullFloat64
		_ = d.QueryRow(`SELECT est_1rm FROM athlete_1rm WHERE athlete_id = ? AND movement_key = ?`,
			in.AthleteID, m.MovementKey).Scan(&prev)

		if !prev.Valid || est > prev.Float64 {
			if prev.Valid {
				p := prev.Float64
				movementPRs = append(movementPRs, PRInfo{
					IsPR: true, Type: "1rm", Context: m.MovementKey,
					Previous: &p, New: &est,
					Description: fmt.Sprintf("Новый расчётный 1ПМ %.1f кг (было %.1f)", est, prev.Float64),
				})
			} else {
				v := est
				movementPRs = append(movementPRs, PRInfo{
					IsPR: true, Type: "1rm", Context: m.MovementKey, New: &v,
					Description: fmt.Sprintf("Первый расчётный 1ПМ: %.1f кг", est),
				})
			}
			if _, err := d.Exec(`
				INSERT INTO athlete_1rm (athlete_id, movement_key, est_1rm, updated_at) VALUES (?, ?, ?, ?)
				ON CONFLICT(athlete_id, movement_key) DO UPDATE SET est_1rm = excluded.est_1rm, updated_at = excluded.updated_at`,
				in.AthleteID, m.MovementKey, est, db.NowDB()); err != nil {
				return nil, err
			}
			if _, err := d.Exec(`
				INSERT INTO personal_records (id, athlete_id, record_type, context, value, achieved_at, slot_id)
				VALUES (?, ?, '1rm', ?, ?, ?, ?)`,
				db.NewUUID(), in.AthleteID, m.MovementKey, est, db.NowDB(), slotID); err != nil {
				return nil, err
			}
		}
	}

	// Главный PR — время, иначе первый 1RM.
	if !pr.IsPR && len(movementPRs) > 0 {
		pr = movementPRs[0]
	}
	if pr.IsPR && pr.Description == "" {
		if pr.Type == "fastest_time" {
			pr.Description = fmt.Sprintf("Новый личный рекорд: %s (%.0f сек)", formatDuration(int(*pr.New)), *pr.New)
			if pr.Previous != nil {
				pr.Description += fmt.Sprintf(" — лучше прошлого на %.0f сек", *pr.Previous-*pr.New)
			}
		}
	}

	return &SaveResultResponse{
		ResultID: resultID,
		PR:       pr,
		PrevPRs:  prevAttempts,
	}, nil
}

func intPtrAny(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func formatDuration(sec int) string {
	m := sec / 60
	s := sec % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

// wTemplateAttempts — прошлые результаты спортсмена по этому шаблону.
func wTemplateAttempts(d *sql.DB, athleteID, templateID string, currentWodID sql.NullString) []PRHistory {
	out := []PRHistory{}
	if templateID == "" {
		return out
	}
	rows, err := d.Query(`
		SELECT r.created_at, r.time_seconds, r.rounds
		FROM workout_results r
		JOIN cycle_slots s ON s.id = r.slot_id
		WHERE r.athlete_id = ? AND s.template_id = ?
		ORDER BY r.created_at ASC`, athleteID, templateID)
	if err != nil {
		return out
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var created sql.NullString
		var t, rounds sql.NullInt64
		if rows.Scan(&created, &t, &rounds) == nil {
			iso := db.ISO(created.String)
			day := iso
			if len(day) >= 10 {
				day = day[:10]
			}
			if t.Valid {
				out = append(out, PRHistory{Date: day, Value: float64(t.Int64), Kind: "time"})
			} else if rounds.Valid {
				out = append(out, PRHistory{Date: day, Value: float64(rounds.Int64), Kind: "rounds"})
			}
		}
	}
	return out
}

// ListSlotResults — результаты слота (GET /api/slots/{id}/results).
func ListSlotResults(d *sql.DB, slotID string) []map[string]any {
	type resRow struct {
		id, athID string
		athlete   string
		t, r, rp  sql.NullInt64
		weight    sql.NullFloat64
		version   sql.NullString
		rpe       sql.NullInt64
		notes     sql.NullString
		created   sql.NullString
	}
	var rowsData []resRow
	{
		rows, err := d.Query(`
			SELECT r.id, r.athlete_id, COALESCE(a.name, ''), r.time_seconds, r.rounds, r.reps, r.weight_kg,
			       r.scaled_version, r.rpe, r.notes, r.created_at
			FROM workout_results r LEFT JOIN athletes a ON a.id = r.athlete_id
			WHERE r.slot_id = ? ORDER BY r.created_at ASC`, slotID)
		if err != nil {
			return nil
		}
		for rows.Next() {
			var x resRow
			if rows.Scan(&x.id, &x.athID, &x.athlete, &x.t, &x.r, &x.rp, &x.weight, &x.version, &x.rpe, &x.notes, &x.created) == nil {
				rowsData = append(rowsData, x)
			}
		}
		_ = rows.Close()
	}

	out := []map[string]any{}
	for _, x := range rowsData {
		item := map[string]any{
			"id":             x.id,
			"athlete_id":     x.athID,
			"athlete_name":   x.athlete,
			"time_seconds":   nullInt64(x.t),
			"rounds":         nullInt64(x.r),
			"reps":           nullInt64(x.rp),
			"weight_kg":      nullFloat(x.weight),
			"scaled_version": nullStrPtrToAny(x.version),
			"rpe":            nullInt64(x.rpe),
			"notes":          nullStrPtrToAny(x.notes),
			"created_at":     isoOrNullLocal(x.created),
		}
		out = append(out, item)
	}
	return out
}

func nullInt64(n sql.NullInt64) any {
	if !n.Valid {
		return nil
	}
	return n.Int64
}

func nullFloat(n sql.NullFloat64) any {
	if !n.Valid {
		return nil
	}
	return n.Float64
}

func nullStrPtrToAny(n sql.NullString) any {
	if !n.Valid {
		return nil
	}
	return n.String
}

func isoOrNullLocal(n sql.NullString) any {
	if !n.Valid || n.String == "" {
		return nil
	}
	return db.ISO(n.String)
}
