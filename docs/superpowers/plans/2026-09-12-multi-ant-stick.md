# Multi-Stick ANT+ Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Поддержка нескольких ANT+ USB-стиков (авто-детект, до 8×N поясов) с независимым рестартом и горячим подключением.

**Architecture:** Подход A из спеки: коллектор получает StickManager — supervisor-горутину, которая тикает по списку стиков (`ant.ListSticks`) и управляет per-stick сессиями (`easy.NewForSerial`). Каждая сессия — свой Node с 8 wildcard-каналами и собственным рестарт-циклом. `knownIDs`/`dedups` остаются общими под mutex.

**Tech Stack:** Go 1.25, openant-go v0.2.0 (патчим), gousb v1.1.3 (`Context.OpenDevices`), anttest (mock-стики для тестов).

**Spec:** `docs/superpowers/specs/2026-09-12-multi-ant-stick-design.md`

## Global Constraints

- Ветка `multi-ant-stick` (уже создана, roadmap+spec закоммичены)
- openant-go — отдельный репозиторий: клон в `~/git/openant-go`, remote `git@github.com:MaxDukov/openant-go.git`, релиз тегом `v0.2.0`
- Публичное API `ant.New(d, maxSensors, cb)` в backend-go НЕ меняется (`app.go` не трогаем)
- Дедуп <2 с и upsert сенсора — без изменений; double-upsert одного device_id допустим
- Тесты: `go test -race ./...`; CI-джоба Go уже ставит libusb-1.0-0-dev
- Коммиты в стиле репо: `feat:`, `test:`, `docs:`; в openant-go — свои коммиты
- Пакет коллектора сам называется `ant` — импорт библиотечного пакета только с алиасом: `openant "github.com/maxdukov/openant-go/ant"`
- Существующее поведение одного стика должно сохраниться (регресс: старые тесты зелёные)
- Shell — zsh: URL с `?` в curl — в кавычках

---

### Task 1: openant-go — перечисление стиков (ListSticks, OpenStickBySerial)

**Files:**
- Create: `~/git/openant-go/ant/list.go`
- Modify: `~/git/openant-go/ant/driver_usb_gousb.go` (поле `serial`, ветка в `Open()`)
- Test: `~/git/openant-go/ant/list_test.go`, `~/git/openant-go/ant/integration_test.go` (дополнить)

**Interfaces:**
- Consumes: `gousb.Context.OpenDevices(opener func(*gousb.DeviceDesc) bool) ([]*gousb.Device, error)`, `(*gousb.Device).SerialNumber() (string, error)`, поле `(*gousb.Device).Descriptor *gousb.DeviceDesc` (Bus, Address, Vendor, Product)
- Produces: `type StickInfo struct{ Serial string }`, `func ListSticks() []StickInfo`, `func OpenStickBySerial(serial string) (Driver, error)` — используются Task 2 (easy) и Task 4 (коллектор)

- [ ] **Step 1: Клонировать репозиторий (если клона ещё нет)**

```bash
test -d ~/git/openant-go || git clone git@github.com:MaxDukov/openant-go.git ~/git/openant-go
git -C ~/git/openant-go checkout -b multi-stick
```

- [ ] **Step 2: Добавить поле serial в usbDriver и ветку выбора в Open()**

В `~/git/openant-go/ant/driver_usb_gousb.go`: в структуру `usbDriver` добавить поле
`serial string`; в начало `Open()` (после захвата mutex, до создания контекста) добавить
ветку: если `u.serial != ""` — перечислить устройства через `OpenDevices` с фильтром
VID `ANTVendorID` и PID ∈ {`ANTProductUSB2`, `ANTProductUSBm`}, у каждого прочитать
`SerialNumber()`, оставить только устройство с совпавшим серийником, остальные
`Close()`; не найдено — `u.ctx.Close()` и ошибка `fmt.Errorf("ant: стик с серийником
%q не найден", u.serial)`. Скелет ветки:

```go
if u.serial != "" {
    var target *gousb.Device
    devs, err := u.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
        return desc.Vendor == ANTVendorID &&
            (desc.Product == ANTProductUSB2 || desc.Product == ANTProductUSBm)
    })
    if err != nil {
        u.ctx.Close()
        return fmt.Errorf("ant: перечисление usb: %w", err)
    }
    for _, dev := range devs {
        if target == nil {
            if sn, serr := dev.SerialNumber(); serr == nil && strings.TrimSpace(sn) == u.serial {
                target = dev
                continue
            }
        }
        dev.Close()
    }
    if target == nil {
        u.ctx.Close()
        return fmt.Errorf("ant: стик с серийником %q не найден", u.serial)
    }
    dev = target
    // далее общий путь: claim интерфейса, эндпоинты (существующий код функции)
}
```

Импортировать `strings`. Существующий путь (`OpenDeviceWithVIDPID`) не трогать.

- [ ] **Step 3: Создать ant/list.go**

```go
package ant

import (
	"fmt"
	"strings"

	"github.com/google/gousb"
)

// StickInfo описывает подключённый ANT+ USB-стик.
type StickInfo struct {
	// Serial — серийный номер устройства; при пустом читаемом серийнике —
	// синтетический ключ "bus.addr".
	Serial string
}

var stickPIDs = []gousb.ID{ANTProductUSB2, ANTProductUSBm}

func isStickDesc(desc *gousb.DeviceDesc) bool {
	if desc.Vendor != ANTVendorID {
		return false
	}
	for _, pid := range stickPIDs {
		if desc.Product == pid {
			return true
		}
	}
	return false
}

// stickInfoFor строит StickInfo из дескриптора и серийника.
// Вынесено для unit-тестирования без USB.
func stickInfoFor(desc *gousb.DeviceDesc, serial string) StickInfo {
	if s := strings.TrimSpace(serial); s != "" {
		return StickInfo{Serial: s}
	}
	return StickInfo{Serial: fmt.Sprintf("%d.%d", desc.Bus, desc.Address)}
}

// ListSticks перечисляет все подключённые ANT+ USB-стики.
// Устройства закрываются после осмотра; владение не удерживается.
func ListSticks() []StickInfo { return listSticks(gousb.NewContext) }

func listSticks(newCtx func() *gousb.Context) []StickInfo {
	ctx := newCtx()
	defer ctx.Close()
	devs, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return isStickDesc(desc)
	})
	if err != nil && len(devs) == 0 {
		return nil
	}
	var sticks []StickInfo
	for _, dev := range devs {
		sn, _ := dev.SerialNumber()
		sticks = append(sticks, stickInfoFor(dev.Descriptor, sn))
		dev.Close()
	}
	return sticks
}

// OpenStickBySerial возвращает неоткрытый драйвер конкретного стика.
// Открытие выполняет реестр (easy.NewWithDriver → Driver.Open).
func OpenStickBySerial(serial string) (Driver, error) {
	s := strings.TrimSpace(serial)
	if s == "" {
		return nil, fmt.Errorf("ant: пустой серийный номер стика")
	}
	return &usbDriver{serial: s}, nil
}
```

- [ ] **Step 4: Написать unit-тест чистой логики (list_test.go)**

```go
package ant

import (
	"testing"

	"github.com/google/gousb"
)

func TestStickInfoForSerial(t *testing.T) {
	desc := &gousb.DeviceDesc{Bus: 1, Address: 7}
	got := stickInfoFor(desc, " 12345 \n")
	if got.Serial != "12345" {
		t.Fatalf("Serial = %q, want %q", got.Serial, "12345")
	}
}

func TestStickInfoForEmptySerialFallsBackToBusAddr(t *testing.T) {
	desc := &gousb.DeviceDesc{Bus: 1, Address: 7}
	got := stickInfoFor(desc, "")
	if got.Serial != "1.7" {
		t.Fatalf("Serial = %q, want %q", got.Serial, "1.7")
	}
}

func TestIsStickDescFiltersVIDPID(t *testing.T) {
	cases := []struct {
		desc gousb.DeviceDesc
		want bool
	}{
		{gousb.DeviceDesc{Vendor: ANTVendorID, Product: ANTProductUSB2}, true},
		{gousb.DeviceDesc{Vendor: ANTVendorID, Product: ANTProductUSBm}, true},
		{gousb.DeviceDesc{Vendor: ANTVendorID, Product: 0x1234}, false},
		{gousb.DeviceDesc{Vendor: 0x1234, Product: ANTProductUSB2}, false},
	}
	for i, c := range cases {
		if got := isStickDesc(&c.desc); got != c.want {
			t.Errorf("case %d: isStickDesc = %v, want %v", i, got, c.want)
		}
	}
}
```

- [ ] **Step 5: Запустить unit-тесты openant-go**

Run: `go test ./ant/ -run 'TestStickInfo|TestIsStickDesc' -v`
Expected: PASS (3 теста)

- [ ] **Step 6: Дополнить integration-тест (реальное железо, под тегом)**

В `~/git/openant-go/ant/integration_test.go` (файл уже имеет build-тег `integration`)
добавить:

```go
func TestListSticksIntegration(t *testing.T) {
	sticks := ListSticks()
	if len(sticks) == 0 {
		t.Skip("нет подключённых ANT+ стиков")
	}
	t.Logf("sticks: %+v", sticks)

	d, err := OpenStickBySerial(sticks[0].Serial)
	if err != nil {
		t.Fatalf("OpenStickBySerial: %v", err)
	}
	node, err := easy.NewWithDriver(d)
	if err != nil {
		t.Fatalf("node for serial %q: %v", sticks[0].Serial, err)
	}
	node.Stop()
}

func TestListSticksEmpty(t *testing.T) {
	// пустой серийник — валидационная ошибка без обращения к USB
	if _, err := OpenStickBySerial("  "); err == nil {
		t.Fatal("ожидалась ошибка на пустой серийник")
	}
}
```

Проверить сборку тегированных тестов (без железа они скипнутся/не запустятся):

Run: `go vet -tags integration ./ant/`
Expected: ошибок нет

- [ ] **Step 7: Коммит в openant-go**

```bash
git -C ~/git/openant-go add ant/list.go ant/list_test.go ant/driver_usb_gousb.go ant/integration_test.go
git -C ~/git/openant-go commit -m "ant: list all attached sticks and open by serial

ListSticks enumerates every 0fcf:1008/1009 device via gousb
OpenDevices; OpenStickBySerial returns a driver bound to a specific
stick (empty serial falls back to a bus.addr synthetic key)."
```

---

### Task 2: openant-go — easy.NewForSerial

**Files:**
- Modify: `~/git/openant-go/easy/node.go` (добавить функцию после `NewWithDriver`)
- Test: `~/git/openant-go/easy/node_test.go` (дополнить)

**Interfaces:**
- Consumes: `ant.OpenStickBySerial(serial string) (Driver, error)` (Task 1)
- Produces: `func NewForSerial(serial string) (*Node, error)` — используется Task 4 (коллектор) и Task 1 integration-тестом (косвенно)

- [ ] **Step 1: Добавить NewForSerial в easy/node.go**

Сразу после `NewWithDriver`:

```go
// NewForSerial finds the ANT stick with the given serial number, opens it
// and returns a started Node. Use ListSticks to discover serials of all
// attached sticks.
func NewForSerial(serial string) (*Node, error) {
	d, err := ant.OpenStickBySerial(serial)
	if err != nil {
		return nil, fmt.Errorf("easy: stick %q: %w", serial, err)
	}
	return NewWithDriver(d)
}
```

- [ ] **Step 2: Сборка и тесты пакета easy**

Run: `go build ./... && go test ./easy/ -run TestNode -v`
Expected: BUILD OK, существующие тесты PASS

- [ ] **Step 3: Коммит в openant-go**

```bash
git -C ~/git/openant-go add easy/node.go
git -C ~/git/openant-go commit -m "easy: NewForSerial builds a Node on a specific stick by serial"
```

---

### Task 3: Релиз openant-go v0.2.0 и обновление зависимости

**Files:**
- Modify: `backend-go/go.mod`, `backend-go/go.sum` (через `go get`)

**Interfaces:**
- Consumes: ветка `multi-stick` в `~/git/openant-go` (Tasks 1-2)
- Produces: `github.com/maxdukov/openant-go v0.2.0` в backend-go — используют Tasks 4-5

- [ ] **Step 1: Финальная проверка openant-go**

Run: `cd ~/git/openant-go && go vet ./... && go test ./...`
Expected: PASS (integration-тесты без тега не запускаются)

- [ ] **Step 2: Пуш и тег**

```bash
git -C ~/git/openant-go push origin multi-stick:master
git -C ~/git/openant-go tag v0.2.0
git -C ~/git/openant-go push origin v0.2.0
```

- [ ] **Step 3: Обновить зависимость в backend-go**

```bash
cd backend-go && go get github.com/maxdukov/openant-go@v0.2.0 && go mod tidy && go build ./...
```

Expected: BUILD OK

- [ ] **Step 4: Коммит в cf**

```bash
git add backend-go/go.mod backend-go/go.sum
git commit -m "build: bump openant-go to v0.2.0 (multi-stick support)"
```

---

### Task 4: Коллектор — StickManager (supervisor + per-stick сессии)

**Files:**
- Modify: `backend-go/internal/collectors/ant/ant.go` (переписать: полный код ниже)
- Test: `backend-go/internal/collectors/ant/ant_test.go` (только `newEnv` — адаптация под новый конструктор)

**Interfaces:**
- Consumes: `ant.StickInfo`, `ant.ListSticks()` (Task 1), `easy.NewForSerial(serial)` (Task 2)
- Produces: конструктор `newCollector(d *sql.DB, maxSensors int, cb collectors.Callbacks, stickLister func() []openant.StickInfo, stickOpener func(serial string) (*easy.Node, error), tick time.Duration) *Collector` — используют тесты Tasks 4-5; публичный `New(d, maxSensors, cb)` сохранён для `app.go`

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
		func() []openant.StickInfo { return []openant.StickInfo{{Serial: "sim"}} },
		func(serial string) (*easy.Node, error) {
			if serial != "sim" {
				return nil, fmt.Errorf("unknown stick %q", serial)
			}
			return easy.NewWithDriver(env.sim)
		},
		50*time.Millisecond)
	return env
}
```

и добавить импорт `openant "github.com/maxdukov/openant-go/ant"`.

Run: `go test ./internal/collectors/ant/ -run TestAntCollector -v`
Expected: FAIL (компиляция: newCollector требует 6 аргументов — конструктор ещё старый)

- [ ] **Step 2: Переписать ant.go — полный код файла**

```go
// Package ant — ANT+-коллектор на базе openant-go.
//
// Поддерживает несколько USB-стиков: supervisor опрашивает список стиков
// (ant.ListSticks) и управляет per-stick сессиями — по 8 wildcard-каналов
// на стик. Упавший стик рестартует независимо (backoff 5 с), подключение
// нового стика подхватывается без перезапуска коллектора.
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

	// stickLister возвращает список подключённых стиков (прод — openant.ListSticks,
	// тесты — фейк). stickOpener строит Node для стика по серийнику.
	stickLister func() []openant.StickInfo
	stickOpener func(serial string) (*easy.Node, error)
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
	key    string
	cancel context.CancelFunc
	done   chan struct{}
}

// New создаёт ANT+-коллектор с авто-детектом всех стиков
// (до maxSensors wildcard-каналов на каждом).
func New(d *sql.DB, maxSensors int, cb collectors.Callbacks) *Collector {
	return newCollector(d, maxSensors, cb,
		openant.ListSticks,
		func(serial string) (*easy.Node, error) { return easy.NewForSerial(serial) },
		10*time.Second)
}

func newCollector(d *sql.DB, maxSensors int, cb collectors.Callbacks,
	stickLister func() []openant.StickInfo,
	stickOpener func(serial string) (*easy.Node, error),
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
	want := make(map[string]bool, len(sticks))
	for _, s := range sticks {
		want[s.Serial] = true
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for key, sess := range c.sessions {
		if !want[key] {
			slog.Info("stick removed", "stick", key)
			delete(c.sessions, key)
			sess.cancel()
		}
	}
	for _, s := range sticks {
		if _, ok := c.sessions[s.Serial]; ok {
			continue
		}
		c.sessions[s.Serial] = c.startStickLocked(s.Serial)
	}
}

// startStickLocked запускает сессию стика. Вызывающий держит c.mu.
func (c *Collector) startStickLocked(key string) *stickSession {
	ctx, cancel := context.WithCancel(context.Background())
	sess := &stickSession{key: key, cancel: cancel, done: make(chan struct{})}
	slog.Info("stick added", "stick", key)
	go func() {
		defer close(sess.done)
		c.stickLoop(ctx, key)
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

// stickLoop — цикл одной сессии: connect → run → рестарт через 5 с при ошибке.
// Завершается по отмене ctx (стик удалён или коллектор остановлен).
func (c *Collector) stickLoop(ctx context.Context, key string) {
	for {
		if err := c.stickSession(ctx, key); err != nil {
			select {
			case <-ctx.Done():
				return
			default:
			}
			slog.Error("ANT stick session error", "stick", key, "err", err)
			slog.Info("restarting stick in 5s", "stick", key)
			select {
			case <-time.After(5 * time.Second):
			case <-ctx.Done():
				return
			}
		}
	}
}

// stickSession — один жизненный цикл Node: подключение, каналы, диспетчер.
func (c *Collector) stickSession(ctx context.Context, key string) error {
	node, err := c.stickOpener(key)
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

Supervisor polls ListSticks every 10s: new sticks start a session,
removed ones stop gracefully. Each stick runs its own restart loop
(5s backoff) with 8 wildcard HR channels. knownIDs/dedups stay
shared. Single-stick behavior unchanged."
```

---

### Task 5: Мультистик-тесты

**Files:**
- Test: `backend-go/internal/collectors/ant/ant_multistick_test.go` (новый)

**Interfaces:**
- Consumes: `newCollector(...)` сигнатура из Task 4; `anttest.NewSimDriver()`, `sim.EmitBroadcast(ch, data)`, `hrExtendedPage(deviceID, hr)` из ant_test.go (тот же пакет)
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

// fakeSticks — управляемый стенд: серийники → SimDriver, историю открытий
// и список стиков можно менять на ходу.
type fakeSticks struct {
	mu     sync.Mutex
	sims   map[string]*anttest.SimDriver
	list   []string
	opens  map[string]int
	failOn map[string]int // стик падает первые N открытий (после — успех)
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
		out = append(out, openant.StickInfo{Serial: s})
	}
	return out
}

func (fs *fakeSticks) opener(serial string) (*easy.Node, error) {
	fs.mu.Lock()
	fs.opens[serial]++
	open := fs.opens[serial]
	failUntil := fs.failOn[serial]
	fs.mu.Unlock()
	if open <= failUntil {
		return nil, fmt.Errorf("stick %q unavailable (attempt %d)", serial, open)
	}
	fs.mu.Lock()
	sim := fs.sims[serial]
	fs.mu.Unlock()
	if sim == nil {
		return nil, fmt.Errorf("unknown stick %q", serial)
	}
	return easy.NewWithDriver(sim)
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

	// bad постоянно ретраится (рестарт-цикл живёт).
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

func TestCollectorStopIdempotentWithSessions(t *testing.T) {
	fs := newFakeSticks("s1", "s2")
	c := newMultiEnv(t, fs)
	c.Start()
	// сессии живы — двойной Stop не должен паниковать/дедлокнуть/глобить
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

### Task 6: Документация и локальные проверки

**Files:**
- Modify: `README.md` (раздел «Железо», строка про лимит 8 каналов)
- Modify: `CHANGELOG.md` (секция v0.4.0)
- Modify: `docs/design/feature-roadmap.md` (статус 1.4 → Выполнено — ПОСЛЕ приёмки Task 7; сейчас только заготовка)
- Modify: `docs/superpowers/specs/2026-09-12-multi-ant-stick-design.md` (строка про fakelibusb)

**Interfaces:**
- Consumes: итоговое поведение Tasks 1-5
- Produces: документация; проверки по AGENTS.md

- [ ] **Step 1: Спека — уточнить тесты библиотеки**

В спеке заменить фразу «unit на fakelibusb (мок libusb в самом gousb)» на:
«unit на чистой логике (фильтр VID/PID, фолбэк серийника — fakelibusb недоступен
извне пакета gousb); интеграционные — на реальных стиках под build-тегом `integration`».

- [ ] **Step 2: README — раздел Железо**

Заменить блок про пояса:

```markdown
- Пояса: любые ANT+ HR-датчики (Garmin, Polar, …) — до 8×N, где N — число
  подключённых ANT+ стиков (8 wildcard-каналов на стик). Стики обнаруживаются
  автоматически; подключение/отключение стика на ходу подхватывается без
  перезапуска (рестарт упавшего — через 5 с, независимо от остальных)
```

- [ ] **Step 3: CHANGELOG — в секцию `[Неопубликовано] — v0.4.0` добавить строку**

```markdown
- Поддержка нескольких ANT+ стиков: авто-детект всех подключённых (до 8×N поясов),
  независимый рестарт стиков, горячее подключение; openant-go v0.2.0
  (`ListSticks` / `OpenStickBySerial` / `easy.NewForSerial`)
```

- [ ] **Step 4: Локальные проверки (vet, тесты, build, smoke по AGENTS.md)**

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

- [ ] **Step 5: Коммит**

```bash
git add README.md CHANGELOG.md docs/superpowers/specs/2026-09-12-multi-ant-stick-design.md
git commit -m "docs: multi-stick support in README and changelog; spec test note"
```

---

### Task 7: Деплой и приёмка на проде (2 стика)

**Files:**
- Modify: ничего в репо (деплой бинарника на 192.168.0.73)

**Interfaces:**
- Consumes: `/tmp/cf-smoke`-сборка Task 6; прод-хост `ma.dukov@192.168.0.73`, контейнер `cf-backend-1`
- Produces: работающий мультистик на проде

- [ ] **Step 1: Кросс-компиляция aarch64-linux-musl**

```bash
mkdir -p /tmp/opencode/libusb-arm64 && cd /tmp/opencode
# (если ещё не собран libusb для arm64 — как в текущем деплое)
cd backend-go && CGO_ENABLED=1 \
  CC=/opt/homebrew/opt/musl-cross/bin/aarch64-linux-musl-gcc \
  PKG_CONFIG_PATH=/tmp/opencode/libusb-arm64/lib/pkgconfig \
  GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" \
  -o /tmp/opencode/cf-server-linux-arm64 ./cmd/server
```

- [ ] **Step 2: Деплой на прод**

```bash
scp /tmp/opencode/cf-server-linux-arm64 ma.dukov@192.168.0.73:/tmp/
ssh ma.dukov@192.168.0.73 'docker cp /tmp/cf-server-linux-arm64 cf-backend-1:/app/cf-server-new && \
  docker exec cf-backend-1 sh -c "mv /app/cf-server-new /app/cf-server && chmod +x /app/cf-server" && \
  docker restart cf-backend-1 && sleep 6 && curl -s http://localhost:8000/api/health'
```

Expected: `{"status":"ok"}`

- [ ] **Step 3: Приёмочный сценарий (2 физических стика в USB)**

Логи (обе сессии):

```bash
ssh ma.dukov@192.168.0.73 'docker logs cf-backend-1 --since 5m 2>&1 | grep -E "stick added|ANT"'
```

Expected: две строки `stick added` с разными серийниками.

9 поясов (8 на стик 1 + 1 на стик 2): все карточки на мониторе `http://192.168.0.73:8000/`.

Выдернуть стик 2 → его датчики пропадают с монитора, стик 1 продолжает работать
(`docker logs` → `stick removed`). Воткнуть обратно → `stick added`, датчики возвращаются.

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
