package completion

// getShowCompletions returns completions for SHOW commands
func (ce *CompletionEngine) getShowCompletions(_ []string, wordPos int) []string {
	if wordPos == 1 {
		return ShowCommands
	}
	return []string{}
}

func (pce *ParserBasedCompletionEngine) getConsistencySuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return ConsistencyLevels
	}
	return []string{}
}

func (pce *ParserBasedCompletionEngine) getOutputSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		// After OUTPUT, suggest format types
		return OutputFormats
	}
	return []string{}
}

func (pce *ParserBasedCompletionEngine) getAutoFetchSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		// After AUTOFETCH, suggest ON or OFF
		return []string{"ON", "OFF"}
	}
	return []string{}
}
