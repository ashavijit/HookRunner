package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Violation struct {
	Rule    string
	Message string
}

type EvalResult struct {
	Passed     bool
	Violations []Violation
}

func Evaluate(rules *PolicyRules, files []string, commitMsg string) EvalResult {
	result := EvalResult{Passed: true}

	if rules == nil {
		return result
	}

	if rules.MaxFilesChanged > 0 && len(files) > rules.MaxFilesChanged {
		result.Violations = append(result.Violations, Violation{
			Rule:    "max_files_changed",
			Message: fmt.Sprintf("too many files: %d (max: %d)", len(files), rules.MaxFilesChanged),
		})
	}

	for _, dir := range rules.ForbidDirectories {
		for _, file := range files {
			if strings.HasPrefix(file, dir) || strings.Contains(file, "/"+dir) || strings.Contains(file, "\\"+dir) {
				result.Violations = append(result.Violations, Violation{
					Rule:    "forbid_directories",
					Message: fmt.Sprintf("forbidden directory: %s (file: %s)", dir, file),
				})
				break
			}
		}
	}

	for _, pattern := range rules.ForbidFiles {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		for _, file := range files {
			if re.MatchString(file) {
				result.Violations = append(result.Violations, Violation{
					Rule:    "forbid_files",
					Message: fmt.Sprintf("forbidden file: %s", file),
				})
			}
		}
	}

	for _, ext := range rules.ForbidFileExtensions {
		for _, file := range files {
			if strings.HasSuffix(file, ext) || strings.HasSuffix(file, "."+ext) {
				result.Violations = append(result.Violations, Violation{
					Rule:    "forbid_file_extensions",
					Message: fmt.Sprintf("forbidden extension %s: %s", ext, file),
				})
			}
		}
	}

	for _, required := range rules.RequiredFiles {
		found := false
		for _, file := range files {
			if strings.HasSuffix(file, required) || file == required {
				found = true
				break
			}
		}
		if !found {
			result.Violations = append(result.Violations, Violation{
				Rule:    "required_files",
				Message: fmt.Sprintf("required file not found: %s", required),
			})
		}
	}

	if rules.MaxFileSizeKB > 0 {
		maxBytes := int64(rules.MaxFileSizeKB * 1024)
		for _, file := range files {
			if info, err := os.Stat(file); err == nil {
				if info.Size() > maxBytes {
					result.Violations = append(result.Violations, Violation{
						Rule:    "max_file_size_kb",
						Message: fmt.Sprintf("file too large: %s (%d KB, max: %d KB)", file, info.Size()/1024, rules.MaxFileSizeKB),
					})
				}
			}
		}
	}

	for _, pattern := range rules.ForbidFileContent {
		re, err := regexp.Compile(pattern.Pattern)
		if err != nil {
			continue
		}
		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			if re.Match(content) {
				desc := pattern.Description
				if desc == "" {
					desc = pattern.Pattern
				}
				result.Violations = append(result.Violations, Violation{
					Rule:    "forbid_file_content",
					Message: fmt.Sprintf("forbidden content in %s: %s", filepath.Base(file), desc),
				})
			}
		}
	}

	// Check regex_block patterns (from remote policies)
	for _, pattern := range rules.RegexBlock {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		for _, file := range files {
			content, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			if re.Match(content) {
				desc := describeSecretPattern(pattern)
				result.Violations = append(result.Violations, Violation{
					Rule:    "secret_detected",
					Message: fmt.Sprintf("%s in %s - remove before committing", desc, filepath.Base(file)),
				})
			}
		}
	}

	if commitMsg != "" && rules.CommitMessage != nil {
		cm := rules.CommitMessage
		if cm.Regex != "" {
			re, err := regexp.Compile(cm.Regex)
			if err == nil && !re.MatchString(commitMsg) {
				errMsg := cm.Error
				if errMsg == "" {
					errMsg = fmt.Sprintf("does not match: %s", cm.Regex)
				}
				result.Violations = append(result.Violations, Violation{
					Rule:    "commit_message",
					Message: errMsg,
				})
			}
		}
	}

	result.Passed = len(result.Violations) == 0
	return result
}

func (r EvalResult) String() string {
	if r.Passed {
		return "All policies passed"
	}

	var sb strings.Builder
	sb.WriteString("Policy violations:\n")
	for _, v := range r.Violations {
		sb.WriteString(fmt.Sprintf("  ✗ [%s] %s\n", v.Rule, v.Message))
	}
	return sb.String()
}

// describeSecretPattern returns a user-friendly description for common secret patterns
func describeSecretPattern(pattern string) string {
	descriptions := map[string]string{
		"AKIA[0-9A-Z]{16}":            "AWS Access Key",
		"-----BEGIN PRIVATE KEY-----": "Private Key",
		"-----BEGIN RSA PRIVATE KEY":  "RSA Private Key",
		"(?i)password=":               "Hardcoded Password",
		"ghp_[A-Za-z0-9_]{36}":        "GitHub Personal Access Token",
		"gho_[A-Za-z0-9_]{36}":        "GitHub OAuth Token",
		"github_pat_[A-Za-z0-9_]{22}": "GitHub PAT",
		"sk-[A-Za-z0-9]{48}":          "OpenAI API Key",
		"xox[baprs]-[A-Za-z0-9-]+":    "Slack Token",
		"(?i)api[_-]?key":             "API Key",
		"(?i)secret[_-]?key":          "Secret Key",
	}

	if desc, ok := descriptions[pattern]; ok {
		return desc + " detected"
	}

	// For unknown patterns, provide a generic but clear message
	return "Potential secret/credential detected"
}
