package system

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
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

type Resources struct {
	CPU       CPUStats    `json:"cpu"`
	Memory    MemoryStats `json:"memory"`
	Load      LoadStats   `json:"load"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

type Monitor struct{}

func NewMonitor() *Monitor { return &Monitor{} }

// Collect samples the host's CPU, memory, and load average.
// The CPU measurement blocks for 500 ms to produce an accurate interval sample.
func (m *Monitor) Collect() (*Resources, error) {
	// cpu.Percent with interval=500ms blocks but gives an accurate reading.
	percents, err := cpu.Percent(500*time.Millisecond, false)
	if err != nil {
		return nil, err
	}

	cores, _ := cpu.Counts(true) // logical cores; non-fatal if it fails

	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// load.Avg is Linux/macOS only; gracefully degrade on unsupported platforms.
	avgStat, err := load.Avg()
	if err != nil {
		avgStat = &load.AvgStat{}
	}

	cpuPct := 0.0
	if len(percents) > 0 {
		cpuPct = percents[0]
	}

	return &Resources{
		CPU: CPUStats{
			Percent: cpuPct,
			Cores:   cores,
		},
		Memory: MemoryStats{
			Total:   vmStat.Total,
			Used:    vmStat.Used,
			Free:    vmStat.Free,
			Percent: vmStat.UsedPercent,
		},
		Load: LoadStats{
			Load1:  avgStat.Load1,
			Load5:  avgStat.Load5,
			Load15: avgStat.Load15,
		},
		UpdatedAt: time.Now(),
	}, nil
}
