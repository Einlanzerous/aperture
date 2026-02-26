package checker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/aperture-dashboard/aperture/internal/config"
)

// Worker runs health checks for all configured services on a fixed interval.
// Results are stored in a thread-safe map and served to the API layer.
type Worker struct {
	cfg      *config.Config
	statuses map[string]*ServiceStatus
	mu       sync.RWMutex
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

func NewWorker(cfg *config.Config) *Worker {
	return &Worker{
		cfg:      cfg,
		statuses: make(map[string]*ServiceStatus, len(cfg.Services)),
		stopCh:   make(chan struct{}),
	}
}

// Start runs an initial check immediately, then re-checks on every interval tick.
func (w *Worker) Start() {
	w.runChecks()

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(time.Duration(w.cfg.CheckInterval) * time.Second)
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
func (w *Worker) Stop() {
	close(w.stopCh)
	w.wg.Wait()
}

// GetStatuses returns a snapshot of all current service statuses.
func (w *Worker) GetStatuses() []*ServiceStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()

	out := make([]*ServiceStatus, 0, len(w.statuses))
	for _, s := range w.statuses {
		cp := *s
		out = append(out, &cp)
	}
	return out
}

// runChecks fans out one goroutine per service, collects results, and stores them.
func (w *Worker) runChecks() {
	results := make(chan *ServiceStatus, len(w.cfg.Services))
	var wg sync.WaitGroup

	for _, svc := range w.cfg.Services {
		wg.Add(1)
		go func(svc config.ServiceConfig) {
			defer wg.Done()
			results <- w.checkService(svc)
		}(svc)
	}

	// Close the channel once all goroutines finish so we can range over it.
	go func() {
		wg.Wait()
		close(results)
	}()

	w.mu.Lock()
	defer w.mu.Unlock()
	for status := range results {
		w.statuses[status.Name] = status
	}
}

func (w *Worker) checkService(svc config.ServiceConfig) *ServiceStatus {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	s := &ServiceStatus{
		Name:      svc.Name,
		Type:      string(svc.Type),
		URL:       svc.URL,
		Container: svc.Container,
		Icon:      svc.Icon,
		Category:  svc.Category,
		Size:      svc.Size,
		CheckedAt: time.Now(),
	}

	switch svc.Type {
	case config.ServiceTypeHTTP:
		s.Status, s.StatusCode, s.ResponseTime, s.Message = checkHTTP(ctx, svc.URL)

	case config.ServiceTypeDocker:
		s.Status, s.Message = checkDocker(ctx, svc.Container)

	default:
		s.Status = StatusUnknown
		s.Message = "unsupported service type"
	}

	log.Printf("[checker] %-24s %-8s → %s", svc.Name, svc.Type, s.Status)
	return s
}
