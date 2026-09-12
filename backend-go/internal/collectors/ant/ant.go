// Package ant — ANT+-коллектор на базе openant-go.
//
// Поддерживает несколько USB-стиков: supervisor опрашивает список стиков
// (ant.Sticks) и управляет per-stick сессиями — по 8 wildcard-каналов
// на стик. Восстановление после сбоя стика выполняет auto-reconnect
// внутри easy.Node.Run; подключение нового стика подхватывается
// supervisor'ом без перезапуска коллектора.
package ant

import (
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	openant "github.com/maxdukov/openant-go/ant"
	"github.com/maxdukov/openant-go/devices"
	"github.com/maxdukov/openant-go/easy"

	"github.com/maxdukov/cf/backend-go/internal/collectors"
)

// Collector — фоновый ANT+-коллектор (один или несколько USB-стиков).
type Collector struct {
	db         *sql.DB
	callbacks  collectors.Callbacks
	maxSensors int

	// stickLister возвращает список подключённых стиков (прод — openant.Sticks,
	// тесты — фейк). stickOpener строит Node для конкретного стика
	// (прод — easy.NewStick, тесты — фейк).
	stickLister func() []openant.StickInfo
	stickOpener func(openant.StickInfo) (*easy.Node, error)
	tick        time.Duration // интервал опроса списка стиков

	// mu защищает knownIDs, dedups и sessions.
	mu       sync.Mutex
	knownIDs map[int]bool
	dedups   map[int]*dedupEntry
	sessions map[string]*stickSession

	stopOnce sync.Once
	stopCh   chan struct{}
	done     chan struct{}
}

// dedupEntry — состояние дедупликатора одного устройства.
type dedupEntry struct {
	dedup *collectors.Dedup
}

// stickSession — жизненный цикл одного стика.
type stickSession struct {
	key    string // StickInfo.String()
	cancel context.CancelFunc
	done   chan struct{}
}

// New создаёт ANT+-коллектор с авто-детектом всех стиков
// (до maxSensors wildcard-каналов на каждом).
func New(d *sql.DB, maxSensors int, cb collectors.Callbacks) *Collector {
	return newCollector(d, maxSensors, cb,
		openant.Sticks,
		func(info openant.StickInfo) (*easy.Node, error) { return easy.NewStick(info) },
		10*time.Second)
}

func newCollector(d *sql.DB, maxSensors int, cb collectors.Callbacks,
	stickLister func() []openant.StickInfo,
	stickOpener func(openant.StickInfo) (*easy.Node, error),
	tick time.Duration) *Collector {
	if maxSensors <= 0 {
		maxSensors = 8
	}
	if tick <= 0 {
		tick = 10 * time.Second
	}
	return &Collector{
		db:          d,
		callbacks:   cb,
		maxSensors:  maxSensors,
		stickLister: stickLister,
		stickOpener: stickOpener,
		tick:        tick,
		knownIDs:    make(map[int]bool),
		dedups:      make(map[int]*dedupEntry),
		sessions:    make(map[string]*stickSession),
		stopCh:      make(chan struct{}),
		done:        make(chan struct{}),
	}
}

// Start запускает supervisor в фоновой горутине.
func (c *Collector) Start() {
	go c.run()
	slog.Info("ANT+ collector started", "max_sensors_per_stick", c.maxSensors)
}

// Stop останавливает supervisor и все сессии (idempotent).
func (c *Collector) Stop() {
	c.stopOnce.Do(func() { close(c.stopCh) })
	<-c.done
	slog.Info("ANT+ collector stopped")
}

// run — supervisor: сверяет список стиков с активными сессиями.
func (c *Collector) run() {
	defer close(c.done)
	ticker := time.NewTicker(c.tick)
	defer ticker.Stop()

	c.reconcileSticks()
	for {
		select {
		case <-c.stopCh:
			c.stopAllSessions()
			return
		case <-ticker.C:
			c.reconcileSticks()
		}
	}
}

// reconcileSticks синхронизирует активные сессии со списком стиков.
func (c *Collector) reconcileSticks() {
	sticks := c.stickLister()
	want := make(map[string]openant.StickInfo, len(sticks))
	for _, s := range sticks {
		want[s.String()] = s
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for key, sess := range c.sessions {
		if _, ok := want[key]; !ok {
			slog.Info("stick removed", "stick", key)
			delete(c.sessions, key)
			sess.cancel()
		}
	}
	// Мёртвые сессии (стик не открылся, ошибка каналов, Node.Run вернулся)
	// вычищаются — add-проход ниже пересоздаст сессию на этом же тике.
	for key, sess := range c.sessions {
		select {
		case <-sess.done:
			delete(c.sessions, key)
		default:
		}
	}
	for key, info := range want {
		if _, ok := c.sessions[key]; ok {
			continue
		}
		c.sessions[key] = c.startStickLocked(key, info)
	}
}

// startStickLocked запускает сессию стика. Вызывающий держит c.mu.
func (c *Collector) startStickLocked(key string, info openant.StickInfo) *stickSession {
	ctx, cancel := context.WithCancel(context.Background())
	sess := &stickSession{key: key, cancel: cancel, done: make(chan struct{})}
	slog.Info("stick added", "stick", key)
	go func() {
		defer close(sess.done)
		if err := c.stickSession(ctx, info); err != nil {
			// Восстановление внутри Node.Run (auto-reconnect); если сессия
			// всё же завершилась (ошибка открытия, каналов или Run вернулся) —
			// следующий тик reconcileSticks вычистит мёртвую запись
			// (sess.done закрыт) и пересоздаст сессию.
			slog.Error("ANT stick session failed", "stick", key, "err", err)
		}
	}()
	return sess
}

func (c *Collector) stopAllSessions() {
	c.mu.Lock()
	dones := make([]chan struct{}, 0, len(c.sessions))
	for key, sess := range c.sessions {
		sess.cancel()
		dones = append(dones, sess.done)
		delete(c.sessions, key)
	}
	c.mu.Unlock()
	for _, done := range dones {
		<-done // Stop() возвращает управление только после полной остановки
	}
}

// stickSession — один жизненный цикл Node: подключение, каналы, диспетчер.
// node.Run блокируется до отмены ctx; сбои стика восстанавливает
// auto-reconnect внутри Node (openant-go v0.1.2).
func (c *Collector) stickSession(ctx context.Context, info openant.StickInfo) error {
	node, err := c.stickOpener(info)
	if err != nil {
		return err
	}
	defer node.Stop()

	if err := node.SetNetworkKey(0x00, devices.ANTPLUS_NETWORK_KEY); err != nil {
		return err
	}

	if err := c.setupChannels(node); err != nil {
		return err
	}

	runDone := make(chan struct{})
	go func() {
		defer close(runDone)
		node.Run(ctx)
	}()

	select {
	case <-ctx.Done():
	case <-runDone:
	}
	// Дренируем диспетчер событий: node.Run может ещё доставлять
	// буферизованные данные в колбэки; Stop не должен возвращаться,
	// пока последний колбэк не отработал (контракт Run v0.1.2 —
	// гарантированный выход по ctx.Done, ожидание не зависает).
	<-runDone
	return nil
}

// setupChannels назначает maxSensors wildcard-каналов HR на Node.
func (c *Collector) setupChannels(node *easy.Node) error {
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

			slog.Info("sensor found", "device_id", deviceID)

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
	return nil
}
