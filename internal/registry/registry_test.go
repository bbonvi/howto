package registry

import (
	"testing"

	"github.com/yourusername/howto/internal/config"
	"github.com/yourusername/howto/internal/parser"
)

func TestBuildRegistry_AllProjectDocs(t *testing.T) {
	projectDocs := []parser.Document{
		{Name: "commits", Description: "Commit rules", Source: parser.SourceProjectScoped},
		{Name: "testing", Description: "Test rules", Source: parser.SourceProjectScoped},
	}

	registry := BuildRegistry(nil, projectDocs, &config.ProjectConfig{})

	if registry.Count() != 2 {
		t.Errorf("expected 2 docs, got %d", registry.Count())
	}

	if !registry.Has("commits") {
		t.Error("expected commits doc to be in registry")
	}
	if !registry.Has("testing") {
		t.Error("expected testing doc to be in registry")
	}
}

func TestBuildRegistry_GlobalRequiredTrue(t *testing.T) {
	globalDocs := []parser.Document{
		{Name: "rust-lang", Description: "Rust rules", Required: true, Source: parser.SourceGlobal},
		{Name: "go-lang", Description: "Go rules", Required: true, Source: parser.SourceGlobal},
	}

	registry := BuildRegistry(globalDocs, nil, &config.ProjectConfig{})

	if registry.Count() != 2 {
		t.Errorf("expected 2 docs (both required=true), got %d", registry.Count())
	}

	if !registry.Has("rust-lang") {
		t.Error("expected rust-lang doc to be in registry")
	}
	if !registry.Has("go-lang") {
		t.Error("expected go-lang doc to be in registry")
	}
}

func TestBuildRegistry_GlobalRequiredFalse_NotInConfig(t *testing.T) {
	globalDocs := []parser.Document{
		{Name: "optional-rule", Description: "Optional rule", Required: false, Source: parser.SourceGlobal},
	}

	registry := BuildRegistry(globalDocs, nil, &config.ProjectConfig{})

	if registry.Count() != 0 {
		t.Errorf("expected 0 docs (required=false, not in config), got %d", registry.Count())
	}

	if registry.Has("optional-rule") {
		t.Error("did not expect optional-rule to be in registry")
	}
}

func TestBuildRegistry_GlobalRequiredFalse_InConfig(t *testing.T) {
	globalDocs := []parser.Document{
		{Name: "important-rule", Description: "Important rule", Required: false, Source: parser.SourceGlobal},
	}

	projectConfig := &config.ProjectConfig{
		Require: []string{"important-rule"},
	}

	registry := BuildRegistry(globalDocs, nil, projectConfig)

	if registry.Count() != 1 {
		t.Errorf("expected 1 doc (required=false but in config), got %d", registry.Count())
	}

	if !registry.Has("important-rule") {
		t.Error("expected important-rule to be in registry")
	}
}

func TestBuildRegistry_MixedRequiredField(t *testing.T) {
	globalDocs := []parser.Document{
		{Name: "always-show", Description: "Always show", Required: true, Source: parser.SourceGlobal},
		{Name: "never-show", Description: "Never show", Required: false, Source: parser.SourceGlobal},
		{Name: "show-if-required", Description: "Show if required", Required: false, Source: parser.SourceGlobal},
	}

	projectConfig := &config.ProjectConfig{
		Require: []string{"show-if-required"},
	}

	registry := BuildRegistry(globalDocs, nil, projectConfig)

	if registry.Count() != 2 {
		t.Errorf("expected 2 docs, got %d", registry.Count())
	}

	if !registry.Has("always-show") {
		t.Error("expected always-show to be in registry")
	}
	if registry.Has("never-show") {
		t.Error("did not expect never-show to be in registry")
	}
	if !registry.Has("show-if-required") {
		t.Error("expected show-if-required to be in registry")
	}
}

func TestBuildRegistry_ProjectOverridesGlobal(t *testing.T) {
	globalDocs := []parser.Document{
		{Name: "commits", Description: "Global commit rules", Content: "Global content", Source: parser.SourceGlobal, Required: true},
	}

	projectDocs := []parser.Document{
		{Name: "commits", Description: "Project commit rules", Content: "Project content", Source: parser.SourceProjectScoped},
	}

	registry := BuildRegistry(globalDocs, projectDocs, &config.ProjectConfig{})

	if registry.Count() != 1 {
		t.Errorf("expected 1 doc (project overrides global), got %d", registry.Count())
	}

	doc, ok := registry.Get("commits")
	if !ok {
		t.Fatal("expected commits doc to exist")
	}

	if doc.Source != parser.SourceProjectScoped {
		t.Error("expected project-scoped doc to override global")
	}
	if doc.Content != "Project content" {
		t.Errorf("expected project content, got '%s'", doc.Content)
	}
}

func TestBuildRegistry_Combined(t *testing.T) {
	globalDocs := []parser.Document{
		{Name: "rust-lang", Description: "Rust", Required: true, Source: parser.SourceGlobal},
		{Name: "go-lang", Description: "Go", Required: true, Source: parser.SourceGlobal},
		{Name: "optional", Description: "Optional", Required: false, Source: parser.SourceGlobal},
	}

	projectDocs := []parser.Document{
		{Name: "commits", Description: "Commits", Source: parser.SourceProjectScoped},
		{Name: "testing", Description: "Testing", Source: parser.SourceProjectScoped},
	}

	registry := BuildRegistry(globalDocs, projectDocs, &config.ProjectConfig{})

	// Should have: rust-lang, go-lang, commits, testing (not optional)
	if registry.Count() != 4 {
		t.Errorf("expected 4 docs, got %d", registry.Count())
	}

	expectedDocs := []string{"rust-lang", "go-lang", "commits", "testing"}
	for _, name := range expectedDocs {
		if !registry.Has(name) {
			t.Errorf("expected %s to be in registry", name)
		}
	}

	if registry.Has("optional") {
		t.Error("did not expect optional to be in registry")
	}
}

func TestRegistry_Get(t *testing.T) {
	projectDocs := []parser.Document{
		{Name: "test-doc", Description: "Test", Content: "Test content", Source: parser.SourceProjectScoped},
	}

	registry := BuildRegistry(nil, projectDocs, &config.ProjectConfig{})

	doc, ok := registry.Get("test-doc")
	if !ok {
		t.Fatal("expected doc to exist")
	}

	if doc.Name != "test-doc" {
		t.Errorf("expected name 'test-doc', got '%s'", doc.Name)
	}
	if doc.Content != "Test content" {
		t.Errorf("expected content 'Test content', got '%s'", doc.Content)
	}

	_, ok = registry.Get("nonexistent")
	if ok {
		t.Error("did not expect nonexistent doc to exist")
	}
}

func TestRegistry_List(t *testing.T) {
	projectDocs := []parser.Document{
		{Name: "zeta", Description: "Z", Source: parser.SourceProjectScoped, FilePath: "2-zeta.md"},
		{Name: "alpha", Description: "A", Source: parser.SourceProjectScoped, FilePath: "1-alpha.md"},
		{Name: "middle", Description: "M", Source: parser.SourceProjectScoped, FilePath: "3-middle.md"},
	}

	registry := BuildRegistry(nil, projectDocs, &config.ProjectConfig{})

	names := registry.List()

	expected := []string{"alpha", "zeta", "middle"}
	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(names))
	}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("expected name[%d] = %s, got %s", i, expected[i], name)
		}
	}
}

func TestRegistry_GetAll(t *testing.T) {
	projectDocs := []parser.Document{
		{Name: "zebra", Description: "Z", Source: parser.SourceProjectScoped, FilePath: "2-zebra.md"},
		{Name: "alpha", Description: "A", Source: parser.SourceProjectScoped, FilePath: "1-alpha.md"},
	}

	registry := BuildRegistry(nil, projectDocs, &config.ProjectConfig{})

	docs := registry.GetAll()

	if len(docs) != 2 {
		t.Fatalf("expected 2 docs, got %d", len(docs))
	}

	// Should be sorted by filename
	if docs[0].Name != "alpha" {
		t.Errorf("expected first doc to be 'alpha', got '%s'", docs[0].Name)
	}
	if docs[1].Name != "zebra" {
		t.Errorf("expected second doc to be 'zebra', got '%s'", docs[1].Name)
	}
}

func TestRegistry_Has(t *testing.T) {
	projectDocs := []parser.Document{
		{Name: "exists", Description: "Exists", Source: parser.SourceProjectScoped},
	}

	registry := BuildRegistry(nil, projectDocs, &config.ProjectConfig{})

	if !registry.Has("exists") {
		t.Error("expected 'exists' to be in registry")
	}

	if registry.Has("does-not-exist") {
		t.Error("did not expect 'does-not-exist' to be in registry")
	}
}

func TestRegistry_Count(t *testing.T) {
	tests := []struct {
		name          string
		globalDocs    []parser.Document
		projectDocs   []parser.Document
		projectConfig *config.ProjectConfig
		expectedCount int
	}{
		{
			name:          "empty",
			globalDocs:    nil,
			projectDocs:   nil,
			projectConfig: &config.ProjectConfig{},
			expectedCount: 0,
		},
		{
			name:       "only project docs",
			globalDocs: nil,
			projectDocs: []parser.Document{
				{Name: "doc1", Description: "D1", Source: parser.SourceProjectScoped},
				{Name: "doc2", Description: "D2", Source: parser.SourceProjectScoped},
			},
			projectConfig: &config.ProjectConfig{},
			expectedCount: 2,
		},
		{
			name: "only global docs",
			globalDocs: []parser.Document{
				{Name: "doc1", Description: "D1", Required: true, Source: parser.SourceGlobal},
			},
			projectDocs:   nil,
			projectConfig: &config.ProjectConfig{},
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := BuildRegistry(tt.globalDocs, tt.projectDocs, tt.projectConfig)
			if registry.Count() != tt.expectedCount {
				t.Errorf("expected count %d, got %d", tt.expectedCount, registry.Count())
			}
		})
	}
}
