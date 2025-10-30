package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCachedRegistryLoaderReloadsOnFileChange(t *testing.T) {
	tempDir := t.TempDir()
	globalDir := filepath.Join(tempDir, "global")
	projectDir := filepath.Join(tempDir, "project")

	mustMkdir(t, globalDir)
	mustMkdir(t, projectDir)

	docPath := filepath.Join(globalDir, "sample.md")
	writeDoc(t, docPath, "sample", "Initial description", "first version")

	loader := NewCachedRegistryLoader(globalDir, projectDir)

	reg1, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	doc1, ok := reg1.Get("sample")
	if !ok {
		t.Fatalf("expected playbook sample to exist")
	}
	if doc1.Content != "first version" {
		t.Fatalf("unexpected content: %q", doc1.Content)
	}

	// Second call should hit the cache and return identical content.
	reg2, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() failed on second call: %v", err)
	}
	doc2, ok := reg2.Get("sample")
	if !ok {
		t.Fatalf("expected playbook sample to exist on second load")
	}
	if doc2.Content != "first version" {
		t.Fatalf("unexpected content on second load: %q", doc2.Content)
	}

	// Update file content to force a reload.
	time.Sleep(20 * time.Millisecond) // ensure modtime changes across filesystems
	writeDoc(t, docPath, "sample", "Initial description", "updated version")

	reg3, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() failed after update: %v", err)
	}
	doc3, ok := reg3.Get("sample")
	if !ok {
		t.Fatalf("expected playbook sample to exist after update")
	}
	if doc3.Content != "updated version" {
		t.Fatalf("expected updated content, got %q", doc3.Content)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
}

func writeDoc(t *testing.T, path, name, description, body string) {
	t.Helper()

	content := []byte("---\n")
	content = append(content, []byte("name: "+name+"\n")...)
	content = append(content, []byte("description: "+description+"\n")...)
	content = append(content, []byte("---\n")...)
	content = append(content, []byte(body)...)

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("failed to write doc %s: %v", path, err)
	}
}
