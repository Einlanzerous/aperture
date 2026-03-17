package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/aperture-dashboard/aperture/internal/checker"
	"github.com/aperture-dashboard/aperture/internal/config"
	"github.com/aperture-dashboard/aperture/internal/store"
	"github.com/aperture-dashboard/aperture/internal/system"
)

const (
	ollamaTagsPath        = "/api/tags"
	maxOllamaResponseBody = 10 << 20 // 10 MB
	handlerTimeout        = 15 * time.Second
)

type Handler struct {
	worker     *checker.Worker
	sysMonitor *system.Monitor
	cfg        *config.Config
	store      store.Store
	httpClient *http.Client
}

func NewHandler(worker *checker.Worker, sysMonitor *system.Monitor, cfg *config.Config, s store.Store) *Handler {
	return &Handler{
		worker:     worker,
		sysMonitor: sysMonitor,
		cfg:        cfg,
		store:      s,
		httpClient: &http.Client{Timeout: handlerTimeout},
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode error", "component", "api", "error", err)
	}
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// Health is a simple liveness probe for the backend itself.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetConfig returns dashboard-level configuration for the frontend.
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"title":         h.cfg.Title,
		"checkInterval": h.cfg.CheckInterval,
		"ollamaEnabled": h.cfg.Ollama.URL != "",
		"systemEnabled": h.cfg.System.Enabled,
	})
}

// GetServices returns the latest health-check result for all configured services.
func (h *Handler) GetServices(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"services":  h.worker.GetStatuses(),
		"updatedAt": time.Now(),
	})
}

// GetSystemResources samples and returns the host's CPU / memory / load stats.
func (h *Handler) GetSystemResources(w http.ResponseWriter, r *http.Request) {
	res, err := h.sysMonitor.Collect()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// GetOllamaModels proxies the Ollama /api/tags endpoint so the frontend avoids CORS issues.
func (h *Handler) GetOllamaModels(w http.ResponseWriter, r *http.Request) {
	if h.cfg.Ollama.URL == "" {
		writeError(w, http.StatusServiceUnavailable, "Ollama URL not configured")
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, h.cfg.Ollama.URL+ollamaTagsPath, nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "invalid Ollama URL: "+err.Error())
		return
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		writeError(w, http.StatusBadGateway, "cannot reach Ollama: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxOllamaResponseBody))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "read error: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

// GetServiceHistory returns raw recent check records for a single service.
// Query params: period (duration string, default "24h")
func (h *Handler) GetServiceHistory(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "storage not configured")
		return
	}

	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "service name is required")
		return
	}

	period := 24 * time.Hour
	if p := r.URL.Query().Get("period"); p != "" {
		parsed, err := time.ParseDuration(p)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid period: "+err.Error())
			return
		}
		period = parsed
	}

	records, err := h.store.GetRecentHistory(r.Context(), name, period)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch history: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"service": name,
		"period":  period.String(),
		"records": records,
	})
}

// GetServiceUptime returns daily summaries for a single service.
// Query params: days (int, default 30)
func (h *Handler) GetServiceUptime(w http.ResponseWriter, r *http.Request) {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "storage not configured")
		return
	}

	name := r.PathValue("name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "service name is required")
		return
	}

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		parsed, err := strconv.Atoi(d)
		if err != nil || parsed <= 0 {
			writeError(w, http.StatusBadRequest, "days must be a positive integer")
			return
		}
		days = parsed
	}

	to := time.Now().UTC()
	from := to.AddDate(0, 0, -days)

	summaries, err := h.store.GetDailySummaries(r.Context(), name, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch uptime: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"service":   name,
		"days":      days,
		"summaries": summaries,
	})
}
