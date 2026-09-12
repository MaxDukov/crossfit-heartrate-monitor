package ant

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/maxdukov/openant-go/anttest"
	"github.com/maxdukov/openant-go/easy"

	openant "github.com/maxdukov/openant-go/ant"

	"github.com/maxdukov/cf/backend-go/internal/collectors"
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
	// :memory:-SQLite в modernc — отдельная БД на соединение; с пулом > 1
	// конкурентные записи сессий попадают в разные пустые БД
	// ("no such table: sensors"). Ограничиваем пул одним соединением.
	d.SetMaxOpenConns(1)
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

	// b открывается асинхронно (первый reconcile в фоновой горутине) — ждём.
	deadline := time.Now().Add(5 * time.Second)
	for fs.opens_("b") < 1 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if fs.opens_("b") < 1 {
		t.Fatal("b не открылся при старте")
	}
	// Выдернули b и ждём ПОЛНОГО завершения старой сессии (sess.done
	// закрывается после node.Stop(), т.е. после закрытия её SimDriver):
	// повторное открытие до этого переиспользовало бы ещё живой инстанс
	// и гонялось бы с его закрытием.
	c.mu.Lock()
	sessB := c.sessions["usb2 serial=b bus=0 addr=0"]
	c.mu.Unlock()
	if sessB == nil {
		t.Fatal("сессия b не найдена до удаления стика")
	}
	fs.setList("a")
	select {
	case <-sessB.done:
	case <-time.After(5 * time.Second):
		t.Fatal("сессия b не завершилась после удаления стика")
	}
	// Вернули — должна открыться НОВАЯ сессия (второе открытие).
	fs.setList("a", "b")

	deadline = time.Now().Add(5 * time.Second)
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
