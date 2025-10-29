# howto

`howto` is a Go CLI that gives language models a deterministic way to pull curated instructions at runtime. Agents call the binary, discover the list of available playbooks, and request the one they need for the current task. Humans maintain the Markdown libraries; agents consume them.

## Why Agents Like It
- Clean, fixed-format output that is easy for LLM tooling to parse.
- Merge rules let you mix global guidance (applies to every repo) with project overrides.
- YAML front matter validation fails fast when a document is malformed, preventing ambiguous agent responses.
- Ships as a single Go binary with no runtime dependencies beyond `gopkg.in/yaml.v3`.

## Requirements
- Go 1.21 or newer
- Markdown documentation with YAML front matter (schema below)

## Installation
### Prebuilt binaries
- Download the latest release assets for macOS (amd64, arm64) or Linux (amd64, arm64) from the GitHub Releases page.
- Verify the `checksums.txt` file if you need integrity guarantees.
- Mark the binary as executable and place it on your `PATH`, e.g. `chmod +x howto-linux-amd64 && mv howto-linux-amd64 /usr/local/bin/howto`.

### Build from source
```bash
git clone https://github.com/yourusername/howto.git
cd howto
go build ./...
# or install into your GOPATH/bin:
go install ./...
```

## Agent Workflow
```bash
# List available playbooks (agents typically do this once)
howto

# Pull the required playbook before acting
howto <playbook>
```

`howto` exits with a non-zero status if configuration is missing, a document fails to parse, or the requested entry does not exist—surface these errors to the human operator so they can fix the library.

## Documentation Libraries

### Global Library
- Location: `~/.config/howto/`
- Any `.md` file is parsed and considered part of the global catalogue.
- Global entries honour the `required` flag. They are included by default unless the flag is `false` and the project config does not opt in.

### Project Library
- Location: `<project root>/.howto/`
- Markdown files in this directory are always included and override global documents that share the same `name`.
- Optional configuration lives beside the docs in `.howto/config.yaml`.

### Front Matter Schema
Every Markdown file must start with YAML front matter:

```yaml
---
name: optional-custom-playbook-name # defaults to the filename without .md
description: concise explanation shown in `howto` listings (required)
required: true # optional, only evaluated for global documents
---
```

Anything after the closing delimiter is rendered verbatim when the playbook is selected. Missing delimiters or an empty `description` field trigger a parsing error so the problematic document never reaches an agent.

## Project Configuration
Create `.howto/config.yaml` in your project to declare additional requirements:

```yaml
require:
  - important-rule
  - security-checklist
```

Documents listed under `require` are pulled in even if the corresponding global Markdown sets `required: false`. This lets you keep optional guidance in your global library and selectively switch it on for certain codebases.

## Development
- Run tests: `go test ./...`
- Integration fixtures live under `testdata/` and mirror the global/project layout so you can iterate without touching a live agent database.

Feel free to open issues or pull requests with ideas for new features or improvements.
