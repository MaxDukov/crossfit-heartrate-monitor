// Package ant — ANT+-коллектор на базе openant-go.
//
// Порт services/ant_collector.py из Python-версии: 8 wildcard-каналов
// (device_id=0), дедупликация одинакового пульса (<2 с), upsert датчиков
// в БД, авто-рестарт через 5 с при ошибке стика.
package ant

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/maxdukov/openant-go/devices"
	"github.com/maxdukov/openant-go/easy"

	"github.com/maxdukov/cf/backend-go/internal/collectors"
)

// Collector — фоновый ANT+-коллектор (USB-стик).
type Collector struct {
	db         *sql.DB
	callbacks  collectors.Callbacks
	maxSensors int

	// nodeFactory создаёт Node (по умолчанию — реальный USB-стик);
	// заменяется в тестах на anttest-симулятор.
	nodeFactory func() (*easy.Node, error)

	mu       sync.Mutex
	knownIDs map[int]bool
	// mu защищает knownIDs и dedups (общие для всех каналов коллектора).
	dedups map[int]*dedupEntry

	stopOnce sync.Once
	stopCh   chan struct{}
	done     chan struct{}
}

// dedupEntry — состояние дедупликатора одного устройства.
type dedupEntry struct {
	dedup *collectors.Dedup
}

// New создаёт ANT+-коллектор на maxSensors wildcard-каналов.
func New(d *sql.DB, maxSensors int, cb collectors.Callbacks) *Collector {
	return newCollector(d, maxSensors, cb, func() (*easy.Node, error) { return easy.New() })
}

func newCollector(d *sql.DB, maxSensors int, cb collectors.Callbacks, factory func() (*easy.Node, error)) *Collector {
	if maxSensors <= 0 {
		maxSensors = 8
	}
	return &Collector{
		db:          d,
		callbacks:   cb,
		maxSensors:  maxSensors,
		nodeFactory: factory,
		knownIDs:    make(map[int]bool),
		dedups:      make(map[int]*dedupEntry),
		stopCh:      make(chan struct{}),
		done:        make(chan struct{}),
	}
}

// Start запускает коллектор в фоновой горутине.
func (c *Collector) Start() {
	go c.run()
	slog.Info("ANT+ collector started", "max_sensors", c.maxSensors)
}

// Stop останавливает коллектор (idempotent).
func (c *Collector) Stop() {
	c.stopOnce.Do(func() { close(c.stopCh) })
	<-c.done
	slog.Info("ANT+ collector stopped")
}

func (c *Collector) stopped() bool {
	select {
	case <-c.stopCh:
		return true
	default:
		return false
	}
}

// run — основной цикл с авто-рестартом через 5 с при ошибке.
func (c *Collector) run() {
	defer close(c.done)
	for {
		if c.stopped() {
			return
		}
		if err := c.session(); err != nil {
			if c.stopped() {
				return
			}
			slog.Error("ANT+ collector error", "err", err)
			slog.Info("restarting ANT+ collector in 5s")
			select {
			case <-time.After(5 * time.Second):
			case <-c.stopCh:
				return
			}
		}
	}
}

// session — один жизненный цикл Node: подключение, каналы, диспетчер.
func (c *Collector) session() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		select {
		case <-c.stopCh:
			cancel()
		case <-ctx.Done():
		}
	}()

	node, err := c.nodeFactory()
	if err != nil {
		return err
	}
	defer node.Stop()

	if err := node.SetNetworkKey(0x00, devices.ANTPLUS_NETWORK_KEY); err != nil {
		return err
	}

	for i := 0; i < c.maxSensors; i++ {
		hr, err := devices.NewHeartRate(node, 0 /* wildcard */, 0)
		if err != nil {
			return err
		}

		hr.OnFound = func() {
			deviceID := hr.DeviceID // реальный ID после wildcard-поиска
			if deviceID == 0 {
				return
			}
			c.mu.Lock()
			isNew := !c.knownIDs[deviceID]
			c.knownIDs[deviceID] = true
			c.mu.Unlock()

			slog.Info("sensor found", "channel", i+1, "device_id", deviceID)

			collectors.UpsertSensor(c.db, deviceID)

			if isNew && c.callbacks.OnNewSensor != nil &&
				!collectors.IsSensorAssigned(c.db, deviceID) &&
				!collectors.IsSensorIgnored(c.db, deviceID) {
				c.callbacks.OnNewSensor(deviceID)
			}
		}

		hr.OnDeviceData = func(_page int, _pageName string, data devices.DeviceData) {
			d, ok := data.(devices.HeartRateData)
			if !ok {
				return
			}
			hrVal := d.HeartRate
			if hrVal == 0 {
				return
			}

			c.mu.Lock()
			st := c.dedups[hr.DeviceID]
			if st == nil {
				st = &dedupEntry{dedup: collectors.NewDedup(2 * time.Second)}
				c.dedups[hr.DeviceID] = st
			}
			pass := st.dedup.Pass(hrVal, time.Now())
			c.mu.Unlock()
			if !pass {
				return
			}

			collectors.UpdateSensorHR(c.db, hr.DeviceID, hrVal, d.BatteryPercentage)

			if c.callbacks.OnHRData != nil {
				c.callbacks.OnHRData(hr.DeviceID, hrVal, d.BatteryPercentage)
			}
		}
	}

	node.Run(ctx) // блокирующий диспетчер до cancel/Stop
	return nil
}
