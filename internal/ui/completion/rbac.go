package completion

func (pce *ParserBasedCompletionEngine) getGrantSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		// Return a subset of common permissions for simple suggestions
		return RBACPermissions
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
		return OnKeyword
	}

	// After ON
	return RBACResourceTypes
}

func (pce *ParserBasedCompletionEngine) getRevokeSuggestions(tokens []string) []string {
	// Similar to GRANT
	return pce.getGrantSuggestions(tokens)
}
