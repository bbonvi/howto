package loader

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/howto/internal/parser"
)

// LoadGlobalDocs loads all markdown documentation from the global config directory
func LoadGlobalDocs(configDir string) ([]parser.Document, error) {
	return loadDocs(configDir, parser.SourceGlobal)
}

// LoadProjectDocs loads all markdown documentation from the project-scoped directory
func LoadProjectDocs(projectDir string) ([]parser.Document, error) {
	return loadDocs(projectDir, parser.SourceProjectScoped)
}

// loadDocs loads all markdown files from a directory
func loadDocs(dir string, source parser.Source) ([]parser.Document, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Directory doesn't exist - not an error, just return empty slice
		return []parser.Document{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat directory %s: %w", dir, err)
	}

	var docs []parser.Document
	var loadErrors []string

	// Walk directory and find all .md files
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Log warning but continue walking
			loadErrors = append(loadErrors, fmt.Sprintf("error accessing path %s: %v", path, err))
			return nil
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .md files
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		// Parse the file
		doc, err := parser.ParseFile(path, source)
		if err != nil {
			// Log warning but continue processing other files
			loadErrors = append(loadErrors, fmt.Sprintf("failed to parse %s: %v", path, err))
			return nil
		}

		docs = append(docs, *doc)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	// If there were load errors, we could log them here
	// For now, we'll just silently continue (graceful degradation)
	// In a production version, you might want to use a logger
	_ = loadErrors

	return docs, nil
}

// GetLoadErrors can be used to retrieve errors that occurred during loading
// This is useful for debugging or verbose mode
func GetLoadErrors(dir string, source parser.Source) []string {
	var loadErrors []string

	// Walk directory and collect errors
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			loadErrors = append(loadErrors, fmt.Sprintf("error accessing path %s: %v", path, err))
			return nil
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		_, err = parser.ParseFile(path, source)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Sprintf("failed to parse %s: %v", path, err))
		}

		return nil
	})

	return loadErrors
}
