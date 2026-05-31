package system

import (
	"testing"
)

// pushPercent appends a synthetic sample carrying the given cpu percent so tests
// can exercise the ring buffer / history math without touching real hardware.
func (s *Sampler) pushPercent(p float64) {
	smp := sample{cpu: &CPUStats{Percent: p}}
	s.mu.Lock()
	s.ring[s.head] = smp
	s.head = (s.head + 1) % ringCapacity
	if s.size < ringCapacity {
		s.size++
	}
	s.mu.Unlock()
}

func TestSnapshot_EmptyReturnsNulls(t *testing.T) {
	s := NewSampler(Metrics{CPU: true})
	res := s.Snapshot(10)
	if res.CPU != nil || res.Memory != nil || res.Load != nil || res.GPU != nil {
		t.Fatalf("empty sampler should return null metrics, got %+v", res)
	}
	if res.History != nil {
		t.Fatalf("empty sampler should not include history, got %+v", res.History)
	}
}

func TestSnapshot_LatestAndHistoryOrder(t *testing.T) {
	s := NewSampler(Metrics{CPU: true})
	for i := 1; i <= 5; i++ {
		s.pushPercent(float64(i * 10)) // 10,20,30,40,50
	}

	res := s.Snapshot(3)
	if res.CPU == nil || res.CPU.Percent != 50 {
		t.Fatalf("latest snapshot should be newest sample (50), got %+v", res.CPU)
	}
	if res.History == nil {
		t.Fatalf("history requested but nil")
	}
	// Last 3 samples, oldest -> newest.
	want := []float64{30, 40, 50}
	if len(res.History.CPU) != len(want) {
		t.Fatalf("history len = %d, want %d (%v)", len(res.History.CPU), len(want), res.History.CPU)
	}
	for i := range want {
		if res.History.CPU[i] != want[i] {
			t.Fatalf("history[%d] = %v, want %v (full %v)", i, res.History.CPU[i], want[i], res.History.CPU)
		}
	}
}

func TestSnapshot_NoHistoryWhenZero(t *testing.T) {
	s := NewSampler(Metrics{CPU: true})
	s.pushPercent(42)
	res := s.Snapshot(0)
	if res.History != nil {
		t.Fatalf("history=0 should omit history block, got %+v", res.History)
	}
}

func TestSnapshot_HistoryClampedToSize(t *testing.T) {
	s := NewSampler(Metrics{CPU: true})
	s.pushPercent(1)
	s.pushPercent(2)
	res := s.Snapshot(100) // ask for more than we have
	if len(res.History.CPU) != 2 {
		t.Fatalf("history should clamp to available samples, got %v", res.History.CPU)
	}
}

func TestRingBuffer_WrapAround(t *testing.T) {
	s := NewSampler(Metrics{CPU: true})
	// Overfill so the buffer wraps: push ringCapacity+5 samples.
	total := ringCapacity + 5
	for i := 0; i < total; i++ {
		s.pushPercent(float64(i))
	}
	if s.size != ringCapacity {
		t.Fatalf("size should cap at %d, got %d", ringCapacity, s.size)
	}
	res := s.Snapshot(ringCapacity)
	// Newest value should be total-1; oldest retained should be total-ringCapacity.
	if res.CPU.Percent != float64(total-1) {
		t.Fatalf("latest after wrap = %v, want %v", res.CPU.Percent, total-1)
	}
	if got := res.History.CPU[0]; got != float64(total-ringCapacity) {
		t.Fatalf("oldest retained = %v, want %v", got, total-ringCapacity)
	}
	if got := res.History.CPU[len(res.History.CPU)-1]; got != float64(total-1) {
		t.Fatalf("newest in history = %v, want %v", got, total-1)
	}
}

func TestGPUUnavailableWhenNilAdapter(t *testing.T) {
	g := collectGPU(nil)
	if g.Available || g.Vendor != "" || g.TempC != nil {
		t.Fatalf("nil adapter should yield unavailable zero GPU stats, got %+v", g)
	}
}
