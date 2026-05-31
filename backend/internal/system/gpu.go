package system

import (
	"context"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// GPUStats is the per-sample GPU snapshot. It is always populated (with zero /
// null values when no GPU is available) so the JSON shape stays stable.
type GPUStats struct {
	Available bool     `json:"available"` // false => no adapter, or probe failed
	Vendor    string   `json:"vendor"`    // "amd" | "nvidia" | ""
	Name      string   `json:"name"`      // best-effort product name, "" if unknown
	Percent   float64  `json:"percent"`   // utilization %
	VRAMUsed  uint64   `json:"vramUsed"`  // bytes
	VRAMTotal uint64   `json:"vramTotal"` // bytes
	TempC     *float64 `json:"tempC"`     // null when temperature unreadable
}

// gpuProbeTimeout bounds each shell-out so a wedged smi binary cannot stall the
// sampler loop.
const gpuProbeTimeout = 4 * time.Second

// gpuAdapter is a vendor-specific GPU reader. Probe selects the active adapter
// once at startup; Sample is then called on the sampler cadence.
type gpuAdapter interface {
	vendor() string
	// sample returns a populated GPUStats; Available is set by the caller.
	sample(ctx context.Context) (GPUStats, error)
}

// newGPUAdapter probes PATH for a supported SMI binary. For AMD it prefers the
// modern amd-smi (the successor to rocm-smi, which reports temperature reliably)
// and falls back to rocm-smi; NVIDIA's nvidia-smi is last. Returns nil when none
// is present.
func newGPUAdapter() gpuAdapter {
	if _, err := exec.LookPath("amd-smi"); err == nil {
		return newAMDSMIAdapter()
	}
	if _, err := exec.LookPath("rocm-smi"); err == nil {
		return &rocmSMIAdapter{}
	}
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		return &nvidiaAdapter{}
	}
	return nil
}

// collectGPU runs the active adapter and returns a stats snapshot. When no
// adapter is configured (no binary on PATH) it returns an unavailable snapshot.
// Any adapter error also degrades to unavailable rather than failing the caller.
func collectGPU(adapter gpuAdapter) GPUStats {
	if adapter == nil {
		return GPUStats{Available: false}
	}

	ctx, cancel := context.WithTimeout(context.Background(), gpuProbeTimeout)
	defer cancel()

	stats, err := adapter.sample(ctx)
	if err != nil {
		return GPUStats{Available: false, Vendor: adapter.vendor()}
	}
	stats.Available = true
	stats.Vendor = adapter.vendor()
	return stats
}

// ─── AMD (amd-smi, preferred) ─────────────────────────────────────────────────
//
// amd-smi is AMD's current unified SMI tool. Unlike rocm-smi 3.x — whose
// --showtemp returns empty on this card — it reports temperature reliably, so it
// is probed first.

const bytesPerGiB = 1024 * 1024 * 1024

type amdSMIAdapter struct {
	name string // product name, read once at construction ("" on failure)
}

func newAMDSMIAdapter() *amdSMIAdapter {
	return &amdSMIAdapter{name: amdSMIProductName()}
}

func (a *amdSMIAdapter) vendor() string { return "amd" }

// sample reads utilization, temperature and VRAM from a single
// `amd-smi monitor --json` invocation — a flat, dashboard-oriented payload.
func (a *amdSMIAdapter) sample(ctx context.Context) (GPUStats, error) {
	out, err := exec.CommandContext(ctx, "amd-smi", "monitor", "--json").Output()
	if err != nil {
		return GPUStats{}, err
	}

	var entries []map[string]json.RawMessage
	if err := json.Unmarshal(out, &entries); err != nil {
		return GPUStats{}, err
	}
	if len(entries) == 0 {
		return GPUStats{}, nil
	}
	m := entries[0]

	stats := GPUStats{Name: a.name}
	if v, ok := amdSMIValueField(m, "gfx"); ok {
		stats.Percent = v
	}
	// Prefer the hotspot sensor, fall back to the memory sensor.
	if v, ok := amdSMIValueField(m, "hotspot_temperature"); ok {
		stats.TempC = &v
	} else if v, ok := amdSMIValueField(m, "memory_temperature"); ok {
		stats.TempC = &v
	}
	// vram_* are reported in GiB; convert to bytes.
	if v, ok := amdSMIValueField(m, "vram_used"); ok {
		stats.VRAMUsed = uint64(v * bytesPerGiB)
	}
	if v, ok := amdSMIValueField(m, "vram_total"); ok {
		stats.VRAMTotal = uint64(v * bytesPerGiB)
	}
	return stats, nil
}

// amdSMIProductName best-effort reads the ASIC market name once. Empty on any
// failure — the name is cosmetic and must never block sampling.
func amdSMIProductName() string {
	ctx, cancel := context.WithTimeout(context.Background(), gpuProbeTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, "amd-smi", "static", "--asic", "--json").Output()
	if err != nil {
		return ""
	}
	var entries []struct {
		ASIC struct {
			MarketName string `json:"market_name"`
		} `json:"asic"`
	}
	if err := json.Unmarshal(out, &entries); err != nil || len(entries) == 0 {
		return ""
	}
	name := entries[0].ASIC.MarketName
	// In a minimal container amd-smi can't always resolve the marketing name and
	// falls back to the raw "0x1002" vendor id. Treat that as no name so the UI
	// shows the clean vendor label instead of a hex string.
	if strings.HasPrefix(name, "0x") {
		return ""
	}
	return name
}

// amdSMIValueField extracts a {"value": <number>, "unit": …} field. Returns
// ok=false when the key is absent or carries a non-numeric placeholder (amd-smi
// emits the bare string "N/A" for unsupported sensors).
func amdSMIValueField(m map[string]json.RawMessage, key string) (float64, bool) {
	raw, ok := m[key]
	if !ok {
		return 0, false
	}
	var v struct {
		Value *float64 `json:"value"`
	}
	if err := json.Unmarshal(raw, &v); err != nil || v.Value == nil {
		return 0, false
	}
	return *v.Value, true
}

// ─── AMD (rocm-smi, fallback) ─────────────────────────────────────────────────

type rocmSMIAdapter struct{}

func (a *rocmSMIAdapter) vendor() string { return "amd" }

// sample queries rocm-smi for utilization and VRAM in one invocation and
// temperature in a second (which is often empty on this card/version). Values
// arrive as JSON strings keyed by card (card0, card1, …); we read the first.
func (a *rocmSMIAdapter) sample(ctx context.Context) (GPUStats, error) {
	var stats GPUStats

	// Utilization + VRAM. These two are reliable on the target hardware.
	useOut, err := exec.CommandContext(ctx,
		"rocm-smi", "--showuse", "--showmemuse", "--showmeminfo", "vram", "--json").Output()
	if err != nil {
		return stats, err
	}

	card, err := firstCard(useOut)
	if err != nil {
		return stats, err
	}

	stats.Percent = parseFloatField(card, "GPU use (%)")
	stats.VRAMTotal = parseUintField(card, "VRAM Total Memory (B)")
	stats.VRAMUsed = parseUintField(card, "VRAM Total Used Memory (B)")

	// Temperature is best-effort: rocm-smi 3.1.0 returns empty --showtemp on this
	// card. Read it defensively and leave TempC nil on any failure or absence.
	if temp, ok := a.readTemp(ctx); ok {
		stats.TempC = &temp
	}

	return stats, nil
}

// readTemp attempts to read the edge sensor temperature. Returns ok=false on any
// failure or when the field is missing/empty (a documented quirk on this host).
func (a *rocmSMIAdapter) readTemp(ctx context.Context) (float64, bool) {
	out, err := exec.CommandContext(ctx, "rocm-smi", "--showtemp", "--json").Output()
	if err != nil {
		return 0, false
	}
	card, err := firstCard(out)
	if err != nil {
		return 0, false
	}
	// Sensor key naming varies; try the common edge sensor labels in order.
	for _, key := range []string{
		"Temperature (Sensor edge) (C)",
		"Temperature (Sensor junction) (C)",
		"Temperature (Sensor memory) (C)",
	} {
		if raw, present := card[key]; present {
			if v, err := strconv.ParseFloat(strings.TrimSpace(asString(raw)), 64); err == nil {
				return v, true
			}
		}
	}
	return 0, false
}

// firstCard unmarshals rocm-smi JSON ({"card0": {...}, ...}) and returns the
// first card's field map. rocm-smi orders keys, but map iteration does not, so
// we pick the lowest-numbered cardN key deterministically.
func firstCard(data []byte) (map[string]any, error) {
	var parsed map[string]map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	var bestKey string
	for k := range parsed {
		if !strings.HasPrefix(k, "card") {
			continue
		}
		if bestKey == "" || k < bestKey {
			bestKey = k
		}
	}
	if bestKey == "" {
		return map[string]any{}, nil
	}
	return parsed[bestKey], nil
}

// ─── NVIDIA (nvidia-smi) ──────────────────────────────────────────────────────
//
// NOTE: written blind — no NVIDIA hardware available in the dev environment.

type nvidiaAdapter struct{}

func (n *nvidiaAdapter) vendor() string { return "nvidia" }

// sample queries nvidia-smi for a single CSV row. memory.used/total are reported
// in MiB and converted to bytes; temperature.gpu is in Celsius.
func (n *nvidiaAdapter) sample(ctx context.Context) (GPUStats, error) {
	var stats GPUStats

	out, err := exec.CommandContext(ctx,
		"nvidia-smi",
		"--query-gpu=utilization.gpu,memory.used,memory.total,temperature.gpu,name",
		"--format=csv,noheader,nounits").Output()
	if err != nil {
		return stats, err
	}

	// First line = first GPU. Fields are comma-separated.
	line := firstLine(string(out))
	fields := strings.Split(line, ",")
	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
	}

	if len(fields) > 0 {
		stats.Percent, _ = strconv.ParseFloat(fields[0], 64)
	}
	if len(fields) > 1 {
		if mib, err := strconv.ParseFloat(fields[1], 64); err == nil {
			stats.VRAMUsed = uint64(mib) * 1024 * 1024
		}
	}
	if len(fields) > 2 {
		if mib, err := strconv.ParseFloat(fields[2], 64); err == nil {
			stats.VRAMTotal = uint64(mib) * 1024 * 1024
		}
	}
	if len(fields) > 3 {
		if t, err := strconv.ParseFloat(fields[3], 64); err == nil {
			stats.TempC = &t
		}
	}
	if len(fields) > 4 {
		stats.Name = fields[4]
	}

	return stats, nil
}

// ─── shared parsing helpers ───────────────────────────────────────────────────

// firstLine returns the first non-empty trimmed line of s.
func firstLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		if trimmed := strings.TrimSpace(line); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// asString coerces a JSON-decoded value to its string form. rocm-smi reports
// every value as a quoted string, but we stay defensive against numbers too.
func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return ""
	}
}

// parseFloatField reads a string-typed numeric field, returning 0 when absent or
// unparseable.
func parseFloatField(card map[string]any, key string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(asString(card[key])), 64)
	if err != nil {
		return 0
	}
	return v
}

// parseUintField reads a string-typed byte-count field, returning 0 when absent
// or unparseable.
func parseUintField(card map[string]any, key string) uint64 {
	v, err := strconv.ParseUint(strings.TrimSpace(asString(card[key])), 10, 64)
	if err != nil {
		return 0
	}
	return v
}
