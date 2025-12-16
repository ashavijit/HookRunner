package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepoRoot(t *testing.T) {
	root, err := FindRepoRoot()
	if err != nil {
		t.Skipf("not in a git repository: %v", err)
	}

	if root == "" {
		t.Error("repo root should not be empty")
	}

	if !filepath.IsAbs(root) {
		t.Error("repo root should be absolute path")
	}
}

func TestGetStagedFiles(t *testing.T) {
	if !IsInsideWorkTree() {
		t.Skip("not in a git repository")
	}

	files, err := GetStagedFiles()
	if err != nil {
		t.Fatalf("failed to get staged files: %v", err)
	}

	if files == nil {
		files = []string{}
	}

	for _, f := range files {
		if f == "" {
			t.Error("file should not be empty string")
		}
	}
}

func TestGetAllFiles(t *testing.T) {
	if !IsInsideWorkTree() {
		t.Skip("not in a git repository")
	}

	files, err := GetAllFiles()
	if err != nil {
		t.Fatalf("failed to get files: %v", err)
	}

	if len(files) == 0 {
		t.Error("expected at least one file in repo")
	}

	for _, f := range files {
		if f == "" {
			t.Error("file should not be empty string")
		}
	}
}

func TestIsInsideWorkTree(t *testing.T) {
	result := IsInsideWorkTree()
	t.Logf("IsInsideWorkTree: %v", result)
}

func TestInstallHook(t *testing.T) {
	if !IsInsideWorkTree() {
		t.Skip("not in a git repository")
	}

	err := InstallHook("pre-commit", "/path/to/binary")
	if err != nil {
		t.Fatalf("failed to install hook: %v", err)
	}

	root, _ := FindRepoRoot()
	hookPath := filepath.Join(root, ".git", "hooks", "pre-commit")

	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		t.Error("hook file should exist")
	}

	content, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("failed to read hook: %v", err)
	}

	if len(content) == 0 {
		t.Error("hook content should not be empty")
	}
}

func TestUninstallHook(t *testing.T) {
	if !IsInsideWorkTree() {
		t.Skip("not in a git repository")
	}

	err := InstallHook("test-hook", "/path/to/binary")
	if err != nil {
		t.Fatalf("failed to install hook: %v", err)
	}

	err = UninstallHook("test-hook")
	if err != nil {
		t.Fatalf("failed to uninstall hook: %v", err)
	}

	root, _ := FindRepoRoot()
	hookPath := filepath.Join(root, ".git", "hooks", "test-hook")

	if _, err := os.Stat(hookPath); !os.IsNotExist(err) {
		t.Error("hook file should not exist after uninstall")
	}
}

func TestUninstallHook_NotExists(t *testing.T) {
	if !IsInsideWorkTree() {
		t.Skip("not in a git repository")
	}

	err := UninstallHook("nonexistent-hook")
	if err != nil {
		t.Errorf("uninstalling nonexistent hook should not error: %v", err)
	}
}
