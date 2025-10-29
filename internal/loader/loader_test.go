package loader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/howto/internal/parser"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "howto-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return tmpDir
}

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}

func TestLoadGlobalDocs_Success(t *testing.T) {
	tmpDir := setupTestDir(t)

	// Create test files
	writeTestFile(t, filepath.Join(tmpDir, "rust-lang.md"), `---
name: rust-lang
description: Rust documentation
---

# Rust content`)

	writeTestFile(t, filepath.Join(tmpDir, "go-lang.md"), `---
description: Go documentation
---

# Go content`)

	docs, err := LoadGlobalDocs(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docs) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(docs))
	}

	// Check that both docs were loaded
	names := make(map[string]bool)
	for _, doc := range docs {
		names[doc.Name] = true
		if doc.Source != parser.SourceGlobal {
			t.Errorf("expected source to be Global, got %s", doc.Source)
		}
	}

	if !names["rust-lang"] {
		t.Error("expected rust-lang doc to be loaded")
	}
	if !names["go-lang"] {
		t.Error("expected go-lang doc to be loaded")
	}
}

func TestLoadProjectDocs_Success(t *testing.T) {
	tmpDir := setupTestDir(t)

	writeTestFile(t, filepath.Join(tmpDir, "commits.md"), `---
name: commits
description: Commit guidelines
---

# Commit rules`)

	docs, err := LoadProjectDocs(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("expected 1 doc, got %d", len(docs))
	}

	if docs[0].Name != "commits" {
		t.Errorf("expected name 'commits', got '%s'", docs[0].Name)
	}
	if docs[0].Source != parser.SourceProjectScoped {
		t.Errorf("expected source to be ProjectScoped, got %s", docs[0].Source)
	}
}

func TestLoadDocs_NonExistentDirectory(t *testing.T) {
	docs, err := LoadGlobalDocs("/nonexistent/directory/that/does/not/exist")
	if err != nil {
		t.Fatalf("expected no error for nonexistent directory, got: %v", err)
	}

	if len(docs) != 0 {
		t.Errorf("expected empty docs for nonexistent directory, got %d docs", len(docs))
	}
}

func TestLoadDocs_EmptyDirectory(t *testing.T) {
	tmpDir := setupTestDir(t)

	docs, err := LoadGlobalDocs(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docs) != 0 {
		t.Errorf("expected empty docs for empty directory, got %d docs", len(docs))
	}
}

func TestLoadDocs_IgnoresNonMarkdownFiles(t *testing.T) {
	tmpDir := setupTestDir(t)

	// Create valid markdown file
	writeTestFile(t, filepath.Join(tmpDir, "valid.md"), `---
description: Valid doc
---
Content`)

	// Create non-markdown files
	writeTestFile(t, filepath.Join(tmpDir, "readme.txt"), "Not a markdown file")
	writeTestFile(t, filepath.Join(tmpDir, "config.yaml"), "key: value")

	docs, err := LoadGlobalDocs(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docs) != 1 {
		t.Errorf("expected 1 doc (only .md file), got %d", len(docs))
	}
}

func TestLoadDocs_HandlesInvalidFiles(t *testing.T) {
	tmpDir := setupTestDir(t)

	// Create valid file
	writeTestFile(t, filepath.Join(tmpDir, "valid.md"), `---
description: Valid doc
---
Content`)

	// Create invalid file (missing description)
	writeTestFile(t, filepath.Join(tmpDir, "invalid.md"), `---
name: invalid
---
Content`)

	docs, err := LoadGlobalDocs(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have 1 doc (the valid one)
	if len(docs) != 1 {
		t.Errorf("expected 1 doc (invalid file should be skipped), got %d", len(docs))
	}

	if docs[0].Name != "valid" {
		t.Errorf("expected 'valid' doc to be loaded, got '%s'", docs[0].Name)
	}
}

func TestLoadDocs_Subdirectories(t *testing.T) {
	tmpDir := setupTestDir(t)

	// Create files in subdirectories
	writeTestFile(t, filepath.Join(tmpDir, "root.md"), `---
description: Root doc
---
Root`)

	writeTestFile(t, filepath.Join(tmpDir, "subdir", "nested.md"), `---
description: Nested doc
---
Nested`)

	docs, err := LoadGlobalDocs(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docs) != 2 {
		t.Fatalf("expected 2 docs (root + nested), got %d", len(docs))
	}

	names := make(map[string]bool)
	for _, doc := range docs {
		names[doc.Name] = true
	}

	if !names["root"] {
		t.Error("expected root doc to be loaded")
	}
	if !names["nested"] {
		t.Error("expected nested doc to be loaded")
	}
}

func TestLoadDocs_CaseInsensitiveMdExtension(t *testing.T) {
	tmpDir := setupTestDir(t)

	// Create files with different case extensions
	writeTestFile(t, filepath.Join(tmpDir, "lower.md"), `---
description: Lower case
---
Content`)

	writeTestFile(t, filepath.Join(tmpDir, "upper.MD"), `---
description: Upper case
---
Content`)

	writeTestFile(t, filepath.Join(tmpDir, "mixed.Md"), `---
description: Mixed case
---
Content`)

	docs, err := LoadGlobalDocs(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docs) != 3 {
		t.Errorf("expected 3 docs (all .md variations), got %d", len(docs))
	}
}

func TestLoadDocs_RequiredField(t *testing.T) {
	tmpDir := setupTestDir(t)

	writeTestFile(t, filepath.Join(tmpDir, "required-true.md"), `---
description: Required true
required: true
---
Content`)

	writeTestFile(t, filepath.Join(tmpDir, "required-false.md"), `---
description: Required false
required: false
---
Content`)

	docs, err := LoadGlobalDocs(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(docs) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(docs))
	}

	for _, doc := range docs {
		if doc.Name == "required-true" && !doc.Required {
			t.Error("expected required-true to have Required=true")
		}
		if doc.Name == "required-false" && doc.Required {
			t.Error("expected required-false to have Required=false")
		}
	}
}
