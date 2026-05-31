package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// unmarshalSystem decodes just the system: block from a YAML snippet.
func unmarshalSystem(t *testing.T, src string) SystemConfig {
	t.Helper()
	var wrapper struct {
		System SystemConfig `yaml:"system"`
	}
	if err := yaml.Unmarshal([]byte(src), &wrapper); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return wrapper.System
}

func TestSystemConfig_BareEnabledEnablesAll(t *testing.T) {
	got := unmarshalSystem(t, "system:\n  enabled: true\n")
	if !got.CPU.Enabled || !got.Memory.Enabled || !got.Load.Enabled || !got.GPU.Enabled {
		t.Fatalf("bare enabled:true should enable all four metrics, got %+v", got)
	}
}

func TestSystemConfig_BareDisabledDisablesAll(t *testing.T) {
	got := unmarshalSystem(t, "system:\n  enabled: false\n")
	if got.CPU.Enabled || got.Memory.Enabled || got.Load.Enabled || got.GPU.Enabled {
		t.Fatalf("bare enabled:false should disable all four metrics, got %+v", got)
	}
}

func TestSystemConfig_PerMetric(t *testing.T) {
	got := unmarshalSystem(t, `
system:
  cpu:    { enabled: true }
  memory: { enabled: false }
  load:   { enabled: true }
  gpu:    { enabled: false }
`)
	if !got.CPU.Enabled || got.Memory.Enabled || !got.Load.Enabled || got.GPU.Enabled {
		t.Fatalf("per-metric flags not honored, got %+v", got)
	}
}

func TestSystemConfig_SubKeyOverridesBare(t *testing.T) {
	// Bare enabled:true sets the base; the present gpu sub-key overrides to false.
	got := unmarshalSystem(t, `
system:
  enabled: true
  gpu: { enabled: false }
`)
	if !got.CPU.Enabled || !got.Memory.Enabled || !got.Load.Enabled {
		t.Fatalf("bare base should still enable unspecified metrics, got %+v", got)
	}
	if got.GPU.Enabled {
		t.Fatalf("present gpu sub-key should override bare base to false, got %+v", got)
	}
}

func TestSystemConfig_AbsentDisablesAll(t *testing.T) {
	got := unmarshalSystem(t, "title: x\n")
	if got.CPU.Enabled || got.Memory.Enabled || got.Load.Enabled || got.GPU.Enabled {
		t.Fatalf("absent system block should leave all metrics disabled, got %+v", got)
	}
}
