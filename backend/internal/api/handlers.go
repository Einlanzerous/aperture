package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
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
	sysSampler *system.Sampler
	cfg        *config.Config
	store      store.Store
	actions    *semaphore.Manager
	httpClient *http.Client
}

func NewHandler(worker *checker.Worker, sysSampler *system.Sampler, cfg *config.Config, s store.Store, actions *semaphore.Manager) *Handler {
	return &Handler{
		worker:     worker,
		sysSampler: sysSampler,
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
		"title":         h.cfg.Title,
		"checkInterval": h.cfg.CheckInterval,
		"ollamaEnabled": h.cfg.Ollama.URL != "",
		// systemEnabled stays for backward compat: true if any metric is on.
		"systemEnabled": h.cfg.System.CPU.Enabled || h.cfg.System.Memory.Enabled ||
			h.cfg.System.Load.Enabled || h.cfg.System.GPU.Enabled,
		// Per-metric flags let the frontend decide which widgets to render.
		"system": map[string]bool{
			"cpu":    h.cfg.System.CPU.Enabled,
			"memory": h.cfg.System.Memory.Enabled,
			"load":   h.cfg.System.Load.Enabled,
			"gpu":    h.cfg.System.GPU.Enabled,
		},
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

// GetSystemResources returns the latest sampled CPU / memory / load / GPU stats
// from the background sampler, with no blocking system calls in the request path.
// The optional ?history=<n> query includes the last n samples (oldest -> newest).
func (h *Handler) GetSystemResources(w http.ResponseWriter, r *http.Request) {
	// history defaults to 0 (omit the history block). Reject malformed values but
	// treat an absent param as "no history".
	history, ok := queryHistory(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, h.sysSampler.Snapshot(history))
}

// queryHistory parses the optional ?history=<n> param. Absent => 0 (no history).
// A non-negative integer is required; anything else is a 400.
func queryHistory(w http.ResponseWriter, r *http.Request) (int, bool) {
	raw := r.URL.Query().Get("history")
	if raw == "" {
		return 0, true
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		writeError(w, http.StatusBadRequest, "history must be a non-negative integer")
		return 0, false
	}
	return v, true
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

	// Drop hidden models before forwarding. On a successful upstream response we
	// rewrite the body; if filtering fails to parse (unexpected shape) we fall
	// back to the raw body so the widget still works.
	if resp.StatusCode == http.StatusOK && len(h.cfg.Ollama.HiddenModels) > 0 {
		if filtered, err := filterOllamaModels(body, h.cfg.Ollama.HiddenModels); err == nil {
			body = filtered
		} else {
			slog.Warn("ollama model filter skipped", "component", "api", "error", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)
}

// filterOllamaModels removes models whose name matches any hidden pattern from an
// Ollama /api/tags payload ({"models":[{"name":...}, …]}), preserving every other
// field of the surviving model objects. Returns an error if the body is not the
// expected shape.
func filterOllamaModels(body []byte, hidden []string) ([]byte, error) {
	var payload struct {
		Models []json.RawMessage `json:"models"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	kept := make([]json.RawMessage, 0, len(payload.Models))
	for _, raw := range payload.Models {
		var meta struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(raw, &meta); err != nil {
			return nil, err
		}
		if modelHidden(meta.Name, hidden) {
			continue
		}
		kept = append(kept, raw)
	}

	return json.Marshal(map[string]any{"models": kept})
}

// modelHidden reports whether name matches any of the hidden patterns.
func modelHidden(name string, patterns []string) bool {
	for _, p := range patterns {
		if matchGlob(p, name) {
			return true
		}
	}
	return false
}

// matchGlob reports whether name matches a simple glob where '*' matches any run
// of characters (including '/', unlike path.Match). A pattern with no '*' is an
// exact, case-sensitive match.
func matchGlob(pattern, name string) bool {
	if !strings.Contains(pattern, "*") {
		return pattern == name
	}
	parts := strings.Split(pattern, "*")
	// The leading and trailing segments anchor the ends; interior segments must
	// appear in order.
	if !strings.HasPrefix(name, parts[0]) {
		return false
	}
	name = name[len(parts[0]):]
	for _, seg := range parts[1 : len(parts)-1] {
		idx := strings.Index(name, seg)
		if idx < 0 {
			return false
		}
		name = name[idx+len(seg):]
	}
	return strings.HasSuffix(name, parts[len(parts)-1])
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
