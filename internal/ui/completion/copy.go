package completion

import "strings"

// getCopyCompletions returns completions for COPY commands
func (ce *CompletionEngine) getCopyCompletions(words []string, wordPos int) []string {
	// Convert all words to uppercase for comparison
	upperWords := make([]string, len(words))
	for i, w := range words {
		upperWords[i] = strings.ToUpper(w)
	}

	// First scan for key tokens to understand context
	hasTo := false
	hasFrom := false
	hasWith := false
	toIndex := -1
	fromIndex := -1
	withIndex := -1

	for i, word := range upperWords {
		switch word {
		case "TO":
			hasTo = true
			toIndex = i
		case "FROM":
			hasFrom = true
			fromIndex = i
		case "WITH":
			hasWith = true
			withIndex = i
		}
	}

	// After COPY, before any TO/FROM
	if !hasTo && !hasFrom {
		if wordPos == 1 {
			// After COPY, suggest both table names and keyspace.table combinations
			return ce.getTableAndKeyspaceTableNames()
		}
		// After table name (could be keyspace.table), suggest TO or FROM
		return CopyDirections
	}

	// If we're in WITH clause, provide context-aware completions
	if hasWith && withIndex >= 0 {
		// Check if we're completing an option value
		if wordPos > withIndex {
			// Check the last few words to understand context
			if wordPos-1 < len(upperWords) {
				lastWord := upperWords[wordPos-1]

				// Check if last word contains an assignment with value (e.g., FORMAT='PARQUET')
				if strings.Contains(lastWord, "=") && strings.Contains(lastWord, "'") {
					// This is a complete option assignment, suggest AND
					return []string{"AND"}
				}

				// If last word is a quoted value, suggest AND to chain more options
				if strings.HasPrefix(lastWord, "'") && strings.HasSuffix(lastWord, "'") {
					return []string{"AND"}
				}

				// Check for specific option keywords
				switch lastWord {
				case "FORMAT", "FORMAT=":
					return CopyFormats
				case "COMPRESSION", "COMPRESSION=":
					return ParquetCompressionTypes
				case "AND":
					// After AND, suggest more options
					return CopyOptions
				}

				// Check if second-to-last word is an option waiting for value
				if wordPos-2 >= 0 && wordPos-2 < len(upperWords) {
					secondLastWord := upperWords[wordPos-2]
					switch secondLastWord {
					case "FORMAT":
						if lastWord == "=" {
							return CopyFormats
						}
						// If we have FORMAT followed by a value, suggest AND
						if strings.HasPrefix(lastWord, "'") && strings.HasSuffix(lastWord, "'") {
							return []string{"AND"}
						}
					case "COMPRESSION":
						if lastWord == "=" {
							return ParquetCompressionTypes
						}
						// If we have COMPRESSION followed by a value, suggest AND
						if strings.HasPrefix(lastWord, "'") && strings.HasSuffix(lastWord, "'") {
							return []string{"AND"}
						}
					}
				}
			}

			// If we're right after WITH, suggest options
			if wordPos == withIndex+1 {
				return CopyOptions
			}

			// Default: suggest both AND and options for flexibility
			// This allows users to either chain with AND or start a new option
			return append([]string{"AND"}, CopyOptions...)
		}
	}

	// If we have WITH or the last word is WITH, suggest options
	if hasWith || (len(upperWords) > 0 && upperWords[len(upperWords)-1] == "WITH") {
		// We're after WITH, suggest COPY options
		return CopyOptions
	}

	// After TO/FROM but before WITH
	if (hasTo || hasFrom) && !hasWith {
		// Check if we're after the filename
		directionIndex := toIndex
		if hasFrom {
			directionIndex = fromIndex
		}

		// If we're right after TO/FROM, suggest file paths
		if wordPos == directionIndex+1 {
			return CopyFileSuggestions
		}

		// If we're at least 2 positions after TO/FROM (table TO 'filename'), suggest WITH
		if wordPos > directionIndex+1 {
			return []string{"WITH"}
		}
	}

	return []string{}
}

func (pce *ParserBasedCompletionEngine) getCopySuggestions(tokens []string) []string {
	if len(tokens) == 1 {
		// After COPY, suggest both table names and keyspace.table combinations
		return pce.getTableAndKeyspaceTableNames()
	}

	// Check for TO/FROM
	hasTo := false
	hasFrom := false
	hasWith := false
	hasDot := false
	withIndex := -1

	for i, token := range tokens {
		switch token {
		case "TO":
			hasTo = true
		case "FROM":
			hasFrom = true
		case "WITH":
			hasWith = true
			withIndex = i
		case ".":
			hasDot = true
		}
	}

	// After table name (and optional column list), suggest TO/FROM
	// Handle both "COPY table" and "COPY keyspace.table" cases
	if !hasTo && !hasFrom {
		// If we have exactly 2 tokens (COPY table) or 4 tokens (COPY keyspace . table)
		if len(tokens) == 2 || (len(tokens) == 4 && hasDot) {
			return CopyDirections
		}
	}

	// After TO/FROM
	if (hasTo || hasFrom) && !hasWith {
		// Find the position of TO/FROM
		toFromPos := -1
		for i, token := range tokens {
			if token == "TO" || token == "FROM" {
				toFromPos = i
				break
			}
		}

		// If we're right after TO/FROM, suggest file paths
		if toFromPos >= 0 && len(tokens) == toFromPos+1 {
			return CopyFileSuggestions
		}

		// If we have more tokens after TO/FROM, suggest WITH
		if len(tokens) >= 4 {
			return []string{"WITH"}
		}
	}

	// After WITH, provide context-aware completions
	if hasWith && withIndex >= 0 {
		// Look for specific option keywords that need value completions
		if len(tokens) > withIndex+1 {
			lastToken := tokens[len(tokens)-1]
			secondLast := ""
			if len(tokens) > withIndex+2 {
				secondLast = tokens[len(tokens)-2]
			}

			// Check for option that needs specific values
			switch strings.ToUpper(secondLast) {
			case "FORMAT":
				if lastToken == "=" {
					return CopyFormats
				}
			case "COMPRESSION":
				if lastToken == "=" {
					return ParquetCompressionTypes
				}
			}

			// Check if last token itself is an option needing values
			switch strings.ToUpper(lastToken) {
			case "FORMAT=":
				return CopyFormats
			case "COMPRESSION=":
				return ParquetCompressionTypes
			case "AND":
				return CopyOptions
			}
		}
		return CopyOptions
	}

	return []string{}
}
