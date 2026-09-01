// Package collectors — интерфейс источников ЧСС-данных (mock / ANT+).
package collectors

import (
	"database/sql"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/maxdukov/cf/backend-go/internal/db"
)

// Callbacks — события, которые источник данных отправляет в приложение.
type Callbacks struct {
	// OnHRData вызывается при каждом новом значении пульса.
	OnHRData func(deviceID int, hr int, battery int)
	// OnNewSensor вызывается при обнаружении нового датчика.
	OnNewSensor func(deviceID int)
}

// Collector — источник ЧСС-данных. Реализации: mock, ant (ANT+ USB-стик).
type Collector interface {
	Start()
	Stop()
}

// UpsertSensor создаёт запись датчика, если её ещё нет.
func UpsertSensor(d *sql.DB, deviceID int) {
	_, err := d.Exec(
		`INSERT INTO sensors (device_id, athlete_id, ignored) VALUES (?, NULL, 0)
		 ON CONFLICT(device_id) DO NOTHING`,
		deviceID,
	)
	if err != nil {
		slog.Error("upsert sensor", "device_id", deviceID, "err", err)
	}
}

// UpdateSensorHR обновляет последние показания датчика.
// battery == 0xFF (255) — маркер «батарея неизвестна», не перезаписываем.
func UpdateSensorHR(d *sql.DB, deviceID, hr, battery int) {
	var sb strings.Builder
	sb.WriteString(`UPDATE sensors SET last_hr = ?, last_seen_at = ?`)
	args := []any{hr, db.NowDB()}
	if battery != 0xFF {
		sb.WriteString(`, battery_level = ?`)
		args = append(args, battery)
	}
	sb.WriteString(` WHERE device_id = ?`)
	args = append(args, deviceID)
	_, err := d.Exec(sb.String(), args...)
	if err != nil {
		slog.Error("update sensor hr", "device_id", deviceID, "err", err)
	}
}

// IsSensorAssigned — привязан ли датчик к спортсмену.
func IsSensorAssigned(d *sql.DB, deviceID int) bool {
	var athleteID sql.NullString
	err := d.QueryRow(`SELECT athlete_id FROM sensors WHERE device_id = ?`, deviceID).Scan(&athleteID)
	if err != nil {
		return false
	}
	return athleteID.Valid && athleteID.String != ""
}

// IsSensorIgnored — помечен ли датчик как проигнорированный.
func IsSensorIgnored(d *sql.DB, deviceID int) bool {
	var ignored bool
	err := d.QueryRow(`SELECT ignored FROM sensors WHERE device_id = ?`, deviceID).Scan(&ignored)
	return err == nil && ignored
}

// Dedup — дедупликатор «тот же пульс чаще, чем раз в N секунд».
type Dedup struct {
	mu      sync.Mutex
	lastHR  int
	hasLast bool
	lastAt  time.Time
	window  time.Duration
}

// NewDedup создаёт дедупликатор с окном (например, 2 секунды).
func NewDedup(window time.Duration) *Dedup { return &Dedup{window: window} }

// Pass возвращает true, если событие нужно обработать.
func (dd *Dedup) Pass(hr int, now time.Time) bool {
	dd.mu.Lock()
	defer dd.mu.Unlock()
	if dd.hasLast && dd.lastHR == hr && now.Sub(dd.lastAt) < dd.window {
		return false
	}
	dd.lastHR = hr
	dd.lastAt = now
	dd.hasLast = true
	return true
}
