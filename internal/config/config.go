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
	Name        string            `yaml:"name" json:"name"`
	Tool        string            `yaml:"tool" json:"tool"`
	Run         string            `yaml:"run" json:"run"`
	Script      string            `yaml:"script" json:"script"`
	Runner      string            `yaml:"runner" json:"runner"`
	Args        []string          `yaml:"args" json:"args"`
	FixArgs     []string          `yaml:"fix_args" json:"fix_args"`
	Files       string            `yaml:"files" json:"files"`
	Glob        string            `yaml:"glob" json:"glob"`
	Exclude     string            `yaml:"exclude" json:"exclude"`
	Root        string            `yaml:"root" json:"root"`
	Timeout     string            `yaml:"timeout" json:"timeout"`
	After       string            `yaml:"after" json:"after"`
	Skip        string            `yaml:"skip" json:"skip"`
	Only        string            `yaml:"only" json:"only"`
	Tags        []string          `yaml:"tags" json:"tags"`
	Env         map[string]string `yaml:"env" json:"env"`
	PassEnv     []string          `yaml:"pass_env" json:"pass_env"`
	FailFast    bool              `yaml:"fail_fast" json:"fail_fast"`
	Interactive bool              `yaml:"interactive" json:"interactive"`
	StageFixed  bool              `yaml:"stage_fixed" json:"stage_fixed"`
	Piped       bool              `yaml:"piped" json:"piped"`
}

type PolicyRef struct {
	URL string `yaml:"url" json:"url"`
}

type ForbiddenContentPattern struct {
	Pattern     string `yaml:"pattern" json:"pattern"`
	Description string `yaml:"description" json:"description"`
}

type CommitMessageRule struct {
	Regex string `yaml:"regex" json:"regex"`
	Error string `yaml:"error" json:"error"`
}

type PolicyRules struct {
	ForbidFiles          []string                  `yaml:"forbid_files" json:"forbid_files"`
	ForbidDirectories    []string                  `yaml:"forbid_directories" json:"forbid_directories"`
	ForbidFileExtensions []string                  `yaml:"forbid_file_extensions" json:"forbid_file_extensions"`
	RequiredFiles        []string                  `yaml:"required_files" json:"required_files"`
	MaxFileSizeKB        int                       `yaml:"max_file_size_kb" json:"max_file_size_kb"`
	MaxFilesChanged      int                       `yaml:"max_files_changed" json:"max_files_changed"`
	ForbidFileContent    []ForbiddenContentPattern `yaml:"forbid_file_content" json:"forbid_file_content"`
	CommitMessage        *CommitMessageRule        `yaml:"commit_message" json:"commit_message"`
	EnforceHooks         []string                  `yaml:"enforce_hooks" json:"enforce_hooks"`
	HookTimeBudgetMs     map[string]int            `yaml:"hook_time_budget_ms" json:"hook_time_budget_ms"`
	MaxParallelHooks     int                       `yaml:"max_parallel_hooks" json:"max_parallel_hooks"`
}

type LocalPolicy struct {
	Name        string            `yaml:"name" json:"name"`
	Version     string            `yaml:"version" json:"version"`
	Description string            `yaml:"description" json:"description"`
	Rules       PolicyRules       `yaml:"rules" json:"rules"`
	Metadata    map[string]string `yaml:"metadata" json:"metadata"`
}

type Policies struct {
	Type          string        `yaml:"type" json:"type"`
	Policies      []PolicyRef   `yaml:"policies" json:"policies"`
	LocalPolicies []LocalPolicy `yaml:"localPolicies" json:"localPolicies"`
}

type Config struct {
	Tools       map[string]Tool   `yaml:"tools" json:"tools"`
	Hooks       map[string][]Hook `yaml:"hooks" json:"hooks"`
	Policies    *Policies         `yaml:"policies" json:"policies"`
	ExcludeTags []string          `yaml:"exclude_tags" json:"exclude_tags"`
	Parallel    bool              `yaml:"parallel" json:"parallel"`
	ScriptsDir  string            `yaml:"scripts_dir" json:"scripts_dir"`
}

func Load(dir string) (*Config, string, error) {
	candidates := []string{"hooks.yaml", "hooks.yml", "hooks.json"}
	for _, name := range candidates {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			cfg, err := loadFile(path)
			if err != nil {
				return nil, path, err
			}
			cfg = mergeLocalConfig(cfg, dir)
			return cfg, path, nil
		}
	}
	return nil, "", fmt.Errorf("no config file found (hooks.yaml, hooks.yml, or hooks.json)")
}

func mergeLocalConfig(cfg *Config, dir string) *Config {
	localFiles := []string{"hooks-local.yaml", "hooks-local.yml", "hooks-local.json"}
	for _, name := range localFiles {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			local, err := loadFile(path)
			if err == nil {
				cfg = mergeConfigs(cfg, local)
			}
			break
		}
	}
	return cfg
}

func mergeConfigs(base, override *Config) *Config {
	if override.ExcludeTags != nil {
		base.ExcludeTags = append(base.ExcludeTags, override.ExcludeTags...)
	}
	if override.Parallel {
		base.Parallel = true
	}
	if override.ScriptsDir != "" {
		base.ScriptsDir = override.ScriptsDir
	}
	for hookType, hooks := range override.Hooks {
		if base.Hooks == nil {
			base.Hooks = make(map[string][]Hook)
		}
		for _, oh := range hooks {
			found := false
			for i, bh := range base.Hooks[hookType] {
				if bh.Name == oh.Name {
					base.Hooks[hookType][i] = oh
					found = true
					break
				}
			}
			if !found {
				base.Hooks[hookType] = append(base.Hooks[hookType], oh)
			}
		}
	}
	return base
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

func (c *Config) HasRemotePolicies() bool {
	if c.Policies == nil {
		return false
	}
	return c.Policies.Type == "raw" && len(c.Policies.Policies) > 0
}

func (c *Config) GetPolicyURLs() []string {
	if c.Policies == nil {
		return nil
	}
	urls := make([]string, len(c.Policies.Policies))
	for i, ref := range c.Policies.Policies {
		urls[i] = ref.URL
	}
	return urls
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
  type: raw
  policies:
    - url: https://policies.example.dev/default.yaml
  localPolicies:
    commit-style:
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
