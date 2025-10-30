package main

import (
	"fmt"
	"os"

	"github.com/yourusername/howto/internal/app"
	"github.com/yourusername/howto/internal/output"
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
	globalPath, err := app.GlobalConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get global config path: %w", err)
	}

	projectPath, err := app.ProjectConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get project path: %w", err)
	}

	// Build registry
	reg, err := app.LoadRegistry(globalPath, projectPath)
	if err != nil {
		return err
	}

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
