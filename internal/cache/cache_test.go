package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	c := New("/tmp/work")
	expected := filepath.Join("/tmp/work", ".hookrunner", "cache")
	if c.dir != expected {
		t.Errorf("expected %s, got %s", expected, c.dir)
	}
}

func TestComputeHookHash(t *testing.T) {
	hash1 := ComputeHookHash("go", []string{"fmt", "./..."}, "\\.go$", "", "")
	hash2 := ComputeHookHash("go", []string{"fmt", "./..."}, "\\.go$", "", "")
	hash3 := ComputeHookHash("go", []string{"vet", "./..."}, "\\.go$", "", "")

	if hash1 != hash2 {
		t.Error("same inputs should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("different inputs should produce different hash")
	}
}

func TestIsCached_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	_ = os.WriteFile(testFile, []byte("package main"), 0644)

	cached, uncached := c.IsCached("lint", []string{testFile}, "abc123")

	if len(cached) != 0 {
		t.Errorf("expected 0 cached, got %d", len(cached))
	}
	if len(uncached) != 1 {
		t.Errorf("expected 1 uncached, got %d", len(uncached))
	}
}

func TestMarkPassed_ThenCached(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	_ = os.WriteFile(testFile, []byte("package main"), 0644)

	hookHash := ComputeHookHash("go", []string{"fmt"}, "", "", "")

	err := c.MarkPassed("format", []string{testFile}, hookHash)
	if err != nil {
		t.Fatalf("MarkPassed failed: %v", err)
	}

	cached, uncached := c.IsCached("format", []string{testFile}, hookHash)

	if len(cached) != 1 {
		t.Errorf("expected 1 cached, got %d", len(cached))
	}
	if len(uncached) != 0 {
		t.Errorf("expected 0 uncached, got %d", len(uncached))
	}
}

func TestInvalidate(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	_ = os.WriteFile(testFile, []byte("package main"), 0644)

	hookHash := "abc123"
	_ = c.MarkPassed("lint", []string{testFile}, hookHash)
	_ = c.Invalidate("lint", []string{testFile}, hookHash)

	cached, _ := c.IsCached("lint", []string{testFile}, hookHash)
	if len(cached) != 0 {
		t.Error("file should not be cached after invalidate")
	}
}

func TestClear(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	_ = os.WriteFile(testFile, []byte("package main"), 0644)

	_ = c.MarkPassed("lint", []string{testFile}, "abc")
	_ = c.Clear()

	if _, err := os.Stat(c.dir); !os.IsNotExist(err) {
		t.Error("cache dir should not exist after clear")
	}
}

func TestFileChange_InvalidatesCache(t *testing.T) {
	tmpDir := t.TempDir()
	c := New(tmpDir)

	testFile := filepath.Join(tmpDir, "test.go")
	_ = os.WriteFile(testFile, []byte("package main"), 0644)

	hookHash := "abc123"
	_ = c.MarkPassed("lint", []string{testFile}, hookHash)

	_ = os.WriteFile(testFile, []byte("package main // modified"), 0644)

	cached, uncached := c.IsCached("lint", []string{testFile}, hookHash)
	if len(cached) != 0 {
		t.Error("modified file should not be cached")
	}
	if len(uncached) != 1 {
		t.Error("modified file should be uncached")
	}
}
