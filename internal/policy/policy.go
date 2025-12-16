package policy

import (
	"fmt"
	"regexp"
	"strings"
)

type CommitMessagePolicy struct {
	Regex     string `yaml:"regex" json:"regex"`
	MaxLength int    `yaml:"max_length" json:"max_length"`
	MinLength int    `yaml:"min_length" json:"min_length"`
}

type Policies struct {
	MaxFilesChanged   int                 `yaml:"max_files_changed" json:"max_files_changed"`
	ForbidDirectories []string            `yaml:"forbid_directories" json:"forbid_directories"`
	ForbidFiles       []string            `yaml:"forbid_files" json:"forbid_files"`
	RequireFiles      []string            `yaml:"require_files" json:"require_files"`
	CommitMessage     CommitMessagePolicy `yaml:"commit_message" json:"commit_message"`
}

type Violation struct {
	Rule    string
	Message string
}

type PolicyResult struct {
	Passed     bool
	Violations []Violation
}

func Evaluate(policies *Policies, files []string, commitMsg string) PolicyResult {
	result := PolicyResult{Passed: true}

	if policies == nil {
		return result
	}

	if policies.MaxFilesChanged > 0 && len(files) > policies.MaxFilesChanged {
		result.Violations = append(result.Violations, Violation{
			Rule:    "max_files_changed",
			Message: fmt.Sprintf("too many files changed: %d (max: %d)", len(files), policies.MaxFilesChanged),
		})
	}

	for _, dir := range policies.ForbidDirectories {
		for _, file := range files {
			if strings.HasPrefix(file, dir) || strings.Contains(file, "/"+dir) {
				result.Violations = append(result.Violations, Violation{
					Rule:    "forbid_directories",
					Message: fmt.Sprintf("forbidden directory modified: %s (file: %s)", dir, file),
				})
				break
			}
		}
	}

	for _, pattern := range policies.ForbidFiles {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}
		for _, file := range files {
			if re.MatchString(file) {
				result.Violations = append(result.Violations, Violation{
					Rule:    "forbid_files",
					Message: fmt.Sprintf("forbidden file modified: %s", file),
				})
			}
		}
	}

	for _, required := range policies.RequireFiles {
		found := false
		for _, file := range files {
			if strings.HasSuffix(file, required) || file == required {
				found = true
				break
			}
		}
		if !found {
			result.Violations = append(result.Violations, Violation{
				Rule:    "require_files",
				Message: fmt.Sprintf("required file not found in commit: %s", required),
			})
		}
	}

	if commitMsg != "" {
		if policies.CommitMessage.MinLength > 0 && len(commitMsg) < policies.CommitMessage.MinLength {
			result.Violations = append(result.Violations, Violation{
				Rule:    "commit_message.min_length",
				Message: fmt.Sprintf("commit message too short: %d chars (min: %d)", len(commitMsg), policies.CommitMessage.MinLength),
			})
		}

		if policies.CommitMessage.MaxLength > 0 && len(commitMsg) > policies.CommitMessage.MaxLength {
			result.Violations = append(result.Violations, Violation{
				Rule:    "commit_message.max_length",
				Message: fmt.Sprintf("commit message too long: %d chars (max: %d)", len(commitMsg), policies.CommitMessage.MaxLength),
			})
		}

		if policies.CommitMessage.Regex != "" {
			re, err := regexp.Compile(policies.CommitMessage.Regex)
			if err == nil && !re.MatchString(commitMsg) {
				result.Violations = append(result.Violations, Violation{
					Rule:    "commit_message.regex",
					Message: fmt.Sprintf("commit message does not match pattern: %s", policies.CommitMessage.Regex),
				})
			}
		}
	}

	result.Passed = len(result.Violations) == 0
	return result
}

func (r PolicyResult) String() string {
	if r.Passed {
		return "All policies passed"
	}

	var sb strings.Builder
	sb.WriteString("Policy violations:\n")
	for _, v := range r.Violations {
		sb.WriteString(fmt.Sprintf("  - [%s] %s\n", v.Rule, v.Message))
	}
	return sb.String()
}
