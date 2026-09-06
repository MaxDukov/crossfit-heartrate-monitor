// Package handlers — маршрутизация REST API.
package handlers

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// NewRouter собирает все маршруты приложения.
func NewRouter(a *App, frontendDir string) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	}))

	r.Get("/api/health", a.Health)
	r.Get("/api/debug/dbstats", a.DBStats)
	// pprof — диагностика зависаний (goroutine-дамп) через docker exec;
	// наружу порт не публикуется.
	r.Get("/debug/pprof/", pprof.Index)
	r.Get("/debug/pprof/cmdline", pprof.Cmdline)
	r.Get("/debug/pprof/profile", pprof.Profile)
	r.Get("/debug/pprof/symbol", pprof.Symbol)
	r.Get("/debug/pprof/trace", pprof.Trace)
	r.Get("/debug/pprof/goroutine", pprof.Index)
	r.Get("/debug/pprof/allocs", pprof.Index)
	r.Get("/debug/pprof/block", pprof.Index)
	r.Get("/debug/pprof/mutex", pprof.Index)
	r.Get("/debug/pprof/heap", pprof.Index)
	r.Get("/ws", a.Hub.HandleWS)

	r.Route("/api/athletes", func(r chi.Router) {
		r.Get("/", a.ListAthletes)
		r.Post("/", a.CreateAthlete)
		r.Put("/{athlete_id}", a.UpdateAthlete)
		r.Delete("/{athlete_id}", a.DeleteAthlete)
	})
	r.Route("/api/sensors", func(r chi.Router) {
		r.Get("/", a.ListSensors)
		r.Post("/{device_id}/assign", a.AssignSensor)
		r.Delete("/{device_id}/assign", a.UnassignSensor)
		r.Post("/{device_id}/ignore", a.IgnoreSensor)
		r.Post("/{device_id}/unignore", a.UnignoreSensor)
	})
	r.Route("/api/sessions", func(r chi.Router) {
		r.Get("/", a.ListSessions)
		r.Post("/", a.CreateSession)
		r.Get("/active", a.GetActiveSession)
		r.Post("/{session_id}/end", a.EndSession)
		r.Post("/{session_id}/athletes", a.AddAthleteToSession)
		r.Delete("/{session_id}/athletes/{athlete_id}", a.RemoveAthleteFromSession)
	})
	r.Route("/api/analytics", func(r chi.Router) {
		r.Get("/athletes/{athlete_id}/stats", a.AthleteStats)
		r.Get("/athletes/{athlete_id}/history", a.AthleteHistory)
	})
	r.Route("/api/equipment", func(r chi.Router) {
		r.Get("/", a.ListEquipment)
		r.Get("/inventory", a.ListInventory)
		r.Put("/inventory", a.UpdateInventory)
	})
	r.Route("/api/wods", func(r chi.Router) {
		r.Post("/generate", a.GenerateWods)
		r.Post("/select", a.SelectWod)
		r.Get("/active", a.GetActiveWod)
		r.Post("/active/end", a.EndActiveWod)
		r.Get("/history", a.ListWodHistory)
		r.Get("/templates", a.ListWodTemplates)
		r.Post("/custom", a.CreateCustomWod)
		r.Get("/templates/{template_id}", a.GetWodTemplate)
	})
	r.Route("/api/cycles", func(r chi.Router) {
		r.Get("/", a.ListCycles)
		r.Post("/", a.CreateCycle)
		r.Get("/{cycle_id}", a.GetCycle)
		r.Put("/{cycle_id}/status", a.UpdateCycleStatus)
		r.Delete("/{cycle_id}", a.DeleteCycle)
		r.Get("/{cycle_id}/analytics", a.GetCycleAnalytics)
		r.Put("/{cycle_id}/groups/{group_id}/third-day-off", a.SetGroupThirdDayOff)
		r.Post("/{cycle_id}/groups/{group_id}/third-day-off/replan", a.ReplanGroupThirdDayOff)
	})
	r.Route("/api/slots", func(r chi.Router) {
		r.Get("/{slot_id}", a.GetSlot)
		r.Get("/{slot_id}/recommendations", a.SlotRecommendations)
		r.Post("/{slot_id}/assign", a.AssignSlot)
		r.Delete("/{slot_id}/assign", a.UnassignSlot)
		r.Delete("/{slot_id}/assign/{wod_id}", a.UnassignSlotWod)
		r.Put("/{slot_id}/movements", a.UpdateSlotMovements)
		r.Post("/{slot_id}/start", a.StartSlot)
		r.Post("/{slot_id}/complete", a.CompleteSlot)
		r.Post("/{slot_id}/results", a.SaveSlotResult)
		r.Get("/{slot_id}/results", a.ListSlotResults)
	})
	r.Get("/api/movements", a.ListMovements)
	r.Post("/api/movements", a.CreateMovement)
	r.Put("/api/movements/{key}", a.UpdateMovement)
	r.Route("/api/system", func(r chi.Router) {
		r.Get("/mode", a.GetMode)
		r.Post("/mode", a.SetMode)
	})

	// SPA-статика (если задана).
	if frontendDir != "" {
		r.Get("/*", NewSPAHandler(frontendDir).ServeHTTP)
	} else {
		r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("CF-Monitor backend (Go)"))
		})
	}
	return r
}
