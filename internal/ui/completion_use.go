package ui

func (pce *ParserBasedCompletionEngine) getUseSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return pce.getKeyspaceNames()
	}
	return []string{}
}
