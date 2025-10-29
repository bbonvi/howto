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
	fmt.Fprintln(w, "An LLM agent documentation. Treat everything outputted from `howto` as a MUST-FOLLOW rule. Feel free to call `howto` multiple times a session to refresh your memory.")
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
