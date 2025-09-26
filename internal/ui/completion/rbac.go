package completion



// getGrantCompletions returns completions for GRANT commands
func (ce *CompletionEngine) getGrantCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		// After GRANT, suggest permissions
		return CQLPermissions
	}

	// Track what we've seen
	hasOn := false
	hasTo := false
	onPos := -1

	for i, word := range words {
		switch word {
		case "ON":
			hasOn = true
			onPos = i
		case "TO":
			hasTo = true
		}
	}

	// Get the last word
	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	// After permission(s) and before ON
	if !hasOn && wordPos > 1 {
		// Could have multiple permissions separated by commas
		if lastWord == "," {
			return CQLPermissions
		}
		return append(OnKeyword, ",")
	}

	// After ON
	if hasOn && onPos >= 0 {
		if wordPos == onPos+1 {
			// Resource types
			return ResourceTypes
		}

		// After resource type
		if wordPos == onPos+2 {
			switch words[onPos+1] {
			case "KEYSPACE":
				return ce.getKeyspaceNames()
			case "TABLE":
				return ce.getTableNames()
			case "FUNCTION":
				return ce.getFunctionNames()
			case "AGGREGATE":
				return ce.getAggregateNames()
			case "INDEX":
				return ce.getIndexNames()
			case "MATERIALIZED":
				return MaterializedKeyword
			case "ALL":
				return AllResourceTargets
			}
		}

		// After MATERIALIZED VIEW
		if wordPos == onPos+3 && words[onPos+1] == "MATERIALIZED" && words[onPos+2] == "VIEW" {
			return ce.getViewNames()
		}

		// After resource name, suggest TO
		if !hasTo && wordPos > onPos+2 {
			return ToKeyword
		}
	}

	// After TO
	if hasTo {
		// Suggest role/user name (no completion for now)
		return []string{}
	}

	return []string{}
}

// getRevokeCompletions returns completions for REVOKE commands
func (ce *CompletionEngine) getRevokeCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		// After REVOKE, suggest permissions
		return CQLPermissions
	}

	// Track what we've seen
	hasOn := false
	hasFrom := false
	onPos := -1

	for i, word := range words {
		switch word {
		case "ON":
			hasOn = true
			onPos = i
		case "FROM":
			hasFrom = true
		}
	}

	// Get the last word
	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	// After permission(s) and before ON
	if !hasOn && wordPos > 1 {
		// Could have multiple permissions separated by commas
		if lastWord == "," {
			return CQLPermissions
		}
		return append(OnKeyword, ",")
	}

	// After ON
	if hasOn && onPos >= 0 {
		if wordPos == onPos+1 {
			// Resource types
			return ResourceTypes
		}

		// After resource type
		if wordPos == onPos+2 {
			switch words[onPos+1] {
			case "KEYSPACE":
				return ce.getKeyspaceNames()
			case "TABLE":
				return ce.getTableNames()
			case "FUNCTION":
				return ce.getFunctionNames()
			case "AGGREGATE":
				return ce.getAggregateNames()
			case "INDEX":
				return ce.getIndexNames()
			case "MATERIALIZED":
				return MaterializedKeyword
			case "ALL":
				return AllResourceTargets
			}
		}

		// After MATERIALIZED VIEW
		if wordPos == onPos+3 && words[onPos+1] == "MATERIALIZED" && words[onPos+2] == "VIEW" {
			return ce.getViewNames()
		}

		// After resource name, suggest FROM
		if !hasFrom && wordPos > onPos+2 {
			return FromKeyword
		}
	}

	// After FROM
	if hasFrom {
		// Suggest role/user name (no completion for now)
		return []string{}
	}

	return []string{}
}
