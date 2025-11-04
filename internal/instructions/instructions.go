package instructions

import "strings"

var llmRules = []string{
	"Start every new task by running `howto` to refresh the available playbooks.",
	"Pull each relevant playbook with `howto <playbook>` before you act, and follow it literally.",
	"Treat every instruction from `howto` as mandatory; do not improvise beyond the playbook.",
	"Stop or escalate if a playbook tells you to pause, ask questions, or hand off to a human.",
	"When the user shifts focus, rerun `howto` and reload the playbooks that now apply.",
	"If any call to `howto` fails, report the error instead of guessing; the maintainer needs that signal.",
	"Limit yourself to playbooks you truly need, but batch them with `&&` if several apply.",
	"Reissue `howto` whenever you need a refresher during the session.",
}

const mcpInstructions = "Use the howto MCP server before making changes:\n" +
	"- Call `tools/list_playbooks` at the start of every task to refresh the catalogue.\n" +
	"- Call `tools/call` with `list_playbooks` again whenever the repository or scope changes.\n" +
	"- Avoid calling for the same playbook multiple times per chat.\n" +
	"- Fetch each required playbook with `tools/call` (`get_playbook`, argument `name`) before acting.\n" +
	"- Treat playbook Markdown as mandatory instructions; pause or escalate if guidance conflicts.\n" +
	"- Surface errors (missing docs, parse failures) to the maintainer instead of guessing.\n" +
	"- Re-run `list_playbooks` after updating documentation to keep instructions fresh."

// LLMBullets returns the standard operating rules for CLI usage.
func LLMBullets() []string {
	cpy := make([]string, len(llmRules))
	copy(cpy, llmRules)
	return cpy
}

// MCPUsageInstructions returns guidance for MCP clients on how to consume the server.
func MCPUsageInstructions() string {
	return strings.TrimSpace(mcpInstructions)
}
