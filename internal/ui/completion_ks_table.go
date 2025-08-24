package ui

import (
	"fmt"
	"os"
	"strings"
)

// handleKeyspaceTableCompletion handles completions for keyspace.table patterns
func (ce *CompletionEngine) handleKeyspaceTableCompletion(input string) []string {
	// Find the last dot
	lastDot := strings.LastIndex(input, ".")
	beforeDot := input[:lastDot]
	afterDot := input[lastDot+1:]

	// Check if we have a keyspace name pattern (word before the dot)
	words := strings.Fields(beforeDot)
	if len(words) >= 2 {
		// Get the potential keyspace name (last word before the dot)
		potentialKeyspace := words[len(words)-1]

		// Check if this is a valid keyspace
		keyspaces := ce.getKeyspaceNames()
		for _, ks := range keyspaces {
			if strings.EqualFold(ks, potentialKeyspace) {
				// We have a valid keyspace - check tables from that keyspace
				tables := ce.getTablesForKeyspace(ks)

				// If there's text after the dot, check if it's a complete table name or partial
				if afterDot != "" {
					// Check for exact match first
					var isCompleteTable bool
					for _, table := range tables {
						if strings.EqualFold(table, afterDot) {
							isCompleteTable = true
							break
						}
					}

					// If it's a complete table name, don't try to complete it further
					// Let the parser-based completion handle what comes next
					if isCompleteTable {
						// Debug logging
						if debugFile, err := os.OpenFile("cqlai_debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
							fmt.Fprintf(debugFile, "[DEBUG] Complete table name detected: %s.%s, falling through to parser\n", ks, afterDot)
							defer debugFile.Close()
						}
						// Return nil to indicate fall through to parser
						return nil
					}

					// It's a partial table name - filter and return matches
					var filtered []string
					upperAfterDot := strings.ToUpper(afterDot)
					for _, table := range tables {
						if strings.HasPrefix(strings.ToUpper(table), upperAfterDot) {
							// Return just the table name, not the full path
							filtered = append(filtered, table)
						}
					}
					if len(filtered) > 0 {
						return filtered
					}
				} else {
					// No text after dot - return just the table names
					// The keyboard handler will append them correctly
					return tables
				}
			}
		}
	}
	return nil
}
