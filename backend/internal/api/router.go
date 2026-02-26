package api

import (
	"net/http"

	"github.com/aperture-dashboard/aperture/internal/checker"
	"github.com/aperture-dashboard/aperture/internal/config"
	"github.com/aperture-dashboard/aperture/internal/system"
	"github.com/rs/cors"
)

// NewRouter wires up all API routes and wraps the mux with a permissive CORS handler
// (suitable for local / self-hosted use; tighten AllowedOrigins in production).
func NewRouter(worker *checker.Worker, sysMonitor *system.Monitor, cfg *config.Config) http.Handler {
	h := NewHandler(worker, sysMonitor, cfg)

	mux := http.NewServeMux()

	// Go 1.22+ pattern syntax: "METHOD /path"
	mux.HandleFunc("GET /api/health", h.Health)
	mux.HandleFunc("GET /api/config", h.GetConfig)
	mux.HandleFunc("GET /api/services", h.GetServices)
	mux.HandleFunc("GET /api/system/resources", h.GetSystemResources)
	mux.HandleFunc("GET /api/ollama/models", h.GetOllamaModels)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	return c.Handler(mux)
}
