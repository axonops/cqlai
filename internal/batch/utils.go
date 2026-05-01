package batch

import "strings"

// stripComments removes SQL-style comments from CQL statements while respecting quoted strings.
// It handles:
// - Single-quoted strings 'like this' (CQL standard)
// - Escaped quotes '' within strings
// - Line comments: -- and //
// - Block comments: /* ... */
func stripComments(input string) string {
	var result strings.Builder
	result.Grow(len(input))

	i := 0
	inSingleQuote := false

	for i < len(input) {
		ch := input[i]

		// Handle single quotes
		if ch == '\'' {
			if inSingleQuote {
				// Check for escaped quote ''
				if i+1 < len(input) && input[i+1] == '\'' {
					result.WriteString("''")
					i += 2
					continue
				}
				// End of quoted string
				inSingleQuote = false
			} else {
				// Start of quoted string
				inSingleQuote = true
			}
			result.WriteByte(ch)
			i++
			continue
		}

		// If inside quotes, just copy the character
		if inSingleQuote {
			result.WriteByte(ch)
			i++
			continue
		}

		// Not in quotes - check for comments

		// Check for -- line comment
		if ch == '-' && i+1 < len(input) && input[i+1] == '-' {
			// Skip until end of line
			for i < len(input) && input[i] != '\n' {
				i++
			}
			// Don't skip the newline itself - let it be processed normally
			continue
		}

		// Check for // line comment
		if ch == '/' && i+1 < len(input) && input[i+1] == '/' {
			// Skip until end of line
			for i < len(input) && input[i] != '\n' {
				i++
			}
			continue
		}

		// Check for /* block comment */
		if ch == '/' && i+1 < len(input) && input[i+1] == '*' {
			i += 2 // Skip /*
			// Find closing */
			for i < len(input) {
				if input[i] == '*' && i+1 < len(input) && input[i+1] == '/' {
					i += 2 // Skip */
					break
				}
				i++
			}
			continue
		}

		// Regular character - copy it
		result.WriteByte(ch)
		i++
	}

	return result.String()
}

// splitStatements intelligently splits CQL statements by semicolons,
// respecting quoted strings and handling BATCH blocks.
func splitStatements(content string) []string {
	var statements []string
	var currentStmt strings.Builder

	inSingleQuote := false
	inBatch := false

	for i := 0; i < len(content); i++ {
		ch := content[i]

		// Handle single quotes
		if ch == '\'' {
			if inSingleQuote {
				// Check for escaped quote ''
				if i+1 < len(content) && content[i+1] == '\'' {
					currentStmt.WriteString("''")
					i++
					continue
				}
				// End of quoted string
				inSingleQuote = false
			} else {
				// Start of quoted string
				inSingleQuote = true
			}
			currentStmt.WriteByte(ch)
			continue
		}

		// If inside quotes, just copy the character
		if inSingleQuote {
			currentStmt.WriteByte(ch)
			continue
		}

		// Check for BATCH start (case-insensitive)
		if !inBatch && (ch == 'B' || ch == 'b') {
			remaining := strings.ToUpper(content[i:])
			if strings.HasPrefix(remaining, "BEGIN BATCH") ||
				strings.HasPrefix(remaining, "BEGIN UNLOGGED BATCH") ||
				strings.HasPrefix(remaining, "BEGIN COUNTER BATCH") {
				inBatch = true
			}
		}

		// Check for APPLY BATCH (end of batch)
		if inBatch && (ch == 'A' || ch == 'a') {
			remaining := strings.ToUpper(content[i:])
			if strings.HasPrefix(remaining, "APPLY BATCH") {
				inBatch = false
			}
		}

		// Handle semicolon - statement separator (unless in batch)
		if ch == ';' {
			currentStmt.WriteByte(ch)

			if !inBatch {
				// End of statement
				stmt := strings.TrimSpace(currentStmt.String())
				if stmt != "" && stmt != ";" {
					statements = append(statements, stmt)
				}
				currentStmt.Reset()
			}
			continue
		}

		// Regular character
		currentStmt.WriteByte(ch)
	}

	// Add any remaining statement
	stmt := strings.TrimSpace(currentStmt.String())
	if stmt != "" && stmt != ";" {
		statements = append(statements, stmt)
	}

	return statements
}
