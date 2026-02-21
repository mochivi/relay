package config

import (
	"io"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Global   *GlobalConfig    `yaml:"global"`
	Services []*ServiceConfig `yaml:"services"`
	Routes   []*RouteConfig   `yaml:"routes"`
}

type GlobalConfig struct {
	Port int    `yaml:"port"`
	Addr string `yaml:"addr"`
}

type ServiceConfig struct {
	Name      string   `yaml:"name"`
	Algorithm string   `yaml:"algorithm"`
	Backends  []string `yaml:"backends"`
}

type RouteConfig struct {
	Pattern string `yaml:"path"`
	Service string `yaml:"service"`
}

func ParseConfig(reader io.Reader) (*Config, error) {
	doc, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	config := Config{}
	if err := yaml.Unmarshal(doc, &config); err != nil {
		return nil, err
	}
	config.handleDefaults()
	return &config, nil
}

func (c *Config) handleDefaults() {
	c.Global.handleDefaults()
	for _, svc := range c.Services {
		svc.handleDefaults()
	}
}

func (c *ServiceConfig) handleDefaults() {
	if c.Algorithm == "" {
		c.Algorithm = "round_robin"
	}
}

func (c *GlobalConfig) handleDefaults() {
	if c.Port == 0 {
		c.Port = 8080
	}
	if c.Addr == "" {
		c.Addr = "127.0.0.1"
	}
}
