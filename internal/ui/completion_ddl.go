package ui

// getCreateCompletions returns completions for CREATE commands
func (ce *CompletionEngine) getCreateCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return []string{"KEYSPACE", "TABLE", "INDEX", "TYPE", "FUNCTION", "AGGREGATE", "MATERIALIZED", "ROLE", "USER"}
	}

	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	if lastWord == "MATERIALIZED" {
		return []string{"VIEW"}
	}

	if wordPos == 2 && len(words) > 1 {
		switch words[1] {
		case "KEYSPACE", "TABLE", "INDEX", "TYPE", "FUNCTION", "AGGREGATE", "ROLE", "USER":
			// After CREATE <type>, expect a name or IF NOT EXISTS
			return []string{"IF"}
		}
	}

	return []string{}
}

// getDropCompletions returns completions for DROP commands
func (ce *CompletionEngine) getDropCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return []string{"KEYSPACE", "TABLE", "INDEX", "TYPE", "FUNCTION", "AGGREGATE", "MATERIALIZED", "ROLE", "USER"}
	}

	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	if lastWord == "MATERIALIZED" {
		return []string{"VIEW"}
	}

	// After DROP <type>
	if wordPos == 2 && len(words) > 1 {
		switch words[1] {
		case "KEYSPACE":
			return append([]string{"IF"}, ce.getKeyspaceNames()...)
		case "TABLE":
			return append([]string{"IF"}, ce.getTableNames()...)
		case "INDEX":
			return append([]string{"IF"}, ce.getIndexNames()...)
		case "TYPE":
			return append([]string{"IF"}, ce.getTypeNames()...)
		case "FUNCTION":
			return append([]string{"IF"}, ce.getFunctionNames()...)
		case "AGGREGATE":
			return append([]string{"IF"}, ce.getAggregateNames()...)
		case "ROLE", "USER":
			return []string{"IF"}
		}
	}

	// After DROP MATERIALIZED VIEW
	if wordPos == 3 && len(words) > 2 && words[1] == "MATERIALIZED" && words[2] == "VIEW" {
		return append([]string{"IF"}, ce.getViewNames()...)
	}

	return []string{}
}

// getAlterCompletions returns completions for ALTER commands
func (ce *CompletionEngine) getAlterCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		return []string{"KEYSPACE", "TABLE", "TYPE", "MATERIALIZED", "ROLE", "USER"}
	}

	lastWord := ""
	if len(words) > 0 {
		lastWord = words[len(words)-1]
	}

	if lastWord == "MATERIALIZED" {
		return []string{"VIEW"}
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
			return []string{"ADD", "DROP", "ALTER", "RENAME", "WITH"}
		case "KEYSPACE":
			return []string{"WITH"}
		case "TYPE":
			return []string{"ADD", "RENAME"}
		}
	}

	return []string{}
}

// getTruncateCompletions returns completions for TRUNCATE commands
func (ce *CompletionEngine) getTruncateCompletions(words []string, wordPos int) []string {
	if wordPos == 1 {
		// After TRUNCATE, can specify TABLE or just table name
		tables := ce.getTableNames()
		return append([]string{"TABLE"}, tables...)
	}

	if wordPos == 2 && len(words) > 1 && words[1] == "TABLE" {
		return ce.getTableNames()
	}

	return []string{}
}

func (pce *ParserBasedCompletionEngine) getCreateSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return []string{
			"AGGREGATE",
			"FUNCTION",
			"INDEX",
			"KEYSPACE",
			"MATERIALIZED",
			"ROLE",
			"TABLE",
			"TRIGGER",
			"TYPE",
			"USER",
		}
	}

	if len(tokens) == 2 && tokens[1] == "MATERIALIZED" {
		return []string{"VIEW"}
	}

	return []string{"IF"}
}

func (pce *ParserBasedCompletionEngine) getDropSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return []string{
			"AGGREGATE",
			"FUNCTION",
			"INDEX",
			"KEYSPACE",
			"MATERIALIZED",
			"ROLE",
			"TABLE",
			"TRIGGER",
			"TYPE",
			"USER",
		}
	}

	if len(tokens) == 2 && tokens[1] == "MATERIALIZED" {
		return []string{"VIEW"}
	}

	// After object type, suggest IF EXISTS
	return []string{"IF"}
}

func (pce *ParserBasedCompletionEngine) getAlterSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return []string{
			"KEYSPACE",
			"MATERIALIZED",
			"ROLE",
			"TABLE",
			"TYPE",
			"USER",
		}
	}

	if len(tokens) == 2 && tokens[1] == "MATERIALIZED" {
		return []string{"VIEW"}
	}

	return []string{}
}
