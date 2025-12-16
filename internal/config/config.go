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
	Name     string            `yaml:"name" json:"name"`
	Tool     string            `yaml:"tool" json:"tool"`
	Args     []string          `yaml:"args" json:"args"`
	FixArgs  []string          `yaml:"fix_args" json:"fix_args"`
	Files    string            `yaml:"files" json:"files"`
	Glob     string            `yaml:"glob" json:"glob"`
	Exclude  string            `yaml:"exclude" json:"exclude"`
	Timeout  string            `yaml:"timeout" json:"timeout"`
	After    string            `yaml:"after" json:"after"`
	Skip     string            `yaml:"skip" json:"skip"`
	Only     string            `yaml:"only" json:"only"`
	Env      map[string]string `yaml:"env" json:"env"`
	PassEnv  []string          `yaml:"pass_env" json:"pass_env"`
	FailFast bool              `yaml:"fail_fast" json:"fail_fast"`
}

type CommitMessagePolicy struct {
	Regex     string `yaml:"regex" json:"regex"`
	MaxLength int    `yaml:"max_length" json:"max_length"`
	MinLength int    `yaml:"min_length" json:"min_length"`
}

type Policies struct {
	MaxFilesChanged   int                 `yaml:"max_files_changed" json:"max_files_changed"`
	ForbidDirectories []string            `yaml:"forbid_directories" json:"forbid_directories"`
	ForbidFiles       []string            `yaml:"forbid_files" json:"forbid_files"`
	RequireFiles      []string            `yaml:"require_files" json:"require_files"`
	CommitMessage     CommitMessagePolicy `yaml:"commit_message" json:"commit_message"`
}

type Config struct {
	Tools    map[string]Tool   `yaml:"tools" json:"tools"`
	Hooks    map[string][]Hook `yaml:"hooks" json:"hooks"`
	Policies *Policies         `yaml:"policies" json:"policies"`
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

func DefaultConfig() string {
	return `tools:
  golangci-lint:
    version: 1.55.2
    install:
      windows: https://github.com/golangci/golangci-lint/releases/download/v1.55.2/golangci-lint-1.55.2-windows-amd64.zip
      linux: https://github.com/golangci/golangci-lint/releases/download/v1.55.2/golangci-lint-1.55.2-linux-amd64.tar.gz
      darwin: https://github.com/golangci/golangci-lint/releases/download/v1.55.2/golangci-lint-1.55.2-darwin-amd64.tar.gz

policies:
  max_files_changed: 50
  forbid_directories: ["vendor/", "generated/"]
  commit_message:
    regex: "^(feat|fix|chore|docs|refactor|test):"
    min_length: 10

hooks:
  pre-commit:
    - name: gofmt
      tool: go
      args: ["fmt", "./..."]
      files: "\\.go$"

    - name: govet
      tool: go
      args: ["vet", "./..."]
      files: "\\.go$"
      after: gofmt

  pre-push:
    - name: test
      tool: go
      args: ["test", "./..."]
      timeout: 5m
`
}
