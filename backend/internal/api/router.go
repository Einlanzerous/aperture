package api

import (
	"net/http"

	"github.com/aperture-dashboard/aperture/internal/checker"
	"github.com/aperture-dashboard/aperture/internal/config"
	"github.com/aperture-dashboard/aperture/internal/system"
	"github.com/rs/cors"
)

// NewRouter wires up all API routes and wraps the mux with a CORS handler.
func NewRouter(worker *checker.Worker, sysMonitor *system.Monitor, cfg *config.Config) http.Handler {
	h := NewHandler(worker, sysMonitor, cfg)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("GET /api/config", h.GetConfig)
	mux.HandleFunc("GET /api/services", h.GetServices)
	mux.HandleFunc("GET /api/system/resources", h.GetSystemResources)
	mux.HandleFunc("GET /api/ollama/models", h.GetOllamaModels)

	origins := cfg.CORSOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	c := cors.New(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: []string{"GET", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	return c.Handler(mux)
}
