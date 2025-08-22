package ui

func (pce *ParserBasedCompletionEngine) getDescribeSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return []string{
			"KEYSPACE",
			"KEYSPACES",
			"TABLE",
			"TABLES",
			"TYPE",
			"TYPES",
			"FUNCTION",
			"FUNCTIONS",
			"AGGREGATE",
			"AGGREGATES",
			"MATERIALIZED",
		}
	}

	if len(tokens) == 2 && tokens[1] == "MATERIALIZED" {
		return []string{"VIEW", "VIEWS"}
	}

	return []string{}
}
