package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"

	appconfig "github.com/maxdukov/cf/backend-go/internal/config"
	"github.com/maxdukov/cf/backend-go/internal/db"
	"github.com/maxdukov/cf/backend-go/internal/services"
	"github.com/maxdukov/cf/backend-go/internal/ws"
)

// testServer — изолированный роутер с временной БД.
type testServer struct {
	r chi.Router
	d *sql.DB
}

func newTestServer(t *testing.T) *testServer {
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
	hub := ws.NewHub()
	hr := services.NewHRProcessor(d, hub)
	app := NewApp(appconfig.Config{DevMode: true}, d, hub, hr)
	return &testServer{r: NewRouter(app, ""), d: d}
}

func (ts *testServer) close() {}

func (ts *testServer) request(t *testing.T, method, path string, wantStatus int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	ts.r.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s: статус %d, хочу %d, тело: %s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	return rec
}

func (ts *testServer) post(t *testing.T, path, body string, wantStatus int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ts.r.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("POST %s: статус %d, хочу %d, тело: %s", path, rec.Code, wantStatus, rec.Body.String())
	}
	return rec
}

func (ts *testServer) getJSON(t *testing.T, path string, out any) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	ts.r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s: статус %d, тело: %s", path, rec.Code, rec.Body.String())
	}
	if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
}

func (ts *testServer) firstAthleteID(t *testing.T) string {
	t.Helper()
	id := db.NewUUID()
	if _, err := ts.d.Exec(
		`INSERT INTO athletes (id, name, max_hr) VALUES (?, 'Тест Спортсмен', 190)`, id,
	); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestCyclesWorkflow(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// ── Создание цикла ──
	body := `{"name":"Тест-цикл","goal":"Сила","weeks":2,"start_date":"2026-09-07",
	          "modality_priority":"strength",
	          "groups":[{"name":"А","weekdays":[1,3,5]},{"name":"Б","weekdays":[2,4]}]}`
	rec := ts.post(t, "/api/cycles", body, http.StatusCreated)
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" {
		t.Fatal("пустой id цикла")
	}

	// ── Календарь: 2 недели (Пн,Ср,Пт)=6 + (Вт,Чт)=4 слотов, все empty ──
	var cycle struct {
		Slots []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Date   string `json:"slot_date"`
		} `json:"slots"`
		SlotsTotal  int     `json:"slots_total"`
		FillPercent float64 `json:"fill_percent"`
	}
	ts.getJSON(t, "/api/cycles/"+created.ID, &cycle)
	if cycle.SlotsTotal != 10 {
		t.Fatalf("ожидалось 10 слотов, получено %d", cycle.SlotsTotal)
	}
	if len(cycle.Slots) != 10 {
		t.Fatalf("ожидалось 10 слотов в календаре, получено %d", len(cycle.Slots))
	}
	if len(cycle.Slots[0].Date) != 10 {
		t.Fatalf("дата слота не нормализована: %q", cycle.Slots[0].Date)
	}
	if cycle.Slots[0].Status != "empty" {
		t.Fatalf("статус нового слота = %q, хочу empty", cycle.Slots[0].Status)
	}
	slotID := cycle.Slots[0].ID

	// ── Рекомендации ──
	var recs []map[string]any
	ts.getJSON(t, "/api/slots/"+slotID+"/recommendations", &recs)
	if len(recs) == 0 {
		t.Fatal("нет рекомендаций для слота")
	}
	if _, ok := recs[0]["reason"]; !ok {
		t.Fatal("у рекомендации нет обоснования")
	}
	tplID, _ := recs[0]["template_id"].(string)

	// ── Назначение ──
	ts.post(t, "/api/slots/"+slotID+"/assign",
		fmt.Sprintf(`{"template_id":%q,"group_level":"intermediate"}`, tplID), http.StatusCreated)
	slot := map[string]any{}
	ts.getJSON(t, "/api/slots/"+slotID, &slot)
	if slot["status"] != "planned" {
		t.Fatalf("статус после назначения = %v, хочу planned", slot["status"])
	}
	wod := slot["wod"].(map[string]any)
	if wod["is_active"] != false {
		t.Fatal("WoD слота не должен быть активным на мониторе")
	}
	if len(wod["movements"].([]any)) == 0 {
		t.Fatal("у назначенного WoD нет движений")
	}

	// ── Старт: сессия + активный WoD ──
	rec = ts.post(t, "/api/slots/"+slotID+"/start", "", http.StatusOK)
	var started struct {
		SessionID string `json:"session_id"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &started)
	if started.SessionID == "" {
		t.Fatal("нет session_id после старта")
	}
	var active map[string]any
	ts.getJSON(t, "/api/wods/active", &active)
	if active == nil || active["name"] == nil {
		t.Fatal("WoD слота не стал активным после старта")
	}

	// ── Повторный старт должен быть ошибкой ──
	ts.post(t, "/api/slots/"+slotID+"/start", "", http.StatusBadRequest)

	// ── Результаты: два раза — второй быстрее (PR) ──
	athID := ts.firstAthleteID(t)
	ts.post(t, "/api/slots/"+slotID+"/results",
		fmt.Sprintf(`{"athlete_id":%q,"time_seconds":700,"rpe":8,"scaled_version":"rx"}`, athID), http.StatusCreated)
	rec = ts.post(t, "/api/slots/"+slotID+"/results",
		fmt.Sprintf(`{"athlete_id":%q,"time_seconds":650,"rpe":9,"scaled_version":"rx"}`, athID), http.StatusCreated)
	var res struct {
		PR struct {
			IsPR bool `json:"is_pr"`
		} `json:"pr"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if !res.PR.IsPR {
		t.Fatal("второй результат быстрее — должен быть PR")
	}

	// ── Завершение ──
	ts.request(t, http.MethodPost, "/api/slots/"+slotID+"/complete", http.StatusNoContent)
	ts.getJSON(t, "/api/wods/active", &active)
	if active != nil {
		t.Fatal("после завершения активный WoD должен быть погашен")
	}

	// ── Аналитика ──
	var analytics struct {
		SlotsCompleted int `json:"slots_completed"`
		Athletes       []struct {
			Workouts int `json:"workouts_done"`
		} `json:"athletes"`
	}
	ts.getJSON(t, "/api/cycles/"+created.ID+"/analytics", &analytics)
	if analytics.SlotsCompleted != 1 {
		t.Fatalf("completed = %d, хочу 1", analytics.SlotsCompleted)
	}
	if len(analytics.Athletes) != 1 || analytics.Athletes[0].Workouts != 1 {
		t.Fatalf("некорректная статистика спортсмена: %+v", analytics.Athletes)
	}
}

func TestCustomWodConstructor(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Каталог движений.
	var movements []map[string]any
	ts.getJSON(t, "/api/movements", &movements)
	if len(movements) == 0 {
		t.Fatal("каталог движений пуст")
	}

	// Конструктор с автоименем.
	rec := ts.post(t, "/api/wods/custom", `{
		"format":"amrap","duration_min":15,"intensity":"high","theme":"full_body",
		"movements":[
			{"movement_key":"wall_ball","reps":20},
			{"movement_key":"pull_up","reps":10},
			{"movement_key":"box_jump","reps":15}
		]}`, http.StatusCreated)
	var created struct {
		TemplateID string   `json:"template_id"`
		Warnings   []string `json:"warnings"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.TemplateID == "" {
		t.Fatal("пустой template_id")
	}

	// Кастомный шаблон доступен в библиотеке и имеет предпросмотр.
	var preview map[string]any
	ts.getJSON(t, "/api/wods/templates/"+created.TemplateID+"?group_level=beginner", &preview)
	if preview["format"] != "amrap" {
		t.Fatalf("предпросмотр кастомного шаблона: %+v", preview)
	}

	// Неизвестное движение — 400.
	ts.post(t, "/api/wods/custom", `{
		"format":"amrap","duration_min":10,"intensity":"medium","theme":"core",
		"movements":[{"movement_key":"nonexistent"}]}`, http.StatusBadRequest)
}

func TestCycle48hWarnings(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// Дни подряд — предупреждение при создании.
	rec := ts.post(t, "/api/cycles", `{
		"name":"Без отдыха","weeks":1,"start_date":"2026-09-07",
		"groups":[{"name":"А","weekdays":[2,3]}]}`, http.StatusCreated)
	var created struct {
		Warnings []string `json:"warnings"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	if len(created.Warnings) == 0 {
		t.Fatal("нет предупреждения о 48 часах при соседних днях")
	}
}
