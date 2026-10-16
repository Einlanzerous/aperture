package checker

import (
	"context"
	"testing"
	"time"

	"github.com/aperture-dashboard/aperture/internal/config"
)

// fakeChecker returns healthy once release is closed (or immediately if nil).
type fakeChecker struct {
	release chan struct{}
}

func (f fakeChecker) Check(ctx context.Context) (Status, int, int64, string) {
	if f.release != nil {
		select {
		case <-f.release:
		case <-ctx.Done():
		}
	}
	return StatusHealthy, 200, 1, ""
}

// APTR-34: a slow in-flight check must not block GetStatuses readers.
func TestGetStatusesDoesNotBlockOnSlowCheck(t *testing.T) {
	release := make(chan struct{})
	w := &Worker{
		entries: []serviceEntry{
			{config: config.ServiceConfig{Name: "fast", Type: config.ServiceTypeHTTP}, checker: fakeChecker{}},
			{config: config.ServiceConfig{Name: "slow", Type: config.ServiceTypeHTTP}, checker: fakeChecker{release: release}},
		},
		statuses:     map[string]*ServiceStatus{},
		compactEvery: 1,
	}

	done := make(chan struct{})
	go func() { w.runChecks(); close(done) }()

	// Give the fast check time to land on the results channel, so the old
	// code would now be holding the write lock while waiting on "slow".
	time.Sleep(50 * time.Millisecond)

	got := make(chan int, 1)
	go func() { got <- len(w.GetStatuses()) }()
	select {
	case <-got:
	case <-time.After(100 * time.Millisecond):
		close(release)
		<-done
		t.Fatal("GetStatuses blocked while a slow check was in flight")
	}

	close(release)
	<-done
	if n := len(w.GetStatuses()); n != 2 {
		t.Fatalf("after cycle: want 2 statuses, got %d", n)
	}
}
