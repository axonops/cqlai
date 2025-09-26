package completion

// getShowCompletions returns completions for SHOW commands
func (ce *CompletionEngine) getShowCompletions(_ []string, wordPos int) []string {
	if wordPos == 1 {
		return ShowCommands
	}
	return []string{}
}



