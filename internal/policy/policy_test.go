package policy

import (
	"testing"
)

func TestPolicyRules_Merge(t *testing.T) {
	base := PolicyRules{
		MaxFilesChanged:   10,
		ForbidDirectories: []string{"vendor/"},
	}
	overlay := PolicyRules{
		MaxFilesChanged:   20,
		ForbidDirectories: []string{"generated/"},
	}

	result := base.Merge(overlay)

	if result.MaxFilesChanged != 20 {
		t.Errorf("expected 20, got %d", result.MaxFilesChanged)
	}
	if len(result.ForbidDirectories) != 2 {
		t.Errorf("expected 2 dirs, got %d", len(result.ForbidDirectories))
	}
}

func TestPolicyRules_MergeCommitMessage(t *testing.T) {
	base := PolicyRules{
		CommitMessage: &CommitMessageRule{
			Regex: "^feat:",
		},
	}
	overlay := PolicyRules{
		CommitMessage: &CommitMessageRule{
			Error: "commit must start with feat:",
		},
	}

	result := base.Merge(overlay)

	if result.CommitMessage == nil {
		t.Fatal("commit message should not be nil")
	}
	if result.CommitMessage.Regex != "^feat:" {
		t.Error("regex should be preserved")
	}
	if result.CommitMessage.Error != "commit must start with feat:" {
		t.Error("error should be set from overlay")
	}
}

func TestPolicyRules_MergeNoDuplicates(t *testing.T) {
	base := PolicyRules{
		ForbidFiles: []string{"\\.env$"},
	}
	overlay := PolicyRules{
		ForbidFiles: []string{"\\.env$", "\\.secret$"},
	}

	result := base.Merge(overlay)

	if len(result.ForbidFiles) != 2 {
		t.Errorf("expected 2 unique files, got %d: %v", len(result.ForbidFiles), result.ForbidFiles)
	}
}

func TestRemotePolicy_Identifier(t *testing.T) {
	tests := []struct {
		name   string
		policy RemotePolicy
		want   string
	}{
		{"with version", RemotePolicy{Name: "security", Version: "1.0"}, "security@1.0"},
		{"no version", RemotePolicy{Name: "basic"}, "basic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.policy.Identifier(); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestEvaluate_MaxFilesChanged(t *testing.T) {
	rules := &PolicyRules{MaxFilesChanged: 3}
	files := []string{"a.go", "b.go", "c.go", "d.go"}

	result := Evaluate(rules, files, "")

	if result.Passed {
		t.Error("expected failure for too many files")
	}
	if len(result.Violations) != 1 {
		t.Errorf("expected 1 violation, got %d", len(result.Violations))
	}
}

func TestEvaluate_ForbidDirectories(t *testing.T) {
	rules := &PolicyRules{ForbidDirectories: []string{"vendor/"}}
	files := []string{"main.go", "vendor/lib.go"}

	result := Evaluate(rules, files, "")

	if result.Passed {
		t.Error("expected failure for forbidden directory")
	}
}

func TestEvaluate_CommitMessageRegex(t *testing.T) {
	rules := &PolicyRules{
		CommitMessage: &CommitMessageRule{Regex: "^(feat|fix):"},
	}

	result := Evaluate(rules, nil, "invalid message")
	if result.Passed {
		t.Error("expected failure for invalid commit message")
	}

	result = Evaluate(rules, nil, "feat: add feature")
	if !result.Passed {
		t.Error("expected pass for valid commit message")
	}
}

func TestEvaluate_NilRules(t *testing.T) {
	result := Evaluate(nil, []string{"a.go"}, "msg")
	if !result.Passed {
		t.Error("nil rules should pass")
	}
}

func TestValidatePolicy(t *testing.T) {
	valid := &RemotePolicy{Name: "test"}
	if err := ValidatePolicy(valid); err != nil {
		t.Errorf("valid policy should pass: %v", err)
	}

	invalid := &RemotePolicy{}
	if err := ValidatePolicy(invalid); err == nil {
		t.Error("invalid policy should fail")
	}
}

func TestParseRemotePolicy(t *testing.T) {
	yaml := []byte("name: test\nversion: \"1.0\"\nrules:\n  max_files_changed: 10\n")
	policy, err := ParseRemotePolicy(yaml)
	if err != nil {
		t.Fatalf("ParseRemotePolicy failed: %v", err)
	}
	if policy.Name != "test" {
		t.Errorf("got name %s, want test", policy.Name)
	}
	if policy.Rules.MaxFilesChanged != 10 {
		t.Errorf("got max_files_changed %d, want 10", policy.Rules.MaxFilesChanged)
	}
}
