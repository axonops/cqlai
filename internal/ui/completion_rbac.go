package ui

func (pce *ParserBasedCompletionEngine) getGrantSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		// Return a subset of common permissions for simple suggestions
		return []string{"ALL", "SELECT", "MODIFY", "CREATE", "ALTER", "DROP", "AUTHORIZE"}
	}

	// Look for ON
	hasOn := false
	for _, t := range tokens {
		if t == "ON" {
			hasOn = true
			break
		}
	}

	if !hasOn {
		return []string{"ON"}
	}

	// After ON
	return []string{"KEYSPACE", "TABLE", "ROLE", "ALL"}
}

func (pce *ParserBasedCompletionEngine) getRevokeSuggestions(tokens []string) []string {
	// Similar to GRANT
	return pce.getGrantSuggestions(tokens)
}
