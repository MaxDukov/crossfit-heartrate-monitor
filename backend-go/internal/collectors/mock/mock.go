// Package mock — генератор виртуальных ЧСС-данных для разработки.
package mock

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"database/sql"

	"github.com/maxdukov/cf/backend-go/internal/collectors"
	"github.com/maxdukov/cf/backend-go/internal/db"
)

var mockNames = []string{"Анна", "Борис", "Вера", "Глеб", "Дина", "Егор", "Жанна", "Захар"}
var mockWeights = []float64{62, 85, 58, 78, 65, 90, 55, 82}
var mockAges = []int{28, 35, 24, 31, 40, 26, 33, 29}

// rangePair — [min, max] пульса для виртуального датчика.
type rangePair struct{ lo, hi int }

var mockRanges = map[int]rangePair{
	1: {60, 90}, 2: {60, 90},
	3: {90, 160}, 4: {90, 160},
	5: {140, 200}, 6: {140, 200},
	7: {160, 220}, 8: {160, 220},
}

var mockOrder = []int{1, 2, 3, 4, 5, 6, 7, 8}

// Collector — фоновый генератор mock-данных (8 виртуальных датчиков).
type Collector struct {
	db        *sql.DB
	callbacks collectors.Callbacks
	stop      context.CancelFunc
	done      chan struct{}
}

// New создаёт mock-коллектор.
func New(d *sql.DB, cb collectors.Callbacks) *Collector {
	return &Collector{db: d, callbacks: cb, done: make(chan struct{})}
}

// Start запускает генерацию в фоновой горутине.
func (m *Collector) Start() {
	go m.run()
	slog.Info("mock collector started (8 virtual sensors)")
}

// Stop останавливает генерацию.
func (m *Collector) Stop() {
	if m.stop != nil {
		m.stop()
	}
	<-m.done
	slog.Info("mock collector stopped")
}

func (m *Collector) run() {
	ctx, cancel := context.WithCancel(context.Background())
	m.stop = cancel
	defer close(m.done)

	m.ensureMockAthletes()

	// Плавный «разогрев»: датчики появляются по одному, как реальные ANT+.
	for _, deviceID := range mockOrder {
		collectors.UpsertSensor(m.db, deviceID)
		if m.callbacks.OnNewSensor != nil {
			m.callbacks.OnNewSensor(deviceID)
		}
		select {
		case <-time.After(300 * time.Millisecond):
		case <-ctx.Done():
			return
		}
	}

	prevHR := make(map[int]int)
	for {
		for _, deviceID := range mockOrder {
			select {
			case <-ctx.Done():
				return
			default:
			}
			r := mockRanges[deviceID]

			prev, ok := prevHR[deviceID]
			var hr int
			if ok {
				drift := rand.Intn(11) - 5 // [-5, 5]
				hr = max(r.lo, min(r.hi, prev+drift))
			} else {
				hr = r.lo + rand.Intn(r.hi-r.lo+1)
			}
			prevHR[deviceID] = hr
			battery := 70 + rand.Intn(31)

			collectors.UpdateSensorHR(m.db, deviceID, hr, battery)
			if m.callbacks.OnHRData != nil {
				m.callbacks.OnHRData(deviceID, hr, battery)
			}
		}
		select {
		case <-time.After(2 * time.Second):
		case <-ctx.Done():
			return
		}
	}
}

// ensureMockAthletes создаёт тестовых спортсменов и привязывает датчики.
func (m *Collector) ensureMockAthletes() {
	tx, err := m.db.Begin()
	if err != nil {
		slog.Error("mock athletes: begin tx", "err", err)
		return
	}
	defer func() { _ = tx.Rollback() }()

	for i, deviceID := range mockOrder {
		var athleteID string
		err := tx.QueryRow(`SELECT athlete_id FROM sensors WHERE device_id = ? AND athlete_id IS NOT NULL`, deviceID).Scan(&athleteID)
		if err == nil {
			continue // уже привязан
		}

		name := mockNames[i]
		var existingID string
		err = tx.QueryRow(`SELECT id FROM athletes WHERE name = ?`, name).Scan(&existingID)
		if err != nil {
			existingID = db.NewUUID()
			_, err = tx.Exec(
				`INSERT INTO athletes (id, name, max_hr, weight_kg, age, created_at, updated_at)
				 VALUES (?, ?, 190, ?, ?, ?, ?)`,
				existingID, name, mockWeights[i], mockAges[i], db.NowDB(), db.NowDB(),
			)
			if err != nil {
				slog.Error("mock athletes: create", "name", name, "err", err)
				continue
			}
		}

		_, err = tx.Exec(`INSERT INTO sensors (device_id, athlete_id, ignored) VALUES (?, ?, 0)
			 ON CONFLICT(device_id) DO UPDATE SET athlete_id = excluded.athlete_id`,
			deviceID, existingID)
		if err != nil {
			slog.Error("mock athletes: assign", "device_id", deviceID, "err", err)
		}
	}

	if err := tx.Commit(); err != nil {
		slog.Error("mock athletes: commit", "err", err)
		return
	}
	slog.Info("mock athletes created and assigned")
}
