package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestSlotResultsAnd1RM(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// ── Цикл с двумя слотами: один для результатов, один пустой ──
	rec := ts.post(t, "/api/cycles", `{"name":"Цикл результатов","weeks":1,"start_date":"2026-09-07",
		"groups":[{"name":"А","weekdays":[1,3]}]}`, http.StatusCreated)
	var created struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)

	var cycle struct {
		Slots []struct {
			ID string `json:"id"`
		} `json:"slots"`
	}
	ts.getJSON(t, "/api/cycles/"+created.ID, &cycle)
	slotID, emptySlotID := cycle.Slots[0].ID, cycle.Slots[1].ID

	// ── Результаты по не начатому слоту — 400 ──
	ts.post(t, "/api/slots/"+emptySlotID+"/results",
		`{"athlete_id":"nope","time_seconds":600}`, http.StatusBadRequest)

	// ── Назначение и старт ──
	var recs []map[string]any
	ts.getJSON(t, "/api/slots/"+slotID+"/recommendations", &recs)
	tplID, _ := recs[0]["template_id"].(string)
	ts.post(t, "/api/slots/"+slotID+"/assign",
		fmt.Sprintf(`{"template_id":%q,"group_level":"intermediate"}`, tplID), http.StatusCreated)
	ts.post(t, "/api/slots/"+slotID+"/start", "", http.StatusOK)
	athID := ts.firstAthleteID(t)

	// ── Первый результат: PR по времени + 1ПМ по движению (Epley) ──
	// 60 кг × 6 повторений → 60 × (1 + 6/30) = 72 кг.
	rec = ts.post(t, "/api/slots/"+slotID+"/results", fmt.Sprintf(`{
		"athlete_id":%q,"time_seconds":700,"rpe":8,"scaled_version":"rx",
		"movements":[{"movement_key":"wall_ball","weight_kg":60,"reps":6}]}`, athID), http.StatusCreated)
	var r1 struct {
		ResultID string `json:"result_id"`
		PR       struct {
			IsPR bool   `json:"is_pr"`
			Type string `json:"type"`
		} `json:"pr"`
		Prev []map[string]any `json:"previous_attempts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &r1); err != nil {
		t.Fatal(err)
	}
	if r1.ResultID == "" {
		t.Fatal("пустой result_id")
	}
	if !r1.PR.IsPR || r1.PR.Type != "fastest_time" {
		t.Fatalf("первый результат — не PR по времени: %+v", r1.PR)
	}
	if len(r1.Prev) != 0 {
		t.Fatalf("у первой попытки есть история: %v", r1.Prev)
	}
	var est float64
	if err := ts.d.QueryRow(
		`SELECT est_1rm FROM athlete_1rm WHERE athlete_id = ? AND movement_key = 'wall_ball'`, athID,
	).Scan(&est); err != nil {
		t.Fatal("нет расчётного 1ПМ после результата:", err)
	}
	if est != 72 {
		t.Fatalf("1ПМ по Эпли = %v, хочу 72", est)
	}

	// ── Второй, медленнее: не PR, но история попыток есть ──
	rec = ts.post(t, "/api/slots/"+slotID+"/results", fmt.Sprintf(`{
		"athlete_id":%q,"time_seconds":750,"scaled_version":"rx",
		"movements":[{"movement_key":"wall_ball","weight_kg":50,"reps":5}]}`, athID), http.StatusCreated)
	var r2 struct {
		PR struct {
			IsPR bool `json:"is_pr"`
		} `json:"pr"`
		Prev []map[string]any `json:"previous_attempts"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &r2); err != nil {
		t.Fatal(err)
	}
	if r2.PR.IsPR {
		t.Fatal("медленный результат отмечен как PR")
	}
	if len(r2.Prev) != 1 {
		t.Fatalf("хочу 1 предыдущую попытку, получено %d", len(r2.Prev))
	}
	// 50 × (1 + 5/30) ≈ 58.3 < 72 — 1ПМ не должен перезаписаться.
	if err := ts.d.QueryRow(
		`SELECT est_1rm FROM athlete_1rm WHERE athlete_id = ? AND movement_key = 'wall_ball'`, athID,
	).Scan(&est); err != nil {
		t.Fatal(err)
	}
	if est != 72 {
		t.Fatalf("1ПМ перезаписан худшим результатом: %v", est)
	}

	// ── Список результатов слота ──
	var results []map[string]any
	ts.getJSON(t, "/api/slots/"+slotID+"/results", &results)
	if len(results) != 2 {
		t.Fatalf("хочу 2 результата в слоте, получено %d", len(results))
	}
}
