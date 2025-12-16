package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ashavijit/hookrunner/internal/config"
	"github.com/ashavijit/hookrunner/internal/executor"
	"github.com/ashavijit/hookrunner/internal/git"
	"github.com/ashavijit/hookrunner/internal/tool"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
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

var runCmd = &cobra.Command{
	Use:   "run [hook-type]",
	Short: "Run specified hook",
	Args:  cobra.ExactArgs(1),
	RunE:  runHook,
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

var allFiles bool

func init() {
	runCmd.Flags().BoolVar(&allFiles, "all-files", false, "Run on all files instead of staged files")
	rootCmd.AddCommand(installCmd, runCmd, listCmd, doctorCmd)
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
	if !allFiles {
		files, err = git.GetStagedFiles()
		if err != nil {
			return err
		}
		if len(files) == 0 {
			fmt.Println("No staged files")
			return nil
		}
	}

	cacheDir := filepath.Join(workDir, ".hooks", "cache")
	toolMgr := tool.NewManager(cacheDir)
	exec := executor.New(cfg, toolMgr, workDir)

	results := exec.Run(hookType, files, allFiles)
	executor.PrintResults(results)

	if executor.HasFailure(results) {
		os.Exit(1)
	}

	return nil
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
			fmt.Printf("  - %s (tool: %s)\n", h.Name, h.Tool)
		}
		fmt.Println()
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
	if _, err := os.Stat(cacheDir); err == nil {
		fmt.Printf("%s Cache directory exists\n", green("[OK]"))
	} else {
		fmt.Printf("%s Cache directory not found (will be created on first run)\n", green("[INFO]"))
	}

	return nil
}
