package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Source indicates where a document came from
type Source int

const (
	SourceGlobal Source = iota
	SourceProjectScoped
)

func (s Source) String() string {
	switch s {
	case SourceGlobal:
		return "global"
	case SourceProjectScoped:
		return "project"
	default:
		return "unknown"
	}
}

// Document represents a parsed markdown file with YAML frontmatter
type Document struct {
	Name        string // From frontmatter or filename
	Description string // Required field
	Required    bool   // Default: true (global only)
	Content     string // Markdown body (no frontmatter)
	Source      Source // Global or ProjectScoped
	FilePath    string // Original file path for debugging
}

// frontmatter represents the YAML metadata structure
type frontmatter struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Required    *bool  `yaml:"required"` // Pointer to distinguish unset vs false
}

// ParseFile reads and parses a markdown file with YAML frontmatter
func ParseFile(path string, source Source) (*Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return ParseContent(content, filepath.Base(path), source, path)
}

// ParseContent parses markdown content with YAML frontmatter
func ParseContent(content []byte, filename string, source Source, filepath string) (*Document, error) {
	// Extract frontmatter and body
	fm, body, err := extractFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("failed to extract frontmatter: %w", err)
	}

	// Parse YAML frontmatter
	var meta frontmatter
	if err := yaml.Unmarshal(fm, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	// Validate required fields
	if meta.Description == "" {
		return nil, fmt.Errorf("missing required field: description")
	}

	// Build document
	doc := &Document{
		Name:        meta.Name,
		Description: meta.Description,
		Required:    true, // Default
		Content:     string(body),
		Source:      source,
		FilePath:    filepath,
	}

	// Apply defaults
	if doc.Name == "" {
		// Default to filename without .md extension
		doc.Name = strings.TrimSuffix(filename, ".md")
	}

	// Handle required field
	if meta.Required != nil {
		doc.Required = *meta.Required
	}

	return doc, nil
}

// extractFrontmatter separates YAML frontmatter from markdown content
// Expected format:
// ---
// yaml: content
// ---
// markdown content
func extractFrontmatter(content []byte) (frontmatter []byte, body []byte, err error) {
	// Check if content starts with ---
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return nil, nil, fmt.Errorf("missing frontmatter delimiter at start")
	}

	// Find the start position (after first ---)
	start := 3
	if bytes.HasPrefix(content, []byte("---\r\n")) {
		start = 5
	} else {
		start = 4 // "---\n"
	}

	// Find the closing --- delimiter
	remaining := content[start:]
	endDelimIndex := bytes.Index(remaining, []byte("\n---\n"))
	if endDelimIndex == -1 {
		endDelimIndex = bytes.Index(remaining, []byte("\r\n---\r\n"))
		if endDelimIndex == -1 {
			endDelimIndex = bytes.Index(remaining, []byte("\n---\r\n"))
			if endDelimIndex == -1 {
				// Check if file ends with just \n--- (no content after)
				if bytes.HasSuffix(remaining, []byte("\n---")) {
					endDelimIndex = len(remaining) - 4 // Position before \n---
				} else if bytes.HasSuffix(remaining, []byte("\r\n---")) {
					endDelimIndex = len(remaining) - 5 // Position before \r\n---
				} else {
					return nil, nil, fmt.Errorf("missing closing frontmatter delimiter")
				}
			}
		}
	}

	// Extract frontmatter (between the --- delimiters)
	frontmatter = remaining[:endDelimIndex]

	// Find where body starts (after closing ---)
	bodyStartIndex := start + endDelimIndex
	// Skip past the closing delimiter and newline
	for bodyStartIndex < len(content) && (content[bodyStartIndex] == '\n' || content[bodyStartIndex] == '\r' || content[bodyStartIndex] == '-') {
		bodyStartIndex++
	}

	// Extract body
	if bodyStartIndex < len(content) {
		body = bytes.TrimSpace(content[bodyStartIndex:])
	}

	return frontmatter, body, nil
}
