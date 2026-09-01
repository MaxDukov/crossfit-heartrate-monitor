// Package hrzones — расчёт пульсовых зон по % от максимальной ЧСС.
//
// Зоны (по ТЗ):
//
//	Zone 1: ≤60% max_hr — Восстановление (blue)
//	Zone 2: 61-80%      — Умеренная      (green)
//	Zone 3: 81-100%     — Высокая        (amber)
//	Zone 4: >100%       — Критическая    (red)
package hrzones

import "math"

// CalcZone возвращает номер зоны (1-4) для текущего пульса.
func CalcZone(hr, maxHR int) int {
	pct := float64(hr) / float64(maxHR) * 100
	switch {
	case pct <= 60:
		return 1
	case pct <= 80:
		return 2
	case pct <= 100:
		return 3
	default:
		return 4
	}
}

// CalcPercent возвращает % от максимальной ЧСС (округлённый до 0.1).
func CalcPercent(hr, maxHR int) float64 {
	return math.Round(float64(hr)/float64(maxHR)*100*10) / 10
}

// CalcCaloriesPerMin — расход калорий (ккал/мин).
//
// Формула Keyt (при наличии веса и возраста):
//
//	(0.6309*HR + 0.1988*вес + 0.2017*возраст - 55.0963) / 4.184
//
// Упрощённая: HR * 0.014.
func CalcCaloriesPerMin(hr int, weightKg *float64, age *int) float64 {
	var kcal float64
	if weightKg != nil && age != nil && *weightKg > 0 {
		kcal = (0.6309*float64(hr) + 0.1988**weightKg + 0.2017*float64(*age) - 55.0963) / 4.184
	} else {
		kcal = float64(hr) * 0.014
	}
	if kcal < 0 {
		kcal = 0
	}
	return math.Round(kcal*100) / 100
}
