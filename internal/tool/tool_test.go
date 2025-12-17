package tool

import (
	"path/filepath"
	"testing"

	"github.com/ashavijit/hookrunner/internal/config"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager("/tmp/cache")

	if mgr == nil {
		t.Fatal("manager should not be nil")
	}

	if mgr.CacheDir != "/tmp/cache" {
		t.Errorf("expected cache dir /tmp/cache, got %s", mgr.CacheDir)
	}
}

func TestEnsureTool_SystemTool(t *testing.T) {
	mgr := NewManager(t.TempDir())

	path, err := mgr.EnsureTool("go", nil)
	if err != nil {
		t.Skipf("go not found: %v", err)
	}

	if path == "" {
		t.Error("path should not be empty")
	}
}

func TestEnsureTool_NotFound(t *testing.T) {
	mgr := NewManager(t.TempDir())

	_, err := mgr.EnsureTool("nonexistent-tool-12345", nil)
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}

func TestEnsureTool_WithConfig_NoURL(t *testing.T) {
	mgr := NewManager(t.TempDir())

	tool := &config.Tool{
		Version: "1.0.0",
		Install: map[string]string{},
	}

	_, err := mgr.EnsureTool("test-tool", tool)
	if err == nil {
		t.Error("expected error for missing URL")
	}
}

func TestGetCachedPath(t *testing.T) {
	mgr := NewManager("/cache")

	path := mgr.getCachedPath("tool", "1.0.0")

	if path == "" {
		t.Error("cached path should not be empty")
	}

	expectedDir := filepath.Join("/cache", "tool-1.0.0")
	// Use strings.HasPrefix instead of deprecated filepath.HasPrefix
	if len(path) < len(expectedDir) || path[:len(expectedDir)] != expectedDir {
		t.Errorf("expected path to start with %s, got %s", expectedDir, path)
	}
}

func TestFindSystemTool(t *testing.T) {
	mgr := NewManager(t.TempDir())

	path, err := mgr.findSystemTool("go")
	if err != nil {
		t.Skipf("go not found: %v", err)
	}

	if path == "" {
		t.Error("path should not be empty")
	}
}

func TestFindSystemTool_NotFound(t *testing.T) {
	mgr := NewManager(t.TempDir())

	_, err := mgr.findSystemTool("nonexistent-tool-xyz-123")
	if err == nil {
		t.Error("expected error for nonexistent tool")
	}
}

func TestExtractTarGz_InvalidFile(t *testing.T) {
	mgr := NewManager(t.TempDir())

	err := mgr.extractTarGz("/nonexistent/file.tar.gz", t.TempDir(), "tool")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestExtractZip_InvalidFile(t *testing.T) {
	mgr := NewManager(t.TempDir())

	err := mgr.extractZip("/nonexistent/file.zip", t.TempDir(), "tool")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
