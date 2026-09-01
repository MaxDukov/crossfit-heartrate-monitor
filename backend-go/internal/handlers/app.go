// Package handlers — HTTP-обработчики REST API.
package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/maxdukov/cf/backend-go/internal/collectors"
	"github.com/maxdukov/cf/backend-go/internal/collectors/ant"
	"github.com/maxdukov/cf/backend-go/internal/collectors/mock"
	"github.com/maxdukov/cf/backend-go/internal/config"
	"github.com/maxdukov/cf/backend-go/internal/services"
	"github.com/maxdukov/cf/backend-go/internal/ws"
)

// App — общее состояние приложения, доступное обработчикам.
type App struct {
	DB  *sql.DB
	Hub *ws.Hub
	HR  *services.HRProcessor
	Cfg config.Config

	// Состояние коллектора (mode switch защищён мьютексом —
	// исправление гонки из Python-версии).
	collectorCh chan collectors.Collector
	mode        string
}

// NewApp создаёт приложение.
func NewApp(cfg config.Config, d *sql.DB, hub *ws.Hub, hr *services.HRProcessor) *App {
	mode := "ant"
	if cfg.DevMode {
		mode = "mock"
	}
	return &App{
		DB:          d,
		Hub:         hub,
		HR:          hr,
		Cfg:         cfg,
		collectorCh: make(chan collectors.Collector, 1),
		mode:        mode,
	}
}

// StartCollector запускает коллектор согласно текущему режиму.
func (a *App) StartCollector() {
	c := a.newCollector(a.mode)
	a.setCollector(c)
	c.Start()
}

// SwitchCollector атомарно останавливает текущий коллектор и запускает новый.
func (a *App) SwitchCollector(mode string) {
	a.HR.ResetState()
	c := a.newCollector(mode)
	a.stopCollector()
	a.setCollector(c)
	a.mode = mode
	c.Start()
}

// StopCollector останавливает активный коллектор (graceful shutdown).
func (a *App) StopCollector() { a.stopCollector() }

// Mode возвращает текущий режим коллектора.
func (a *App) Mode() string { return a.mode }

func (a *App) newCollector(mode string) collectors.Collector {
	if mode == "ant" {
		return ant.New(a.DB, 8, collectors.Callbacks{
			OnHRData:    a.HR.OnHRData,
			OnNewSensor: a.HR.OnNewSensor,
		})
	}
	return mock.New(a.DB, collectors.Callbacks{
		OnHRData:    a.HR.OnHRData,
		OnNewSensor: a.HR.OnNewSensor,
	})
}

func (a *App) currentCollector() collectors.Collector {
	select {
	case c := <-a.collectorCh:
		select {
		case a.collectorCh <- c:
			return c
		default:
			return c
		}
	default:
		return nil
	}
}

func (a *App) setCollector(c collectors.Collector) {
	select {
	case <-a.collectorCh:
	default:
	}
	a.collectorCh <- c
}

func (a *App) stopCollector() {
	if c := a.currentCollector(); c != nil {
		c.Stop()
	}
}

// ── helpers ──────────────────────────────────────────────────

// writeJSON отправляет ответ в формате FastAPI-совместимого JSON.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// httpError имитирует формат ошибок FastAPI: {"detail": "..."}.
func httpError(w http.ResponseWriter, status int, detail string) {
	writeJSON(w, status, map[string]string{"detail": detail})
}

// decodeJSON строго декодирует тело запроса (лишние/невалидные поля — ошибка).
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		httpError(w, http.StatusUnprocessableEntity, err.Error())
		return false
	}
	return true
}
