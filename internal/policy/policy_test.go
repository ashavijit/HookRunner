package policy

import (
	"testing"
)

func TestEvaluate_MaxFilesChanged(t *testing.T) {
	policies := &Policies{
		MaxFilesChanged: 3,
	}

	files := []string{"a.go", "b.go", "c.go", "d.go"}
	result := Evaluate(policies, files, "")

	if result.Passed {
		t.Error("expected policy to fail for too many files")
	}

	if len(result.Violations) != 1 {
		t.Errorf("expected 1 violation, got %d", len(result.Violations))
	}

	if result.Violations[0].Rule != "max_files_changed" {
		t.Errorf("expected rule max_files_changed, got %s", result.Violations[0].Rule)
	}
}

func TestEvaluate_MaxFilesChanged_Pass(t *testing.T) {
	policies := &Policies{
		MaxFilesChanged: 5,
	}

	files := []string{"a.go", "b.go"}
	result := Evaluate(policies, files, "")

	if !result.Passed {
		t.Error("expected policy to pass")
	}
}

func TestEvaluate_ForbidDirectories(t *testing.T) {
	policies := &Policies{
		ForbidDirectories: []string{"vendor/"},
	}

	files := []string{"main.go", "vendor/lib.go"}
	result := Evaluate(policies, files, "")

	if result.Passed {
		t.Error("expected policy to fail for forbidden directory")
	}

	if result.Violations[0].Rule != "forbid_directories" {
		t.Errorf("expected rule forbid_directories, got %s", result.Violations[0].Rule)
	}
}

func TestEvaluate_ForbidDirectories_Pass(t *testing.T) {
	policies := &Policies{
		ForbidDirectories: []string{"vendor/"},
	}

	files := []string{"main.go", "lib/util.go"}
	result := Evaluate(policies, files, "")

	if !result.Passed {
		t.Error("expected policy to pass")
	}
}

func TestEvaluate_CommitMessageRegex(t *testing.T) {
	policies := &Policies{
		CommitMessage: CommitMessagePolicy{
			Regex: "^(feat|fix|chore):",
		},
	}

	result := Evaluate(policies, nil, "invalid commit message")

	if result.Passed {
		t.Error("expected policy to fail for invalid commit message")
	}

	if result.Violations[0].Rule != "commit_message.regex" {
		t.Errorf("expected rule commit_message.regex, got %s", result.Violations[0].Rule)
	}
}

func TestEvaluate_CommitMessageRegex_Pass(t *testing.T) {
	policies := &Policies{
		CommitMessage: CommitMessagePolicy{
			Regex: "^(feat|fix|chore):",
		},
	}

	result := Evaluate(policies, nil, "feat: add new feature")

	if !result.Passed {
		t.Error("expected policy to pass")
	}
}

func TestEvaluate_CommitMessageMinLength(t *testing.T) {
	policies := &Policies{
		CommitMessage: CommitMessagePolicy{
			MinLength: 20,
		},
	}

	result := Evaluate(policies, nil, "short msg")

	if result.Passed {
		t.Error("expected policy to fail for short message")
	}

	if result.Violations[0].Rule != "commit_message.min_length" {
		t.Errorf("expected rule commit_message.min_length, got %s", result.Violations[0].Rule)
	}
}

func TestEvaluate_CommitMessageMaxLength(t *testing.T) {
	policies := &Policies{
		CommitMessage: CommitMessagePolicy{
			MaxLength: 10,
		},
	}

	result := Evaluate(policies, nil, "this is a very long commit message")

	if result.Passed {
		t.Error("expected policy to fail for long message")
	}

	if result.Violations[0].Rule != "commit_message.max_length" {
		t.Errorf("expected rule commit_message.max_length, got %s", result.Violations[0].Rule)
	}
}

func TestEvaluate_NilPolicies(t *testing.T) {
	result := Evaluate(nil, []string{"a.go"}, "msg")

	if !result.Passed {
		t.Error("expected nil policies to pass")
	}
}

func TestEvaluate_MultiplePolicies(t *testing.T) {
	policies := &Policies{
		MaxFilesChanged:   2,
		ForbidDirectories: []string{"vendor/"},
		CommitMessage: CommitMessagePolicy{
			Regex: "^feat:",
		},
	}

	files := []string{"a.go", "b.go", "c.go", "vendor/d.go"}
	result := Evaluate(policies, files, "bad message")

	if result.Passed {
		t.Error("expected policy to fail")
	}

	if len(result.Violations) != 3 {
		t.Errorf("expected 3 violations, got %d", len(result.Violations))
	}
}

func TestPolicyResult_String(t *testing.T) {
	result := PolicyResult{Passed: true}
	if result.String() != "All policies passed" {
		t.Error("expected 'All policies passed'")
	}

	result = PolicyResult{
		Passed: false,
		Violations: []Violation{
			{Rule: "test", Message: "test message"},
		},
	}
	str := result.String()
	if str == "" {
		t.Error("expected non-empty string for violations")
	}
}
