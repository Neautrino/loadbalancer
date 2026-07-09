package config

import (
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

type Duration time.Duration

type Config struct {
	Port int `yaml:"port"`
	Algorithm string `yaml:"algorithm"`
	Health HealthConfig `yaml:"health"`
	CircuitBreaker CBConfig `yaml:"circuit_breaker"`
	Backends []BackendConfig `yaml:"backends"`
}

type HealthConfig struct {
	Interval Duration `yaml:"interval"`
	Path string `yaml:"path"`
}

type CBConfig struct {
	Threshold int `yaml:"threshold"`
	Cooldown Duration `yaml:"cooldown"`
}

type BackendConfig struct {
	URL string `yaml:"url"`
	Weight int `yaml:"weight"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return  nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}

	*d = Duration(parsed)
	return nil
}