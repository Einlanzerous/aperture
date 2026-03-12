package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

const (
	DefaultPort          = 8888
	DefaultCheckInterval = 30
	DefaultTitle         = "Aperture"
	DefaultDockerSocket  = "/var/run/docker.sock"
)

type ServiceType string

const (
	ServiceTypeHTTP   ServiceType = "http"
	ServiceTypeDocker ServiceType = "docker"
)

var validSizes = map[string]bool{"": true, "s": true, "m": true, "l": true}

type ServiceConfig struct {
	Name      string      `yaml:"name"`
	Type      ServiceType `yaml:"type"`
	URL       string      `yaml:"url,omitempty"`
	Container string      `yaml:"container,omitempty"`
	Icon      string      `yaml:"icon,omitempty"`
	Category  string      `yaml:"category,omitempty"`
	Href      string      `yaml:"href,omitempty"`
	Size      string      `yaml:"size,omitempty"` // s | m | l
}

type OllamaConfig struct {
	URL string `yaml:"url"`
}

type SystemConfig struct {
	Enabled bool `yaml:"enabled"`
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

	return nil
}
