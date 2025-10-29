package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourusername/howto/internal/config"
	"github.com/yourusername/howto/internal/parser"
	"github.com/yourusername/howto/internal/registry"
)

func TestPrintHelp_WithCommands(t *testing.T) {
	globalDocs := []parser.Document{
		{Name: "rust-lang", Description: "Documentation for Rust projects", Required: true, Source: parser.SourceGlobal},
	}
	projectDocs := []parser.Document{
		{Name: "commits", Description: "Commit guidelines", Source: parser.SourceProjectScoped},
	}

	reg := registry.BuildRegistry(globalDocs, projectDocs, &config.ProjectConfig{})

	var buf bytes.Buffer
	PrintHelp(&buf, reg)

	output := buf.String()

	// Check for expected elements
	if !strings.Contains(output, "Usage: howto [COMMAND]") {
		t.Error("expected usage line in output")
	}

	if !strings.Contains(output, "An LLM agent documentation") {
		t.Error("expected description line in output")
	}

	if !strings.Contains(output, "Commands:") {
		t.Error("expected 'Commands:' header in output")
	}

	if !strings.Contains(output, "commits:") {
		t.Error("expected 'commits:' command in output")
	}

	if !strings.Contains(output, "rust-lang:") {
		t.Error("expected 'rust-lang:' command in output")
	}

	if !strings.Contains(output, "Documentation for Rust projects") {
		t.Error("expected rust-lang description in output")
	}

	if !strings.Contains(output, "Commit guidelines") {
		t.Error("expected commits description in output")
	}
}

func TestPrintHelp_Empty(t *testing.T) {
	reg := registry.BuildRegistry(nil, nil, &config.ProjectConfig{})

	var buf bytes.Buffer
	PrintHelp(&buf, reg)

	output := buf.String()

	if !strings.Contains(output, "No commands available.") {
		t.Error("expected 'No commands available.' for empty registry")
	}
}

func TestPrintHelp_Sorted(t *testing.T) {
	docs := []parser.Document{
		{Name: "zebra", Description: "Z", Source: parser.SourceProjectScoped},
		{Name: "alpha", Description: "A", Source: parser.SourceProjectScoped},
		{Name: "middle", Description: "M", Source: parser.SourceProjectScoped},
	}

	reg := registry.BuildRegistry(nil, docs, &config.ProjectConfig{})

	var buf bytes.Buffer
	PrintHelp(&buf, reg)

	output := buf.String()

	// Find positions of command names
	alphaPos := strings.Index(output, "alpha:")
	middlePos := strings.Index(output, "middle:")
	zebraPos := strings.Index(output, "zebra:")

	if alphaPos == -1 || middlePos == -1 || zebraPos == -1 {
		t.Fatal("expected all commands to be in output")
	}

	// Check they're in alphabetical order
	if !(alphaPos < middlePos && middlePos < zebraPos) {
		t.Error("expected commands to be sorted alphabetically")
	}
}

func TestPrintCommand_Success(t *testing.T) {
	docs := []parser.Document{
		{
			Name:        "test-doc",
			Description: "Test",
			Content:     "# Test Content\n\nThis is a test.",
			Source:      parser.SourceProjectScoped,
		},
	}

	reg := registry.BuildRegistry(nil, docs, &config.ProjectConfig{})

	var buf bytes.Buffer
	err := PrintCommand(&buf, reg, "test-doc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	expectedContent := "# Test Content\n\nThis is a test.\n"
	if output != expectedContent {
		t.Errorf("expected output:\n%s\ngot:\n%s", expectedContent, output)
	}
}

func TestPrintCommand_NotFound(t *testing.T) {
	reg := registry.BuildRegistry(nil, nil, &config.ProjectConfig{})

	var buf bytes.Buffer
	err := PrintCommand(&buf, reg, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent command")
	}

	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("expected 'unknown command' error, got: %v", err)
	}
}

func TestPrintCommand_OnlyContent(t *testing.T) {
	// Ensure frontmatter is not included in output
	docs := []parser.Document{
		{
			Name:        "doc",
			Description: "This should not appear in output",
			Content:     "Only this content should appear",
			Source:      parser.SourceProjectScoped,
		},
	}

	reg := registry.BuildRegistry(nil, docs, &config.ProjectConfig{})

	var buf bytes.Buffer
	err := PrintCommand(&buf, reg, "doc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if strings.Contains(output, "This should not appear in output") {
		t.Error("description should not be in command output")
	}

	if !strings.Contains(output, "Only this content should appear") {
		t.Error("expected content to be in command output")
	}
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		indent   int
		expected string
	}{
		{
			name:     "simple text",
			text:     "Hello world",
			indent:   4,
			expected: "    Hello world",
		},
		{
			name:     "empty text",
			text:     "",
			indent:   4,
			expected: "    (no description)",
		},
		{
			name:     "multiline text",
			text:     "Line 1\nLine 2",
			indent:   2,
			expected: "  Line 1\n  Line 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.indent, 80)
			if result != tt.expected {
				t.Errorf("expected:\n%s\ngot:\n%s", tt.expected, result)
			}
		})
	}
}

func TestPrintHelp_Indentation(t *testing.T) {
	docs := []parser.Document{
		{Name: "test", Description: "Test description", Source: parser.SourceProjectScoped},
	}

	reg := registry.BuildRegistry(nil, docs, &config.ProjectConfig{})

	var buf bytes.Buffer
	PrintHelp(&buf, reg)

	output := buf.String()

	// Command name should be indented with 4 spaces
	if !strings.Contains(output, "    test:") {
		t.Error("expected command name to be indented with 4 spaces")
	}

	// Description should be indented with 8 spaces
	if !strings.Contains(output, "        Test description") {
		t.Error("expected description to be indented with 8 spaces")
	}
}
