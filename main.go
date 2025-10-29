package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yourusername/howto/internal/config"
	"github.com/yourusername/howto/internal/loader"
	"github.com/yourusername/howto/internal/output"
	"github.com/yourusername/howto/internal/registry"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse playbook arguments
	args := os.Args[1:] // Skip program name

	if len(args) > 1 {
		return fmt.Errorf("too many arguments (expected 0 or 1, got %d)", len(args))
	}

	if len(args) == 1 && args[0] == "--version" {
		fmt.Fprintln(os.Stdout, version)
		return nil
	}

	// Resolve paths
	globalPath, err := getGlobalConfigPath()
	if err != nil {
		return fmt.Errorf("failed to get global config path: %w", err)
	}

	projectPath, err := getProjectPath()
	if err != nil {
		return fmt.Errorf("failed to get project path: %w", err)
	}

	// Load documents
	globalDocs, err := loader.LoadGlobalDocs(globalPath)
	if err != nil {
		return fmt.Errorf("failed to load global docs: %w", err)
	}

	projectDocs, err := loader.LoadProjectDocs(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load project docs: %w", err)
	}

	// Load project config
	projectConfig, err := config.LoadProjectConfig(projectPath)
	if err != nil {
		return fmt.Errorf("failed to load project config: %w", err)
	}

	// Build registry
	reg := registry.BuildRegistry(globalDocs, projectDocs, projectConfig)

	if len(args) == 0 {
		// No arguments - print help
		output.PrintHelp(os.Stdout, reg)
		return nil
	}

	// Print specific playbook
	playbookName := args[0]
	if err := output.PrintPlaybook(os.Stdout, reg, playbookName); err != nil {
		return err
	}

	return nil
}

// getGlobalConfigPath returns the global config directory path
// Default: ~/.config/howto/
func getGlobalConfigPath() (string, error) {
	// Check for HOME environment variable
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("HOME environment variable not set")
	}

	return filepath.Join(home, ".config", "howto"), nil
}

// getProjectPath returns the project-scoped config directory path
// Default: $(pwd)/.howto/
func getProjectPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	return filepath.Join(cwd, ".howto"), nil
}
