package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/yourusername/howto/internal/instructions"
	"github.com/yourusername/howto/internal/registry"
)

// PrintHelp outputs the help text listing all available playbooks
func PrintHelp(w io.Writer, reg registry.Registry) {
	fmt.Fprintln(w, "Usage: howto [PLAYBOOK]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "`howto` lets language models pull the exact playbooks their operators prepared.")
	fmt.Fprintln(w, "Run it to list playbooks, then fetch the one you need with `howto <playbook>`.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "LLM operating rules:")
	for _, rule := range instructions.LLMBullets() {
		fmt.Fprintf(w, "- %s\n", rule)
	}
	fmt.Fprintln(w)

	docs := reg.GetAll()
	if len(docs) == 0 {
		fmt.Fprintln(w, "No playbooks available.")
		return
	}

	fmt.Fprintln(w, "Playbooks:")
	for _, doc := range docs {
		description := oneLineDescription(doc.Description)
		fmt.Fprintf(w, "  %s: %s\n", doc.Name, description)
	}
}

// PrintPlaybook outputs the full content of a specific playbook
func PrintPlaybook(w io.Writer, doc registry.Registry, name string) error {
	d, ok := doc.Get(name)
	if !ok {
		return fmt.Errorf("unknown playbook: %s", name)
	}

	// Output just the markdown content (no frontmatter)
	fmt.Fprintln(w, d.Content)
	return nil
}

// oneLineDescription collapses whitespace so the description prints on one line
func oneLineDescription(text string) string {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "(no description)"
	}

	return strings.Join(fields, " ")
}
