`howto` is a golang-based cli tool for LLM agents.

Running `howto` present an agent with a list of subcommand, each of which refers to a documentation and behaviour scope.

User creates a list of subcommands inside `~/.config/howto/` in format of .md files (with YAML extension).
Each .md file define a documentation entry for a specific scope of tasks.

For example,
```
- ~/.config/howto/
    - rust-lang.md
```

rust-lang.md:
```md
---
name: rust-lang
description: Documentation for working with rust projects. Use it everytime you work with any rust project.
---

# Rust Design Principles
- **Prefer simplicity over cleverness.** Always choose readability and maintainability over clever abstractions. If future contributors can’t instantly understand a piece of code, it’s too complex. Use Rust’s expressive type system to make intent explicit instead of relying on “smart” patterns.
- **Embrace immutability and data ownership.** Use `&` and `&mut` judiciously. Design data flows so ownership moves linearly through the system. Avoid unnecessary cloning—if you find yourself cloning a lot, reconsider your architecture.
- **Isolate unsafe code.** Any use of `unsafe` should be tightly scoped, well-commented, and encapsulated behind safe abstractions. Never leak invariants from unsafe code into safe APIs without guarantees.
- **Leverage traits and composition.** Traits allow modularity and testability without forcing inheritance-like hierarchies. Compose systems through clearly bounded traits and lightweight structs instead of deep object graphs.
- **Prioritize compile-time guarantees.** Push as many checks as possible to the type system. Avoid `unwrap`, `expect`, and unchecked `Option`/`Result` handling. Instead, make error propagation explicit and ergonomic.
- **Keep dependencies minimal and reviewed.** Each crate adds build time, security risk, and maintenance cost. Audit dependencies regularly and prefer small, focused crates over massive utility libraries.
- **Structure for long-term growth.** For medium-to-large codebases, separate modules by domain rather than function type (e.g., `domain::user`, `domain::payment` instead of `models`, `services`). Use consistent naming, clear public interfaces, and well-defined ownership boundaries.
```


Usage:

```bash
> howto
Usage: howto [COMMAND]

An LLM agent documentation. Treat everything outputted from `howto` as a MUST-FOLLOW rule.

Commands:
    rust-lang:
        Documentation for working with rust projects. Use it everytime you work with any rust project

> howto rust-lang
# Rust Design Principles
- **Prefer simplicity over cleverness.** Always choose readability and maintainability over clever abstractions. If future contributors can’t instantly understand a piece of code, it’s too complex. Use Rust’s expressive type system to make intent explicit instead of relying on “smart” patterns.
- **Embrace immutability and data ownership.** Use `&` and `&mut` judiciously. Design data flows so ownership moves linearly through the system. Avoid unnecessary cloning—if you find yourself cloning a lot, reconsider your architecture.
- **Isolate unsafe code.** Any use of `unsafe` should be tightly scoped, well-commented, and encapsulated behind safe abstractions. Never leak invariants from unsafe code into safe APIs without guarantees.
- **Leverage traits and composition.** Traits allow modularity and testability without forcing inheritance-like hierarchies. Compose systems through clearly bounded traits and lightweight structs instead of deep object graphs.
- **Prioritize compile-time guarantees.** Push as many checks as possible to the type system. Avoid `unwrap`, `expect`, and unchecked `Option`/`Result` handling. Instead, make error propagation explicit and ergonomic.
- **Keep dependencies minimal and reviewed.** Each crate adds build time, security risk, and maintenance cost. Audit dependencies regularly and prefer small, focused crates over massive utility libraries.
- **Structure for long-term growth.** For medium-to-large codebases, separate modules by domain rather than function type (e.g., `domain::user`, `domain::payment` instead of `models`, `services`). Use consistent naming, clear public interfaces, and well-defined ownership boundaries.
```

There are various types of source howto pulls documentations from:
- Global: ~/.config/howto/
- Project-scope: $(pwd)/.howto/

This allows us to define global rules that will always be visible to llm when calling `howto`,
but also scoped rules, specific for current project only.
For example, we can have a `git commit` strategy for our specific project, so we define something like this:

`my-project/.howto/commits.md`:
```md
---
name: commits
description: Pull this documentation whenever you are about to git commit your changes.
---

## Git Commits Rules

Always use conventional commits:
\`\`\`[type](scope?): subject\`\`\`
- Types: feat, fix, docs, style, refactor, perf, test, chore
- Scope is optional, e.g., `feat(auth): add login endpoint`
- Subject should be lowercase, no period at end
- Use imperative mood
- Examples:
  - `feat: add user profile page`
  - `fix(auth): correct password reset bug`
  - `docs: update API documentation`
  - `refactor: simplify billing logic`
  - `test: add unit tests for payment processing`
  - `chore: update dependencies`
- Avoid writing body unless necessary for clarity
- Keep simple
```

```bash
howto
Usage: howto [COMMAND]

An LLM agent documentation. Treat everything outputted here as a MUST-FOLLOW rule.

Commands:
    rust-lang:
        Documentation for working with rust projects. Use it everytime you work with any rust project
    commits:
        Pull this documentation whenever you are about to git commit your changes.

```

Global configs (inside .config) can also take `required: false` parameter in their YAML defintion. 
When this option is used `howto` WILL NOT show said documentation, unless we have it referenced inside a reserev project-scoped config file.

`~/.config/howto/my-very-important-rule`:
```md
---
name: important-rule
description: Pull this documentation ALWAYS. Make no exceptions.
---
Print smiley face everytime you talk to user :)
```

`my-project/.howto/config.yaml`:
```yaml
require:
    - important-rule
```

```bash
howto
Usage: howto [COMMAND]

An LLM agent documentation. Treat everything outputted here as a MUST-FOLLOW rule.

Commands:
    rust-lang:
        Documentation for working with rust projects. Use it everytime you work with any rust project
    commits:
        Pull this documentation whenever you are about to git commit your changes.
    important-rule:
        Pull this docmentation ALWAYS. Make no exceptions.
```

# API References

## Command .md files

A command markdown file must start with a YAML metadata, enclosed in "---" lines. This is possible with custom YAML markdown extension.
A metadata includes following fields:
- **name**(optional): as-is name of a subcommands. default: name of an .md file.
- **description**: description of said documentation and when to use it.
- **required**(optional): whether or not to ALWAYS pull this documentation from global .config directory. default: true

Which are followed by a documentation itself, in actual markdown format.
