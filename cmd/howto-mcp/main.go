package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yourusername/howto/internal/app"
	"github.com/yourusername/howto/internal/mcp"
)

var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	globalDir, err := app.GlobalConfigDir()
	if err != nil {
		return fmt.Errorf("failed to resolve global config directory: %w", err)
	}

	projectDir, err := app.ProjectConfigDir()
	if err != nil {
		return fmt.Errorf("failed to resolve project config directory: %w", err)
	}

	loader := app.NewCachedRegistryLoader(globalDir, projectDir)
	logger := log.New(os.Stderr, "howto-mcp: ", log.LstdFlags)

	server := mcp.NewServer(os.Stdin, os.Stdout, loader, version, logger)
	return server.Serve()
}
