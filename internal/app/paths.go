package app

import (
	"fmt"
	"os"
	"path/filepath"
)

// GlobalConfigDir returns the global configuration directory path.
// Default: ~/.config/howto/
func GlobalConfigDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("HOME environment variable not set")
	}

	return filepath.Join(home, ".config", "howto"), nil
}

// ProjectConfigDir returns the project-scoped configuration directory path.
// Default: $(pwd)/.howto/
func ProjectConfigDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	return ProjectConfigDirFrom(cwd), nil
}

// ProjectConfigDirFrom returns the project-scoped configuration directory path for the provided working directory.
func ProjectConfigDirFrom(cwd string) string {
	return filepath.Join(cwd, ".howto")
}
