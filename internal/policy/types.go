package policy

type PolicyRef struct {
	URL string `json:"url" yaml:"url"`
}

type UserConfig struct {
	Type          string        `json:"type" yaml:"type"`
	Policies      []PolicyRef   `json:"policies" yaml:"policies"`
	LocalPolicies []LocalPolicy `json:"localPolicies" yaml:"localPolicies"`
}

type LocalPolicy struct {
	Name        string            `json:"name" yaml:"name"`
	Version     string            `json:"version" yaml:"version"`
	Description string            `json:"description" yaml:"description"`
	Rules       PolicyRules       `json:"rules" yaml:"rules"`
	Metadata    map[string]string `json:"metadata" yaml:"metadata"`
}

type RemotePolicy struct {
	Name        string            `json:"name" yaml:"name"`
	Version     string            `json:"version" yaml:"version"`
	Description string            `json:"description" yaml:"description"`
	Rules       PolicyRules       `json:"rules" yaml:"rules"`
	Metadata    map[string]string `json:"metadata" yaml:"metadata"`
}

type ForbiddenContentPattern struct {
	Pattern     string `json:"pattern" yaml:"pattern"`
	Description string `json:"description" yaml:"description"`
}

type CommitMessageRule struct {
	Regex string `json:"regex" yaml:"regex"`
	Error string `json:"error" yaml:"error"`
}

type PolicyRules struct {
	ForbidFiles          []string                  `json:"forbid_files" yaml:"forbid_files"`
	ForbidDirectories    []string                  `json:"forbid_directories" yaml:"forbid_directories"`
	ForbidFileExtensions []string                  `json:"forbid_file_extensions" yaml:"forbid_file_extensions"`
	RequiredFiles        []string                  `json:"required_files" yaml:"required_files"`
	MaxFileSizeKB        int                       `json:"max_file_size_kb" yaml:"max_file_size_kb"`
	MaxFilesChanged      int                       `json:"max_files_changed" yaml:"max_files_changed"`
	ForbidFileContent    []ForbiddenContentPattern `json:"forbid_file_content" yaml:"forbid_file_content"`
	CommitMessage        *CommitMessageRule        `json:"commit_message" yaml:"commit_message"`
	EnforceHooks         []string                  `json:"enforce_hooks" yaml:"enforce_hooks"`
	HookTimeBudgetMs     map[string]int            `json:"hook_time_budget_ms" yaml:"hook_time_budget_ms"`
	MaxParallelHooks     int                       `json:"max_parallel_hooks" yaml:"max_parallel_hooks"`
}

type MergedPolicy struct {
	RemotePolicies []RemotePolicy
	LocalPolicies  []LocalPolicy
	EffectiveRules PolicyRules
}

func (p *RemotePolicy) Identifier() string {
	if p.Version != "" {
		return p.Name + "@" + p.Version
	}
	return p.Name
}

func (p *LocalPolicy) Identifier() string {
	if p.Version != "" {
		return p.Name + "@" + p.Version
	}
	return p.Name
}

func (r *PolicyRules) Merge(other PolicyRules) PolicyRules {
	result := *r

	if other.MaxFilesChanged > 0 {
		result.MaxFilesChanged = other.MaxFilesChanged
	}
	if other.MaxFileSizeKB > 0 {
		result.MaxFileSizeKB = other.MaxFileSizeKB
	}
	if other.MaxParallelHooks > 0 {
		result.MaxParallelHooks = other.MaxParallelHooks
	}

	result.ForbidDirectories = appendUnique(result.ForbidDirectories, other.ForbidDirectories)
	result.ForbidFiles = appendUnique(result.ForbidFiles, other.ForbidFiles)
	result.ForbidFileExtensions = appendUnique(result.ForbidFileExtensions, other.ForbidFileExtensions)
	result.RequiredFiles = appendUnique(result.RequiredFiles, other.RequiredFiles)
	result.EnforceHooks = appendUnique(result.EnforceHooks, other.EnforceHooks)

	for _, pattern := range other.ForbidFileContent {
		result.ForbidFileContent = append(result.ForbidFileContent, pattern)
	}

	if other.CommitMessage != nil {
		if result.CommitMessage == nil {
			result.CommitMessage = &CommitMessageRule{}
		}
		if other.CommitMessage.Regex != "" {
			result.CommitMessage.Regex = other.CommitMessage.Regex
		}
		if other.CommitMessage.Error != "" {
			result.CommitMessage.Error = other.CommitMessage.Error
		}
	}

	if other.HookTimeBudgetMs != nil {
		if result.HookTimeBudgetMs == nil {
			result.HookTimeBudgetMs = make(map[string]int)
		}
		for k, v := range other.HookTimeBudgetMs {
			result.HookTimeBudgetMs[k] = v
		}
	}

	return result
}

func appendUnique(base, items []string) []string {
	seen := make(map[string]bool)
	for _, s := range base {
		seen[s] = true
	}
	result := base
	for _, s := range items {
		if !seen[s] {
			result = append(result, s)
			seen[s] = true
		}
	}
	return result
}
