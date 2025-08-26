package completion

func (pce *ParserBasedCompletionEngine) getDescribeSuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		return DescribeObjectsBasic
	}

	if len(tokens) == 2 && tokens[1] == "MATERIALIZED" {
		return MaterializedViews
	}

	return []string{}
}
