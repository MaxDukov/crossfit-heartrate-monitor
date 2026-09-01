package ant

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/maxdukov/openant-go/anttest"
	"github.com/maxdukov/openant-go/easy"

	"github.com/maxdukov/cf/backend-go/internal/collectors"
	"github.com/maxdukov/cf/backend-go/internal/db"
)

// hrExtendedPage — extended broadcast-страница HR (13 байт):
// 8 байт страницы + флаг 0x80 + deviceID LE (2) + device type + trans type.
func hrExtendedPage(deviceID int, hr int) []byte {
	return []byte{
		0x00, // page 0 (common HR)
		0xFF, 0xFF, 0xFF,
		0x00, 0x00, // beat time
		0x01,                                // beat count
		byte(hr),                            // heart rate
		0x80,                                // extended flag
		byte(deviceID), byte(deviceID >> 8), // device id LE
		0x78, // device type 120 (HR)
		0x00, // trans type
	}
}

type testEnv struct {
	collector *Collector
	sim       *anttest.SimDriver
	d         *sql.DB

	mu        sync.Mutex
	hrEvents  []string
	newSensor []int
}

func (e *testEnv) waitFor(t *testing.T, cond func() bool, what string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		e.mu.Lock()
		ok := cond()
		e.mu.Unlock()
		if ok {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", what)
}

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
	env.collector = newCollector(d, 8, collectorsCallbacks(env), func() (*easy.Node, error) {
		return easy.NewWithDriver(env.sim)
	})
	return env
}

func collectorsCallbacks(env *testEnv) collectors.Callbacks {
	return collectors.Callbacks{
		OnHRData: func(deviceID, hr, battery int) {
			env.mu.Lock()
			env.hrEvents = append(env.hrEvents, fmt.Sprintf("dev=%d hr=%d bat=%d", deviceID, hr, battery))
			env.mu.Unlock()
		},
		OnNewSensor: func(deviceID int) {
			env.mu.Lock()
			env.newSensor = append(env.newSensor, deviceID)
			env.mu.Unlock()
		},
	}
}

func TestAntCollectorLifecycle(t *testing.T) {
	env := newEnv(t)
	env.collector.Start()
	defer env.collector.Stop()

	// Первый датчик: extended-страница с device_id=12345, HR=120.
	env.sim.EmitBroadcast(0, hrExtendedPage(12345, 120))

	env.waitFor(t, func() bool { return len(env.hrEvents) > 0 }, "первое HR-событие")
	env.mu.Lock()
	first := env.hrEvents[0]
	newSensors := append([]int(nil), env.newSensor...)
	env.mu.Unlock()

	if first != "dev=12345 hr=120 bat=255" {
		t.Errorf("first HR event = %q", first)
	}
	if len(newSensors) != 1 || newSensors[0] != 12345 {
		t.Errorf("OnNewSensor = %v, want [12345]", newSensors)
	}

	// Датчик появился в БД с последним пульсом.
	var lastHr sql.NullInt64
	var ignored bool
	err := env.d.QueryRow(`SELECT last_hr, COALESCE(ignored,0) FROM sensors WHERE device_id = 12345`).Scan(&lastHr, &ignored)
	if err != nil {
		t.Fatalf("sensor in db: %v", err)
	}
	if !lastHr.Valid || lastHr.Int64 != 120 {
		t.Errorf("last_hr = %v, want 120", lastHr)
	}
	if ignored {
		t.Error("sensor не должен быть ignored")
	}
}

func TestAntCollectorDedup(t *testing.T) {
	env := newEnv(t)
	env.collector.Start()
	defer env.collector.Stop()

	env.sim.EmitBroadcast(0, hrExtendedPage(4242, 100))
	env.waitFor(t, func() bool { return len(env.hrEvents) >= 1 }, "первое событие")

	// Тот же пульс в течение 2 с — дедуплицируется.
	env.sim.EmitBroadcast(0, hrExtendedPage(4242, 100))
	time.Sleep(300 * time.Millisecond)

	env.mu.Lock()
	count := len(env.hrEvents)
	env.mu.Unlock()
	if count != 1 {
		t.Errorf("после дедупа событий %d, want 1", count)
	}

	// Другой пульс — проходит.
	env.sim.EmitBroadcast(0, hrExtendedPage(4242, 105))
	env.waitFor(t, func() bool {
		return len(env.hrEvents) >= 2
	}, "второе событие после смены пульса")
}

func TestAntCollectorSecondSensorNewOnlyOnce(t *testing.T) {
	env := newEnv(t)
	env.collector.Start()
	defer env.collector.Stop()

	// Канал 0 уже занят первым датчиком — второй приходит на канал 1.
	env.sim.EmitBroadcast(0, hrExtendedPage(111, 90))
	env.sim.EmitBroadcast(1, hrExtendedPage(222, 95))

	env.waitFor(t, func() bool {
		return len(env.newSensor) >= 2
	}, "оба OnNewSensor")

	env.mu.Lock()
	had := env.newSensor
	env.mu.Unlock()
	if len(had) != 2 || had[0] != 111 || had[1] != 222 {
		t.Errorf("OnNewSensor = %v, want [111 222]", had)
	}

	// Повторные страницы тех же датчиков — новых событий быть не должно.
	env.sim.EmitBroadcast(0, hrExtendedPage(111, 92))
	env.waitFor(t, func() bool { return len(env.hrEvents) >= 3 }, "hr после повтора")
	env.mu.Lock()
	newSensors := len(env.newSensor)
	env.mu.Unlock()
	if newSensors != 2 {
		t.Errorf("OnNewSensor дублируется: %d событий, want 2", newSensors)
	}
}
