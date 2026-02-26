package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ServiceType string

const (
	ServiceTypeHTTP   ServiceType = "http"
	ServiceTypeDocker ServiceType = "docker"
)

type ServiceConfig struct {
	Name      string      `yaml:"name"`
	Type      ServiceType `yaml:"type"`
	URL       string      `yaml:"url,omitempty"`
	Container string      `yaml:"container,omitempty"`
	Icon      string      `yaml:"icon,omitempty"`
	Category  string      `yaml:"category,omitempty"`
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
		Port:          8080,
		CheckInterval: 30,
		Title:         "Aperture",
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
