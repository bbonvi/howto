package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yourusername/howto/internal/app"
	"github.com/yourusername/howto/internal/config"
	"github.com/yourusername/howto/internal/loader"
	"github.com/yourusername/howto/internal/mcp"
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

	// Test 1: Registry should contain expected playbooks
	expectedPlaybooks := []string{"rust-lang", "go-lang", "commits", "optional-rule"}
	for _, pb := range expectedPlaybooks {
		if !reg.Has(pb) {
			t.Errorf("expected playbook '%s' to be in registry", pb)
		}
	}

	// Test 2: Help output should list all playbooks
	var helpBuf bytes.Buffer
	output.PrintHelp(&helpBuf, reg)
	helpOutput := helpBuf.String()

	if !strings.Contains(helpOutput, "  rust-lang:") {
		t.Error("expected rust-lang in help output")
	}
	if !strings.Contains(helpOutput, "  commits:") {
		t.Error("expected commits in help output")
	}
	if !strings.Contains(helpOutput, "  optional-rule:") {
		t.Error("expected optional-rule in help output (it's required by project config)")
	}

	// Test 3: Playbook output should show content only
	var cmdBuf bytes.Buffer
	if err := output.PrintPlaybook(&cmdBuf, reg, "rust-lang"); err != nil {
		t.Fatalf("failed to print rust-lang playbook: %v", err)
	}

	cmdOutput := cmdBuf.String()
	if !strings.Contains(cmdOutput, "# Rust Design Principles") {
		t.Error("expected Rust content in playbook output")
	}
	if strings.Contains(cmdOutput, "---") {
		t.Error("did not expect YAML frontmatter in playbook output")
	}

	// Test 4: Project-scoped playbooks should be present
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

func TestIntegration_UnknownPlaybook(t *testing.T) {
	reg := registry.BuildRegistry(nil, nil, &config.ProjectConfig{})

	var buf bytes.Buffer
	err := output.PrintPlaybook(&buf, reg, "nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown playbook")
	}

	if !strings.Contains(err.Error(), "unknown playbook") {
		t.Errorf("expected 'unknown playbook' error, got: %v", err)
	}
}

func TestIntegration_MCPServer(t *testing.T) {
	globalPath := filepath.Join("testdata", ".config", "howto")
	projectPath := filepath.Join("testdata", "project", ".howto")

	loader := app.NewCachedRegistryLoader(globalPath, projectPath)

	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"integration-test"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`,
		`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_playbooks","arguments":{}}}`,
		`{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_playbook","arguments":{"name":"rust-lang"}}}`,
	}, "\n")

	var output bytes.Buffer
	server := mcp.NewServer(strings.NewReader(input), &output, loader, "integration", log.New(io.Discard, "", 0))

	if err := server.Serve(); err != nil {
		t.Fatalf("Serve() returned error: %v", err)
	}

	responses := decodeMCPResponses(t, output.String())
	if len(responses) != 4 {
		t.Fatalf("expected 4 responses, got %d", len(responses))
	}

	if responses[1].Error != nil {
		t.Fatalf("tools/list returned error: %+v", responses[1].Error)
	}
	tools, ok := responses[1].Result["tools"].([]any)
	if !ok || len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %#v", responses[1].Result["tools"])
	}

	assertContentContains(t, responses[2].Result, "rust-lang")
	assertContentContains(t, responses[3].Result, "Rust Design Principles")
}

func TestGlobalConfigDir(t *testing.T) {
	// Save original HOME
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Test with HOME set
	os.Setenv("HOME", "/home/testuser")
	path, err := app.GlobalConfigDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join("/home/testuser", ".config", "howto")
	if path != expected {
		t.Errorf("expected path '%s', got '%s'", expected, path)
	}

	// Test with HOME unset
	os.Unsetenv("HOME")
	_, err = app.GlobalConfigDir()
	if err == nil {
		t.Error("expected error when HOME is not set")
	}
}

func TestProjectConfigDir(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	path, err := app.ProjectConfigDir()
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

type mcpResponse struct {
	ID     any            `json:"id"`
	Result map[string]any `json:"result"`
	Error  *mcpError      `json:"error"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func decodeMCPResponses(t *testing.T, raw string) []mcpResponse {
	t.Helper()
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	lines := strings.Split(raw, "\n")
	out := make([]mcpResponse, 0, len(lines))
	for _, line := range lines {
		var msg mcpResponse
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			t.Fatalf("failed to decode MCP response %q: %v", line, err)
		}
		if msg.Result == nil {
			msg.Result = map[string]any{}
		}
		out = append(out, msg)
	}
	return out
}

func assertContentContains(t *testing.T, result map[string]any, expected string) {
	t.Helper()
	contentRaw, ok := result["content"]
	if !ok {
		t.Fatalf("result missing content field: %#v", result)
	}

	contentSlice, ok := contentRaw.([]any)
	if !ok || len(contentSlice) == 0 {
		t.Fatalf("content has unexpected shape: %#v", contentRaw)
	}

	firstEntry, ok := contentSlice[0].(map[string]any)
	if !ok {
		t.Fatalf("content entry has unexpected type: %#v", contentSlice[0])
	}

	text, _ := firstEntry["text"].(string)
	if !strings.Contains(text, expected) {
		t.Fatalf("expected content to contain %q, got %q", expected, text)
	}
}
