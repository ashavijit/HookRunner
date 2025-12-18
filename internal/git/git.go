package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func FindRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func GetStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only", "--diff-filter=ACMR")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func GetAllFiles() ([]string, error) {
	cmd := exec.Command("git", "ls-files")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func InstallHook(hookType string, binaryPath string) error {
	repoRoot, err := FindRepoRoot()
	if err != nil {
		return err
	}

	hooksDir := filepath.Join(repoRoot, ".git", "hooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}

	hookPath := filepath.Join(hooksDir, hookType)
	binaryPath = strings.ReplaceAll(binaryPath, "\\", "/")
	content := fmt.Sprintf(`#!/bin/sh
exec "%s" run %s
`, binaryPath, hookType)

	//nolint:gosec // G306: Hook script must be executable (0755)
	if err := os.WriteFile(hookPath, []byte(content), 0755); err != nil {
		return fmt.Errorf("failed to write hook: %w", err)
	}

	return nil
}

func UninstallHook(hookType string) error {
	repoRoot, err := FindRepoRoot()
	if err != nil {
		return err
	}

	hookPath := filepath.Join(repoRoot, ".git", "hooks", hookType)
	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		return nil
	}

	if err := os.Remove(hookPath); err != nil {
		return fmt.Errorf("failed to remove hook: %w", err)
	}

	return nil
}

func IsInsideWorkTree() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}


func CreateCleanRoom() (string, error) {
	repoRoot, err := FindRepoRoot()
	if err != nil {
		return "", err
	}

	tempDir, err := os.MkdirTemp("", "hookrunner-cleanroom-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	cmd := exec.Command("git", "checkout-index", "--all", "--prefix="+tempDir+"/")
	cmd.Dir = repoRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to extract staged files: %w\n%s", err, string(output))
	}

	return tempDir, nil
}

func CleanupCleanRoom(path string) error {
	if path == "" {
		return nil
	}
	return os.RemoveAll(path)
}
