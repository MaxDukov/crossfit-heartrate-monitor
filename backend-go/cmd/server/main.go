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

	r := handlers.NewRouter(app, cfg.FrontendDir)
	if cfg.FrontendDir != "" {
		slog.Info("mounting static files", "dir", cfg.FrontendDir)
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
