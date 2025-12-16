package presets

import (
	"testing"
)

func TestList(t *testing.T) {
	presets := List()

	if len(presets) == 0 {
		t.Error("expected at least one preset")
	}

	if len(presets) != 6 {
		t.Errorf("expected 6 presets, got %d", len(presets))
	}
}

func TestGet_Exists(t *testing.T) {
	languages := []string{"go", "nodejs", "python", "java", "ruby", "rust"}

	for _, lang := range languages {
		preset, ok := Get(lang)
		if !ok {
			t.Errorf("preset %s should exist", lang)
		}

		if preset.Name == "" {
			t.Errorf("preset %s should have name", lang)
		}

		if preset.Config == "" {
			t.Errorf("preset %s should have config", lang)
		}

		if preset.Description == "" {
			t.Errorf("preset %s should have description", lang)
		}
	}
}

func TestGet_NotExists(t *testing.T) {
	_, ok := Get("nonexistent")
	if ok {
		t.Error("nonexistent preset should not exist")
	}
}

func TestAvailableLanguages(t *testing.T) {
	languages := AvailableLanguages()

	if len(languages) != 6 {
		t.Errorf("expected 6 languages, got %d", len(languages))
	}

	expected := map[string]bool{
		"go":     true,
		"nodejs": true,
		"python": true,
		"java":   true,
		"ruby":   true,
		"rust":   true,
	}

	for _, lang := range languages {
		if !expected[lang] {
			t.Errorf("unexpected language: %s", lang)
		}
	}
}

func TestPresetConfigs_Valid(t *testing.T) {
	for name, preset := range Languages {
		if len(preset.Config) < 50 {
			t.Errorf("preset %s config seems too short", name)
		}

		if preset.Name == "" {
			t.Errorf("preset %s missing Name", name)
		}
	}
}
