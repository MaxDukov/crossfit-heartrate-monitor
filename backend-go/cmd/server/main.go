// CF-Monitor Backend — Go-версия.
//
// Запускает коллектор ЧСС в фоновой горутине, обслуживает REST API
// и WebSocket для real-time трансляции пульса на фронтенд.
// Полностью совместим по API-контракту с Python (FastAPI) версией.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	appconfig "github.com/maxdukov/cf/backend-go/internal/config"
	"github.com/maxdukov/cf/backend-go/internal/db"
	"github.com/maxdukov/cf/backend-go/internal/handlers"
	"github.com/maxdukov/cf/backend-go/internal/services"
	"github.com/maxdukov/cf/backend-go/internal/ws"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	cfg := appconfig.Load()

	d, err := db.Open(cfg.DBPath)
	if err != nil {
		slog.Error("db open", "err", err)
		os.Exit(1)
	}
	defer func() { _ = d.Close() }()

	if err := db.Migrate(d); err != nil {
		slog.Error("migrate", "err", err)
		os.Exit(1)
	}
	if err := db.Seed(d); err != nil {
		slog.Error("seed", "err", err)
		os.Exit(1)
	}
	slog.Info("database initialized and seeded", "path", cfg.DBPath)

	hub := ws.NewHub()
	hr := services.NewHRProcessor(d, hub)
	app := handlers.NewApp(cfg, d, hub, hr)

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	}))

	r.Get("/api/health", app.Health)
	r.Get("/ws", hub.HandleWS)

	r.Route("/api/athletes", func(r chi.Router) {
		r.Get("/", app.ListAthletes)
		r.Post("/", app.CreateAthlete)
		r.Put("/{athlete_id}", app.UpdateAthlete)
		r.Delete("/{athlete_id}", app.DeleteAthlete)
	})
	r.Route("/api/sensors", func(r chi.Router) {
		r.Get("/", app.ListSensors)
		r.Post("/{device_id}/assign", app.AssignSensor)
		r.Delete("/{device_id}/assign", app.UnassignSensor)
		r.Post("/{device_id}/ignore", app.IgnoreSensor)
		r.Post("/{device_id}/unignore", app.UnignoreSensor)
	})
	r.Route("/api/sessions", func(r chi.Router) {
		r.Get("/", app.ListSessions)
		r.Post("/", app.CreateSession)
		r.Get("/active", app.GetActiveSession)
		r.Post("/{session_id}/end", app.EndSession)
		r.Post("/{session_id}/athletes", app.AddAthleteToSession)
		r.Delete("/{session_id}/athletes/{athlete_id}", app.RemoveAthleteFromSession)
	})
	r.Route("/api/analytics", func(r chi.Router) {
		r.Get("/athletes/{athlete_id}/stats", app.AthleteStats)
		r.Get("/athletes/{athlete_id}/history", app.AthleteHistory)
	})
	r.Route("/api/equipment", func(r chi.Router) {
		r.Get("/", app.ListEquipment)
		r.Get("/inventory", app.ListInventory)
		r.Put("/inventory", app.UpdateInventory)
	})
	r.Route("/api/wods", func(r chi.Router) {
		r.Post("/generate", app.GenerateWods)
		r.Post("/select", app.SelectWod)
		r.Get("/active", app.GetActiveWod)
		r.Post("/active/end", app.EndActiveWod)
		r.Get("/history", app.ListWodHistory)
	})
	r.Route("/api/system", func(r chi.Router) {
		r.Get("/mode", app.GetMode)
		r.Post("/mode", app.SetMode)
	})

	// SPA-статика (если задан CF_FRONTEND_DIR).
	if cfg.FrontendDir != "" {
		slog.Info("mounting static files", "dir", cfg.FrontendDir)
		r.Get("/*", handlers.NewSPAHandler(cfg.FrontendDir).ServeHTTP)
	} else {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("CF-Monitor backend (Go)"))
		})
	}

	app.StartCollector()

	srv := &http.Server{
		Addr:              ":" + itoa(cfg.Port),
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("listening", "port", cfg.Port, "mode", app.Mode())
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	app.StopCollector()
	slog.Info("shutdown complete")
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [20]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + v%10)
		v /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
