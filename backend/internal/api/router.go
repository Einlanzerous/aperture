package api

import (
	"net/http"

	"github.com/aperture-dashboard/aperture/internal/checker"
	"github.com/aperture-dashboard/aperture/internal/config"
	"github.com/aperture-dashboard/aperture/internal/semaphore"
	"github.com/aperture-dashboard/aperture/internal/store"
	"github.com/aperture-dashboard/aperture/internal/system"
	"github.com/rs/cors"
)

// NewRouter wires up all API routes and wraps the mux with a CORS handler.
func NewRouter(worker *checker.Worker, sysMonitor *system.Monitor, cfg *config.Config, s store.Store, actions *semaphore.Manager) http.Handler {
	h := NewHandler(worker, sysMonitor, cfg, s, actions)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("GET /api/config", h.GetConfig)
	mux.HandleFunc("GET /api/services", h.GetServices)
	mux.HandleFunc("GET /api/services/{name}/history", h.GetServiceHistory)
	mux.HandleFunc("GET /api/services/{name}/uptime", h.GetServiceUptime)
	mux.HandleFunc("GET /api/system/resources", h.GetSystemResources)
	mux.HandleFunc("GET /api/ollama/models", h.GetOllamaModels)
	mux.HandleFunc("GET /api/actions", h.GetActions)
	mux.HandleFunc("POST /api/actions/{name}/trigger", h.TriggerAction)
	mux.HandleFunc("GET /api/actions/{name}/status", h.GetActionStatus)

	origins := cfg.CORSOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	c := cors.New(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	return c.Handler(mux)
}
