package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourusername/howto/internal/config"
	"github.com/yourusername/howto/internal/loader"
	"github.com/yourusername/howto/internal/output"
	"github.com/yourusername/howto/internal/registry"
)

func TestIntegration_FullWorkflow(t *testing.T) {
	// Use testdata fixtures
	globalPath := filepath.Join("testdata", ".config", "howto")
	projectPath := filepath.Join("testdata", "project", ".howto")

	// Load documents
	globalDocs, err := loader.LoadGlobalDocs(globalPath)
	if err != nil {
		t.Fatalf("failed to load global docs: %v", err)
	}

	projectDocs, err := loader.LoadProjectDocs(projectPath)
	if err != nil {
		t.Fatalf("failed to load project docs: %v", err)
	}

	// Load project config
	projectConfig, err := config.LoadProjectConfig(projectPath)
	if err != nil {
		t.Fatalf("failed to load project config: %v", err)
	}

	// Build registry
	reg := registry.BuildRegistry(globalDocs, projectDocs, projectConfig)

	// Test 1: Registry should contain expected commands
	expectedCommands := []string{"rust-lang", "go-lang", "commits", "optional-rule"}
	for _, cmd := range expectedCommands {
		if !reg.Has(cmd) {
			t.Errorf("expected command '%s' to be in registry", cmd)
		}
	}

	// Test 2: Help output should list all commands
	var helpBuf bytes.Buffer
	output.PrintHelp(&helpBuf, reg)
	helpOutput := helpBuf.String()

	if !strings.Contains(helpOutput, "rust-lang:") {
		t.Error("expected rust-lang in help output")
	}
	if !strings.Contains(helpOutput, "commits:") {
		t.Error("expected commits in help output")
	}
	if !strings.Contains(helpOutput, "optional-rule:") {
		t.Error("expected optional-rule in help output (it's in project config)")
	}

	// Test 3: Command output should show content only
	var cmdBuf bytes.Buffer
	if err := output.PrintCommand(&cmdBuf, reg, "rust-lang"); err != nil {
		t.Fatalf("failed to print rust-lang command: %v", err)
	}

	cmdOutput := cmdBuf.String()
	if !strings.Contains(cmdOutput, "# Rust Design Principles") {
		t.Error("expected Rust content in command output")
	}
	if strings.Contains(cmdOutput, "---") {
		t.Error("did not expect YAML frontmatter in command output")
	}

	// Test 4: Project-scoped commands should be present
	doc, ok := reg.Get("commits")
	if !ok {
		t.Fatal("expected commits doc to exist")
	}
	if doc.Source != 1 { // parser.SourceProjectScoped = 1
		t.Error("expected commits to be project-scoped")
	}

	// Test 5: Optional rule should be included (it's in project config)
	if !reg.Has("optional-rule") {
		t.Error("expected optional-rule to be included (it's required by project config)")
	}
}

func TestIntegration_WithoutProjectConfig(t *testing.T) {
	globalPath := filepath.Join("testdata", ".config", "howto")

	// Load only global docs
	globalDocs, err := loader.LoadGlobalDocs(globalPath)
	if err != nil {
		t.Fatalf("failed to load global docs: %v", err)
	}

	// Empty project config
	emptyConfig := &config.ProjectConfig{}

	// Build registry without project docs
	reg := registry.BuildRegistry(globalDocs, nil, emptyConfig)

	// Should have rust-lang and go-lang (required=true)
	if !reg.Has("rust-lang") {
		t.Error("expected rust-lang in registry")
	}
	if !reg.Has("go-lang") {
		t.Error("expected go-lang in registry")
	}

	// Should NOT have optional-rule (required=false, not in config)
	if reg.Has("optional-rule") {
		t.Error("did not expect optional-rule in registry without project config")
	}
}

func TestIntegration_UnknownCommand(t *testing.T) {
	reg := registry.BuildRegistry(nil, nil, &config.ProjectConfig{})

	var buf bytes.Buffer
	err := output.PrintCommand(&buf, reg, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown command")
	}

	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("expected 'unknown command' error, got: %v", err)
	}
}

func TestGetGlobalConfigPath(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Test with HOME set
	os.Setenv("HOME", "/home/testuser")
	path, err := getGlobalConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join("/home/testuser", ".config", "howto")
	if path != expected {
		t.Errorf("expected path '%s', got '%s'", expected, path)
	}

	// Test with HOME unset
	os.Unsetenv("HOME")
	_, err = getGlobalConfigPath()
	if err == nil {
		t.Error("expected error when HOME is not set")
	}
}

func TestGetProjectPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	path, err := getProjectPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(cwd, ".howto")
	if path != expected {
		t.Errorf("expected path '%s', got '%s'", expected, path)
	}
}

func TestRunVersion(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	os.Args = []string{"howto", "--version"}

	originalVersion := version
	version = "v1.2.3"
	defer func() { version = originalVersion }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}

	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()
	os.Stdout = w

	if err := run(); err != nil {
		t.Fatalf("expected no error when printing version, got: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close write end of pipe: %v", err)
	}

	outputBytes, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read version output: %v", err)
	}

	if string(outputBytes) != "v1.2.3\n" {
		t.Fatalf("expected version output 'v1.2.3', got %q", string(outputBytes))
	}
}
