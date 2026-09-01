// Package services — бизнес-логика: обработка ЧСС-пайплайна, генератор WoD.
package services

import (
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/maxdukov/cf/backend-go/internal/db"
	"github.com/maxdukov/cf/backend-go/internal/hrzones"
	"github.com/maxdukov/cf/backend-go/internal/ws"
)

// HrUpdate — формат WebSocket-сообщения hr_update клиенту.
type HrUpdate struct {
	Type        string  `json:"type"`
	DeviceID    int     `json:"device_id"`
	AthleteID   *string `json:"athlete_id"`
	AthleteName *string `json:"athlete_name"`
	HeartRate   int     `json:"heart_rate"`
	Zone        int     `json:"zone"`
	ZonePercent float64 `json:"zone_percent"`
	MaxHr       int     `json:"max_hr"`
	Calories    float64 `json:"calories"`
}

// NewSensorEvent — WebSocket-событие: обнаружен новый датчик.
type NewSensorEvent struct {
	Type     string `json:"type"`
	DeviceID int    `json:"device_id"`
}

// HRProcessor — обработчик событий ЧСС от коллекторов.
// Порт _on_hr_data/_on_new_sensor из FastAPI-версии + исправленный
// дефект: точки hr_readings теперь пишутся в БД (аналитика живая).
type HRProcessor struct {
	db  *sql.DB
	hub *ws.Hub

	mu          sync.Mutex
	caloriesAcc map[int]float64
	lastHrTime  map[int]time.Time // по wall clock, как time.monotonic() в Python
}

// NewHRProcessor создаёт процессор ЧСС.
func NewHRProcessor(d *sql.DB, hub *ws.Hub) *HRProcessor {
	return &HRProcessor{
		db:          d,
		hub:         hub,
		caloriesAcc: make(map[int]float64),
		lastHrTime:  make(map[int]time.Time),
	}
}

// ResetState сбрасывает накопленные калории (при переключении режима).
func (p *HRProcessor) ResetState() {
	p.mu.Lock()
	p.caloriesAcc = make(map[int]float64)
	p.lastHrTime = make(map[int]time.Time)
	p.mu.Unlock()
}

// OnHRData — callback из коллектора: рассылает HR через WebSocket,
// пишет точку в hr_readings и обновляет калории.
func (p *HRProcessor) OnHRData(deviceID, hr, battery int) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("hr callback panic", "recover", r)
		}
	}()

	var (
		athleteID   sql.NullString
		athleteName sql.NullString
		maxHr       = 190
		weightKg    sql.NullFloat64
		age         sql.NullInt64
	)
	err := p.db.QueryRow(`
		SELECT s.athlete_id, a.name, a.max_hr, a.weight_kg, a.age
		FROM sensors s
		LEFT JOIN athletes a ON a.id = s.athlete_id
		WHERE s.device_id = ?`, deviceID,
	).Scan(&athleteID, &athleteName, &maxHr, &weightKg, &age)
	if err != nil {
		slog.Error("hr callback: sensor query", "device_id", deviceID, "err", err)
		return
	}

	zone := hrzones.CalcZone(hr, maxHr)
	pct := hrzones.CalcPercent(hr, maxHr)

	// ── Калории: накопление ккал/мин × dt ──
	var wp *float64
	if weightKg.Valid {
		wp = &weightKg.Float64
	}
	var ap *int
	if age.Valid {
		v := int(age.Int64)
		ap = &v
	}

	now := time.Now()
	p.mu.Lock()
	if prev, ok := p.lastHrTime[deviceID]; ok {
		dtMin := now.Sub(prev).Minutes()
		rate := hrzones.CalcCaloriesPerMin(hr, wp, ap)
		p.caloriesAcc[deviceID] += rate * dtMin
	}
	p.lastHrTime[deviceID] = now
	calories := round1(p.caloriesAcc[deviceID])
	p.mu.Unlock()

	payload := HrUpdate{
		Type:        "hr_update",
		DeviceID:    deviceID,
		AthleteID:   nullStrPtr(athleteID),
		AthleteName: nullStrPtr(athleteName),
		HeartRate:   hr,
		Zone:        zone,
		ZonePercent: pct,
		MaxHr:       maxHr,
		Calories:    calories,
	}
	p.hub.Broadcast(payload)

	// ── Запись точки hr_readings (исправление дефекта Python-версии) ──
	if athleteID.Valid && athleteID.String != "" {
		var sessionID sql.NullString
		_ = p.db.QueryRow(`
			SELECT sa.session_id
			FROM session_athletes sa
			JOIN sessions s ON s.id = sa.session_id
			WHERE sa.athlete_id = ? AND sa.left_at IS NULL AND s.ended_at IS NULL
			ORDER BY sa.joined_at DESC LIMIT 1`, athleteID.String,
		).Scan(&sessionID)

		var sess any
		if sessionID.Valid {
			sess = sessionID.String
		}
		if _, err := p.db.Exec(
			`INSERT INTO hr_readings (athlete_id, session_id, heart_rate, zone, timestamp)
			 VALUES (?, ?, ?, ?, ?)`,
			athleteID.String, sess, hr, zone, db.NowDB(),
		); err != nil {
			slog.Error("hr callback: insert reading", "err", err)
		}
	}
}

// OnNewSensor — callback: новый датчик обнаружен, уведомляем фронтенд.
func (p *HRProcessor) OnNewSensor(deviceID int) {
	p.hub.Broadcast(NewSensorEvent{Type: "new_sensor", DeviceID: deviceID})
}

func nullStrPtr(ns sql.NullString) *string {
	if !ns.Valid || ns.String == "" {
		return nil
	}
	return &ns.String
}

func round1(v float64) float64 {
	return float64(int(v*10+0.5*sign(v))) / 10
}

func sign(v float64) float64 {
	if v < 0 {
		return -1
	}
	return 1
}
