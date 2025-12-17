package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ashavijit/hookrunner/internal/config"
	"github.com/ashavijit/hookrunner/internal/dag"
	"github.com/ashavijit/hookrunner/internal/policy"
	"github.com/ashavijit/hookrunner/internal/tool"
	"github.com/fatih/color"
)

type Result struct {
	Name     string
	Success  bool
	Skipped  bool
	Duration time.Duration
	Output   string
	Error    error
}

type Options struct {
	Verbose   bool
	Quiet     bool
	Fix       bool
	FailFast  bool
	SkipHooks []string
	CommitMsg string
}

type Executor struct {
	toolMgr *tool.Manager
	config  *config.Config
	workDir string
	opts    Options
}

func New(cfg *config.Config, toolMgr *tool.Manager, workDir string) *Executor {
	return &Executor{
		toolMgr: toolMgr,
		config:  cfg,
		workDir: workDir,
		opts:    Options{FailFast: true},
	}
}

func (e *Executor) SetOptions(opts Options) {
	e.opts = opts
}

func (e *Executor) Run(hookType string, files []string, allFiles bool) []Result {
	hooks := e.config.GetHooks(hookType)
	if len(hooks) == 0 {
		return nil
	}

	graph := dag.BuildGraph(hooks)
	if graph.HasCycle() {
		return []Result{{
			Name:    "dag",
			Success: false,
			Error:   fmt.Errorf("circular dependency detected in hook graph"),
		}}
	}

	executionPlan := graph.GetExecutionPlan()
	var results []Result
	failed := false

	for _, batch := range executionPlan {
		if failed && e.opts.FailFast {
			break
		}
		batchResults := e.runBatch(batch, files, allFiles)
		results = append(results, batchResults...)

		for _, r := range batchResults {
			if !r.Success && !r.Skipped {
				failed = true
				if e.opts.FailFast {
					break
				}
			}
		}
	}

	return results
}

func (e *Executor) CheckPolicies(files []string, commitMsg string) *policy.EvalResult {
	if e.config.Policies == nil {
		return nil
	}

	cacheDir := filepath.Join(e.workDir, ".hooks", "cache")
	registry := policy.NewRegistry(e.workDir, cacheDir)

	p := e.config.Policies
	userCfg := &policy.UserConfig{
		Type: p.Type,
	}

	for _, ref := range p.Policies {
		userCfg.Policies = append(userCfg.Policies, policy.PolicyRef{URL: ref.URL})
	}

	for _, lp := range p.LocalPolicies {
		userCfg.LocalPolicies = append(userCfg.LocalPolicies, policy.LocalPolicy{
			Name:        lp.Name,
			Version:     lp.Version,
			Description: lp.Description,
			Metadata:    lp.Metadata,
			Rules:       convertRules(lp.Rules),
		})
	}

	merged, err := registry.Load(userCfg)
	if err != nil {
		return &policy.EvalResult{
			Passed:     false,
			Violations: []policy.Violation{{Rule: "load", Message: err.Error()}},
		}
	}

	if merged == nil {
		return nil
	}

	result := policy.Evaluate(&merged.EffectiveRules, files, commitMsg)
	return &result
}

func convertRules(r config.PolicyRules) policy.PolicyRules {
	var cm *policy.CommitMessageRule
	if r.CommitMessage != nil {
		cm = &policy.CommitMessageRule{
			Regex: r.CommitMessage.Regex,
			Error: r.CommitMessage.Error,
		}
	}

	patterns := make([]policy.ForbiddenContentPattern, 0, len(r.ForbidFileContent))
	for _, p := range r.ForbidFileContent {
		patterns = append(patterns, policy.ForbiddenContentPattern{
			Pattern:     p.Pattern,
			Description: p.Description,
		})
	}

	return policy.PolicyRules{
		ForbidFiles:          r.ForbidFiles,
		ForbidDirectories:    r.ForbidDirectories,
		ForbidFileExtensions: r.ForbidFileExtensions,
		RequiredFiles:        r.RequiredFiles,
		MaxFileSizeKB:        r.MaxFileSizeKB,
		MaxFilesChanged:      r.MaxFilesChanged,
		ForbidFileContent:    patterns,
		CommitMessage:        cm,
		EnforceHooks:         r.EnforceHooks,
		HookTimeBudgetMs:     r.HookTimeBudgetMs,
		MaxParallelHooks:     r.MaxParallelHooks,
	}
}

func (e *Executor) runBatch(hooks []config.Hook, files []string, allFiles bool) []Result {
	var wg sync.WaitGroup
	results := make([]Result, len(hooks))

	for i, hook := range hooks {
		wg.Add(1)
		go func(idx int, h config.Hook) {
			defer wg.Done()
			results[idx] = e.runHook(h, files, allFiles)
		}(i, hook)
	}

	wg.Wait()
	return results
}

func (e *Executor) shouldSkip(hook config.Hook) (bool, string) {
	for _, skip := range e.opts.SkipHooks {
		if skip == hook.Name {
			return true, "SKIP env"
		}
	}

	if hook.Skip != "" {
		if val := os.Getenv(hook.Skip); val != "" && val != "0" && val != "false" {
			return true, "skip condition"
		}
	}

	if hook.Only != "" {
		if val := os.Getenv(hook.Only); val == "" || val == "0" || val == "false" {
			return true, "only condition"
		}
	}

	return false, ""
}

func (e *Executor) runHook(hook config.Hook, files []string, allFiles bool) Result {
	start := time.Now()
	result := Result{Name: hook.Name}

	if skip, reason := e.shouldSkip(hook); skip {
		result.Skipped = true
		result.Success = true
		result.Duration = time.Since(start)
		result.Output = fmt.Sprintf("skipped (%s)", reason)
		return result
	}

	if !allFiles {
		matched := e.filterFiles(files, hook)
		if len(matched) == 0 {
			result.Skipped = true
			result.Success = true
			result.Duration = time.Since(start)
			result.Output = "skipped (no matching files)"
			return result
		}
	}

	toolPath, err := e.toolMgr.EnsureTool(hook.Tool, e.config.GetTool(hook.Tool))
	if err != nil {
		result.Error = err
		result.Duration = time.Since(start)
		return result
	}

	timeout := 5 * time.Minute
	if hook.Timeout != "" {
		if parsed, parseErr := time.ParseDuration(hook.Timeout); parseErr == nil {
			timeout = parsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	args := hook.Args
	if e.opts.Fix && len(hook.FixArgs) > 0 {
		args = hook.FixArgs
	}

	cmd := exec.CommandContext(ctx, toolPath, args...)
	cmd.Dir = e.workDir
	cmd.Env = e.buildEnv(hook)

	output, err := cmd.CombinedOutput()

	result.Duration = time.Since(start)
	result.Output = string(output)

	if ctx.Err() == context.DeadlineExceeded {
		result.Error = fmt.Errorf("timeout after %v", timeout)
		return result
	}

	if err != nil {
		result.Error = err
		return result
	}

	result.Success = true
	return result
}

func (e *Executor) buildEnv(hook config.Hook) []string {
	env := os.Environ()

	for k, v := range hook.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	return env
}

func (e *Executor) filterFiles(files []string, hook config.Hook) []string {
	matched := make([]string, 0, len(files))

	for _, f := range files {
		if hook.Files != "" {
			re, err := regexp.Compile(hook.Files)
			if err != nil || !re.MatchString(f) {
				continue
			}
		}

		if hook.Glob != "" {
			ok, matchErr := filepath.Match(hook.Glob, filepath.Base(f))
			if matchErr != nil || !ok {
				continue
			}
		}

		if hook.Exclude != "" {
			re, err := regexp.Compile(hook.Exclude)
			if err == nil && re.MatchString(f) {
				continue
			}
		}

		matched = append(matched, f)
	}

	if hook.Files == "" && hook.Glob == "" {
		return files
	}

	return matched
}

func PrintResults(results []Result, verbose bool, quiet bool) {
	if quiet {
		return
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	for _, r := range results {
		if r.Skipped {
			if verbose {
				fmt.Printf("%s %s (%v) - %s\n", yellow("[SKIP]"), r.Name, r.Duration.Round(time.Millisecond), r.Output)
			}
			continue
		}

		if r.Success {
			fmt.Printf("%s %s (%v)\n", green("[PASS]"), r.Name, r.Duration.Round(time.Millisecond))
			if verbose && r.Output != "" {
				fmt.Printf("  Output:\n%s\n", indent(r.Output))
			}
		} else {
			fmt.Printf("%s %s (%v)\n", red("[FAIL]"), r.Name, r.Duration.Round(time.Millisecond))
			if r.Error != nil {
				fmt.Printf("  Error: %v\n", r.Error)
			}
			if r.Output != "" {
				fmt.Printf("  Output:\n%s\n", indent(r.Output))
			}
		}
	}
}

func PrintPolicyResult(result *policy.EvalResult, quiet bool) {
	if quiet || result == nil {
		return
	}

	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	if result.Passed {
		fmt.Printf("%s policies\n", green("[PASS]"))
	} else {
		fmt.Printf("%s policies\n", red("[FAIL]"))
		for _, v := range result.Violations {
			fmt.Printf("  ✗ [%s] %s\n", v.Rule, v.Message)
		}
	}
}

func indent(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = "    " + line
	}
	return strings.Join(lines, "\n")
}

func HasFailure(results []Result) bool {
	for _, r := range results {
		if !r.Success && !r.Skipped {
			return true
		}
	}
	return false
}

func ParseSkipEnv() []string {
	skip := os.Getenv("SKIP")
	if skip == "" {
		return nil
	}
	return strings.Split(skip, ",")
}
