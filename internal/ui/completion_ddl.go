package ui

// getCreateCompletions returns completions for CREATE commands
func (ce *CompletionEngine) getCreateCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return CreateDropObjectTypes
	}

	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	if lastWord == "MATERIALIZED" {
		return MaterializedKeyword
	}

	if wordPos == 2 && len(words) > 1 {
		// Check if it's a valid object type (excluding MATERIALIZED which needs VIEW after it)
		for _, objType := range CreateDropObjectTypesNoMaterialized {
			if words[1] == objType {
				// After CREATE <type>, expect a name or IF NOT EXISTS
				return IfKeyword
			}
		}
	}

	return []string{}
}

// getDropCompletions returns completions for DROP commands
func (ce *CompletionEngine) getDropCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return CreateDropObjectTypes
	}

	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	if lastWord == "MATERIALIZED" {
		return MaterializedKeyword
	}

	// After DROP <type>
	if wordPos == 2 && len(words) > 1 {
		switch words[1] {
		case "KEYSPACE":
			return append(IfKeyword, ce.getKeyspaceNames()...) // "IF"
		case "TABLE":
			return append(IfKeyword, ce.getTableNames()...) // "IF"
		case "INDEX":
			return append(IfKeyword, ce.getIndexNames()...) // "IF"
		case "TYPE":
			return append(IfKeyword, ce.getTypeNames()...) // "IF"
		case "FUNCTION":
			return append(IfKeyword, ce.getFunctionNames()...) // "IF"
		case "AGGREGATE":
			return append(IfKeyword, ce.getAggregateNames()...) // "IF"
		case "ROLE", "USER":
			return IfKeyword
		}
	}

	// After DROP MATERIALIZED VIEW
	if wordPos == 3 && len(words) > 2 && words[1] == "MATERIALIZED" && words[2] == "VIEW" {
		return append(IfKeyword, ce.getViewNames()...)
	}

	return []string{}
}

// getAlterCompletions returns completions for ALTER commands
func (ce *CompletionEngine) getAlterCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return AlterObjectTypes
	}

	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	if lastWord == "MATERIALIZED" {
		return MaterializedKeyword
	}

	// After ALTER <type>
	if wordPos == 2 && len(words) > 1 {
		switch words[1] {
		case "KEYSPACE":
			return ce.getKeyspaceNames()
		case "TABLE":
			return ce.getTableNames()
		case "TYPE":
			return ce.getTypeNames()
		case "ROLE", "USER":
			return []string{}
		}
	}

	// After ALTER MATERIALIZED VIEW
	if wordPos == 3 && len(words) > 2 && words[1] == "MATERIALIZED" && words[2] == "VIEW" {
		return ce.getViewNames()
	}

	// After the object name, suggest modification keywords
	if wordPos == 3 && len(words) > 2 {
		switch words[1] {
		case "TABLE":
			return AlterTableOperations
		case "KEYSPACE":
			return WithKeyword
		case "TYPE":
			return AlterTypeOperations
		}
	}

	return []string{}
}

// getTruncateCompletions returns completions for TRUNCATE commands
func (ce *CompletionEngine) getTruncateCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		// After TRUNCATE, can specify TABLE or just table name
		tables := ce.getTableNames()
		return append(TableKeyword, tables...)
	}

	if wordPos == 2 && len(words) > 1 && words[1] == "TABLE" {
		return ce.getTableNames()
	}

	return []string{}
}

func (pce *ParserBasedCompletionEngine) getCreateSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return CreateObjectTypes
	}

	if len(tokens) == 2 && tokens[1] == "MATERIALIZED" {
		return MaterializedKeyword
	}

	return IfKeyword
}

func (pce *ParserBasedCompletionEngine) getDropSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return CreateObjectTypes // DROP supports same objects as CREATE
	}

	if len(tokens) == 2 && tokens[1] == "MATERIALIZED" {
		return MaterializedKeyword
	}

	// After object type, suggest IF EXISTS
	return IfKeyword
}

func (pce *ParserBasedCompletionEngine) getAlterSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return AlterObjectTypes
	}

	if len(tokens) == 2 && tokens[1] == "MATERIALIZED" {
		return MaterializedKeyword
	}

	return []string{}
}
