package parser

import (
	"testing"
)

func TestParseContent_Valid(t *testing.T) {
	content := []byte(`---
name: rust-lang
description: Documentation for Rust projects
---

# Rust Design Principles
- Prefer simplicity over cleverness`)

	doc, err := ParseContent(content, "test.md", SourceGlobal, "/test/test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Name != "rust-lang" {
		t.Errorf("expected name 'rust-lang', got '%s'", doc.Name)
	}
	if doc.Description != "Documentation for Rust projects" {
		t.Errorf("expected description 'Documentation for Rust projects', got '%s'", doc.Description)
	}
	if !doc.Required {
		t.Errorf("expected required to be true by default")
	}
	if doc.Source != SourceGlobal {
		t.Errorf("expected source to be Global")
	}
	expectedContent := "# Rust Design Principles\n- Prefer simplicity over cleverness"
	if doc.Content != expectedContent {
		t.Errorf("expected content:\n%s\ngot:\n%s", expectedContent, doc.Content)
	}
}

func TestParseContent_DefaultName(t *testing.T) {
	content := []byte(`---
description: Test documentation
---

Content here`)

	doc, err := ParseContent(content, "my-doc.md", SourceProjectScoped, "/test/my-doc.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Name != "my-doc" {
		t.Errorf("expected name to default to 'my-doc' (filename without .md), got '%s'", doc.Name)
	}
}

func TestParseContent_RequiredFalse(t *testing.T) {
	content := []byte(`---
name: optional-doc
description: Optional documentation
required: false
---

Content`)

	doc, err := ParseContent(content, "test.md", SourceGlobal, "/test/test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Required {
		t.Errorf("expected required to be false, got true")
	}
}

func TestParseContent_RequiredTrue(t *testing.T) {
	content := []byte(`---
name: required-doc
description: Required documentation
required: true
---

Content`)

	doc, err := ParseContent(content, "test.md", SourceGlobal, "/test/test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !doc.Required {
		t.Errorf("expected required to be true")
	}
}

func TestParseContent_MissingDescription(t *testing.T) {
	content := []byte(`---
name: test
---

Content`)

	_, err := ParseContent(content, "test.md", SourceGlobal, "/test/test.md")
	if err == nil {
		t.Fatal("expected error for missing description")
	}
}

func TestParseContent_MissingFrontmatterStart(t *testing.T) {
	content := []byte(`name: test
description: Test
---

Content`)

	_, err := ParseContent(content, "test.md", SourceGlobal, "/test/test.md")
	if err == nil {
		t.Fatal("expected error for missing frontmatter start delimiter")
	}
}

func TestParseContent_MissingFrontmatterEnd(t *testing.T) {
	content := []byte(`---
name: test
description: Test

Content`)

	_, err := ParseContent(content, "test.md", SourceGlobal, "/test/test.md")
	if err == nil {
		t.Fatal("expected error for missing frontmatter end delimiter")
	}
}

func TestParseContent_EmptyBody(t *testing.T) {
	content := []byte(`---
name: test
description: Test documentation
---`)

	doc, err := ParseContent(content, "test.md", SourceGlobal, "/test/test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Content != "" {
		t.Errorf("expected empty content, got '%s'", doc.Content)
	}
}

func TestParseContent_CRLFLineEndings(t *testing.T) {
	// Test with Windows-style line endings
	content := []byte("---\r\nname: test\r\ndescription: Test documentation\r\n---\r\n\r\nContent here")

	doc, err := ParseContent(content, "test.md", SourceGlobal, "/test/test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", doc.Name)
	}
	if doc.Content != "Content here" {
		t.Errorf("expected content 'Content here', got '%s'", doc.Content)
	}
}

func TestParseContent_ComplexMarkdown(t *testing.T) {
	content := []byte(`---
name: complex
description: Complex markdown test
---

# Heading 1

## Heading 2

- List item 1
- List item 2

` + "```go\nfunc main() {}\n```" + `

More content`)

	doc, err := ParseContent(content, "test.md", SourceGlobal, "/test/test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedContent := "# Heading 1\n\n## Heading 2\n\n- List item 1\n- List item 2\n\n```go\nfunc main() {}\n```\n\nMore content"
	if doc.Content != expectedContent {
		t.Errorf("content mismatch.\nExpected:\n%s\n\nGot:\n%s", expectedContent, doc.Content)
	}
}

func TestSource_String(t *testing.T) {
	tests := []struct {
		source   Source
		expected string
	}{
		{SourceGlobal, "global"},
		{SourceProjectScoped, "project"},
		{Source(999), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.source.String(); got != tt.expected {
			t.Errorf("Source(%d).String() = %s, want %s", tt.source, got, tt.expected)
		}
	}
}

func TestParseContent_WhitespaceHandling(t *testing.T) {
	content := []byte(`---
name: whitespace-test
description: Test whitespace handling
---



# Content with leading blank lines


`)

	doc, err := ParseContent(content, "test.md", SourceGlobal, "/test/test.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Content should be trimmed
	expectedContent := "# Content with leading blank lines"
	if doc.Content != expectedContent {
		t.Errorf("expected content:\n'%s'\ngot:\n'%s'", expectedContent, doc.Content)
	}
}
