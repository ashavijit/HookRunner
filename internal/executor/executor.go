package executor

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"sync"
	"time"

	"github.com/ashavijit/hookrunner/internal/config"
	"github.com/ashavijit/hookrunner/internal/tool"
	"github.com/fatih/color"
)

type Result struct {
	Name     string
	Success  bool
	Duration time.Duration
	Output   string
	Error    error
}

type Executor struct {
	toolMgr *tool.Manager
	config  *config.Config
	workDir string
}

func New(cfg *config.Config, toolMgr *tool.Manager, workDir string) *Executor {
	return &Executor{
		toolMgr: toolMgr,
		config:  cfg,
		workDir: workDir,
	}
}

func (e *Executor) Run(hookType string, files []string, allFiles bool) []Result {
	hooks := e.config.GetHooks(hookType)
	if len(hooks) == 0 {
		return nil
	}

	ordered := e.orderHooks(hooks)
	var results []Result

	for _, batch := range ordered {
		batchResults := e.runBatch(batch, files, allFiles)
		results = append(results, batchResults...)

		for _, r := range batchResults {
			if !r.Success {
				return results
			}
		}
	}

	return results
}

func (e *Executor) orderHooks(hooks []config.Hook) [][]config.Hook {
	var independent []config.Hook
	dependent := make(map[string][]config.Hook)

	for _, h := range hooks {
		if h.After == "" {
			independent = append(independent, h)
		} else {
			dependent[h.After] = append(dependent[h.After], h)
		}
	}

	var ordered [][]config.Hook
	if len(independent) > 0 {
		ordered = append(ordered, independent)
	}

	for len(dependent) > 0 {
		var nextBatch []config.Hook
		for after, deps := range dependent {
			found := false
			for _, batch := range ordered {
				for _, h := range batch {
					if h.Name == after {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if found {
				nextBatch = append(nextBatch, deps...)
				delete(dependent, after)
			}
		}
		if len(nextBatch) > 0 {
			ordered = append(ordered, nextBatch)
		} else {
			break
		}
	}

	return ordered
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

func (e *Executor) runHook(hook config.Hook, files []string, allFiles bool) Result {
	start := time.Now()
	result := Result{Name: hook.Name}

	if !allFiles && hook.Files != "" {
		matched := e.filterFiles(files, hook.Files)
		if len(matched) == 0 {
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
		if parsed, err := time.ParseDuration(hook.Timeout); err == nil {
			timeout = parsed
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, toolPath, hook.Args...)
	cmd.Dir = e.workDir
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

func (e *Executor) filterFiles(files []string, pattern string) []string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return files
	}

	var matched []string
	for _, f := range files {
		if re.MatchString(f) {
			matched = append(matched, f)
		}
	}
	return matched
}

func PrintResults(results []Result) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	for _, r := range results {
		if r.Success {
			fmt.Printf("%s %s (%v)\n", green("[PASS]"), r.Name, r.Duration.Round(time.Millisecond))
		} else {
			fmt.Printf("%s %s (%v)\n", red("[FAIL]"), r.Name, r.Duration.Round(time.Millisecond))
			if r.Error != nil {
				fmt.Printf("  Error: %v\n", r.Error)
			}
			if r.Output != "" {
				fmt.Printf("  Output:\n%s\n", r.Output)
			}
		}
	}
}

func HasFailure(results []Result) bool {
	for _, r := range results {
		if !r.Success {
			return true
		}
	}
	return false
}
