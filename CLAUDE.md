# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`howto` is a Go-based CLI tool that provides LLM agents with contextual documentation based on subcommands. It allows users to define documentation scopes that agents should follow when working on specific tasks.

## Architecture

### Documentation Sources
The tool pulls documentation from two locations:
- **Global**: `~/.config/howto/` - documentation available across all projects
- **Project-scoped**: `$(pwd)/.howto/` - documentation specific to the current project

### Documentation Format
Each documentation entry is a markdown file with YAML frontmatter:

```md
---
name: command-name  # optional, defaults to filename
description: When to use this documentation
required: true  # optional, for global configs only (default: true)
---

# Documentation content here
```

### Configuration System
- Project-scoped config at `.howto/config.yaml` can require specific global docs:
```yaml
require:
    - important-rule
```
- Global docs with `required: false` only appear when explicitly required by project config

## Expected Behavior

When `howto` is run without arguments, it lists all available subcommands with their descriptions. When run with a subcommand (e.g., `howto rust-lang`), it outputs the full documentation content.

The tool treats documentation hierarchically:
1. Always shows project-scoped `.howto/` docs
2. Shows global `~/.config/howto/` docs that have `required: true` (default)
3. Shows global docs with `required: false` only if listed in `.howto/config.yaml`

## Development Notes

This is a Go project. When implementing:
- Keep the CLI interface simple and predictable
- Documentation lookup should be fast (no complex parsing)
- Support both global and project-scoped documentation discovery
- Parse YAML frontmatter cleanly from markdown files
- Output should be clean markdown suitable for agent consumption
