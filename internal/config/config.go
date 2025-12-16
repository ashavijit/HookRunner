package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Tool struct {
	Version  string            `yaml:"version" json:"version"`
	Install  map[string]string `yaml:"install" json:"install"`
	Checksum string            `yaml:"checksum" json:"checksum"`
}

type Hook struct {
	Name    string   `yaml:"name" json:"name"`
	Tool    string   `yaml:"tool" json:"tool"`
	Args    []string `yaml:"args" json:"args"`
	Files   string   `yaml:"files" json:"files"`
	Timeout string   `yaml:"timeout" json:"timeout"`
	After   string   `yaml:"after" json:"after"`
}

type Config struct {
	Tools map[string]Tool   `yaml:"tools" json:"tools"`
	Hooks map[string][]Hook `yaml:"hooks" json:"hooks"`
}

func Load(dir string) (*Config, string, error) {
	candidates := []string{"hooks.yaml", "hooks.yml", "hooks.json"}
	for _, name := range candidates {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			cfg, err := loadFile(path)
			return cfg, path, err
		}
	}
	return nil, "", fmt.Errorf("no config file found (hooks.yaml, hooks.yml, or hooks.json)")
}

func loadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("invalid JSON config: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("invalid YAML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format: %s", ext)
	}

	return &cfg, nil
}

func (c *Config) GetHooks(hookType string) []Hook {
	if c.Hooks == nil {
		return nil
	}
	return c.Hooks[hookType]
}

func (c *Config) GetTool(name string) *Tool {
	if c.Tools == nil {
		return nil
	}
	if tool, ok := c.Tools[name]; ok {
		return &tool
	}
	return nil
}
