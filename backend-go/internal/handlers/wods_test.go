package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// createTemplate — кастомный шаблон через конструктор (POST /api/wods/custom).
func createTemplate(t *testing.T, ts *testServer, name string) string {
	t.Helper()
	rec := ts.post(t, "/api/wods/custom", fmt.Sprintf(`{
		"name":%q,"format":"amrap","duration_min":12,"intensity":"high","theme":"gymnastics",
		"movements":[{"movement_key":"burpee","reps":15}]}`, name), http.StatusCreated)
	var out struct {
		TemplateID string `json:"template_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.TemplateID == "" {
		t.Fatal("пустой template_id")
	}
	return out.TemplateID
}

func (ts *testServer) postJSON(t *testing.T, path, body string, wantStatus int) map[string]any {
	t.Helper()
	rec := ts.post(t, path, body, wantStatus)
	var out map[string]any
	if len(rec.Body.Bytes()) > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("POST %s: %v", path, err)
		}
	}
	return out
}

func (ts *testServer) doJSON(t *testing.T, method, path string, wantStatus int) map[string]any {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	ts.r.ServeHTTP(rec, req)
	if rec.Code != wantStatus {
		t.Fatalf("%s %s: статус %d, хочу %d, тело: %s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	var out map[string]any
	if len(rec.Body.Bytes()) > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("%s %s: %v", method, path, err)
		}
	}
	return out
}

func templateNames(t *testing.T, ts *testServer, query string) map[string]bool {
	t.Helper()
	var list []struct {
		Name     string `json:"name"`
		Archived bool   `json:"archived"`
	}
	ts.getJSON(t, "/api/wods/templates?search=Тест&limit=200"+query, &list)
	names := map[string]bool{}
	for _, tpl := range list {
		names[tpl.Name] = tpl.Archived
	}
	return names
}

func TestTemplateDeleteArchiveRestore(t *testing.T) {
	ts := newTestServer(t)
	defer ts.close()

	// ── 1. Жёсткое удаление без истории; повторное — 404 ──
	tpl0 := createTemplate(t, ts, "Тест Удаление")
	out := ts.doJSON(t, http.MethodDelete, "/api/wods/templates/"+tpl0, http.StatusOK)
	if out["deleted"] != true {
		t.Fatalf("удаление без истории: %v", out)
	}
	ts.doJSON(t, http.MethodDelete, "/api/wods/templates/"+tpl0, http.StatusNotFound)

	// ── 2. Цикл с тремя слотами ──
	rec := ts.post(t, "/api/cycles", `{"name":"Цикл архива","weeks":1,"start_date":"2026-09-07",
		"groups":[{"name":"А","weekdays":[1,3,5]}]}`, http.StatusCreated)
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
	if len(cycle.Slots) != 3 {
		t.Fatalf("хочу 3 слота, получено %d", len(cycle.Slots))
	}
	slot0, slot1, slot2 := cycle.Slots[0].ID, cycle.Slots[1].ID, cycle.Slots[2].ID

	tpl1 := createTemplate(t, ts, "Тест В Плане")
	tpl2 := createTemplate(t, ts, "Тест Архивный")
	ts.postJSON(t, "/api/slots/"+slot0+"/assign",
		fmt.Sprintf(`{"template_id":%q,"group_level":"intermediate"}`, tpl1), http.StatusCreated)
	ts.postJSON(t, "/api/slots/"+slot1+"/assign",
		fmt.Sprintf(`{"template_id":%q,"group_level":"intermediate"}`, tpl2), http.StatusCreated)

	// ── 3. Шаблон в запланированном слоте — 409 со списком циклов ──
	conflict := ts.doJSON(t, http.MethodDelete, "/api/wods/templates/"+tpl1, http.StatusConflict)
	if detail, _ := conflict["detail"].(string); !strings.Contains(detail, "включена в план") {
		t.Fatalf("409 без объяснения: %v", conflict)
	}
	if refs, _ := conflict["cycles"].([]any); len(refs) == 0 {
		t.Fatalf("409 без списка циклов: %v", conflict)
	}

	// ── 4. Выполнявшийся шаблон (есть сессия) — архивация ──
	ts.postJSON(t, "/api/slots/"+slot1+"/start", "", http.StatusOK)
	out = ts.doJSON(t, http.MethodDelete, "/api/wods/templates/"+tpl2, http.StatusOK)
	if out["archived"] != true {
		t.Fatalf("удаление выполнявшегося: %v", out)
	}

	// ── 5. Архивный скрыт по умолчанию, виден с include_archived=1 ──
	if _, ok := templateNames(t, ts, "")["Тест Архивный"]; ok {
		t.Fatal("архивный шаблон виден в списке по умолчанию")
	}
	archived := templateNames(t, ts, "&include_archived=1")
	if !archived["Тест Архивный"] {
		t.Fatal("архивный шаблон не найден в include_archived=1")
	}

	// ── 6. Восстановление из архива ──
	ts.doJSON(t, http.MethodPost, "/api/wods/templates/"+tpl2+"/restore", http.StatusOK)
	if _, ok := templateNames(t, ts, "")["Тест Архивный"]; !ok {
		t.Fatal("шаблон не вернулся в активные после restore")
	}

	// ── 7. Назначение архивного шаблона — ошибка ──
	if _, err := ts.d.Exec(`UPDATE wod_templates SET archived = 1 WHERE id = ?`, tpl2); err != nil {
		t.Fatal(err)
	}
	out = ts.postJSON(t, "/api/slots/"+slot2+"/assign",
		fmt.Sprintf(`{"template_id":%q,"group_level":"intermediate"}`, tpl2), http.StatusBadRequest)
	if detail, _ := out["detail"].(string); !strings.Contains(detail, "в архиве") {
		t.Fatalf("назначение архивного прошло без ошибки: %v", out)
	}
}
