package services

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/maxdukov/cf/backend-go/internal/db"
)

func TestPythonRound(t *testing.T) {
	cases := []struct {
		in   float64
		want int
	}{
		{0.4, 0}, {0.5, 0}, {0.6, 1},
		{1.5, 2}, {2.5, 2}, {10.5, 10},
		{-0.5, 0}, {-1.5, -2},
		{7, 7}, {7.49, 7}, {7.51, 8},
	}
	for _, c := range cases {
		if got := pythonRound(c.in); got != c.want {
			t.Errorf("pythonRound(%v) = %d, хочу %d", c.in, got, c.want)
		}
	}
}

func TestSnapToKettlebellSet(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{0, 12}, {12, 12}, {13, 12}, {15, 14}, {17, 16}, {19, 18},
		{21, 20}, {22, 20}, {23, 24}, {26, 24}, {30, 28}, {33, 32}, {100, 32},
	}
	for _, c := range cases {
		if got := snapToKettlebellSet(c.in); got != c.want {
			t.Errorf("snapToKettlebellSet(%d) = %d, хочу %d", c.in, got, c.want)
		}
	}
}

func TestScaleReps(t *testing.T) {
	n := func(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }

	// 20 повторений: множители 0.5/0.75/1.0/1.25.
	for level, want := range map[string]int64{
		"beginner": 10, "intermediate": 15, "advanced": 20, "elite": 25,
	} {
		if got := scaleReps(n(20), level); got.Int64 != want || !got.Valid {
			t.Errorf("scaleReps(20, %s) = %v, хочу %d", level, got, want)
		}
	}
	// Минимум 1 повторение.
	if got := scaleReps(n(1), "beginner"); got.Int64 != 1 {
		t.Errorf("scaleReps(1, beginner) = %v, хочу 1", got)
	}
	// Неизвестный уровень — без скалирования.
	if got := scaleReps(n(20), ""); got.Int64 != 20 {
		t.Errorf("scaleReps(20, \"\") = %v, хочу 20", got)
	}
	// Пустое значение остаётся пустым.
	if got := scaleReps(sql.NullInt64{}, "beginner"); got.Valid {
		t.Errorf("scaleReps(invalid) = %v, хочу invalid", got)
	}
}

func TestScaleWeight(t *testing.T) {
	n := func(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }

	cases := []struct {
		weight int64
		level  string
		step   float64
		want   int64
	}{
		// Штанга: округление к 2.5 кг.
		{100, "advanced", 2.5, 100},
		{100, "elite", 2.5, 125},
		{40, "beginner", 2.5, 20},
		{35, "intermediate", 2.5, 25},
		// Гири: снап к набору 12-14-16-18-20-24-28-32.
		{24, "advanced", 2.0, 24},
		{24, "elite", 2.0, 28},
		{24, "beginner", 2.0, 12},
		{32, "intermediate", 2.0, 24},
		{28, "elite", 2.0, 32},
	}
	for _, c := range cases {
		got := scaleWeight(n(c.weight), c.level, c.step)
		if got.Int64 != c.want || !got.Valid {
			t.Errorf("scaleWeight(%d, %s, %.1f) = %v, хочу %d", c.weight, c.level, c.step, got, c.want)
		}
	}
	if got := scaleWeight(sql.NullInt64{}, "elite", 2.5); got.Valid {
		t.Errorf("scaleWeight(invalid) = %v, хочу invalid", got)
	}
}

func seededDB(t *testing.T) *sql.DB {
	t.Helper()
	d, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	if err := db.Migrate(d); err != nil {
		t.Fatal(err)
	}
	if err := db.Seed(d); err != nil {
		t.Fatal(err)
	}
	// Полный инвентарь: фильтр по оборудованию не отсекает шаблоны.
	if _, err := d.Exec(`INSERT INTO gym_inventory (id, equipment_key, quantity)
		SELECT hex(randomblob(16)), "key", 5 FROM equipment`); err != nil {
		t.Fatal(err)
	}
	return d
}

func TestGenerateWods(t *testing.T) {
	d := seededDB(t)

	variants := GenerateWods(d, "gymnastics", "intermediate")
	if len(variants) == 0 || len(variants) > 3 {
		t.Fatalf("хочу от 1 до 3 вариантов, получено %d", len(variants))
	}
	seen := map[string]bool{}
	for _, v := range variants {
		if v.Name == "" {
			t.Fatal("у варианта пустое имя")
		}
		if seen[v.TemplateID] {
			t.Fatalf("дубликат варианта %q (%s)", v.Name, v.TemplateID)
		}
		seen[v.TemplateID] = true
		if len(v.Movements) == 0 {
			t.Fatalf("у варианта %q нет движений", v.Name)
		}
		for _, m := range v.Movements {
			if m.Reps != nil && *m.Reps < 1 {
				t.Fatalf("у варианта %q движение с reps=%d", v.Name, *m.Reps)
			}
		}
	}

	// Архивный шаблон не участвует в генерации.
	reps := 15
	id, _, err := CreateCustomWod(d, CustomWodInput{
		Format: "amrap", DurationMin: 12, Intensity: "high", Theme: "gymnastics",
		Movements: []CustomMovementInput{{MovementKey: "burpee", Reps: &reps}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`UPDATE wod_templates SET archived = 1 WHERE id = ?`, id); err != nil {
		t.Fatal(err)
	}
	variants = GenerateWods(d, "gymnastics", "intermediate")
	for _, v := range variants {
		if v.TemplateID == id {
			t.Fatal("архивный шаблон попал в генерацию")
		}
	}
}
