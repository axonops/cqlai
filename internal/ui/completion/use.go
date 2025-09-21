package completion

func (pce *ParserBasedCompletionEngine) getUseSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return pce.getKeyspaceNames()
	}
	return []string{}
}

// getUseCompletions returns completions for USE commands
func (ce *CompletionEngine) getUseCompletions(_ []string, wordPos int) []string {
	if wordPos == 1 {
		return ce.getKeyspaceNames()
	}
	return []string{}
}
