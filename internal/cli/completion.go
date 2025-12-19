package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var setupCompletionCmd = &cobra.Command{
	Use:   "setup-completion",
	Short: "Install shell auto-completion",
	Long: `Automatically configures shell auto-completion for the current user.
Supported shells: bash, zsh, fish, powershell.

This command will attempt to modify your shell configuration file
to enable auto-completion for hookrunner.`,
	RunE: runSetupCompletion,
}

func init() {
	rootCmd.AddCommand(setupCompletionCmd)
}

func runSetupCompletion(cmd *cobra.Command, args []string) error {
	shell := identifyShell()
	if len(args) > 0 {
		shell = args[0]
	}

	fmt.Printf("Detected shell: %s\n", shell)

	switch shell {
	case "powershell", "pwsh":
		return setupPowerShell()
	case "bash":
		return setupBash()
	case "zsh":
		return setupZsh()
	case "fish":
		return setupFish()
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

func identifyShell() string {
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	shell := os.Getenv("SHELL")
	if shell != "" {
		return filepath.Base(shell)
	}
	return "bash"
}

func setupPowerShell() error {
	cmd := exec.Command("powershell", "-NoProfile", "-Command", "Write-Host $PROFILE")
	output, err := cmd.Output()
	var profilePath string

	if err == nil && len(output) > 0 {
		profilePath = strings.TrimSpace(string(output))
	}

	if profilePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		profilePath = filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
	}

	if err := os.MkdirAll(filepath.Dir(profilePath), 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	script := "\n# HookRunner completion\nif (Get-Command hookrunner -ErrorAction SilentlyContinue) {\n    hookrunner completion powershell | Out-String | Invoke-Expression\n}\n"

	return appendToProfile(profilePath, script)
}

func setupBash() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	rcFile := filepath.Join(home, ".bashrc")
	script := "\n# HookRunner completion\nsource <(hookrunner completion bash)\n"

	return appendToProfile(rcFile, script)
}

func setupZsh() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	rcFile := filepath.Join(home, ".zshrc")
	script := "\n# HookRunner completion\nsource <(hookrunner completion zsh)\n"

	return appendToProfile(rcFile, script)
}

func setupFish() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ".config", "fish", "completions")
	if mkErr := os.MkdirAll(configDir, 0755); mkErr != nil {
		return mkErr
	}

	target := filepath.Join(configDir, "hookrunner.fish")

	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	return rootCmd.GenFishCompletion(f, true)
}

func appendToProfile(path string, content string) error {
	fmt.Printf("Updating %s...\n", path)

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := os.ReadFile(path)
	if err == nil {
		if strings.Contains(string(data), "# HookRunner completion") {
			fmt.Println("Completion already configured.")
			return nil
		}
	} else if !os.IsNotExist(err) {
		// If read fails for other reasons, report it
		return fmt.Errorf("failed to read profile for duplicate check: %w", err)
	}

	if _, err := f.WriteString(content); err != nil {
		return err
	}

	fmt.Println("Success! Please restart your shell.")
	return nil
}
