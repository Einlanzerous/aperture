package system

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

const (
	// sampleInterval is how often the background sampler reads host metrics.
	sampleInterval = 5 * time.Second
	// ringCapacity is the number of samples retained (240 × 5s = 20 min).
	ringCapacity = 240
)

type CPUStats struct {
	Percent float64 `json:"percent"`
	Cores   int     `json:"cores"`
}

type MemoryStats struct {
	Total   uint64  `json:"total"`   // bytes
	Used    uint64  `json:"used"`    // bytes
	Free    uint64  `json:"free"`    // bytes
	Percent float64 `json:"percent"` // 0–100
}

type LoadStats struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// Metrics enables/disables each sampled metric. Disabled metrics are neither
// sampled nor included in the response (they serialize to null).
type Metrics struct {
	CPU    bool
	Memory bool
	Load   bool
	GPU    bool
}

// Resources is the response payload. Per-metric fields are pointers so a
// disabled (or unread) metric serializes to null per the API contract.
type Resources struct {
	CPU       *CPUStats    `json:"cpu"`
	Memory    *MemoryStats `json:"memory"`
	Load      *LoadStats   `json:"load"`
	GPU       *GPUStats    `json:"gpu"`
	History   *History     `json:"history,omitempty"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

// History holds per-metric percent series, oldest -> newest.
type History struct {
	CPU    []float64 `json:"cpu"`
	Memory []float64 `json:"memory"`
	Load1  []float64 `json:"load1"`
	GPU    []float64 `json:"gpu"`
}

// sample is one point-in-time reading retained in the ring buffer. Disabled
// metrics leave their pointer nil.
type sample struct {
	cpu  *CPUStats
	mem  *MemoryStats
	load *LoadStats
	gpu  *GPUStats
	at   time.Time
}

// Sampler periodically reads host metrics into a fixed-size ring buffer. It
// replaces the old per-request blocking collect: requests read the latest
// snapshot (and optional history window) without any blocking system calls.
type Sampler struct {
	metrics Metrics
	gpu     gpuAdapter

	mu   sync.RWMutex
	ring []sample // ring buffer of the last ringCapacity samples
	head int      // index of the next write
	size int      // number of valid entries (<= ringCapacity)

	stop chan struct{}
	done chan struct{}
}

// NewSampler builds a sampler honoring the supplied per-metric flags. The GPU
// adapter is probed once here (nil when no smi binary is on PATH or GPU is off).
func NewSampler(metrics Metrics) *Sampler {
	s := &Sampler{
		metrics: metrics,
		ring:    make([]sample, ringCapacity),
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
	if metrics.GPU {
		s.gpu = newGPUAdapter()
	}
	return s
}

// Start launches the sampling goroutine. It primes the non-blocking cpu.Percent
// counter immediately so subsequent ticks report meaningful deltas, then samples
// on a fixed interval until Stop is called.
func (s *Sampler) Start() {
	// Prime the gopsutil CPU delta counter. The first non-blocking call seeds
	// internal state and typically returns 0/garbage, so we discard it.
	_, _ = cpu.Percent(0, false)

	go s.run()
}

func (s *Sampler) run() {
	defer close(s.done)

	ticker := time.NewTicker(sampleInterval)
	defer ticker.Stop()

	// Take an initial sample right away so the first request has data without
	// waiting a full interval.
	s.collectInto()

	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.collectInto()
		}
	}
}

// Stop signals the goroutine to exit and waits for it to finish.
func (s *Sampler) Stop() {
	select {
	case <-s.stop:
		// already stopped
	default:
		close(s.stop)
	}
	<-s.done
}

// collectInto reads each enabled metric and appends a sample to the ring buffer.
// It never blocks on CPU: cpu.Percent(0, …) returns the delta since the previous
// call, which is exactly the sampler cadence.
func (s *Sampler) collectInto() {
	smp := sample{at: time.Now()}

	if s.metrics.CPU {
		cpuStats := &CPUStats{}
		if percents, err := cpu.Percent(0, false); err == nil && len(percents) > 0 {
			cpuStats.Percent = percents[0]
		}
		if cores, err := cpu.Counts(true); err == nil {
			cpuStats.Cores = cores
		}
		smp.cpu = cpuStats
	}

	if s.metrics.Memory {
		if vmStat, err := mem.VirtualMemory(); err == nil {
			smp.mem = &MemoryStats{
				Total:   vmStat.Total,
				Used:    vmStat.Used,
				Free:    vmStat.Free,
				Percent: vmStat.UsedPercent,
			}
		}
	}

	if s.metrics.Load {
		// load.Avg is Linux/macOS only; degrade gracefully elsewhere.
		avgStat, err := load.Avg()
		if err != nil {
			avgStat = &load.AvgStat{}
		}
		smp.load = &LoadStats{
			Load1:  avgStat.Load1,
			Load5:  avgStat.Load5,
			Load15: avgStat.Load15,
		}
	}

	if s.metrics.GPU {
		g := collectGPU(s.gpu)
		smp.gpu = &g
	}

	s.mu.Lock()
	s.ring[s.head] = smp
	s.head = (s.head + 1) % ringCapacity
	if s.size < ringCapacity {
		s.size++
	}
	s.mu.Unlock()
}

// Snapshot returns the most recent reading plus an optional history window of the
// last historyN samples (oldest -> newest). historyN <= 0 omits the history block.
func (s *Sampler) Snapshot(historyN int) *Resources {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := &Resources{UpdatedAt: time.Now()}

	if s.size == 0 {
		// No sample yet; return nulls with the current timestamp.
		return res
	}

	latest := s.at(s.size - 1)
	res.CPU = latest.cpu
	res.Memory = latest.mem
	res.Load = latest.load
	res.GPU = latest.gpu
	res.UpdatedAt = latest.at

	if historyN > 0 {
		res.History = s.historyLocked(historyN)
	}

	return res
}

// at returns the i-th oldest valid sample (0 == oldest). Caller holds the lock.
func (s *Sampler) at(i int) sample {
	start := (s.head - s.size + ringCapacity) % ringCapacity
	return s.ring[(start+i)%ringCapacity]
}

// historyLocked builds the per-metric percent series for the last n samples,
// oldest -> newest. Caller holds the read lock.
func (s *Sampler) historyLocked(n int) *History {
	if n > s.size {
		n = s.size
	}
	h := &History{
		CPU:    make([]float64, 0, n),
		Memory: make([]float64, 0, n),
		Load1:  make([]float64, 0, n),
		GPU:    make([]float64, 0, n),
	}
	first := s.size - n
	for i := first; i < s.size; i++ {
		smp := s.at(i)
		if smp.cpu != nil {
			h.CPU = append(h.CPU, smp.cpu.Percent)
		}
		if smp.mem != nil {
			h.Memory = append(h.Memory, smp.mem.Percent)
		}
		if smp.load != nil {
			h.Load1 = append(h.Load1, smp.load.Load1)
		}
		if smp.gpu != nil {
			h.GPU = append(h.GPU, smp.gpu.Percent)
		}
	}
	return h
}
