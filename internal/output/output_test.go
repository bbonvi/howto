package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourusername/howto/internal/config"
	"github.com/yourusername/howto/internal/parser"
	"github.com/yourusername/howto/internal/registry"
)

func TestPrintHelp_WithPlaybooks(t *testing.T) {
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
	if !strings.Contains(output, "Usage: howto [PLAYBOOK]") {
		t.Error("expected usage line in output")
	}

	if !strings.Contains(output, "`howto` lets language models pull the exact playbooks their operators prepared.") {
		t.Error("expected overview line in output")
	}

	if !strings.Contains(output, "Run it to list playbooks, then fetch the one you need with `howto <playbook>`.") {
		t.Error("expected usage hint in output")
	}

	if !strings.Contains(output, "Playbooks:") {
		t.Error("expected 'Playbooks:' header in output")
	}

	if !strings.Contains(output, "- commits: Commit guidelines") {
		t.Error("expected 'commits' playbook line in output")
	}

	if !strings.Contains(output, "- rust-lang: Documentation for Rust projects") {
		t.Error("expected 'rust-lang' playbook line in output")
	}
}

func TestPrintHelp_Empty(t *testing.T) {
	reg := registry.BuildRegistry(nil, nil, &config.ProjectConfig{})

	var buf bytes.Buffer
	PrintHelp(&buf, reg)

	output := buf.String()

	if !strings.Contains(output, "No playbooks available.") {
		t.Error("expected 'No playbooks available.' for empty registry")
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

	// Find positions of playbook names
	alphaPos := strings.Index(output, "- alpha:")
	middlePos := strings.Index(output, "- middle:")
	zebraPos := strings.Index(output, "- zebra:")

	if alphaPos == -1 || middlePos == -1 || zebraPos == -1 {
		t.Fatal("expected all playbooks to be in output")
	}

	// Check they're in alphabetical order
	if !(alphaPos < middlePos && middlePos < zebraPos) {
		t.Error("expected playbooks to be sorted alphabetically")
	}
}

func TestPrintPlaybook_Success(t *testing.T) {
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
	err := PrintPlaybook(&buf, reg, "test-doc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	expectedContent := "# Test Content\n\nThis is a test.\n"
	if output != expectedContent {
		t.Errorf("expected output:\n%s\ngot:\n%s", expectedContent, output)
	}
}

func TestPrintPlaybook_NotFound(t *testing.T) {
	reg := registry.BuildRegistry(nil, nil, &config.ProjectConfig{})

	var buf bytes.Buffer
	err := PrintPlaybook(&buf, reg, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent playbook")
	}

	if !strings.Contains(err.Error(), "unknown playbook") {
		t.Errorf("expected 'unknown playbook' error, got: %v", err)
	}
}

func TestPrintPlaybook_OnlyContent(t *testing.T) {
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
	err := PrintPlaybook(&buf, reg, "doc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	if strings.Contains(output, "This should not appear in output") {
		t.Error("description should not be in playbook output")
	}

	if !strings.Contains(output, "Only this content should appear") {
		t.Error("expected content to be in playbook output")
	}
}

func TestOneLineDescription(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "simple text",
			text:     "Hello world",
			expected: "Hello world",
		},
		{
			name:     "empty text",
			text:     "",
			expected: "(no description)",
		},
		{
			name:     "multiline text",
			text:     "Line 1\nLine 2\tLine 3",
			expected: "Line 1 Line 2 Line 3",
		},
		{
			name:     "excess whitespace",
			text:     "   spaced   out   ",
			expected: "spaced out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := oneLineDescription(tt.text)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestPrintHelp_PlaybookListing(t *testing.T) {
	docs := []parser.Document{
		{Name: "test", Description: "Test description", Source: parser.SourceProjectScoped},
	}

	reg := registry.BuildRegistry(nil, docs, &config.ProjectConfig{})

	var buf bytes.Buffer
	PrintHelp(&buf, reg)

	output := buf.String()

	expectedLine := "- test: Test description"
	if !strings.Contains(output, expectedLine) {
		t.Errorf("expected playbook listing %q in output, got:\n%s", expectedLine, output)
	}
}
