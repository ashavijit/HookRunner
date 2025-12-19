package cli

import (
	"testing"
)

func TestRootCmd(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}
	if rootCmd.Use != "hookrunner" {
		t.Errorf("expected 'hookrunner', got '%s'", rootCmd.Use)
	}
}

func TestVersionCmd(t *testing.T) {
	if versionCmd == nil {
		t.Fatal("versionCmd should not be nil")
	}
	if versionCmd.Use != "version" {
		t.Errorf("expected 'version', got '%s'", versionCmd.Use)
	}
}

func TestPresetsCmd(t *testing.T) {
	if presetsCmd == nil {
		t.Fatal("presetsCmd should not be nil")
	}
}

func TestPolicyCmd(t *testing.T) {
	if policyCmd == nil {
		t.Fatal("policyCmd should not be nil")
	}
	if len(policyCmd.Commands()) != 3 {
		t.Errorf("expected 3 subcommands, got %d", len(policyCmd.Commands()))
	}
}

func TestCacheCmd(t *testing.T) {
	if cacheCmd == nil {
		t.Fatal("cacheCmd should not be nil")
	}
}

func TestRunCmdFlags(t *testing.T) {
	flags := runCmd.Flags()

	if flags.Lookup("all-files") == nil {
		t.Error("missing --all-files flag")
	}
	if flags.Lookup("verbose") == nil {
		t.Error("missing --verbose flag")
	}
	if flags.Lookup("quiet") == nil {
		t.Error("missing --quiet flag")
	}
	if flags.Lookup("fix") == nil {
		t.Error("missing --fix flag")
	}
	if flags.Lookup("dry-run") == nil {
		t.Error("missing --dry-run flag")
	}
	if flags.Lookup("cached") == nil {
		t.Error("missing --cached flag")
	}
	if flags.Lookup("clean-room") == nil {
		t.Error("missing --clean-room flag")
	}
}

func TestInitCmdFlags(t *testing.T) {
	flags := initCmd.Flags()

	if flags.Lookup("lang") == nil {
		t.Error("missing --lang flag")
	}
}

func TestPromptConfirm(t *testing.T) {
	result := promptConfirm("test")
	if result {
		t.Error("promptConfirm should return false without input")
	}
}

func TestExecute(t *testing.T) {
	oldArgs := rootCmd.Args
	rootCmd.SetArgs([]string{"--help"})
	err := Execute()
	if err != nil {
		t.Logf("Execute returned: %v", err)
	}
	rootCmd.Args = oldArgs
}
