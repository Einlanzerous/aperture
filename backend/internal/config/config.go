package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultPort              = 8888
	DefaultCheckInterval     = 30
	DefaultTitle             = "Aperture"
	DefaultDockerSocket      = "/var/run/docker.sock"
	DefaultRetentionRaw      = 48 * time.Hour
	DefaultRetentionSummary  = 30 * 24 * time.Hour // 30 days
	DefaultCompactCycleCount = 100
)

type ServiceType string

const (
	ServiceTypeHTTP   ServiceType = "http"
	ServiceTypeDocker ServiceType = "docker"
)

var validSizes = map[string]bool{"": true, "s": true, "m": true, "l": true}

type SemaphoreConfig struct {
	URL      string `yaml:"url"`
	TokenEnv string `yaml:"token_env"` // env var name holding the API token
}

type ActionConfig struct {
	Name       string `yaml:"name"`
	ProjectID  int    `yaml:"project_id"`
	TemplateID int    `yaml:"template_id"`
	Category   string `yaml:"category,omitempty"`
	Icon       string `yaml:"icon,omitempty"`
	Size       string `yaml:"size,omitempty"`
}

type ServiceConfig struct {
	Name          string      `yaml:"name"`
	Type          ServiceType `yaml:"type"`
	URL           string      `yaml:"url,omitempty"`
	Container     string      `yaml:"container,omitempty"`
	Icon          string      `yaml:"icon,omitempty"`
	Category      string      `yaml:"category,omitempty"`
	Href          string      `yaml:"href,omitempty"`
	Size          string      `yaml:"size,omitempty"`           // s | m | l
	DetailDefault bool        `yaml:"detail_default,omitempty"` // show detailed history view by default
	SkipVerify    bool        `yaml:"skip_verify,omitempty"`    // skip TLS certificate verification
}

type OllamaConfig struct {
	URL string `yaml:"url"`
}

type SystemConfig struct {
	Enabled bool `yaml:"enabled"`
}

type RetentionConfig struct {
	Raw     time.Duration `yaml:"raw"`
	Summary time.Duration `yaml:"summary"`
}

type StorageConfig struct {
	Driver    string          `yaml:"driver"` // "sqlite", "postgres", or "" (disabled)
	DSN       string          `yaml:"dsn"`
	Retention RetentionConfig `yaml:"retention"`
}

type Config struct {
	Port          int             `yaml:"port"`
	CheckInterval int             `yaml:"check_interval"` // seconds
	Title         string          `yaml:"title"`
	DockerSocket  string          `yaml:"docker_socket"`
	CORSOrigins   []string        `yaml:"cors_origins"`
	Services      []ServiceConfig `yaml:"services"`
	Ollama        OllamaConfig    `yaml:"ollama"`
	System        SystemConfig    `yaml:"system"`
	Storage       StorageConfig   `yaml:"storage"`
	Semaphore     SemaphoreConfig `yaml:"semaphore"`
	Actions       []ActionConfig  `yaml:"actions"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Port:          DefaultPort,
		CheckInterval: DefaultCheckInterval,
		Title:         DefaultTitle,
		DockerSocket:  DefaultDockerSocket,
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	applyEnvOverrides(cfg)
	applyStorageDefaults(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("APERTURE_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			cfg.Port = p
		}
	}
	if v := os.Getenv("APERTURE_CHECK_INTERVAL"); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.CheckInterval = i
		}
	}
	if v := os.Getenv("APERTURE_TITLE"); v != "" {
		cfg.Title = v
	}
}

func applyStorageDefaults(cfg *Config) {
	if cfg.Storage.Retention.Raw == 0 {
		cfg.Storage.Retention.Raw = DefaultRetentionRaw
	}
	if cfg.Storage.Retention.Summary == 0 {
		cfg.Storage.Retention.Summary = DefaultRetentionSummary
	}
}

var validDrivers = map[string]bool{"": true, "sqlite": true, "postgres": true}

func (c *Config) Validate() error {
	if c.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be > 0, got %d", c.CheckInterval)
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be 1–65535, got %d", c.Port)
	}

	seen := make(map[string]bool, len(c.Services))
	for i, svc := range c.Services {
		if svc.Name == "" {
			return fmt.Errorf("services[%d]: name is required", i)
		}
		if seen[svc.Name] {
			return fmt.Errorf("services[%d]: duplicate name %q", i, svc.Name)
		}
		seen[svc.Name] = true

		switch svc.Type {
		case ServiceTypeHTTP:
			if svc.URL == "" {
				return fmt.Errorf("service %q: http type requires a url", svc.Name)
			}
		case ServiceTypeDocker:
			if svc.Container == "" {
				return fmt.Errorf("service %q: docker type requires a container", svc.Name)
			}
		default:
			return fmt.Errorf("service %q: unsupported type %q", svc.Name, svc.Type)
		}

		if !validSizes[svc.Size] {
			return fmt.Errorf("service %q: size must be s, m, or l; got %q", svc.Name, svc.Size)
		}
	}

	if !validDrivers[c.Storage.Driver] {
		return fmt.Errorf("storage.driver must be sqlite, postgres, or empty; got %q", c.Storage.Driver)
	}
	if c.Storage.Driver != "" && c.Storage.DSN == "" {
		return fmt.Errorf("storage.dsn is required when driver is %q", c.Storage.Driver)
	}
	if c.Storage.Retention.Raw <= 0 {
		return fmt.Errorf("storage.retention.raw must be positive, got %s", c.Storage.Retention.Raw)
	}
	if c.Storage.Retention.Summary <= 0 {
		return fmt.Errorf("storage.retention.summary must be positive, got %s", c.Storage.Retention.Summary)
	}

	if len(c.Actions) > 0 {
		if c.Semaphore.URL == "" {
			return fmt.Errorf("semaphore.url is required when actions are configured")
		}
		if c.Semaphore.TokenEnv == "" {
			return fmt.Errorf("semaphore.token_env is required when actions are configured")
		}
		if os.Getenv(c.Semaphore.TokenEnv) == "" {
			return fmt.Errorf("environment variable %q (semaphore.token_env) is not set", c.Semaphore.TokenEnv)
		}
	}

	for i, action := range c.Actions {
		if action.Name == "" {
			return fmt.Errorf("actions[%d]: name is required", i)
		}
		if seen[action.Name] {
			return fmt.Errorf("actions[%d]: duplicate name %q (names must be unique across services and actions)", i, action.Name)
		}
		seen[action.Name] = true
		if action.ProjectID <= 0 {
			return fmt.Errorf("action %q: project_id must be > 0", action.Name)
		}
		if action.TemplateID <= 0 {
			return fmt.Errorf("action %q: template_id must be > 0", action.Name)
		}
		if !validSizes[action.Size] {
			return fmt.Errorf("action %q: size must be s, m, or l; got %q", action.Name, action.Size)
		}
	}

	return nil
}
