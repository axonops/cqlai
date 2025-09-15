package batch

import "strings"

// stripComments removes SQL-style comments from CQL statements while preserving newlines
func stripComments(input string) string {
	var result strings.Builder
	lines := strings.Split(input, "\n")

	inBlockComment := false
	for _, line := range lines {
		// Handle block comments /* ... */
		for {
			if inBlockComment {
				endIdx := strings.Index(line, "*/")
				if endIdx >= 0 {
					line = line[endIdx+2:]
					inBlockComment = false
				} else {
					// Entire line is within block comment
					line = ""
					break
				}
			}

			startIdx := strings.Index(line, "/*")
			if startIdx >= 0 {
				endIdx := strings.Index(line[startIdx:], "*/")
				if endIdx >= 0 {
					// Block comment on same line
					line = line[:startIdx] + line[startIdx+endIdx+2:]
				} else {
					// Block comment starts but doesn't end on this line
					line = line[:startIdx]
					inBlockComment = true
					break
				}
			} else {
				break
			}
		}

		// Handle line comments -- and //
		if idx := strings.Index(line, "--"); idx >= 0 {
			line = line[:idx]
		}
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = line[:idx]
		}

		// Trim trailing whitespace but preserve the line structure
		line = strings.TrimRight(line, " \t\r")

		// Add the line (even if empty) to preserve line breaks
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString(line)
	}

	return result.String()
}

// splitStatements intelligently splits CQL statements, handling BATCH blocks
func splitStatements(content string) []string {
	var statements []string
	var currentStmt strings.Builder

	// Track if we're inside a BATCH statement
	inBatch := false

	// Track if we're accumulating a non-empty statement
	hasContent := false

	// Process line by line to handle BATCH blocks and regular statements
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines unless we're in a batch
		if trimmedLine == "" && !inBatch && !hasContent {
			continue
		}

		// Convert to uppercase for comparison
		upperLine := strings.ToUpper(trimmedLine)

		// Check for BATCH start
		if strings.HasPrefix(upperLine, "BEGIN BATCH") ||
		   strings.HasPrefix(upperLine, "BEGIN UNLOGGED BATCH") ||
		   strings.HasPrefix(upperLine, "BEGIN COUNTER BATCH") {
			inBatch = true
			hasContent = true
		}

		// Add line to current statement
		if currentStmt.Len() > 0 {
			currentStmt.WriteString("\n")
		}
		currentStmt.WriteString(line)

		if trimmedLine != "" {
			hasContent = true
		}

		// Check if we have a complete statement
		if strings.HasSuffix(trimmedLine, ";") {
			if inBatch {
				// Check if this ends the batch
				if strings.HasPrefix(upperLine, "APPLY BATCH") {
					inBatch = false
					stmt := strings.TrimSpace(currentStmt.String())
					if stmt != "" && stmt != ";" {
						statements = append(statements, stmt)
					}
					currentStmt.Reset()
					hasContent = false
				}
				// Otherwise, continue accumulating the batch
			} else {
				// Regular statement ended
				stmt := strings.TrimSpace(currentStmt.String())
				if stmt != "" && stmt != ";" {
					statements = append(statements, stmt)
				}
				currentStmt.Reset()
				hasContent = false
			}
		}
	}

	// Add any remaining statement
	if hasContent {
		stmt := strings.TrimSpace(currentStmt.String())
		if stmt != "" && stmt != ";" {
			statements = append(statements, stmt)
		}
	}

	return statements
}