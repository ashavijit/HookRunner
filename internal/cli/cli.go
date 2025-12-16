package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ashavijit/hookrunner/internal/config"
	"github.com/ashavijit/hookrunner/internal/executor"
	"github.com/ashavijit/hookrunner/internal/git"
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
)

var rootCmd = &cobra.Command{
	Use:   "hookrunner",
	Short: "Cross-platform pre-commit hook system",
	Long:  "A cross-platform pre-commit hook system with YAML/JSON configuration",
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
	Short: "Create sample config file",
	RunE:  runInit,
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

	rootCmd.AddCommand(installCmd, uninstallCmd, runCmd, runCmdCmd, listCmd, doctorCmd, initCmd, versionCmd)
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

	if err := os.WriteFile(configPath, []byte(config.DefaultConfig()), 0644); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	fmt.Printf("Created %s\n", configPath)
	fmt.Println("Run 'hookrunner install' to install git hooks")
	return nil
}
