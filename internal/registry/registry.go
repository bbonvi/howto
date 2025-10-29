package registry

import (
	"sort"

	"github.com/yourusername/howto/internal/config"
	"github.com/yourusername/howto/internal/parser"
)

// Registry maps playbook names to their documentation
type Registry map[string]parser.Document

// BuildRegistry creates a unified playbook registry with filtering logic
// Rules:
// 1. Always include all project-scoped docs
// 2. For global docs:
//   - Include if required=true (default)
//   - Include if required=false AND name is in projectConfig.Require
//   - Exclude if required=false AND name is NOT in projectConfig.Require
//
// 3. If name conflicts: project-scoped overrides global
func BuildRegistry(globalDocs, projectDocs []parser.Document, projectConfig *config.ProjectConfig) Registry {
	registry := make(Registry)

	// First, add global docs based on filtering rules
	for _, doc := range globalDocs {
		// Skip if required=false and not in project config require list
		if !doc.Required && !projectConfig.HasRequire(doc.Name) {
			continue
		}

		registry[doc.Name] = doc
	}

	// Then, add project-scoped docs (they override global docs with same name)
	for _, doc := range projectDocs {
		registry[doc.Name] = doc
	}

	return registry
}

// Get retrieves a document by name
func (r Registry) Get(name string) (parser.Document, bool) {
	doc, ok := r[name]
	return doc, ok
}

// List returns all document names sorted alphabetically
func (r Registry) List() []string {
	names := make([]string, 0, len(r))
	for name := range r {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetAll returns all documents sorted by name
func (r Registry) GetAll() []parser.Document {
	names := r.List()
	docs := make([]parser.Document, 0, len(names))
	for _, name := range names {
		docs = append(docs, r[name])
	}
	return docs
}

// Count returns the number of documents in the registry
func (r Registry) Count() int {
	return len(r)
}

// Has checks if a document with the given name exists
func (r Registry) Has(name string) bool {
	_, ok := r[name]
	return ok
}
