// Package services — аналитика цикла (Экран 7 draft1.MD).
package services

import (
	"database/sql"

	"github.com/maxdukov/cf/backend-go/internal/db"
)

// AthleteCycleStats — статистика спортсмена в цикле.
type AthleteCycleStats struct {
	AthleteID   string   `json:"athlete_id"`
	Name        string   `json:"name"`
	Workouts    int      `json:"workouts_done"`
	AvgRPE      *float64 `json:"avg_rpe"`
	TotalWeight float64  `json:"total_volume_kg"`
	PRCount     int      `json:"pr_count"`
}

// GroupCycleStats — план/факт по группе.
type GroupCycleStats struct {
	GroupID    string `json:"group_id"`
	Name       string `json:"name"`
	SlotsTotal int    `json:"slots_total"`
	Filled     int    `json:"slots_filled"`
	Completed  int    `json:"slots_completed"`
	Skipped    int    `json:"slots_skipped"`
}

// ThemeStat — запланировано/проведено по темам.
type ThemeStat struct {
	Theme     string `json:"theme"`
	ThemeName string `json:"theme_name"`
	Planned   int    `json:"planned"`
	Done      int    `json:"done"`
}

// RMProgress — точка динамики 1ПМ.
type RMProgress struct {
	Athlete     string  `json:"athlete_name"`
	MovementKey string  `json:"movement_key"`
	Movement    string  `json:"movement_name"`
	Date        string  `json:"date"`
	Value       float64 `json:"value"`
}

// CycleAnalytics — полный отчёт по циклу (Экран 7).
type CycleAnalytics struct {
	SlotsTotal   int                 `json:"slots_total"`
	SlotsFilled  int                 `json:"slots_filled"`
	SlotsDone    int                 `json:"slots_completed"`
	SlotsSkipped int                 `json:"slots_skipped"`
	FillPercent  float64             `json:"fill_percent"`
	DonePercent  float64             `json:"completion_percent"`
	Groups       []GroupCycleStats   `json:"groups"`
	Themes       []ThemeStat         `json:"themes"`
	Athletes     []AthleteCycleStats `json:"athletes"`
	RMProgress   []RMProgress        `json:"rm_progress"`
	Advice       []string            `json:"advice"`
}

// GetCycleAnalytics собирает аналитику цикла.
func GetCycleAnalytics(d *sql.DB, cycleID string) (*CycleAnalytics, error) {
	var exists int
	if err := d.QueryRow(`SELECT COUNT(*) FROM training_cycles WHERE id = ?`, cycleID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, sql.ErrNoRows
	}

	a := &CycleAnalytics{}

	// Группы: план/факт.
	{
		rows, err := d.Query(`
			SELECT g.id, g.name,
			       COUNT(s.id),
			       COALESCE(SUM(CASE WHEN s.status != 'empty' THEN 1 ELSE 0 END), 0),
			       COALESCE(SUM(CASE WHEN s.status = 'completed' THEN 1 ELSE 0 END), 0),
			       COALESCE(SUM(CASE WHEN s.status = 'skipped' THEN 1 ELSE 0 END), 0)
			FROM cycle_groups g LEFT JOIN cycle_slots s ON s.group_id = g.id
			WHERE g.cycle_id = ? GROUP BY g.id ORDER BY g.name`, cycleID)
		if err == nil {
			for rows.Next() {
				var g GroupCycleStats
				if rows.Scan(&g.GroupID, &g.Name, &g.SlotsTotal, &g.Filled, &g.Completed, &g.Skipped) == nil {
					a.Groups = append(a.Groups, g)
					a.SlotsTotal += g.SlotsTotal
					a.SlotsFilled += g.Filled
					a.SlotsDone += g.Completed
					a.SlotsSkipped += g.Skipped
				}
			}
			_ = rows.Close()
		}
	}
	if a.SlotsTotal > 0 {
		a.FillPercent = float64(a.SlotsFilled) * 100 / float64(a.SlotsTotal)
		a.DonePercent = float64(a.SlotsDone) * 100 / float64(a.SlotsTotal)
	}

	// Темы: план vs факт.
	{
		rows, err := d.Query(`
			SELECT COALESCE(w.theme, ''), COUNT(*),
			       COALESCE(SUM(CASE WHEN s.status = 'completed' THEN 1 ELSE 0 END), 0)
			FROM cycle_slots s JOIN wods w ON w.id = s.wod_id
			WHERE s.cycle_id = ? GROUP BY w.theme ORDER BY COUNT(*) DESC`, cycleID)
		if err == nil {
			for rows.Next() {
				var ts ThemeStat
				if rows.Scan(&ts.Theme, &ts.Planned, &ts.Done) == nil {
					ts.ThemeName = ThemeName(ts.Theme)
					a.Themes = append(a.Themes, ts)
				}
			}
			_ = rows.Close()
		}
	}

	// Спортсмены: тренировки, RPE, объём, PR.
	{
		rows, err := d.Query(`
			SELECT r.athlete_id, COALESCE(at.name, ''), COUNT(DISTINCT r.slot_id),
			       AVG(r.rpe),
			       COALESCE((SELECT SUM(rm.weight_kg * rm.reps) FROM result_movements rm
			         JOIN workout_results r2 ON r2.id = rm.result_id
			         WHERE r2.athlete_id = r.athlete_id AND r2.slot_id IN
			           (SELECT id FROM cycle_slots WHERE cycle_id = ?)), 0),
			       (SELECT COUNT(*) FROM personal_records p
			         WHERE p.athlete_id = r.athlete_id AND p.slot_id IN
			           (SELECT id FROM cycle_slots WHERE cycle_id = ?))
			FROM workout_results r
			JOIN cycle_slots s ON s.id = r.slot_id
			LEFT JOIN athletes at ON at.id = r.athlete_id
			WHERE s.cycle_id = ?
			GROUP BY r.athlete_id ORDER BY COUNT(DISTINCT r.slot_id) DESC`,
			cycleID, cycleID, cycleID)
		if err == nil {
			for rows.Next() {
				var st AthleteCycleStats
				var avgRPE sql.NullFloat64
				if rows.Scan(&st.AthleteID, &st.Name, &st.Workouts, &avgRPE, &st.TotalWeight, &st.PRCount) == nil {
					if avgRPE.Valid {
						v := avgRPE.Float64
						st.AvgRPE = &v
					}
					a.Athletes = append(a.Athletes, st)
				}
			}
			_ = rows.Close()
		}
	}

	// Динамика 1ПМ в цикле.
	{
		rows, err := d.Query(`
			SELECT COALESCE(at.name, ''), p.context, COALESCE(m.name, p.context),
			       p.achieved_at, p.value
			FROM personal_records p
			JOIN cycle_slots s ON s.id = p.slot_id
			LEFT JOIN athletes at ON at.id = p.athlete_id
			LEFT JOIN movements m ON m."key" = p.context
			WHERE s.cycle_id = ? AND p.record_type = '1rm'
			ORDER BY p.achieved_at ASC`, cycleID)
		if err == nil {
			for rows.Next() {
				var rp RMProgress
				var achieved sql.NullString
				if rows.Scan(&rp.Athlete, &rp.MovementKey, &rp.Movement, &achieved, &rp.Value) == nil {
					rp.Date = isoDateStr(achieved)
					a.RMProgress = append(a.RMProgress, rp)
				}
			}
			_ = rows.Close()
		}
	}

	// Рекомендации (Экран 7).
	a.Advice = cycleAdvice(a)
	return a, nil
}

func isoDateStr(n sql.NullString) string {
	if !n.Valid || n.String == "" {
		return ""
	}
	iso := db.ISO(n.String)
	if len(iso) >= 10 {
		return iso[:10]
	}
	return iso
}

func cycleAdvice(a *CycleAnalytics) []string {
	var out []string
	// Отстающие модальности: план есть, факт заметно ниже.
	for _, t := range a.Themes {
		if t.Planned >= 2 && t.Done == 0 {
			out = append(out, "Модальность «"+t.ThemeName+"» запланирована, но не проведена ни разу — проверьте баланс")
		} else if t.Planned >= 3 && t.Done*2 < t.Planned {
			out = append(out, "Модальность «"+t.ThemeName+"» отстаёт от плана ("+itoaLocal(t.Done)+" из "+itoaLocal(t.Planned)+")")
		}
	}
	if a.SlotsFilled > 0 && a.SlotsDone*3 < a.SlotsFilled && a.SlotsDone < a.SlotsFilled {
		out = append(out, "Многие запланированные тренировки не завершены — фиксируйте результаты для точной аналитики")
	}
	if len(a.RMProgress) > 0 {
		out = append(out, "Рекорды силы растут — можно повысить базовые веса в следующем блоке")
	}
	if len(out) == 0 {
		out = append(out, "Цикл идёт по плану — отклонений не найдено")
	}
	return out
}

func itoaLocal(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [12]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
