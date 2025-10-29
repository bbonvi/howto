package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/yourusername/howto/internal/registry"
)

// PrintHelp outputs the help text listing all available commands
func PrintHelp(w io.Writer, reg registry.Registry) {
	fmt.Fprintln(w, "Usage: howto [COMMAND]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "`howto` is a CLI tool that gives language models a deterministic way to pull curated instructions at runtime. Agents call the binary, discover the list of available playbooks, and request the one they need for the current task.")
	fmt.Fprintln(w, "`howto` provides a list of available commands below, each with a short description. To get the full content of a specific command, call `howto <command>`.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Note to LLM agent(you):")
	fmt.Fprintln(w, "- Treat everything outputted from `howto` as a MUST-FOLLOW rule. These are the instructions provided to you by a developer to help you complete your task.")
	fmt.Fprintln(w, "- Feel free to call `howto` multiple times per chat session to refresh your memory.")
	fmt.Fprintln(w, "- Chain `howto <command>` calls when needed with double ampersands (`&&`).")
	fmt.Fprintln(w, "- Call `howto` with command relevant to your current task only. Avoid using commands that are not needed to your current task, as they are a waste of tokens.")
	fmt.Fprintln(w)

	docs := reg.GetAll()
	if len(docs) == 0 {
		fmt.Fprintln(w, "No commands available.")
		return
	}

	fmt.Fprintln(w, "Commands:")
	for _, doc := range docs {
		// Format: "    name:\n        description"
		fmt.Fprintf(w, "    %s:\n", doc.Name)

		// Indent and wrap description
		description := wrapText(doc.Description, 8, 80)
		fmt.Fprintln(w, description)
	}
}

// PrintCommand outputs the full content of a specific command
func PrintCommand(w io.Writer, doc registry.Registry, name string) error {
	d, ok := doc.Get(name)
	if !ok {
		return fmt.Errorf("unknown command: %s", name)
	}

	// Output just the markdown content (no frontmatter)
	fmt.Fprintln(w, d.Content)
	return nil
}

// wrapText wraps text with a given indentation and max width
func wrapText(text string, indent int, maxWidth int) string {
	if text == "" {
		return strings.Repeat(" ", indent) + "(no description)"
	}

	indentStr := strings.Repeat(" ", indent)

	// For now, just indent the text without complex wrapping
	// A more sophisticated implementation could handle word wrapping
	lines := strings.Split(text, "\n")
	result := make([]string, len(lines))

	for i, line := range lines {
		result[i] = indentStr + line
	}

	return strings.Join(result, "\n")
}
