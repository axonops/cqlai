package router

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/db"
	grammar "github.com/axonops/cqlai/internal/parser/grammar"
)

// VisitKwDescribe visits a describe statement.
// Deprecated: These functions are kept for backward compatibility
// Use db.CQLTypeHandler instead for standardized type handling

func (v *CqlCommandVisitorImpl) VisitKwDescribe(ctx *grammar.KwDescribeContext) interface{} {
	return "DESCRIBE statement"
}

// VisitConsistencyCommand visits a consistency command.
func (v *CqlCommandVisitorImpl) VisitConsistencyCommand(ctx *grammar.ConsistencyCommandContext) interface{} {
	// For now, return current consistency level
	// In a real implementation, this would get/set the session's consistency
	return "Current consistency level: " + v.session.Consistency()
}

func (v *CqlCommandVisitorImpl) VisitUse_(ctx *grammar.Use_Context) interface{} {
	keyspace := ctx.Keyspace().GetText()
	// In a real application, you would handle the session change here.
	// For now, we just return a confirmation message.
	return "Now using keyspace " + keyspace
}

func (v *CqlCommandVisitorImpl) VisitSelect_(ctx *grammar.Select_Context) interface{} {
	// Immediately print to stderr to confirm we're in this function
	fmt.Fprintf(os.Stderr, "[TRACE] VisitSelect_ called\n")

	// Get the full text of the SELECT statement
	query := ctx.GetText()

	// Get current working directory for logging
	cwd, _ := os.Getwd()
	logPath := cwd + "/cqlai_debug.log"

	// Open or create debug log file in current working directory
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// If we can't open the log file, fall back to stdout
		fmt.Printf("[WARNING] Could not open debug log file at %s: %v\n", logPath, err)
		logFile = os.Stdout
	} else {
		// Notify user where the log file is
		fmt.Fprintf(os.Stderr, "[INFO] Debug logging to: %s\n", logPath)
	}
	defer func() {
		if logFile != os.Stdout {
			logFile.Close()
		}
	}()

	// Create a logger function
	logDebug := func(format string, args ...interface{}) {
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		msg := fmt.Sprintf(format, args...)
		fmt.Fprintf(logFile, "[%s] %s\n", timestamp, msg)
		logFile.Sync() // Ensure it's written immediately
	}

	// Add logging for debugging
	logDebug("========================================")
	logDebug("[DEBUG] Executing query: %s", query)

	// Special handling for problematic queries
	// Check if this might be a query with vector or custom types
	// Note: Some tables with vector columns or custom types may fail with standard scanning
	isProblematicQuery := strings.Contains(strings.ToLower(query), "sai_vector_test") ||
		strings.Contains(strings.ToLower(query), "vector") ||
		strings.Contains(strings.ToLower(query), "embedding")

	logDebug("[DEBUG] Is problematic query: %v", isProblematicQuery)

	// Check if this is a JSON query
	isJsonQuery := false
	upperQuery := strings.ToUpper(query)
	if strings.Contains(upperQuery, "SELECT JSON") || strings.Contains(upperQuery, "SELECT DISTINCT JSON") {
		isJsonQuery = true
	}

	logDebug("[DEBUG] Is JSON query: %v", isJsonQuery)

	// Execute the query
	iter := v.session.Query(query).Iter()

	// Get column info
	columns := iter.Columns()
	logDebug("[DEBUG] Number of columns: %d", len(columns))

	// Log column details
	for i, col := range columns {
		logDebug("[DEBUG] Column %d: Name=%s, Type=%v, TypeInfo=%T",
			i, col.Name, col.TypeInfo.Type(), col.TypeInfo)
	}

	if len(columns) == 0 {
		if err := iter.Close(); err != nil {
			logDebug("[DEBUG] Error closing iterator: %v", err)
			return fmt.Errorf("query failed: %v", err)
		}
		return "No results"
	}

	// Prepare headers
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Name
	}

	// Collect results
	results := [][]string{headers}

	// For JSON queries, use simple string scanning
	if isJsonQuery {
		scanners := make([]interface{}, len(columns))
		for i := range scanners {
			var s string
			scanners[i] = &s
		}

		for iter.Scan(scanners...) {
			row := make([]string, len(columns))
			for i, scanner := range scanners {
				if v, ok := scanner.(*string); ok && v != nil {
					row[i] = *v
				} else {
					row[i] = "null"
				}
			}
			results = append(results, row)
		}
	} else {
		// For non-JSON queries
		// Create a standardized type handler
		typeHandler := db.NewCQLTypeHandler()

		// For problematic queries (with vectors, etc.), use a special approach
		if isProblematicQuery {
			logDebug("[DEBUG] Using string scanning approach for problematic query")

			// Try to handle row by row with explicit column handling
			rowNum := 0
			for {
				// Create scanners based on column types, but be very permissive
				scanners := make([]interface{}, len(columns))
				for i := range columns {
					// Use string pointers for everything to avoid unmarshaling issues
					var s string
					scanners[i] = &s
				}

				logDebug("[DEBUG] Attempting to scan row %d", rowNum)

				if !iter.Scan(scanners...) {
					logDebug("[DEBUG] Scan returned false at row %d", rowNum)
					// Check for scan error
					if err := iter.Close(); err != nil {
						logDebug("[DEBUG] Iterator error: %v", err)
					}
					break
				}

				logDebug("[DEBUG] Successfully scanned row %d", rowNum)
				rowNum++

				row := make([]string, len(columns))
				for i, scanner := range scanners {
					if str, ok := scanner.(*string); ok {
						row[i] = *str
						logDebug("[DEBUG] Row %d, Col %d (%s): %s", rowNum-1, i, columns[i].Name, *str)
					} else {
						row[i] = fmt.Sprintf("%v", scanner)
						logDebug("[DEBUG] Row %d, Col %d (%s): %v (type: %T)", rowNum-1, i, columns[i].Name, scanner, scanner)
					}
				}
				results = append(results, row)
			}
		} else {
			logDebug("[DEBUG] Using MapScan approach for normal query")

			// For normal queries, use MapScan which handles most types well
			rowNum := 0
			for {
				rowMap := make(map[string]interface{})

				logDebug("[DEBUG] Attempting MapScan for row %d", rowNum)

				if !iter.MapScan(rowMap) {
					logDebug("[DEBUG] MapScan returned false at row %d", rowNum)
					// Check for scan error
					if err := iter.Close(); err != nil {
						logDebug("[DEBUG] Iterator error after MapScan: %v", err)
					}
					break
				}

				logDebug("[DEBUG] Successfully MapScanned row %d", rowNum)
				rowNum++

				// Process the row
				row := make([]string, len(columns))
				for i, col := range columns {
					if val, ok := rowMap[col.Name]; ok {
						logDebug("[DEBUG] Row %d, Col %d (%s): value type = %T", rowNum-1, i, col.Name, val)
						// Use the standardized type handler with type info when available
						row[i] = typeHandler.FormatValue(val, col.TypeInfo)
					} else {
						logDebug("[DEBUG] Row %d, Col %d (%s): NULL", rowNum-1, i, col.Name)
						row[i] = typeHandler.NullString
					}
				}
				results = append(results, row)
			}
		}

	}

	logDebug("[DEBUG] Query processing complete. Total rows: %d", len(results)-1) // -1 for header
	logDebug("========================================")

	if err := iter.Close(); err != nil {
		logDebug("[DEBUG] Final iterator close error: %v", err)
		return fmt.Errorf("query failed: %v", err)
	}

	if len(results) == 1 {
		return "No rows returned"
	}

	return results
}
