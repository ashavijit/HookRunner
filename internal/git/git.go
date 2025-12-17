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

type BlameInfo struct {
	Author  string
	Email   string
	Date    string
	Commit  string
	Summary string
	Line    int
	Content string
}

func GetBlame(file string, line int) (*BlameInfo, error) {
	lineRange := fmt.Sprintf("%d,%d", line, line)
	cmd := exec.Command("git", "blame", "-L", lineRange, "--porcelain", file)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	info := &BlameInfo{Line: line}
	lines := strings.Split(string(out), "\n")

	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "author "):
			info.Author = strings.TrimPrefix(l, "author ")
		case strings.HasPrefix(l, "author-mail "):
			info.Email = strings.Trim(strings.TrimPrefix(l, "author-mail "), "<>")
		case strings.HasPrefix(l, "author-time "):
			info.Date = strings.TrimPrefix(l, "author-time ")
		case strings.HasPrefix(l, "summary "):
			info.Summary = strings.TrimPrefix(l, "summary ")
		case len(l) == 40:
			info.Commit = l[:8]
		case strings.HasPrefix(l, "\t"):
			info.Content = strings.TrimPrefix(l, "\t")
		}
	}

	return info, nil
}

func FormatBlame(info *BlameInfo) string {
	if info == nil {
		return ""
	}
	return fmt.Sprintf("Blame: %s <%s> commit %s: \"%s\"",
		info.Author, info.Email, info.Commit, info.Summary)
}
