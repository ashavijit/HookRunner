package lua

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewRunner(t *testing.T) {
	r := NewRunner("/tmp/work")
	if r == nil {
		t.Fatal("runner should not be nil")
	}
	if r.workDir != "/tmp/work" {
		t.Errorf("expected /tmp/work, got %s", r.workDir)
	}
}

func TestRunPolicy_PassingScript(t *testing.T) {
	tmpDir := t.TempDir()
	r := NewRunner(tmpDir)

	scriptPath := filepath.Join(tmpDir, "pass.lua")
	if err := os.WriteFile(scriptPath, []byte(`
function check(file, content)
	return true, ""
end
`), 0644); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := r.RunPolicy(scriptPath, []string{"test.go"})
	if err != nil {
		t.Fatalf("RunPolicy failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 violations, got %d", len(results))
	}
}

func TestRunPolicy_FailingScript(t *testing.T) {
	tmpDir := t.TempDir()
	r := NewRunner(tmpDir)

	scriptPath := filepath.Join(tmpDir, "fail.lua")
	if err := os.WriteFile(scriptPath, []byte(`
function check(file, content)
	return false, "test violation"
end
`), 0644); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := r.RunPolicy(scriptPath, []string{"test.go"})
	if err != nil {
		t.Fatalf("RunPolicy failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 violation, got %d", len(results))
	}
	if results[0].Message != "test violation" {
		t.Errorf("expected 'test violation', got '%s'", results[0].Message)
	}
}

func TestRunPolicy_BlockFunction(t *testing.T) {
	tmpDir := t.TempDir()
	r := NewRunner(tmpDir)

	scriptPath := filepath.Join(tmpDir, "block.lua")
	if err := os.WriteFile(scriptPath, []byte(`block("blocked message", "file.go")`), 0644); err != nil {
		t.Fatal(err)
	}

	results, err := r.RunPolicy(scriptPath, []string{})
	if err != nil {
		t.Fatalf("RunPolicy failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 violation, got %d", len(results))
	}
	if results[0].Message != "blocked message" {
		t.Errorf("expected 'blocked message', got '%s'", results[0].Message)
	}
}

func TestRunPolicy_InvalidScript(t *testing.T) {
	tmpDir := t.TempDir()
	r := NewRunner(tmpDir)

	scriptPath := filepath.Join(tmpDir, "invalid.lua")
	if err := os.WriteFile(scriptPath, []byte(`invalid lua syntax !!!`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := r.RunPolicy(scriptPath, []string{})
	if err == nil {
		t.Error("expected error for invalid script")
	}
}

func TestRunPolicy_NonExistentScript(t *testing.T) {
	r := NewRunner("/tmp")
	_, err := r.RunPolicy("/nonexistent.lua", []string{})
	if err == nil {
		t.Error("expected error for non-existent script")
	}
}

func TestRunScript_Simple(t *testing.T) {
	tmpDir := t.TempDir()
	r := NewRunner(tmpDir)

	scriptPath := filepath.Join(tmpDir, "simple.lua")
	if err := os.WriteFile(scriptPath, []byte(`local x = 1 + 1`), 0644); err != nil {
		t.Fatal(err)
	}

	err := r.RunScript(scriptPath)
	if err != nil {
		t.Errorf("RunScript failed: %v", err)
	}
}

func TestRunScript_Invalid(t *testing.T) {
	tmpDir := t.TempDir()
	r := NewRunner(tmpDir)

	scriptPath := filepath.Join(tmpDir, "bad.lua")
	if err := os.WriteFile(scriptPath, []byte(`syntax error !!!`), 0644); err != nil {
		t.Fatal(err)
	}

	err := r.RunScript(scriptPath)
	if err == nil {
		t.Error("expected error for invalid script")
	}
}
