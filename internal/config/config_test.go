package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "howto-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return tmpDir
}

func writeConfigFile(t *testing.T, dir string, content string) {
	t.Helper()
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}

func TestLoadProjectConfig_ValidConfig(t *testing.T) {
	tmpDir := setupTestDir(t)
	writeConfigFile(t, tmpDir, `require:
  - important-rule
  - another-rule`)

	config, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Require) != 2 {
		t.Fatalf("expected 2 required items, got %d", len(config.Require))
	}

	if config.Require[0] != "important-rule" {
		t.Errorf("expected first item to be 'important-rule', got '%s'", config.Require[0])
	}
	if config.Require[1] != "another-rule" {
		t.Errorf("expected second item to be 'another-rule', got '%s'", config.Require[1])
	}
}

func TestLoadProjectConfig_EmptyRequire(t *testing.T) {
	tmpDir := setupTestDir(t)
	writeConfigFile(t, tmpDir, `require: []`)

	config, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Require) != 0 {
		t.Errorf("expected empty require list, got %d items", len(config.Require))
	}
}

func TestLoadProjectConfig_NoRequireField(t *testing.T) {
	tmpDir := setupTestDir(t)
	writeConfigFile(t, tmpDir, `other_field: value`)

	config, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.Require == nil || len(config.Require) != 0 {
		t.Errorf("expected empty require list, got %v", config.Require)
	}
}

func TestLoadProjectConfig_MissingFile(t *testing.T) {
	tmpDir := setupTestDir(t)

	config, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("expected no error for missing config file, got: %v", err)
	}

	if len(config.Require) != 0 {
		t.Errorf("expected empty require list for missing config, got %d items", len(config.Require))
	}
}

func TestLoadProjectConfig_InvalidYAML(t *testing.T) {
	tmpDir := setupTestDir(t)
	writeConfigFile(t, tmpDir, `invalid: yaml: content:`)

	_, err := LoadProjectConfig(tmpDir)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadProjectConfig_NonExistentDirectory(t *testing.T) {
	config, err := LoadProjectConfig("/nonexistent/directory")
	if err != nil {
		t.Fatalf("expected no error for nonexistent directory, got: %v", err)
	}

	if len(config.Require) != 0 {
		t.Errorf("expected empty require list, got %d items", len(config.Require))
	}
}

func TestHasRequire(t *testing.T) {
	config := &ProjectConfig{
		Require: []string{"rule1", "rule2", "rule3"},
	}

	tests := []struct {
		name     string
		expected bool
	}{
		{"rule1", true},
		{"rule2", true},
		{"rule3", true},
		{"rule4", false},
		{"", false},
		{"Rule1", false}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.HasRequire(tt.name)
			if result != tt.expected {
				t.Errorf("HasRequire(%q) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestHasRequire_EmptyConfig(t *testing.T) {
	config := &ProjectConfig{
		Require: []string{},
	}

	if config.HasRequire("anything") {
		t.Error("expected false for empty require list")
	}
}

func TestLoadProjectConfig_SingleItem(t *testing.T) {
	tmpDir := setupTestDir(t)
	writeConfigFile(t, tmpDir, `require:
  - single-rule`)

	config, err := LoadProjectConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Require) != 1 {
		t.Fatalf("expected 1 required item, got %d", len(config.Require))
	}

	if config.Require[0] != "single-rule" {
		t.Errorf("expected 'single-rule', got '%s'", config.Require[0])
	}
}
