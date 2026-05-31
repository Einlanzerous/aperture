package api

import (
	"encoding/json"
	"testing"
)

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		pattern, name string
		want          bool
	}{
		{"llama3:8b", "llama3:8b", true},                       // exact
		{"llama3:8b", "llama3:70b", false},                     // exact mismatch
		{"hf.co/*", "hf.co/TheBloke/foo:Q4", true},             // '*' crosses '/'
		{"hf.co/*", "ollama.com/library/llama3", false},        // prefix mismatch
		{"*duplicate*", "some-duplicate-model:latest", true},   // interior match
		{"*:Q4_K_M", "hf.co/x/y:Q4_K_M", true},                 // suffix anchor
		{"*:Q4_K_M", "hf.co/x/y:Q8_0", false},                  // suffix mismatch
		{"a*b*c", "axxbyyc", true},                             // multiple wildcards
		{"a*b*c", "axxc", false},                               // missing interior seg
	}
	for _, c := range cases {
		if got := matchGlob(c.pattern, c.name); got != c.want {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", c.pattern, c.name, got, c.want)
		}
	}
}

func TestFilterOllamaModels(t *testing.T) {
	body := []byte(`{"models":[
		{"name":"llama3:8b","size":123,"details":{"family":"llama"}},
		{"name":"hf.co/TheBloke/foo:Q4","size":456},
		{"name":"qwen:7b","size":789}
	]}`)

	out, err := filterOllamaModels(body, []string{"hf.co/*"})
	if err != nil {
		t.Fatalf("filterOllamaModels: %v", err)
	}

	var parsed struct {
		Models []struct {
			Name    string         `json:"name"`
			Size    int            `json:"size"`
			Details map[string]any `json:"details"`
		} `json:"models"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if len(parsed.Models) != 2 {
		t.Fatalf("expected 2 models after filter, got %d", len(parsed.Models))
	}
	for _, m := range parsed.Models {
		if m.Name == "hf.co/TheBloke/foo:Q4" {
			t.Errorf("hidden model leaked through filter")
		}
	}
	// Surviving objects must keep their other fields intact.
	if parsed.Models[0].Size != 123 || parsed.Models[0].Details["family"] != "llama" {
		t.Errorf("non-name fields not preserved: %+v", parsed.Models[0])
	}
}

func TestFilterOllamaModelsBadShape(t *testing.T) {
	if _, err := filterOllamaModels([]byte(`not json`), []string{"x"}); err == nil {
		t.Errorf("expected error on malformed body")
	}
}
