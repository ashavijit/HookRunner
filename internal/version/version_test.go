package version

import (
	"testing"
)

func TestString(t *testing.T) {
	result := String()

	if result == "" {
		t.Error("version string should not be empty")
	}

	if result != Version {
		t.Errorf("expected %s, got %s", Version, result)
	}
}

func TestFull(t *testing.T) {
	result := Full()

	if result == "" {
		t.Error("full version should not be empty")
	}

	if len(result) < len(Version) {
		t.Error("full version should be longer than version")
	}
}

func TestVersionVariables(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}

	if GitCommit == "" {
		t.Error("GitCommit should not be empty")
	}

	if BuildDate == "" {
		t.Error("BuildDate should not be empty")
	}
}
