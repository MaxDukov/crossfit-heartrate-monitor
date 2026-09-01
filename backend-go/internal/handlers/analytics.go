package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// ── Analytics: агрегаты по hr_readings ────────────────────────

// AthleteStats: GET /api/analytics/athletes/{athlete_id}/stats
func (a *App) AthleteStats(w http.ResponseWriter, r *http.Request) {
	athleteID := chi.URLParam(r, "athlete_id")
	if a.getAthleteRow(w, athleteID) == nil {
		return
	}

	var totalSessions int
	_ = a.DB.QueryRow(
		`SELECT COUNT(DISTINCT session_id) FROM hr_readings
		 WHERE athlete_id = ? AND session_id IS NOT NULL`, athleteID,
	).Scan(&totalSessions)

	var (
		cnt   sql.NullInt64
		avgHr sql.NullFloat64
		maxHr sql.NullInt64
	)
	_ = a.DB.QueryRow(
		`SELECT COUNT(*), AVG(heart_rate), MAX(heart_rate) FROM hr_readings WHERE athlete_id = ?`,
		athleteID,
	).Scan(&cnt, &avgHr, &maxHr)

	var duration sql.NullFloat64
	_ = a.DB.QueryRow(
		`SELECT SUM(strftime('%s', left_at) - strftime('%s', joined_at))
		 FROM session_athletes WHERE athlete_id = ? AND left_at IS NOT NULL`, athleteID,
	).Scan(&duration)

	avg := 0.0
	if avgHr.Valid && avgHr.Float64 > 0 {
		avg = round1(avgHr.Float64)
	}
	maxHrEver := 0
	if maxHr.Valid {
		maxHrEver = int(maxHr.Int64)
	}
	totalDur := 0
	if duration.Valid {
		totalDur = int(duration.Float64)
	}

	writeJSON(w, 200, map[string]any{
		"total_sessions":         totalSessions,
		"total_duration_seconds": totalDur,
		"avg_hr":                 avg,
		"max_hr_ever":            maxHrEver,
	})
}

// AthleteHistory: GET /api/analytics/athletes/{athlete_id}/history?limit=20
func (a *App) AthleteHistory(w http.ResponseWriter, r *http.Request) {
	athleteID := chi.URLParam(r, "athlete_id")
	if a.getAthleteRow(w, athleteID) == nil {
		return
	}
	limit := intQueryParam(r, "limit", 20)

	// Важно: MaxOpenConns(1) — сначала вычитываем агрегаты и закрываем rows,
	// только потом делаем запросы сессий (иначе дедлок на единственном коннекте).
	type aggRow struct {
		sessionID      string
		avgHr          sql.NullFloat64
		maxHr, minHr   sql.NullInt64
		z1, z2, z3, z4 sql.NullInt64
	}
	aggs := []aggRow{}
	{
		rows, err := a.DB.Query(`
			SELECT session_id, AVG(heart_rate), MAX(heart_rate), MIN(heart_rate),
			       SUM(CASE WHEN zone = 1 THEN 1 ELSE 0 END),
			       SUM(CASE WHEN zone = 2 THEN 1 ELSE 0 END),
			       SUM(CASE WHEN zone = 3 THEN 1 ELSE 0 END),
			       SUM(CASE WHEN zone = 4 THEN 1 ELSE 0 END)
			FROM hr_readings
			WHERE athlete_id = ? AND session_id IS NOT NULL
			GROUP BY session_id
			ORDER BY session_id DESC
			LIMIT ?`, athleteID, limit,
		)
		if err != nil {
			httpError(w, 500, err.Error())
			return
		}
		for rows.Next() {
			var g aggRow
			if err := rows.Scan(&g.sessionID, &g.avgHr, &g.maxHr, &g.minHr, &g.z1, &g.z2, &g.z3, &g.z4); err != nil {
				_ = rows.Close()
				httpError(w, 500, err.Error())
				return
			}
			aggs = append(aggs, g)
		}
		_ = rows.Close()
	}

	out := []map[string]any{}
	for _, g := range aggs {
		var (
			sessName  sql.NullString
			startedAt sql.NullString
			endedAt   sql.NullString
		)
		err := a.DB.QueryRow(`SELECT name, started_at, ended_at FROM sessions WHERE id = ?`, g.sessionID).
			Scan(&sessName, &startedAt, &endedAt)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			httpError(w, 500, err.Error())
			return
		}

		duration := 0
		if endedAt.Valid && startedAt.Valid {
			st := parseDBTime(startedAt.String)
			en := parseDBTime(endedAt.String)
			if !st.IsZero() && !en.IsZero() {
				duration = int(en.Sub(st).Seconds())
			}
		}

		avg := 0.0
		if g.avgHr.Valid && g.avgHr.Float64 > 0 {
			avg = round1(g.avgHr.Float64)
		}

		out = append(out, map[string]any{
			"session_id":       g.sessionID,
			"session_name":     nullStr(sessName),
			"avg_hr":           avg,
			"max_hr":           nullInt64(g.maxHr),
			"min_hr":           nullInt64(g.minHr),
			"duration_seconds": duration,
			"zones": map[string]int{
				"zone_1_seconds": nullInt0(g.z1),
				"zone_2_seconds": nullInt0(g.z2),
				"zone_3_seconds": nullInt0(g.z3),
				"zone_4_seconds": nullInt0(g.z4),
			},
		})
	}
	writeJSON(w, 200, out)
}

func parseDBTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	for _, layout := range []string{"2006-01-02 15:04:05.999999", "2006-01-02 15:04:05"} {
		if t, err := time.ParseInLocation(layout, s, time.UTC); err == nil {
			return t
		}
	}
	return time.Time{}
}
