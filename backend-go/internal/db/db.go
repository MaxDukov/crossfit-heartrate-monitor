// Package db — подключение SQLite, миграции, сид-данные.
package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"

	"github.com/maxdukov/cf/backend-go/internal/data"
)

// DBTimeLayout — формат дат в SQLite, идентичный формату SQLAlchemy
// ("2006-01-02 15:04:05.000000"), чтобы Python- и Go-версии
// могли параллельно работать с одной БД.
const DBTimeLayout = "2006-01-02 15:04:05.000000"

const dbTimeLayout = DBTimeLayout

// NowDB возвращает текущее UTC-время в формате БД.
func NowDB() string { return time.Now().UTC().Format(dbTimeLayout) }

// ISO конвертирует дату из формата БД в ISO 8601 для JSON
// (как Pydantic сериализует datetime).
func ISO(dbTime string) string {
	if dbTime == "" {
		return ""
	}
	// modernc/sqlite может вернуть уже RFC3339-строку.
	if t, err := time.Parse(time.RFC3339Nano, dbTime); err == nil {
		return t.Format("2006-01-02T15:04:05.000000Z07:00")
	}
	t, err := time.ParseInLocation(dbTimeLayout, dbTime, time.UTC)
	if err != nil {
		// возможен формат без микросекунд после ручных миграций
		t, err = time.ParseInLocation("2006-01-02 15:04:05", dbTime, time.UTC)
		if err != nil {
			return dbTime
		}
	}
	return t.Format("2006-01-02T15:04:05.000000Z07:00")
}

// Open открывает SQLite с параметрами, безопасными для
// многопоточного доступа (Go) и параллельных процессов.
func Open(path string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", path)
	d, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// WAL позволяет конкурентное чтение; записи сериализует busy_timeout.
	// Пул > 1: одна потерянная по любой причине коннект не морозит сервер.
	d.SetMaxOpenConns(4)
	// Брошенные idle-коннекты возвращаются системе.
	d.SetConnMaxIdleTime(5 * time.Minute)
	return d, nil
}

// Migrate создаёт таблицы при первом запуске и применяет
// исторические миграции (совместимо со старой Python-БД).
func Migrate(d *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS athletes (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			max_hr INTEGER NOT NULL,
			weight_kg FLOAT,
			age INTEGER,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS sensors (
			device_id INTEGER NOT NULL PRIMARY KEY,
			athlete_id VARCHAR(36),
			last_hr INTEGER,
			last_seen_at DATETIME,
			battery_level INTEGER,
			ignored BOOLEAN,
			FOREIGN KEY(athlete_id) REFERENCES athletes (id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			name VARCHAR(200),
			started_at DATETIME,
			ended_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS session_athletes (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			session_id VARCHAR(36) NOT NULL,
			athlete_id VARCHAR(36) NOT NULL,
			joined_at DATETIME,
			left_at DATETIME,
			FOREIGN KEY(session_id) REFERENCES sessions (id) ON DELETE CASCADE,
			FOREIGN KEY(athlete_id) REFERENCES athletes (id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS ix_sa_session_athlete ON session_athletes (session_id, athlete_id)`,
		`CREATE TABLE IF NOT EXISTS hr_readings (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			athlete_id VARCHAR(36) NOT NULL,
			session_id VARCHAR(36),
			heart_rate INTEGER NOT NULL,
			zone INTEGER NOT NULL,
			timestamp DATETIME,
			FOREIGN KEY(athlete_id) REFERENCES athletes (id) ON DELETE CASCADE,
			FOREIGN KEY(session_id) REFERENCES sessions (id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS ix_hr_athlete_ts ON hr_readings (athlete_id, timestamp)`,
		`CREATE TABLE IF NOT EXISTS equipment (
			"key" VARCHAR(50) NOT NULL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			category VARCHAR(50) NOT NULL,
			icon VARCHAR(10) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS gym_inventory (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			equipment_key VARCHAR(50) NOT NULL,
			quantity INTEGER NOT NULL,
			FOREIGN KEY(equipment_key) REFERENCES equipment ("key") ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS movements (
			"key" VARCHAR(80) NOT NULL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			modality VARCHAR(30) NOT NULL,
			muscle_group VARCHAR(30) NOT NULL,
			themes TEXT NOT NULL,
			equipment_keys TEXT NOT NULL,
			difficulty VARCHAR(20) NOT NULL,
			scaling_beginner TEXT,
			scaling_intermediate TEXT,
			is_custom INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS wod_templates (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			format VARCHAR(30) NOT NULL,
			duration_min INTEGER NOT NULL,
			intensity VARCHAR(20) NOT NULL,
			theme VARCHAR(30) NOT NULL,
			is_benchmark BOOLEAN,
			description TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS wod_template_movements (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			template_id VARCHAR(36) NOT NULL,
			movement_key VARCHAR(80) NOT NULL,
			movement_name VARCHAR(100) NOT NULL,
			reps INTEGER,
			weight_male INTEGER,
			weight_female INTEGER,
			sort_order INTEGER NOT NULL,
			rounds_note VARCHAR(100),
			FOREIGN KEY(template_id) REFERENCES wod_templates (id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS wods (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			format VARCHAR(30) NOT NULL,
			duration_min INTEGER NOT NULL,
			intensity VARCHAR(20) NOT NULL,
			theme VARCHAR(30) NOT NULL,
			group_level VARCHAR(20) NOT NULL,
			description TEXT,
			is_active BOOLEAN,
			created_at DATETIME,
			session_id VARCHAR(36),
			FOREIGN KEY(session_id) REFERENCES sessions (id) ON DELETE SET NULL
		)`,
		`CREATE TABLE IF NOT EXISTS wod_movements (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			wod_id VARCHAR(36) NOT NULL,
			movement_key VARCHAR(80) NOT NULL,
			movement_name VARCHAR(100) NOT NULL,
			reps INTEGER,
			weight_male INTEGER,
			weight_female INTEGER,
			sort_order INTEGER NOT NULL,
			scaling_note TEXT,
			rounds_note VARCHAR(100),
			FOREIGN KEY(wod_id) REFERENCES wods (id) ON DELETE CASCADE
		)`,
		// ── Планирование тренировочного процесса (draft1.MD) ──
		`CREATE TABLE IF NOT EXISTS training_cycles (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			name VARCHAR(200) NOT NULL,
			goal TEXT,
			weeks INTEGER NOT NULL DEFAULT 8,
			start_date DATE NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'planned',
			modality_priority VARCHAR(30),
			created_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS cycle_groups (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			cycle_id VARCHAR(36) NOT NULL,
			name VARCHAR(100) NOT NULL,
			weekdays TEXT NOT NULL,
			third_day_off BOOLEAN DEFAULT 0,
			FOREIGN KEY(cycle_id) REFERENCES training_cycles (id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS cycle_slots (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			cycle_id VARCHAR(36) NOT NULL,
			group_id VARCHAR(36) NOT NULL,
			slot_date DATE NOT NULL,
			day_number INTEGER NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'empty',
			kind VARCHAR(20) NOT NULL DEFAULT 'regular',
			wod_id VARCHAR(36),
			template_id VARCHAR(36),
			session_id VARCHAR(36),
			notes TEXT,
			FOREIGN KEY(cycle_id) REFERENCES training_cycles (id) ON DELETE CASCADE,
			FOREIGN KEY(group_id) REFERENCES cycle_groups (id) ON DELETE CASCADE,
			FOREIGN KEY(wod_id) REFERENCES wods (id) ON DELETE SET NULL,
			FOREIGN KEY(session_id) REFERENCES sessions (id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS ix_cs_cycle_date ON cycle_slots (cycle_id, slot_date)`,
		// Несколько тренировок в одном дне (60 мин: разминка 10 + WOD ≤45 + заминка 5).
		`CREATE TABLE IF NOT EXISTS slot_wods (
			slot_id VARCHAR(36) NOT NULL,
			wod_id VARCHAR(36) NOT NULL,
			position INTEGER NOT NULL DEFAULT 0,
			PRIMARY KEY (slot_id, wod_id),
			FOREIGN KEY(slot_id) REFERENCES cycle_slots (id) ON DELETE CASCADE,
			FOREIGN KEY(wod_id) REFERENCES wods (id) ON DELETE CASCADE
		)`,
		// Бэкфилл: существующие дни с одной тренировкой (идемпотентно).
		`INSERT OR IGNORE INTO slot_wods (slot_id, wod_id, position)
			SELECT id, wod_id, 0 FROM cycle_slots WHERE wod_id IS NOT NULL`,
		`CREATE TABLE IF NOT EXISTS workout_results (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			slot_id VARCHAR(36),
			athlete_id VARCHAR(36) NOT NULL,
			wod_id VARCHAR(36),
			time_seconds INTEGER,
			rounds INTEGER,
			reps INTEGER,
			weight_kg FLOAT,
			scaled_version VARCHAR(20),
			rpe INTEGER,
			notes TEXT,
			created_at DATETIME,
			FOREIGN KEY(slot_id) REFERENCES cycle_slots (id) ON DELETE CASCADE,
			FOREIGN KEY(athlete_id) REFERENCES athletes (id) ON DELETE CASCADE,
			FOREIGN KEY(wod_id) REFERENCES wods (id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS ix_wr_athlete ON workout_results (athlete_id, created_at)`,
		`CREATE TABLE IF NOT EXISTS result_movements (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			result_id VARCHAR(36) NOT NULL,
			movement_key VARCHAR(80) NOT NULL,
			weight_kg FLOAT,
			reps INTEGER,
			FOREIGN KEY(result_id) REFERENCES workout_results (id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS personal_records (
			id VARCHAR(36) NOT NULL PRIMARY KEY,
			athlete_id VARCHAR(36) NOT NULL,
			record_type VARCHAR(20) NOT NULL,
			context VARCHAR(120) NOT NULL,
			value FLOAT NOT NULL,
			achieved_at DATETIME,
			slot_id VARCHAR(36),
			FOREIGN KEY(athlete_id) REFERENCES athletes (id) ON DELETE CASCADE,
			FOREIGN KEY(slot_id) REFERENCES cycle_slots (id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS ix_pr_athlete ON personal_records (athlete_id, record_type, context)`,
		`CREATE TABLE IF NOT EXISTS athlete_1rm (
			athlete_id VARCHAR(36) NOT NULL,
			movement_key VARCHAR(80) NOT NULL,
			est_1rm FLOAT NOT NULL,
			updated_at DATETIME,
			PRIMARY KEY (athlete_id, movement_key),
			FOREIGN KEY(athlete_id) REFERENCES athletes (id) ON DELETE CASCADE
		)`,
	}
	for _, s := range stmts {
		if _, err := d.Exec(s); err != nil {
			return fmt.Errorf("migrate: %w\nstatement: %s", err, s)
		}
	}

	// Исторические миграции Python-версии (ALTER TABLE для старых БД).
	alters := []struct {
		table, column, ddl string
	}{
		{"sensors", "ignored", "ALTER TABLE sensors ADD COLUMN ignored BOOLEAN DEFAULT 0 NOT NULL"},
		{"athletes", "weight_kg", "ALTER TABLE athletes ADD COLUMN weight_kg FLOAT"},
		{"athletes", "age", "ALTER TABLE athletes ADD COLUMN age INTEGER"},
		{"wods", "template_id", "ALTER TABLE wods ADD COLUMN template_id VARCHAR(36)"},
		{"cycle_slots", "kind", "ALTER TABLE cycle_slots ADD COLUMN kind VARCHAR(20) NOT NULL DEFAULT 'regular'"},
		{"cycle_groups", "third_day_off", "ALTER TABLE cycle_groups ADD COLUMN third_day_off BOOLEAN DEFAULT 0"},
		// Правки тренера не перетираются seed-refresh'ем.
		{"movements", "is_custom", "ALTER TABLE movements ADD COLUMN is_custom INTEGER NOT NULL DEFAULT 0"},
	}
	for _, a := range alters {
		exists, err := columnExists(d, a.table, a.column)
		if err != nil {
			return err
		}
		if !exists {
			if _, err := d.Exec(a.ddl); err != nil {
				return fmt.Errorf("alter %s.%s: %w", a.table, a.column, err)
			}
		}
	}
	return nil
}

func columnExists(d *sql.DB, table, column string) (bool, error) {
	rows, err := d.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notNull, pk int
		var dfltValue any
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dfltValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

// Seed заполняет справочные таблицы, если они пусты.
func Seed(d *sql.DB) error {
	// ── Инвентарь ──
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM equipment`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		for _, it := range data.EquipmentSeed {
			if _, err := d.Exec(
				`INSERT INTO equipment ("key", name, category, icon) VALUES (?, ?, ?, ?)`,
				it.Key, it.Name, it.Category, it.Icon,
			); err != nil {
				return fmt.Errorf("seed equipment: %w", err)
			}
		}
		slog.Info("seeded equipment", "count", len(data.EquipmentSeed))
	} else {
		for _, it := range data.EquipmentSeed {
			if _, err := d.Exec(
				`INSERT INTO equipment ("key", name, category, icon) VALUES (?, ?, ?, ?)
				 ON CONFLICT("key") DO UPDATE SET name = excluded.name, category = excluded.category, icon = excluded.icon`,
				it.Key, it.Name, it.Category, it.Icon,
			); err != nil {
				return fmt.Errorf("refresh equipment: %w", err)
			}
		}
	}

	// ── Движения ──
	if err := d.QueryRow(`SELECT COUNT(*) FROM movements`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		for _, mv := range data.MovementsSeed {
			if _, err := d.Exec(
				`INSERT INTO movements ("key", name, modality, muscle_group, themes, equipment_keys, difficulty, scaling_beginner, scaling_intermediate)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				mv.Key, mv.Name, mv.Modality, mv.MuscleGroup,
				joinCSV(mv.Themes), joinCSV(mv.EquipmentKeys),
				mv.Difficulty, nullStr(mv.ScalingBeginner), nullStr(mv.ScalingIntermediate),
			); err != nil {
				return fmt.Errorf("seed movements: %w", err)
			}
		}
		slog.Info("seeded movements", "count", len(data.MovementsSeed))
	} else {
		for _, mv := range data.MovementsSeed {
			// Refresh только штатного каталога: отредактированные тренером
			// движения (is_custom = 1) не перетираются.
			if _, err := d.Exec(
				`INSERT INTO movements ("key", name, modality, muscle_group, themes, equipment_keys, difficulty, scaling_beginner, scaling_intermediate, is_custom)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
				 ON CONFLICT("key") DO UPDATE SET name = excluded.name, modality = excluded.modality,
				   muscle_group = excluded.muscle_group, themes = excluded.themes,
				   equipment_keys = excluded.equipment_keys, difficulty = excluded.difficulty,
				   scaling_beginner = excluded.scaling_beginner, scaling_intermediate = excluded.scaling_intermediate
				 WHERE movements.is_custom = 0`,
				mv.Key, mv.Name, mv.Modality, mv.MuscleGroup,
				joinCSV(mv.Themes), joinCSV(mv.EquipmentKeys),
				mv.Difficulty, nullStr(mv.ScalingBeginner), nullStr(mv.ScalingIntermediate),
			); err != nil {
				return fmt.Errorf("refresh movements: %w", err)
			}
		}
	}

	// ── Шаблоны тренировок ──
	if err := d.QueryRow(`SELECT COUNT(*) FROM wod_templates`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		for _, tpl := range data.WodTemplatesSeed {
			if err := insertWodTemplate(d, tpl); err != nil {
				return err
			}
		}
		slog.Info("seeded wod templates", "count", len(data.WodTemplatesSeed))
	} else {
		refreshed, err := refreshSeedTemplates(d)
		if err != nil {
			return err
		}
		if refreshed > 0 {
			slog.Info("refreshed seed wod templates", "count", refreshed)
		}
	}
	return nil
}

// refreshSeedTemplates обновляет контент сид-шаблонов в существующей БД
// (по имени): длительность, описание и набор движений. Пользовательские
// шаблоны (автоимя «формат · тема · дата») не совпадают с сид-именами
// и остаются нетронутыми. Возвращает число обновлённых шаблонов.
func refreshSeedTemplates(d *sql.DB) (int, error) {
	updated := 0
	for _, tpl := range data.WodTemplatesSeed {
		var id string
		err := d.QueryRow(`SELECT id FROM wod_templates WHERE name = ?`, tpl.Name).Scan(&id)
		if err == sql.ErrNoRows {
			if err := insertWodTemplate(d, tpl); err != nil {
				return updated, err
			}
			updated++
			continue
		}
		if err != nil {
			return updated, err
		}
		desc := any(nil)
		if tpl.Description != "" {
			desc = tpl.Description
		}
		if _, err := d.Exec(
			`UPDATE wod_templates SET format = ?, duration_min = ?, intensity = ?, theme = ?, is_benchmark = ?, description = ? WHERE id = ?`,
			tpl.Format, tpl.DurationMin, tpl.Intensity, tpl.Theme, tpl.IsBenchmark, desc, id,
		); err != nil {
			return updated, fmt.Errorf("refresh wod template %s: %w", tpl.Name, err)
		}
		if _, err := d.Exec(`DELETE FROM wod_template_movements WHERE template_id = ?`, id); err != nil {
			return updated, fmt.Errorf("clear movements %s: %w", tpl.Name, err)
		}
		for _, m := range tpl.Movements {
			rn := any(nil)
			if m.RoundsNote != "" {
				rn = m.RoundsNote
			}
			if _, err := d.Exec(
				`INSERT INTO wod_template_movements (id, template_id, movement_key, movement_name, reps, weight_male, weight_female, sort_order, rounds_note)
				 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				NewUUID(), id, m.MovementKey, m.MovementName,
				nullIntPtr(m.Reps), nullIntPtr(m.WeightMale), nullIntPtr(m.WeightFemale),
				m.SortOrder, rn,
			); err != nil {
				return updated, fmt.Errorf("refresh wod movement %s: %w", m.MovementKey, err)
			}
		}
		updated++
	}
	return updated, nil
}

func insertWodTemplate(d *sql.DB, tpl data.WodTemplateItem) error {
	tplID := NewUUID()
	desc := any(nil)
	if tpl.Description != "" {
		desc = tpl.Description
	}
	if _, err := d.Exec(
		`INSERT INTO wod_templates (id, name, format, duration_min, intensity, theme, is_benchmark, description)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		tplID, tpl.Name, tpl.Format, tpl.DurationMin, tpl.Intensity, tpl.Theme, tpl.IsBenchmark, desc,
	); err != nil {
		return fmt.Errorf("seed wod template %s: %w", tpl.Name, err)
	}
	for _, m := range tpl.Movements {
		rn := any(nil)
		if m.RoundsNote != "" {
			rn = m.RoundsNote
		}
		if _, err := d.Exec(
			`INSERT INTO wod_template_movements (id, template_id, movement_key, movement_name, reps, weight_male, weight_female, sort_order, rounds_note)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			NewUUID(), tplID, m.MovementKey, m.MovementName,
			nullIntPtr(m.Reps), nullIntPtr(m.WeightMale), nullIntPtr(m.WeightFemale),
			m.SortOrder, rn,
		); err != nil {
			return fmt.Errorf("seed wod movement %s: %w", m.MovementKey, err)
		}
	}
	return nil
}

func joinCSV(items []string) string {
	out := ""
	for i, s := range items {
		if i > 0 {
			out += ","
		}
		out += s
	}
	return out
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func nullIntPtr(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}
