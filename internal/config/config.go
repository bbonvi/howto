package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProjectConfig represents the .howto/config.yaml structure
type ProjectConfig struct {
	Require []string `yaml:"require"`
}

// LoadProjectConfig loads the project-scoped config.yaml file
// Returns empty config if file doesn't exist (not an error)
func LoadProjectConfig(projectDir string) (*ProjectConfig, error) {
	configPath := filepath.Join(projectDir, "config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// No config file - return empty config (not an error)
		return &ProjectConfig{
			Require: []string{},
		}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to stat config file: %w", err)
	}

	// Read config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config ProjectConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config YAML: %w", err)
	}

	// Ensure Require is not nil
	if config.Require == nil {
		config.Require = []string{}
	}

	return &config, nil
}

// HasRequire checks if a specific doc name is in the require list
func (c *ProjectConfig) HasRequire(name string) bool {
	for _, req := range c.Require {
		if req == name {
			return true
		}
	}
	return false
}
