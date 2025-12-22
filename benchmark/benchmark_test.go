package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ashavijit/hookrunner/internal/config"
	"github.com/ashavijit/hookrunner/internal/executor"
	"github.com/ashavijit/hookrunner/internal/tool"
)

func createDummyFiles(dir string, count int) ([]string, error) {
	var files []string
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("file_%d.go", i)
		path := filepath.Join(dir, name)
		content := fmt.Sprintf("package main\n\nfunc Foo%d() {}\n", i)
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			return nil, err
		}
		files = append(files, name)
	}
	return files, nil
}

func BenchmarkHookRunner_Execution(b *testing.B) {
	// Setup temp workspace
	workDir := b.TempDir()
	files, err := createDummyFiles(workDir, 100)
	if err != nil {
		b.Fatal(err)
	}

	// Setup config with a complex DAG
	// Hook A (parallel)
	// Hook B (parallel)
	// Hook C (parallel)
	// Hook D (after A)
	// Hook E (after B, C)
	// Hook F (after D, E)
	cfg := &config.Config{
		Hooks: map[string][]config.Hook{
			"pre-commit": {
				{Name: "hook-A", Tool: "echo", Args: []string{"running A"}},
				{Name: "hook-B", Tool: "echo", Args: []string{"running B"}},
				{Name: "hook-C", Tool: "echo", Args: []string{"running C"}},
				{Name: "hook-D", Tool: "echo", Args: []string{"running D"}, After: "hook-A"},
				{Name: "hook-E", Tool: "echo", Args: []string{"running E"}, After: "hook-B"}, // simplified dep
				{Name: "hook-F", Tool: "echo", Args: []string{"running F"}, After: "hook-D"},
			},
		},
	}

	toolMgr := tool.NewManager(filepath.Join(workDir, ".cache"))
	// Pre-warm tool manager if needed, but echo is specific.
	// Actually, HookRunner might search for "echo" using LookPath if Tool is specified?
	// The executor calls exec.Command(toolPath, args...) where toolPath is from toolMgr.EnsureTool
	// or system path if not managed. "echo" is special on Windows (built-in cmd).
	// Let's use "go" as the tool since it's likely installed, or a simple command.
	// But "echo" might fail if it's not a binary.
	// On Windows, Executor handles "Run" command differently (cmd /c).
	// Let's use "Run" field instead of Tool for cross-platform echo.

	cfg.Hooks["pre-commit"] = []config.Hook{
		{Name: "hook-A", Run: "echo running A"},
		{Name: "hook-B", Run: "echo running B"},
		{Name: "hook-C", Run: "echo running C"},
		{Name: "hook-D", Run: "echo running D", After: "hook-A"},
		{Name: "hook-E", Run: "echo running E", After: "hook-B"},
		{Name: "hook-F", Run: "echo running F", After: "hook-D"},
	}

	exec := executor.New(cfg, toolMgr, workDir)
	opts := executor.Options{
		Quiet:    true, // reduce I/O impact
		FailFast: true,
	}
	exec.SetOptions(opts)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exec.Run("pre-commit", files, true)
	}
}

func BenchmarkHookRunner_PolicyEngine(b *testing.B) {
	workDir := b.TempDir()
	files, err := createDummyFiles(workDir, 50)
	if err != nil {
		b.Fatal(err)
	}

	cfg := &config.Config{
		Policies: &config.Policies{
			Type: "raw",
			LocalPolicies: []config.LocalPolicy{
				{
					Name: "benchmark-policy",
					Rules: config.PolicyRules{
						MaxFilesChanged:   1000,
						ForbidDirectories: []string{"vendor/"},
						ForbidFileContent: []config.ForbiddenContentPattern{
							{Pattern: "TODO.*FIXME", Description: "No complicated todos"},
						},
					},
				},
			},
		},
	}

	toolMgr := tool.NewManager(filepath.Join(workDir, ".cache"))
	exec := executor.New(cfg, toolMgr, workDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exec.CheckPolicies(files, "feat: benchmark commit")
	}
}
