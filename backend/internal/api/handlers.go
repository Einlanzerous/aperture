package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/aperture-dashboard/aperture/internal/checker"
	"github.com/aperture-dashboard/aperture/internal/config"
	"github.com/aperture-dashboard/aperture/internal/semaphore"
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
	actions    *semaphore.Manager
	httpClient *http.Client
}

func NewHandler(worker *checker.Worker, sysMonitor *system.Monitor, cfg *config.Config, s store.Store, actions *semaphore.Manager) *Handler {
	return &Handler{
		worker:     worker,
		sysMonitor: sysMonitor,
		cfg:        cfg,
		store:      s,
		actions:    actions,
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

// requireStore writes a 503 and returns false when storage is not configured.
func (h *Handler) requireStore(w http.ResponseWriter) bool {
	if h.store == nil {
		writeError(w, http.StatusServiceUnavailable, "storage not configured")
		return false
	}
	return true
}

// requireActions writes a 503 and returns false when actions are not configured.
func (h *Handler) requireActions(w http.ResponseWriter) bool {
	if h.actions == nil {
		writeError(w, http.StatusServiceUnavailable, "actions not configured")
		return false
	}
	return true
}

// pathName extracts a named path parameter and writes a 400 if it is empty.
func pathName(w http.ResponseWriter, r *http.Request, param string) (string, bool) {
	name := r.PathValue(param)
	if name == "" {
		writeError(w, http.StatusBadRequest, param+" is required")
		return "", false
	}
	return name, true
}

// queryDuration parses a duration query parameter, returning fallback when absent.
func queryDuration(w http.ResponseWriter, r *http.Request, key string, fallback time.Duration) (time.Duration, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback, true
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid %s: %v", key, err))
		return 0, false
	}
	return d, true
}

// queryInt parses a positive integer query parameter, returning fallback when absent.
func queryInt(w http.ResponseWriter, r *http.Request, key string, fallback int) (int, bool) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback, true
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		writeError(w, http.StatusBadRequest, key+" must be a positive integer")
		return 0, false
	}
	return v, true
}

// Health is a simple liveness probe for the backend itself.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GetConfig returns dashboard-level configuration for the frontend.
func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"title":          h.cfg.Title,
		"checkInterval":  h.cfg.CheckInterval,
		"ollamaEnabled":  h.cfg.Ollama.URL != "",
		"systemEnabled":  h.cfg.System.Enabled,
		"actionsEnabled": h.actions != nil,
		"storageEnabled": h.store != nil,
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
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("invalid Ollama URL: %v", err))
		return
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		writeError(w, http.StatusBadGateway, fmt.Sprintf("cannot reach Ollama: %v", err))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxOllamaResponseBody))
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("read error: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

// GetServiceHistory returns raw recent check records for a single service.
// Query params: period (duration string, default "24h")
func (h *Handler) GetServiceHistory(w http.ResponseWriter, r *http.Request) {
	if !h.requireStore(w) {
		return
	}
	name, ok := pathName(w, r, "name")
	if !ok {
		return
	}
	period, ok := queryDuration(w, r, "period", 24*time.Hour)
	if !ok {
		return
	}

	records, err := h.store.GetRecentHistory(r.Context(), name, period)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to fetch history: %v", err))
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
	if !h.requireStore(w) {
		return
	}
	name, ok := pathName(w, r, "name")
	if !ok {
		return
	}
	days, ok := queryInt(w, r, "days", 30)
	if !ok {
		return
	}

	to := time.Now().UTC()
	from := to.AddDate(0, 0, -days)

	summaries, err := h.store.GetDailySummaries(r.Context(), name, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to fetch uptime: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"service":   name,
		"days":      days,
		"summaries": summaries,
	})
}

// GetActions returns the current state of all configured actions.
func (h *Handler) GetActions(w http.ResponseWriter, r *http.Request) {
	if !h.requireActions(w) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"actions": h.actions.ListActions(),
	})
}

// TriggerAction fires a Semaphore task for the named action.
func (h *Handler) TriggerAction(w http.ResponseWriter, r *http.Request) {
	if !h.requireActions(w) {
		return
	}
	name, ok := pathName(w, r, "name")
	if !ok {
		return
	}

	state, err := h.actions.Trigger(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusAccepted, state)
}

// GetActionStatus returns the current status of a specific action.
func (h *Handler) GetActionStatus(w http.ResponseWriter, r *http.Request) {
	if !h.requireActions(w) {
		return
	}
	name, ok := pathName(w, r, "name")
	if !ok {
		return
	}

	state, err := h.actions.GetStatus(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, state)
}
