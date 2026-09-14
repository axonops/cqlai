package completion

import "strings"

// Completing a CREATE INDEX.
//
// The statement is a name, the table it is on, what it is on in brackets, and
// then which implementation to use and what to give it. Completion stopped at
// the name: ON, the bracket, the target functions, USING and the options were
// all left to be typed from memory.

// indexOptionValues is what each option of a storage attached index takes.
var indexOptionValues = map[string]whatGoesHere{
	"ascii":                    {values: quoted(BooleanValues)},
	"case_sensitive":           {values: quoted(BooleanValues)},
	"normalize":                {values: quoted(BooleanValues)},
	"construction_beam_width":  {note: Hint("number")},
	"maximum_node_connections": {note: Hint("number")},
	"optimize_for":             {values: quoted(OptimizeForValues)},
	"similarity_function":      {values: quoted(SimilarityFunctions)},
}

// indexStatementCompletions is what to offer in a CREATE INDEX, and whether the
// statement is one.
func indexStatementCompletions(input string) ([]string, bool) {
	words := strings.Fields(strings.ToUpper(input))
	if len(words) == 0 || words[0] != "CREATE" {
		return nil, false
	}

	// CREATE CUSTOM, before the word it is custom of.
	if len(words) == 2 && words[1] == "CUSTOM" && strings.HasSuffix(input, " ") {
		return []string{"INDEX"}, true
	}
	if !isCreateIndex(words) {
		return nil, false
	}

	if suggestions, inOptions := indexOptionCompletions(input); inOptions {
		return suggestions, true
	}
	if suggestions, inTarget := indexTargetCompletions(input); inTarget {
		return suggestions, true
	}
	return indexShapeCompletions(input, words)
}

// isCreateIndex reports whether the statement is creating an index.
func isCreateIndex(words []string) bool {
	if len(words) < 2 || words[0] != "CREATE" {
		return false
	}
	return words[1] == "INDEX" || (words[1] == "CUSTOM" && len(words) > 2 && words[2] == "INDEX")
}

// indexShapeCompletions is the word that comes next in the statement itself.
func indexShapeCompletions(input string, words []string) ([]string, bool) {
	afterSpace := strings.HasSuffix(input, " ")

	switch {
	case !containsWord(input, "ON"):
		// The name is the user's to invent, and the note says so. What follows
		// one is the table it is on.
		if indexNamed(words) && afterSpace {
			return []string{"ON"}, true
		}
		return nil, false

	case !strings.Contains(input, "("):
		// The table is looked up, and the bracket follows it.
		if tableNamed(words) && afterSpace {
			return []string{"("}, true
		}
		return nil, false

	case !strings.Contains(input, ")"):
		return nil, false // inside the brackets, which is answered above

	case !containsWord(input, "USING"):
		if afterSpace {
			return []string{"USING"}, true
		}
		return nil, false

	case !classNamed(input):
		return quoted(IndexClasses), true

	case !containsWord(input, "WITH"):
		if afterSpace {
			return []string{"WITH"}, true
		}
		return nil, false

	case !containsWord(input, "OPTIONS"):
		return []string{"OPTIONS = "}, true
	}
	return []string{"{'"}, true
}

// indexTargetCompletions is what to offer inside the brackets: the column, or
// the function that picks a part of a collection.
func indexTargetCompletions(input string) ([]string, bool) {
	depth, start := 0, 0
	for i := 0; i < len(input); i++ {
		switch input[i] {
		case '(':
			depth++
			if depth == 1 {
				start = i + 1
			}
		case ')':
			depth--
		}
	}

	if depth == 0 {
		return nil, false
	}
	if depth > 1 {
		return []string{columnHint}, true // inside keys(, values(, entries( or full(
	}

	target := strings.TrimSpace(input[start:])
	switch {
	case target == "":
		return append(IndexTargets, columnHint), true
	case strings.HasSuffix(target, ")"), len(finishedWords(input[start:])) > 0:
		return []string{")"}, true
	}
	return []string{columnHint}, true // the column name, being typed
}

// indexOptionCompletions is what to offer inside the OPTIONS map.
func indexOptionCompletions(input string) ([]string, bool) {
	open := strings.LastIndex(input, "{")
	if open < 0 || strings.Contains(input[open:], "}") {
		return nil, false
	}

	if key := keyAwaitingValue(input); key != "" {
		if answer, known := indexOptionValues[key]; known {
			return answer.suggestions(), true
		}
		return []string{Hint("value")}, true
	}
	return keysLeftOf(IndexOptions, input[open+1:]), true
}

// indexNamed reports whether the index has been given a name, which is
// everything after INDEX that is not IF NOT EXISTS.
func indexNamed(words []string) bool {
	rest := words[indexOf(words, "INDEX")+1:]
	for _, word := range rest {
		switch word {
		case "IF", "NOT", "EXISTS":
		default:
			return true
		}
	}
	return false
}

// tableNamed reports whether the table the index is on has been named.
func tableNamed(words []string) bool {
	at := indexOf(words, "ON")
	return at >= 0 && len(words) > at+1
}

// classNamed reports whether USING has been given the implementation to use.
func classNamed(input string) bool {
	at := strings.Index(strings.ToUpper(input), "USING")
	if at < 0 {
		return false
	}

	after := strings.TrimSpace(input[at+len("USING"):])
	return strings.Count(after, "'") >= 2 || strings.Count(after, "\"") >= 2
}

// indexOf is where a word is among them, or -1.
func indexOf(words []string, wanted string) int {
	for i, word := range words {
		if word == wanted {
			return i
		}
	}
	return -1
}
