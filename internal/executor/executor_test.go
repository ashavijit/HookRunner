package executor

import (
	"os"
	"testing"
	"time"

	"github.com/ashavijit/hookrunner/internal/config"
	"github.com/ashavijit/hookrunner/internal/tool"
)

func TestNew(t *testing.T) {
	cfg := &config.Config{}
	toolMgr := tool.NewManager("/tmp/cache")
	exec := New(cfg, toolMgr, "/tmp/work")

	if exec == nil {
		t.Fatal("executor should not be nil")
	}

	if exec.opts.FailFast != true {
		t.Error("default FailFast should be true")
	}
}

func TestSetOptions(t *testing.T) {
	cfg := &config.Config{}
	toolMgr := tool.NewManager("/tmp/cache")
	exec := New(cfg, toolMgr, "/tmp/work")

	opts := Options{
		Verbose:   true,
		Quiet:     false,
		Fix:       true,
		FailFast:  false,
		UseCache:  true,
		SkipHooks: []string{"lint"},
	}
	exec.SetOptions(opts)

	if exec.opts.Verbose != true {
		t.Error("Verbose should be true")
	}

	if exec.opts.Fix != true {
		t.Error("Fix should be true")
	}

	if exec.opts.UseCache != true {
		t.Error("UseCache should be true")
	}
}

func TestRun_NoHooks(t *testing.T) {
	cfg := &config.Config{}
	toolMgr := tool.NewManager("/tmp/cache")
	exec := New(cfg, toolMgr, "/tmp/work")

	results := exec.Run("pre-commit", nil, false)

	if results != nil {
		t.Error("expected nil results for no hooks")
	}
}

func TestRun_WithCycleDetection(t *testing.T) {
	cfg := &config.Config{
		Hooks: map[string][]config.Hook{
			"pre-commit": {
				{Name: "a", Tool: "echo", After: "b"},
				{Name: "b", Tool: "echo", After: "a"},
			},
		},
	}
	toolMgr := tool.NewManager("/tmp/cache")
	exec := New(cfg, toolMgr, "/tmp/work")

	results := exec.Run("pre-commit", []string{"test.go"}, false)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Success {
		t.Error("expected failure due to cycle")
	}
}

func TestRun_DryRun(t *testing.T) {
	cfg := &config.Config{
		Hooks: map[string][]config.Hook{
			"pre-commit": {
				{Name: "test", Tool: "echo", Args: []string{"hello"}},
			},
		},
	}
	toolMgr := tool.NewManager("/tmp/cache")
	exec := New(cfg, toolMgr, t.TempDir())
	exec.SetOptions(Options{DryRun: true, Quiet: true})

	results := exec.Run("pre-commit", []string{"test.go"}, true)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestCheckPolicies_NilPolicies(t *testing.T) {
	cfg := &config.Config{}
	toolMgr := tool.NewManager("/tmp/cache")
	exec := New(cfg, toolMgr, "/tmp/work")

	result := exec.CheckPolicies([]string{"a.go"}, "msg")

	if result != nil {
		t.Error("expected nil result for nil policies")
	}
}

func TestCheckPolicies_WithPolicies(t *testing.T) {
	cfg := &config.Config{
		Policies: &config.Policies{
			Type: "raw",
			LocalPolicies: []config.LocalPolicy{
				{
					Name:    "max-files",
					Version: "local",
					Rules: config.PolicyRules{
						MaxFilesChanged: 2,
					},
				},
			},
		},
	}
	toolMgr := tool.NewManager("/tmp/cache")
	exec := New(cfg, toolMgr, t.TempDir())

	files := []string{"a.go", "b.go", "c.go"}
	result := exec.CheckPolicies(files, "msg")

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Passed {
		t.Error("expected policy to fail")
	}
}

func TestHasFailure(t *testing.T) {
	results := []Result{
		{Name: "a", Success: true},
		{Name: "b", Success: true},
	}

	if HasFailure(results) {
		t.Error("should not have failure")
	}

	results = []Result{
		{Name: "a", Success: true},
		{Name: "b", Success: false},
	}

	if !HasFailure(results) {
		t.Error("should have failure")
	}
}

func TestHasFailure_Skipped(t *testing.T) {
	results := []Result{
		{Name: "a", Success: true},
		{Name: "b", Success: false, Skipped: true},
	}

	if HasFailure(results) {
		t.Error("skipped failures should not count")
	}
}

func TestParseSkipEnv(t *testing.T) {
	_ = os.Setenv("SKIP", "lint,test")
	result := ParseSkipEnv()
	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	_ = os.Unsetenv("SKIP")
}

func TestParseSkipEnv_Empty(t *testing.T) {
	_ = os.Unsetenv("SKIP")
	result := ParseSkipEnv()
	if result != nil {
		t.Error("expected nil for empty SKIP env")
	}
}

func TestPrintResults_Quiet(t *testing.T) {
	results := []Result{
		{Name: "test", Success: true},
	}
	PrintResults(results, false, true)
}

func TestPrintResults_Verbose(t *testing.T) {
	results := []Result{
		{Name: "pass", Success: true, Duration: time.Millisecond * 100, Output: "output"},
		{Name: "fail", Success: false, Duration: time.Millisecond * 200, Output: "error"},
		{Name: "skip", Success: true, Skipped: true, Duration: time.Millisecond * 50},
	}
	PrintResults(results, true, false)
}

func TestPrintPolicyResult_Nil(t *testing.T) {
	PrintPolicyResult(nil, false)
}

func TestClearCache(t *testing.T) {
	cfg := &config.Config{}
	toolMgr := tool.NewManager("/tmp/cache")
	exec := New(cfg, toolMgr, t.TempDir())

	err := exec.ClearCache()
	if err != nil {
		t.Errorf("ClearCache failed: %v", err)
	}
}

func TestResult_Fields(t *testing.T) {
	r := Result{
		Name:     "test",
		Success:  true,
		Skipped:  false,
		Duration: time.Second,
		Output:   "output",
		Error:    nil,
	}

	if r.Name != "test" {
		t.Error("Name mismatch")
	}
	if !r.Success {
		t.Error("Success should be true")
	}
}

func TestOptions_Fields(t *testing.T) {
	o := Options{
		Verbose:    true,
		Quiet:      false,
		Fix:        true,
		FailFast:   true,
		DryRun:     true,
		JSONOutput: true,
		NoColor:    true,
		UseCache:   true,
		SkipHooks:  []string{"a", "b"},
		CommitMsg:  "test",
	}

	if !o.Verbose {
		t.Error("Verbose should be true")
	}
	if !o.UseCache {
		t.Error("UseCache should be true")
	}
}
