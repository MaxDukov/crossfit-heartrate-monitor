# Multi-Stick ANT+ Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Поддержка нескольких ANT+ USB-стиков (авто-детект, до 8×N поясов) с независимым восстановлением и горячим подключением.

**Architecture:** Подход A из спеки: коллектор получает StickManager — supervisor-горутину, которая тикает по списку стиков (`ant.Sticks`, v0.1.2) и управляет per-stick сессиями (`easy.NewStick`). Каждый сессия — свой Node с 8 wildcard-каналами; восстановление после сбоя стика делегировано auto-reconnect внутри `Node.Run(ctx)`. `knownIDs`/`dedups` остаются общими под mutex.

**Tech Stack:** Go 1.25, openant-go v0.1.2 (multi-dongle + auto-reconnect уже реализованы upstream, патчи не нужны), anttest (mock-стики для тестов).

**Spec:** `docs/superpowers/specs/2026-09-12-multi-ant-stick-design.md`

## Global Constraints

- Ветка `multi-ant-stick` (уже создана, roadmap+spec закоммичены)
- Библиотека `github.com/maxdukov/openant-go` v0.1.2 — multi-dongle (`ant.Sticks`, `ant.NewDriverForStick`, `easy.NewStick`) и automatic reconnect уже в ней; в библиотеку изменений НЕ вносить
- Публичное API `ant.New(d, maxSensors, cb)` в backend-go НЕ меняется (`app.go` не трогаем)
- Дедуп <2 с и upsert сенсора — без изменений; double-upsert одного device_id допустим
- Тесты: `go test -race ./...`; CI-джоба Go уже ставит libusb-1.0-0-dev
- Коммиты в стиле репо: `feat:`, `test:`, `docs:`
- Пакет коллектора сам называется `ant` — импорт библиотечного пакета только с алиасом: `openant "github.com/maxdukov/openant-go/ant"`
- Существующее поведение одного стика должно сохраниться (регресс: старые тесты зелёные)
- Shell — zsh: URL с `?` в curl — в кавычках
- Ключ сессии в коллекторе — `StickInfo.String()` (`"usb2 serial=X bus=N addr=N"`), стабилен между тиками

---

### Task 1: Обновление openant-go до v0.1.2

**Files:**
- Modify: `backend-go/go.mod`, `backend-go/go.sum`

**Interfaces:**
- Consumes: тег `v0.1.2` репозитория `github.com/MaxDukov/openant-go` (уже запушен)
- Produces: `ant.StickInfo{Serial string; Product string; Bus, Address int}`, `ant.Sticks() []StickInfo`, `easy.NewStick(info ant.StickInfo, opts ...NodeOption) (*Node, error)` — используют Tasks 2-3

- [ ] **Step 1: Поднять зависимость и проверить совместимость**

```bash
cd backend-go && go get github.com/maxdukov/openant-go@v0.1.2 && go mod tidy && go build ./...
```

Expected: BUILD OK. Если компиляция падает в существующем коде (`ant.go`, `ant_test.go`) —
ознакомиться с ошибками: API `devices.NewHeartRate`, `easy.NewWithDriver`,
`anttest.NewSimDriver().EmitBroadcast` в v0.1.2 сохранены (проверено по исходникам).

- [ ] **Step 2: Прогнать существующие тесты на v0.1.2**

Run: `cd backend-go && go test -race ./internal/collectors/ant/ -v`
Expected: все `TestAntCollector*` PASS на v0.1.2 (regression-база перед рефакторингом)

- [ ] **Step 3: Коммит**

```bash
git add backend-go/go.mod backend-go/go.sum
git commit -m "build: bump openant-go to v0.1.2 (multi-dongle, auto-reconnect)"
```

---

### Task 2: Коллектор — StickManager (supervisor + per-stick сессии)

**Files:**
- Modify: `backend-go/internal/collectors/ant/ant.go` (переписать: полный код ниже)
- Test: `backend-go/internal/collectors/ant/ant_test.go` (только `newEnv` — адаптация под новый конструктор)

**Interfaces:**
- Consumes: `ant.StickInfo`, `ant.Sticks()` (Task 1), `easy.NewStick(info)` (Task 1)
- Produces: конструктор `newCollector(d *sql.DB, maxSensors int, cb collectors.Callbacks, stickLister func() []openant.StickInfo, stickOpener func(openant.StickInfo) (*easy.Node, error), tick time.Duration) *Collector` — используют тесты Tasks 2-3; публичный `New(d, maxSensors, cb)` сохранён для `app.go`

- [ ] **Step 1: Обновить адаптер теста newEnv под новый конструктор (сначала — красный)**

В `ant_test.go` заменить тело `newEnv` (строки 58-76) на:

```go
func newEnv(t *testing.T) *testEnv {
	t.Helper()

	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	if err := db.Migrate(d); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	env := &testEnv{d: d}
	env.sim = anttest.NewSimDriver()
	env.collector = newCollector(d, 8, collectorsCallbacks(env),
		func() []openant.StickInfo { return []openant.StickInfo{{Serial: "sim", Product: "usb2"}} },
		func(info openant.StickInfo) (*easy.Node, error) {
			if info.Serial != "sim" {
				return nil, fmt.Errorf("unknown stick %q", info.Serial)
			}
			return easy.NewWithDriver(env.sim)
		},
		50*time.Millisecond)
	return env
}
```

и добавить импорт `openant "github.com/maxdukov/openant-go/ant"`.

Run: `go test ./internal/collectors/ant/ -run TestAntCollector -v`
Expected: FAIL (компиляция: у newCollector другая сигнатура)

- [ ] **Step 2: Переписать ant.go — полный код файла**

```go
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
			// всё же завершилась ошибкой — supervisor пересоздаст её на
			// следующем тике (сессии в c.sessions уже нет).
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
		return nil
	case <-runDone:
		return nil
	}
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
```

- [ ] **Step 3: Регресс существующих тестов (один стик)**

Run: `cd backend-go && go build ./... && go test -race ./internal/collectors/ant/ -v`
Expected: все `TestAntCollector*` PASS

- [ ] **Step 4: Коммит**

```bash
git add backend-go/internal/collectors/ant/ant.go backend-go/internal/collectors/ant/ant_test.go
git commit -m "feat: multi-stick ANT+ collector (supervisor + per-stick sessions)

Supervisor polls ant.Sticks every 10s: new sticks start a session
(easy.NewStick), removed ones stop gracefully. Stick-level failure
recovery is delegated to easy.Node auto-reconnect (v0.1.2); if a
session still ends, the supervisor recreates it on the next tick.
knownIDs/dedups stay shared. Single-stick behavior unchanged."
```

---

### Task 3: Мультистик-тесты

**Files:**
- Test: `backend-go/internal/collectors/ant/ant_multistick_test.go` (новый)

**Interfaces:**
- Consumes: `newCollector(...)` сигнатура из Task 2; `anttest.NewSimDriver()`, `sim.EmitBroadcast(ch, data)`, `hrExtendedPage(deviceID, hr)` из ant_test.go (тот же пакет)
- Produces: ничего (только тесты)

- [ ] **Step 1: Создать ant_multistick_test.go с фейком нескольких стиков**

```go
package ant

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/maxdukov/openant-go/anttest"
	"github.com/maxdukov/openant-go/easy"

	openant "github.com/maxdukov/openant-go/ant"

	"github.com/maxdukov/cf/backend-go/internal/db"
)

// fakeSticks — управляемый стенд: серийники → SimDriver, список стиков
// и сбои открытия можно менять на ходу.
type fakeSticks struct {
	mu     sync.Mutex
	sims   map[string]*anttest.SimDriver
	list   []string
	opens  map[string]int
	failOn map[string]int // стик недоступен первые N открытий (после — успех)
}

func newFakeSticks(serials ...string) *fakeSticks {
	fs := &fakeSticks{
		sims:   make(map[string]*anttest.SimDriver),
		opens:  make(map[string]int),
		failOn: make(map[string]int),
	}
	for _, s := range serials {
		fs.sims[s] = anttest.NewSimDriver()
		fs.list = append(fs.list, s)
	}
	return fs
}

func (fs *fakeSticks) lister() []openant.StickInfo {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	out := make([]openant.StickInfo, 0, len(fs.list))
	for _, s := range fs.list {
		out = append(out, openant.StickInfo{Serial: s, Product: "usb2"})
	}
	return out
}

func (fs *fakeSticks) opener(info openant.StickInfo) (*easy.Node, error) {
	fs.mu.Lock()
	fs.opens[info.Serial]++
	open := fs.opens[info.Serial]
	failUntil := fs.failOn[info.Serial]
	fs.mu.Unlock()
	if open <= failUntil {
		return nil, fmt.Errorf("stick %q unavailable (attempt %d)", info.Serial, open)
	}
	fs.mu.Lock()
	sim := fs.sims[info.Serial]
	fs.mu.Unlock()
	if sim == nil {
		return nil, fmt.Errorf("unknown stick %q", info.Serial)
	}
	node, err := easy.NewWithDriver(sim)
	if err != nil {
		// anttest-драйвер после Node.Stop() закрыт необратно
		// ("mock: driver permanently closed") — пересоздаём.
		fs.mu.Lock()
		sim = anttest.NewSimDriver()
		fs.sims[info.Serial] = sim
		fs.mu.Unlock()
		node, err = easy.NewWithDriver(sim)
		if err != nil {
			return nil, err
		}
	}
	return node, nil
}

// setList заменяет список стиков (hot plug / удаление).
func (fs *fakeSticks) setList(serials ...string) {
	fs.mu.Lock()
	fs.list = append([]string(nil), serials...)
	fs.mu.Unlock()
}

func (fs *fakeSticks) sim(serial string) *anttest.SimDriver { return fs.sims[serial] }

func (fs *fakeSticks) opens_(serial string) int {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.opens[serial]
}

func newMultiEnv(t *testing.T, fs *fakeSticks) *Collector {
	t.Helper()
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	if err := db.Migrate(d); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return newCollector(d, 8, collectors.Callbacks{}, fs.lister, fs.opener, 50*time.Millisecond)
}

func TestMultiStickTwoSticksHR(t *testing.T) {
	fs := newFakeSticks("stick-a", "stick-b")
	c := newMultiEnv(t, fs)
	c.Start()
	defer c.Stop()

	fs.sim("stick-a").EmitBroadcast(0, hrExtendedPage(101, 120))
	fs.sim("stick-b").EmitBroadcast(0, hrExtendedPage(202, 150))

	// Оба датчика должны появиться в БД (HR-обработка с обоих стиков).
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var n int
		_ = c.db.QueryRow(`SELECT COUNT(*) FROM sensors WHERE device_id IN (101, 202)`).Scan(&n)
		if n == 2 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("не оба датчика дошли до БД с двух стиков")
}

func TestMultiStickIndependentRestart(t *testing.T) {
	fs := newFakeSticks("good", "bad")
	fs.mu.Lock()
	fs.failOn["bad"] = 1_000_000 // "bad" недоступен всегда
	fs.mu.Unlock()
	c := newMultiEnv(t, fs)
	c.Start()
	defer c.Stop()

	// good шлёт HR — должно работать несмотря на постоянные ошибки bad.
	fs.sim("good").EmitBroadcast(0, hrExtendedPage(333, 100))

	// bad постоянно ретраится (supervisor пересоздаёт сессию на тиках).
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if fs.opens_("bad") >= 2 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if fs.opens_("bad") < 2 {
		t.Fatal("bad не ретраится")
	}
	if fs.opens_("good") < 1 {
		t.Fatal("good вообще не открывался")
	}
}

func TestMultiStickHotPlug(t *testing.T) {
	fs := newFakeSticks("a")
	c := newMultiEnv(t, fs)
	c.Start()
	defer c.Stop()

	// Воткнули второй стик — сессия должна стартовать без Stop().
	fs.mu.Lock()
	fs.sims["b"] = anttest.NewSimDriver()
	fs.mu.Unlock()
	fs.setList("a", "b")

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if fs.opens_("b") >= 1 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("новый стик b не был подхвачен supervisor'ом")
}

func TestMultiStickRemovedThenReadded(t *testing.T) {
	fs := newFakeSticks("a", "b")
	c := newMultiEnv(t, fs)
	c.Start()
	defer c.Stop()

	if fs.opens_("b") < 1 {
		t.Fatal("b не открылся при старте")
	}
	// Выдернули b.
	fs.setList("a")
	// Вернули — должна открыться НОВАЯ сессия (второе открытие).
	fs.setList("a", "b")

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if fs.opens_("b") >= 2 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("после переподключения стика новая сессия не открылась")
}

func TestMultiStickRecoveryAfterFailure(t *testing.T) {
	fs := newFakeSticks("flaky")
	fs.mu.Lock()
	fs.failOn["flaky"] = 2 // первые 2 открытия — ошибка, потом успех
	fs.mu.Unlock()
	c := newMultiEnv(t, fs)
	c.Start()
	defer c.Stop()

	fs.sim("flaky").EmitBroadcast(0, hrExtendedPage(404, 130))

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var n int
		_ = c.db.QueryRow(`SELECT COUNT(*) FROM sensors WHERE device_id = 404`).Scan(&n)
		if n == 1 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("стик не восстановился после временных ошибок открытия")
}

func TestCollectorStopIdempotentWithSessions(t *testing.T) {
	fs := newFakeSticks("s1", "s2")
	c := newMultiEnv(t, fs)
	c.Start()
	// сессии живы — двойной Stop не должен паниковать/дедлокнуть
	c.Stop()
	done := make(chan struct{})
	go func() {
		c.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("повторный Stop() не вернулся за 5 с")
	}
}
```

- [ ] **Step 2: Запустить мультистик-тесты**

Run: `cd backend-go && go test -race ./internal/collectors/ant/ -run 'TestMultiStick|TestCollectorStop' -v`
Expected: PASS (6 тестов). Если тесты фланкируют на таймингах — увеличить deadline с 5 с до 10 с, не менять логику.

- [ ] **Step 3: Полный прогон и vet**

Run: `cd backend-go && go vet ./... && go test -race ./...`
Expected: всё PASS

- [ ] **Step 4: Коммит**

```bash
git add backend-go/internal/collectors/ant/ant_multistick_test.go
git commit -m "test: multi-stick collector — two sticks, independent restart, hot plug, recovery"
```

---

### Task 4: Документация и локальные проверки

**Files:**
- Modify: `README.md` (раздел «Железо», строка про лимит 8 каналов)
- Modify: `CHANGELOG.md` (секция v0.4.0)
- Modify: `docs/design/feature-roadmap.md` (статус 1.4 → Выполнено — ПОСЛЕ приёмки Task 5)

**Interfaces:**
- Consumes: итоговое поведение Tasks 1-3
- Produces: документация; проверки по AGENTS.md

- [ ] **Step 1: README — раздел Железо**

Заменить блок про пояса:

```markdown
- Пояса: любые ANT+ HR-датчики (Garmin, Polar, …) — до 8×N, где N — число
  подключённых ANT+ стиков (8 wildcard-каналов на стик). Стики обнаруживаются
  автоматически; восстановление после сбоя стика — auto-reconnect внутри Node,
  подключение/отключение стика на ходу подхватывается без перезапуска
```

- [ ] **Step 2: CHANGELOG — в секцию `[Неопубликовано] — v0.4.0` добавить строку**

```markdown
- Поддержка нескольких ANT+ стиков: авто-детект всех подключённых (до 8×N поясов),
  горячее подключение, восстановление стика через auto-reconnect (openant-go v0.1.2:
  `ant.Sticks`, `easy.NewStick`)
```

- [ ] **Step 3: Локальные проверки (vet, тесты, build, smoke по AGENTS.md)**

```bash
cd backend-go && go vet ./... && go test -race ./... && go build -o /tmp/cf-smoke ./cmd/server
```

Smoke (фронт+бэк+БД одной командой):

```bash
rm -f /tmp/smoke-multi.db && CF_DB_PATH=/tmp/smoke-multi.db CF_DEV_MODE=1 CF_PORT=8023 \
  CF_FRONTEND_DIR=/Users/maxdukov/git/cf/frontend/dist /tmp/cf-smoke > /tmp/smoke-multi.log 2>&1 & sleep 3
curl -s http://localhost:8023/api/health
curl -s http://localhost:8023/api/debug/dbstats
curl -s -o /dev/null -w "%{http_code}" http://localhost:8023/
```

Expected: `{"status":"ok"}`, `in_use:0`, `200`

- [ ] **Step 4: Коммит**

```bash
git add README.md CHANGELOG.md
git commit -m "docs: multi-stick support in README and changelog"
```

---

### Task 5: Деплой и приёмка на проде (2 стика)

**Files:**
- Modify: ничего в репо (деплой бинарника на 192.168.0.73)

**Interfaces:**
- Consumes: `/tmp/cf-smoke`-сборка Task 4; прод-хост `ma.dukov@192.168.0.73`, контейнер `cf-backend-1`
- Produces: работающий мультистик на проде

- [ ] **Step 1: Кросс-компиляция aarch64-linux-musl**

```bash
cd backend-go && CGO_ENABLED=1 \
  CC=/opt/homebrew/opt/musl-cross/bin/aarch64-linux-musl-gcc \
  PKG_CONFIG_PATH=/tmp/opencode/libusb-arm64/lib/pkgconfig \
  GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" \
  -o /tmp/opencode/cf-server-linux-arm64 ./cmd/server
```

- [ ] **Step 2: Деплой на прод**

```bash
scp /tmp/opencode/cf-server-linux-arm64 ma.dukov@192.168.0.73:/tmp/
ssh ma.dukov@192.168.0.73 'docker cp /tmp/cf-server-linux-arm64 cf-backend-1:/usr/local/bin/cf-server-new && \
  docker exec cf-backend-1 sh -c "mv /usr/local/bin/cf-server-new /usr/local/bin/cf-server && chmod +x /usr/local/bin/cf-server" && \
  docker restart cf-backend-1 && sleep 6 && curl -s http://localhost:8000/api/health'
```

Expected: `{"status":"ok"}`

- [ ] **Step 3: Приёмочный сценарий (2 физических стика в USB)**

Логи (обе сессии):

```bash
ssh ma.dukov@192.168.0.73 'docker logs cf-backend-1 --since 5m 2>&1 | grep -E "stick added|ANT"'
```

Expected: две строки `stick added` с разными стиками.

9 поясов (8 на стик 1 + 1 на стик 2): все карточки на мониторе `http://192.168.0.73:8000/`.

Выдернуть стик 2 → его датчики пропадают с монитора, стик 1 продолжает работать
(внутри Node идёт reconnect; при исчезновении — `stick removed` в логах).
Воткнуть обратно → сессия подхватывается (`stick added`), датчики возвращаются.

- [ ] **Step 4: Финальные проверки и статус roadmap**

```bash
curl -s http://192.168.0.73:8000/api/debug/dbstats   # in_use:0 стабильно
```

Обновить в `docs/design/feature-roadmap.md`: статус 1.4 → `✅ Выполнено (multi-ant-stick, <дата>)`
и в сводной таблице. Коммит:

```bash
git add docs/design/feature-roadmap.md
git commit -m "roadmap: 1.4 multi-stick ANT+ — done (verified on production with 2 sticks)"
```

- [ ] **Step 5: Пуш ветки (PR/merge — по решению пользователя)**

```bash
git push -u origin multi-ant-stick
```
