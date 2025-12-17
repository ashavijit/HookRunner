package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_YAML(t *testing.T) {
	dir := t.TempDir()
	content := `
hooks:
  pre-commit:
    - name: test
      tool: go
      args: ["test"]
`
	err := os.WriteFile(filepath.Join(dir, "hooks.yaml"), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, path, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg == nil {
		t.Fatal("config is nil")
	}

	if !filepath.IsAbs(path) {
		t.Error("path should be absolute")
	}

	hooks := cfg.GetHooks("pre-commit")
	if len(hooks) != 1 {
		t.Errorf("expected 1 hook, got %d", len(hooks))
	}

	if hooks[0].Name != "test" {
		t.Errorf("expected hook name 'test', got '%s'", hooks[0].Name)
	}
}

func TestLoad_JSON(t *testing.T) {
	dir := t.TempDir()
	content := `{
  "hooks": {
    "pre-commit": [
      {"name": "test", "tool": "go", "args": ["test"]}
    ]
  }
}`
	err := os.WriteFile(filepath.Join(dir, "hooks.json"), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, _, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	hooks := cfg.GetHooks("pre-commit")
	if len(hooks) != 1 {
		t.Errorf("expected 1 hook, got %d", len(hooks))
	}
}

func TestLoad_NotFound(t *testing.T) {
	dir := t.TempDir()

	_, _, err := Load(dir)
	if err == nil {
		t.Error("expected error for missing config")
	}
}

func TestLoad_WithPolicies(t *testing.T) {
	dir := t.TempDir()
	content := `
policies:
  type: raw
  policies:
    - url: https://example.com/policy.yaml
  localPolicies:
    - name: commit-style
      version: local
      rules:
        commit_message:
          regex: "^feat:"

hooks:
  pre-commit:
    - name: test
      tool: go
`
	err := os.WriteFile(filepath.Join(dir, "hooks.yaml"), []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, _, err := Load(dir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Policies == nil {
		t.Fatal("policies is nil")
	}

	if cfg.Policies.Type != "raw" {
		t.Errorf("expected type raw, got %s", cfg.Policies.Type)
	}

	if len(cfg.Policies.Policies) != 1 {
		t.Errorf("expected 1 policy URL, got %d", len(cfg.Policies.Policies))
	}

	if len(cfg.Policies.LocalPolicies) != 1 {
		t.Error("expected 1 local policy")
	}

	if !cfg.HasRemotePolicies() {
		t.Error("expected HasRemotePolicies to return true")
	}

	urls := cfg.GetPolicyURLs()
	if len(urls) != 1 || urls[0] != "https://example.com/policy.yaml" {
		t.Errorf("unexpected URLs: %v", urls)
	}
}

func TestGetHooks_Empty(t *testing.T) {
	cfg := &Config{}

	hooks := cfg.GetHooks("pre-commit")
	if hooks != nil {
		t.Error("expected nil for empty config")
	}
}

func TestGetTool(t *testing.T) {
	cfg := &Config{
		Tools: map[string]Tool{
			"lint": {Version: "1.0.0"},
		},
	}

	tool := cfg.GetTool("lint")
	if tool == nil {
		t.Fatal("expected tool to be found")
	}

	if tool.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", tool.Version)
	}

	tool = cfg.GetTool("nonexistent")
	if tool != nil {
		t.Error("expected nil for nonexistent tool")
	}
}

func TestGetTool_Empty(t *testing.T) {
	cfg := &Config{}

	tool := cfg.GetTool("lint")
	if tool != nil {
		t.Error("expected nil for empty tools")
	}
}

func TestDefaultConfig(t *testing.T) {
	content := DefaultConfig()

	if content == "" {
		t.Error("default config should not be empty")
	}

	if len(content) < 100 {
		t.Error("default config seems too short")
	}
}
