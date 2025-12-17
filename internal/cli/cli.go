package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ashavijit/hookrunner/internal/config"
	"github.com/ashavijit/hookrunner/internal/executor"
	"github.com/ashavijit/hookrunner/internal/git"
	"github.com/ashavijit/hookrunner/internal/policy"
	"github.com/ashavijit/hookrunner/internal/presets"
	"github.com/ashavijit/hookrunner/internal/tool"
	"github.com/ashavijit/hookrunner/internal/version"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	allFiles   bool
	verbose    bool
	quiet      bool
	fix        bool
	noFailFast bool
	language   string
)

var rootCmd = &cobra.Command{
	Use:   "hookrunner",
	Short: "Cross-platform pre-commit hook system",
	Long:  "A cross-platform pre-commit hook system with YAML/JSON configuration\nSupports: Go, Node.js, Python, Java, Ruby, Rust",
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install git hooks",
	RunE:  runInstall,
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove installed git hooks",
	RunE:  runUninstall,
}

var runCmd = &cobra.Command{
	Use:   "run [hook-type]",
	Short: "Run specified hook",
	Args:  cobra.ExactArgs(1),
	RunE:  runHook,
}

var runCmdCmd = &cobra.Command{
	Use:   "run-cmd [tool] [args...]",
	Short: "Run a tool directly",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runDirectCmd,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured hooks",
	RunE:  runList,
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose setup",
	RunE:  runDoctor,
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create config file (use --lang for language preset)",
	Long:  "Create hooks.yaml config file. Use --lang to specify language:\n  go, nodejs, python, java, ruby, rust",
	RunE:  runInit,
}

var presetsCmd = &cobra.Command{
	Use:   "presets",
	Short: "List available language presets",
	Run:   runPresets,
}

var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage policies",
}

var policyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured policies",
	RunE:  runPolicyList,
}

var policyFetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Force refresh remote policies",
	RunE:  runPolicyFetch,
}

var policyClearCmd = &cobra.Command{
	Use:   "clear-cache",
	Short: "Clear policy cache",
	RunE:  runPolicyClearCache,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("hookrunner %s\n", version.Full())
	},
}

func init() {
	runCmd.Flags().BoolVar(&allFiles, "all-files", false, "Run on all files")
	runCmd.Flags().BoolVar(&verbose, "verbose", false, "Verbose output")
	runCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Quiet output")
	runCmd.Flags().BoolVar(&fix, "fix", false, "Run in fix mode")
	runCmd.Flags().BoolVar(&noFailFast, "no-fail-fast", false, "Continue on failure")

	initCmd.Flags().StringVar(&language, "lang", "", "Language preset (go, nodejs, python, java, ruby, rust)")

	policyCmd.AddCommand(policyListCmd, policyFetchCmd, policyClearCmd)
	rootCmd.AddCommand(installCmd, uninstallCmd, runCmd, runCmdCmd, listCmd, doctorCmd, initCmd, presetsCmd, policyCmd, versionCmd)
}

func Execute() error {
	return rootCmd.Execute()
}

func runInstall(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not inside a git repository")
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	cfg, _, err := config.Load(workDir)
	if err != nil {
		return err
	}

	hookTypes := []string{"pre-commit", "pre-push", "commit-msg"}
	installed := 0

	for _, hookType := range hookTypes {
		if hooks := cfg.GetHooks(hookType); len(hooks) > 0 {
			if err := git.InstallHook(hookType, executable); err != nil {
				return fmt.Errorf("failed to install %s hook: %w", hookType, err)
			}
			fmt.Printf("Installed %s hook\n", hookType)
			installed++
		}
	}

	if installed == 0 {
		fmt.Println("No hooks to install")
	}

	return nil
}

func runUninstall(cmd *cobra.Command, args []string) error {
	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not inside a git repository")
	}

	hookTypes := []string{"pre-commit", "pre-push", "commit-msg"}
	removed := 0

	for _, hookType := range hookTypes {
		if err := git.UninstallHook(hookType); err != nil {
			return fmt.Errorf("failed to uninstall %s hook: %w", hookType, err)
		}
		removed++
	}

	fmt.Printf("Removed %d hooks\n", removed)
	return nil
}

func runHook(cmd *cobra.Command, args []string) error {
	hookType := args[0]
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, _, err := config.Load(workDir)
	if err != nil {
		return err
	}

	hooks := cfg.GetHooks(hookType)
	if len(hooks) == 0 {
		return fmt.Errorf("no hooks configured for %s", hookType)
	}

	var files []string
	if allFiles {
		files, err = git.GetAllFiles()
	} else {
		files, err = git.GetStagedFiles()
	}
	if err != nil {
		return err
	}

	if len(files) == 0 && !allFiles {
		fmt.Println("No staged files")
		return nil
	}

	cacheDir := filepath.Join(workDir, ".hooks", "cache")
	toolMgr := tool.NewManager(cacheDir)
	exec := executor.New(cfg, toolMgr, workDir)

	opts := executor.Options{
		Verbose:   verbose,
		Quiet:     quiet,
		Fix:       fix,
		FailFast:  !noFailFast,
		SkipHooks: executor.ParseSkipEnv(),
	}
	exec.SetOptions(opts)

	results := exec.Run(hookType, files, allFiles)
	executor.PrintResults(results, verbose, quiet)

	if executor.HasFailure(results) {
		os.Exit(1)
	}

	return nil
}

func runDirectCmd(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cacheDir := filepath.Join(workDir, ".hooks", "cache")
	toolMgr := tool.NewManager(cacheDir)

	cfg, _, _ := config.Load(workDir)
	var toolCfg *config.Tool
	if cfg != nil {
		toolCfg = cfg.GetTool(args[0])
	}

	toolPath, err := toolMgr.EnsureTool(args[0], toolCfg)
	if err != nil {
		return fmt.Errorf("tool not found: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	execCmd := exec.CommandContext(ctx, toolPath, args[1:]...)
	execCmd.Dir = workDir
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin

	return execCmd.Run()
}

func runList(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, cfgPath, err := config.Load(workDir)
	if err != nil {
		return err
	}

	fmt.Printf("Config: %s\n\n", cfgPath)

	hookTypes := []string{"pre-commit", "pre-push", "commit-msg"}
	for _, hookType := range hookTypes {
		hooks := cfg.GetHooks(hookType)
		if len(hooks) == 0 {
			continue
		}

		fmt.Printf("%s:\n", hookType)
		for _, h := range hooks {
			extra := ""
			if h.After != "" {
				extra = fmt.Sprintf(" (after: %s)", h.After)
			}
			fmt.Printf("  - %s (tool: %s)%s\n", h.Name, h.Tool, extra)
		}
		fmt.Println()
	}

	if len(cfg.Tools) > 0 {
		fmt.Println("Tools:")
		for name, t := range cfg.Tools {
			fmt.Printf("  - %s v%s\n", name, t.Version)
		}
	}

	return nil
}

func runDoctor(cmd *cobra.Command, args []string) error {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	fmt.Println("HookRunner Doctor")
	fmt.Println("=================")
	fmt.Println()

	if git.IsInsideWorkTree() {
		fmt.Printf("%s Git repository detected\n", green("[OK]"))
	} else {
		fmt.Printf("%s Not a git repository\n", red("[FAIL]"))
	}

	cfg, cfgPath, err := config.Load(workDir)
	if err != nil {
		fmt.Printf("%s Config file: %v\n", red("[FAIL]"), err)
	} else {
		fmt.Printf("%s Config file: %s\n", green("[OK]"), cfgPath)

		hookCount := 0
		for _, hooks := range cfg.Hooks {
			hookCount += len(hooks)
		}
		fmt.Printf("%s Hooks configured: %d\n", green("[OK]"), hookCount)

		toolCount := len(cfg.Tools)
		fmt.Printf("%s Tools defined: %d\n", green("[OK]"), toolCount)
	}

	cacheDir := filepath.Join(workDir, ".hooks", "cache")
	if info, err := os.Stat(cacheDir); err == nil && info.IsDir() {
		fmt.Printf("%s Cache directory exists\n", green("[OK]"))
	} else {
		fmt.Printf("%s Cache directory not found\n", green("[INFO]"))
	}

	fmt.Printf("\nVersion: %s\n", version.String())
	fmt.Printf("Supported: %s\n", strings.Join(presets.AvailableLanguages(), ", "))

	return nil
}

func runInit(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	configPath := filepath.Join(workDir, "hooks.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists: %s", configPath)
	}

	var configContent string
	if language != "" {
		preset, ok := presets.Get(language)
		if !ok {
			return fmt.Errorf("unknown language: %s\nAvailable: %s", language, strings.Join(presets.AvailableLanguages(), ", "))
		}
		configContent = preset.Config
		fmt.Printf("Using %s preset\n", preset.Name)
	} else {
		configContent = config.DefaultConfig()
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	fmt.Printf("Created %s\n", configPath)
	fmt.Println("Run 'hookrunner install' to install git hooks")
	return nil
}

func runPresets(cmd *cobra.Command, args []string) {
	fmt.Println("Available Language Presets:")
	fmt.Println("===========================")
	fmt.Println()

	for _, lang := range presets.AvailableLanguages() {
		p, _ := presets.Get(lang)
		fmt.Printf("  %-10s %s\n", lang, p.Description)
	}

	fmt.Println()
	fmt.Println("Usage: hookrunner init --lang <language>")
}

func runPolicyList(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, _, err := config.Load(workDir)
	if err != nil {
		return err
	}

	fmt.Println("Policy Configuration:")
	fmt.Println("=====================")
	fmt.Println()

	if cfg.Policies == nil {
		fmt.Println("No policies configured")
		return nil
	}

	p := cfg.Policies
	fmt.Printf("Type: %s\n", p.Type)

	if cfg.HasRemotePolicies() {
		fmt.Println("\nRemote Policies:")
		for _, ref := range p.Policies {
			fmt.Printf("  - %s\n", ref.URL)
		}
	}

	if len(p.LocalPolicies) > 0 {
		fmt.Println("\nLocal Policies:")
		for _, lp := range p.LocalPolicies {
			fmt.Printf("  - %s\n", lp.Name)
		}
	}

	if cfg.HasRemotePolicies() {
		cacheDir := filepath.Join(workDir, ".hooks", "cache")
		registry := policy.NewRegistry(workDir, cacheDir)

		userCfg := buildUserConfig(p)

		merged, err := registry.Load(userCfg)
		if err != nil {
			fmt.Printf("\n⚠ Failed to load policies: %v\n", err)
		} else if merged != nil {
			fmt.Println("\nLoaded Remote Policies:")
			for _, rp := range merged.RemotePolicies {
				fmt.Printf("  ✓ %s\n", rp.Identifier())
			}
		}
	}

	return nil
}

func runPolicyFetch(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cfg, _, err := config.Load(workDir)
	if err != nil {
		return err
	}

	if !cfg.HasRemotePolicies() {
		fmt.Println("No remote policies configured")
		return nil
	}

	cacheDir := filepath.Join(workDir, ".hooks", "cache")
	registry := policy.NewRegistry(workDir, cacheDir)

	p := cfg.Policies
	userCfg := buildUserConfig(p)

	if err := registry.Refresh(userCfg); err != nil {
		return fmt.Errorf("failed to refresh: %w", err)
	}

	fmt.Println("Policies refreshed successfully")
	return nil
}

func runPolicyClearCache(cmd *cobra.Command, args []string) error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	cacheDir := filepath.Join(workDir, ".hooks", "cache")
	registry := policy.NewRegistry(workDir, cacheDir)

	if err := registry.ClearCache(); err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	fmt.Println("Policy cache cleared")
	return nil
}

func buildUserConfig(p *config.Policies) *policy.UserConfig {
	userCfg := &policy.UserConfig{Type: p.Type}

	for _, ref := range p.Policies {
		userCfg.Policies = append(userCfg.Policies, policy.PolicyRef{URL: ref.URL})
	}

	for _, lp := range p.LocalPolicies {
		userCfg.LocalPolicies = append(userCfg.LocalPolicies, policy.LocalPolicy{
			Name:        lp.Name,
			Version:     lp.Version,
			Description: lp.Description,
			Metadata:    lp.Metadata,
			Rules:       convertConfigRules(lp.Rules),
		})
	}

	return userCfg
}

func convertConfigRules(r config.PolicyRules) policy.PolicyRules {
	var cm *policy.CommitMessageRule
	if r.CommitMessage != nil {
		cm = &policy.CommitMessageRule{
			Regex: r.CommitMessage.Regex,
			Error: r.CommitMessage.Error,
		}
	}

	var patterns []policy.ForbiddenContentPattern
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
