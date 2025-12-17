package executor

import (
	"testing"

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
		SkipHooks: []string{"lint"},
	}
	exec.SetOptions(opts)

	if exec.opts.Verbose != true {
		t.Error("Verbose should be true")
	}

	if exec.opts.Fix != true {
		t.Error("Fix should be true")
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
	result := ParseSkipEnv()
	if result != nil && len(result) > 0 {
		t.Logf("SKIP env: %v", result)
	}
}

func TestPrintResults_Quiet(t *testing.T) {
	results := []Result{
		{Name: "test", Success: true},
	}
	PrintResults(results, false, true)
}

func TestPrintPolicyResult_Nil(t *testing.T) {
	PrintPolicyResult(nil, false)
}
