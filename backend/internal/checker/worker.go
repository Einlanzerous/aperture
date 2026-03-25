package checker

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aperture-dashboard/aperture/internal/config"
	"github.com/aperture-dashboard/aperture/internal/store"
)

// serviceEntry pairs a config-derived identity with its runtime checker.
type serviceEntry struct {
	config  config.ServiceConfig
	checker Checker
}

// Worker runs health checks for all configured services on a fixed interval.
// Results are stored in a thread-safe map and served to the API layer.
// If a Store is configured, each check result is persisted and compaction
// runs every compactEvery cycles.
type Worker struct {
	entries      []serviceEntry
	interval     time.Duration
	statuses     map[string]*ServiceStatus
	mu           sync.RWMutex
	stopCh       chan struct{}
	wg           sync.WaitGroup
	store        store.Store
	cycleCount   atomic.Int64
	compactEvery int64
	rawMaxAge    time.Duration
	summaryMaxAge time.Duration
}

func NewWorker(cfg *config.Config) *Worker {
	entries := make([]serviceEntry, 0, len(cfg.Services))
	for _, svc := range cfg.Services {
		var c Checker
		switch svc.Type {
		case config.ServiceTypeHTTP:
			c = NewHTTPChecker(svc.URL, svc.SkipVerify, svc.CheckConnectionOnly)
		case config.ServiceTypeDocker:
			c = NewDockerChecker(svc.Container, cfg.DockerSocket)
		default:
			continue
		}
		entries = append(entries, serviceEntry{config: svc, checker: c})
	}

	return &Worker{
		entries:       entries,
		interval:      time.Duration(cfg.CheckInterval) * time.Second,
		statuses:      make(map[string]*ServiceStatus, len(entries)),
		stopCh:        make(chan struct{}),
		compactEvery:  int64(config.DefaultCompactCycleCount),
		rawMaxAge:     cfg.Storage.Retention.Raw,
		summaryMaxAge: cfg.Storage.Retention.Summary,
	}
}

// SetStore attaches a persistent store to the worker.
// Must be called before Start.
func (w *Worker) SetStore(s store.Store) {
	w.store = s
}

// Start runs an initial check immediately, then re-checks on every interval tick.
func (w *Worker) Start() {
	w.runChecks()

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				w.runChecks()
			case <-w.stopCh:
				return
			}
		}
	}()
}

// Stop signals the background goroutine to exit and waits for it to finish.
// The provided context can enforce a deadline on the wait.
func (w *Worker) Stop(ctx context.Context) {
	close(w.stopCh)

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		slog.Warn("worker stop timed out", "error", ctx.Err())
	}
}

// GetStatuses returns a snapshot of all current service statuses, sorted by name.
func (w *Worker) GetStatuses() []*ServiceStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()

	out := make([]*ServiceStatus, 0, len(w.statuses))
	for _, s := range w.statuses {
		cp := *s
		out = append(out, &cp)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// runChecks fans out one goroutine per service, collects results, stores them,
// and periodically triggers compaction if a Store is configured.
func (w *Worker) runChecks() {
	results := make(chan *ServiceStatus, len(w.entries))
	var wg sync.WaitGroup

	for _, e := range w.entries {
		wg.Add(1)
		go func(e serviceEntry) {
			defer wg.Done()
			results <- w.checkService(e)
		}(e)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	collected := make([]*ServiceStatus, 0, len(w.entries))
	w.mu.Lock()
	for status := range results {
		w.statuses[status.Name] = status
		collected = append(collected, status)
	}
	w.mu.Unlock()

	// Persist check results if a store is configured.
	if w.store != nil {
		w.persistResults(collected)

		cycle := w.cycleCount.Add(1)
		if cycle%w.compactEvery == 0 {
			w.runCompaction()
		}
	}
}

func (w *Worker) persistResults(results []*ServiceStatus) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, s := range results {
		record := store.CheckRecord{
			ServiceName:  s.Name,
			ServiceType:  s.Type,
			Status:       string(s.Status),
			StatusCode:   s.StatusCode,
			ResponseTime: s.ResponseTime,
			Message:      s.Message,
			CheckedAt:    s.CheckedAt,
		}
		if err := w.store.RecordCheck(ctx, record); err != nil {
			slog.Error("failed to persist check",
				"component", "checker",
				"service", s.Name,
				"error", err,
			)
		}
	}
}

func (w *Worker) runCompaction() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := w.store.Compact(ctx, w.rawMaxAge); err != nil {
		slog.Error("compaction failed", "component", "checker", "error", err)
	} else {
		slog.Info("compaction complete", "component", "checker")
	}

	if err := w.store.Prune(ctx, w.summaryMaxAge); err != nil {
		slog.Error("prune failed", "component", "checker", "error", err)
	}
}

func (w *Worker) checkService(e serviceEntry) *ServiceStatus {
	ctx, cancel := context.WithTimeout(context.Background(), defaultCheckTimeout)
	defer cancel()

	s := newServiceStatus(e.config)
	s.Status, s.StatusCode, s.ResponseTime, s.Message = e.checker.Check(ctx)

	slog.Info("check complete",
		"component", "checker",
		"service", e.config.Name,
		"type", e.config.Type,
		"status", s.Status,
	)
	return s
}
