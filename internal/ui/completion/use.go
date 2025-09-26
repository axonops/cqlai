package completion


// getUseCompletions returns completions for USE commands
func (ce *CompletionEngine) getUseCompletions(_ []string, wordPos int) []string {
	if wordPos == 1 {
		return ce.getKeyspaceNames()
	}
	return []string{}
}
