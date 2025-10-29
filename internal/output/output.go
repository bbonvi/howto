package output

import (
	"fmt"
	"io"
	"strings"

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
	fmt.Fprintln(w, "- Start every new task by running `howto` to refresh the available playbooks.")
	fmt.Fprintln(w, "- Pull each relevant playbook with `howto <playbook>` before you act, and follow it literally.")
	fmt.Fprintln(w, "- Treat every instruction from `howto` as mandatory; do not improvise beyond the playbook.")
	fmt.Fprintln(w, "- Stop or escalate if a playbook tells you to pause, ask questions, or hand off to a human.")
	fmt.Fprintln(w, "- When the user shifts focus, rerun `howto` and reload the playbooks that now apply.")
	fmt.Fprintln(w, "- If any call to `howto` fails, report the error instead of guessing; the maintainer needs that signal.")
	fmt.Fprintln(w, "- Limit yourself to playbooks you truly need, but batch them with `&&` if several apply.")
	fmt.Fprintln(w, "- Reissue `howto` whenever you need a refresher during the session.")
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
