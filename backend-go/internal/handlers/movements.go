package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

// Допустимые значения признаков движения (каталог, draft1.MD).
var (
	movementModalities   = map[string]bool{"gymnastics": true, "weightlifting": true, "monostructural": true}
	movementMuscleGroups = map[string]bool{"legs": true, "upper_pull": true, "upper_push": true, "chest": true, "core": true, "full_body": true}
	movementDifficulties = map[string]bool{"beginner": true, "intermediate": true, "advanced": true}
	movementKeyRe        = regexp.MustCompile(`^[a-z0-9_]{2,80}$`)
)

type movementReq struct {
	Key                 string   `json:"key"`
	Name                string   `json:"name"`
	Modality            string   `json:"modality"`
	MuscleGroup         string   `json:"muscle_group"`
	Themes              []string `json:"themes"`
	EquipmentKeys       []string `json:"equipment_keys"`
	Difficulty          string   `json:"difficulty"`
	ScalingBeginner     string   `json:"scaling_beginner"`
	ScalingIntermediate string   `json:"scaling_intermediate"`
}

// ruTranslit — транслитерация для генерации ключа из русского названия.
var ruTranslit = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e", 'ж': "zh",
	'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o",
	'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "h", 'ц': "ts",
	'ч': "ch", 'ш': "sh", 'щ': "sch", 'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
}

// strPtr — пустая строка → NULL.
func strPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

// movementSlug — ключ нового движения из названия («Бёрпи» → burpee).
func movementSlug(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		if s, ok := ruTranslit[r]; ok {
			b.WriteString(s)
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			continue
		}
		b.WriteByte('_')
	}
	s := b.String()
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	return strings.Trim(s, "_")
}

// validateMovement проверяет поля и приводит темы/инвентарь к CSV.
func (a *App) validateMovement(req *movementReq) (themes, equipment string, err error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return "", "", fmt.Errorf("укажите название упражнения")
	}
	if !movementModalities[req.Modality] {
		return "", "", fmt.Errorf("недопустимая модальность: %s", req.Modality)
	}
	if !movementMuscleGroups[req.MuscleGroup] {
		return "", "", fmt.Errorf("недопустимая группа мышц: %s", req.MuscleGroup)
	}
	if !movementDifficulties[req.Difficulty] {
		return "", "", fmt.Errorf("недопустимая сложность: %s", req.Difficulty)
	}
	// Темы — только известные (используются генератором и фильтрами).
	themeList := []string{}
	seen := map[string]bool{}
	for _, t := range req.Themes {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		themeList = append(themeList, t)
	}
	themes = strings.Join(themeList, ",")
	// Инвентарь должен существовать в каталоге оборудования.
	eqList := []string{}
	seenEq := map[string]bool{}
	for _, e := range req.EquipmentKeys {
		e = strings.ToLower(strings.TrimSpace(e))
		if e == "" || seenEq[e] {
			continue
		}
		var n int
		if err := a.DB.QueryRow(`SELECT COUNT(*) FROM equipment WHERE "key" = ?`, e).Scan(&n); err != nil {
			return "", "", err
		}
		if n == 0 {
			return "", "", fmt.Errorf("неизвестное оборудование: %s", e)
		}
		seenEq[e] = true
		eqList = append(eqList, e)
	}
	equipment = strings.Join(eqList, ",")
	return themes, equipment, nil
}

// CreateMovement: POST /api/movements — новое движение (is_custom = 1).
func (a *App) CreateMovement(w http.ResponseWriter, r *http.Request) {
	var req movementReq
	if !decodeJSON(w, r, &req) {
		return
	}
	themes, equipment, err := a.validateMovement(&req)
	if err != nil {
		httpError(w, 400, err.Error())
		return
	}
	key := movementSlug(req.Key)
	if key == "" {
		key = movementSlug(req.Name)
	}
	if !movementKeyRe.MatchString(key) {
		httpError(w, 400, "ключ должен быть из латиницы, цифр и «_» (2–80 символов)")
		return
	}
	var n int
	if err := a.DB.QueryRow(`SELECT COUNT(*) FROM movements WHERE "key" = ?`, key).Scan(&n); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if n > 0 {
		httpError(w, 409, fmt.Sprintf("упражнение с ключом «%s» уже есть", key))
		return
	}
	if _, err := a.DB.Exec(
		`INSERT INTO movements ("key", name, modality, muscle_group, themes, equipment_keys, difficulty, scaling_beginner, scaling_intermediate, is_custom)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1)`,
		key, req.Name, req.Modality, req.MuscleGroup, themes, equipment,
		req.Difficulty, strPtr(req.ScalingBeginner), strPtr(req.ScalingIntermediate),
	); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, map[string]any{"key": key})
}

// DeleteMovement: DELETE /api/movements/{key} — убрать дубли и лишнее.
// Движение, входящее в шаблоны тренировок, удалить нельзя (сломает веса);
// в ответе 409 возвращается список шаблонов для перехода.
func (a *App) DeleteMovement(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	rows, err := a.DB.Query(`
		SELECT t.id, t.name FROM wod_template_movements wm
		JOIN wod_templates t ON t.id = wm.template_id
		WHERE wm.movement_key = ? GROUP BY t.id, t.name ORDER BY t.name`, key)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	type tplRef struct {
		ID   string `json:"template_id"`
		Name string `json:"name"`
	}
	used := []tplRef{}
	for rows.Next() {
		var t tplRef
		if rows.Scan(&t.ID, &t.Name) == nil {
			used = append(used, t)
		}
	}
	_ = rows.Close()
	if len(used) > 0 {
		writeJSON(w, 409, map[string]any{
			"detail":    fmt.Sprintf("движение входит в %d шаблон(ов) тренировок", len(used)),
			"templates": used,
		})
		return
	}
	res, err := a.DB.Exec(`DELETE FROM movements WHERE "key" = ?`, key)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		httpError(w, 404, "упражнение не найдено")
		return
	}
	// Надгробие: seed-refresh не должен вернуть движение после рестарта.
	if _, err := a.DB.Exec(`INSERT OR IGNORE INTO movements_deleted ("key") VALUES (?)`, key); err != nil {
		httpError(w, 500, err.Error())
		return
	}
	w.WriteHeader(204)
}

// UpdateMovement: PUT /api/movements/{key} — правка признаков (is_custom = 1).
func (a *App) UpdateMovement(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var req movementReq
	if !decodeJSON(w, r, &req) {
		return
	}
	themes, equipment, err := a.validateMovement(&req)
	if err != nil {
		httpError(w, 400, err.Error())
		return
	}
	res, err := a.DB.Exec(
		`UPDATE movements SET name = ?, modality = ?, muscle_group = ?, themes = ?,
		   equipment_keys = ?, difficulty = ?, scaling_beginner = ?, scaling_intermediate = ?, is_custom = 1
		 WHERE "key" = ?`,
		req.Name, req.Modality, req.MuscleGroup, themes, equipment,
		req.Difficulty, strPtr(req.ScalingBeginner), strPtr(req.ScalingIntermediate), key,
	)
	if err != nil {
		httpError(w, 500, err.Error())
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		httpError(w, 404, "упражнение не найдено")
		return
	}
	writeJSON(w, 200, map[string]any{"key": key})
}
