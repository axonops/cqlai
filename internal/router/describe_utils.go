package router

import (
	"strings"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
)

// extractCreateStatements extracts the create_statement column from various result types
// This is used for Cassandra 4.0+ DESCRIBE commands that return table data with a create_statement column
func (v *CqlCommandVisitorImpl) extractCreateStatements(serverResult interface{}, command string) interface{} {
	if serverResult == nil {
		return "No results"
	}

	var statements []string
	
	// Log the type of result we got
	logger.DebugfToFile("extractCreateStatements", "%s result type: %T", command, serverResult)
	
	switch result := serverResult.(type) {
	case [][]string:
		// Table data format - extract all create_statement values
		if len(result) > 1 {
			// Find the create_statement column index
			createStmtIdx := -1
			for i, header := range result[0] {
				if header == "create_statement" {
					createStmtIdx = i
					break
				}
			}
			
			if createStmtIdx >= 0 {
				// Extract all create statements
				for i := 1; i < len(result); i++ {
					if len(result[i]) > createStmtIdx && result[i][createStmtIdx] != "" {
						statements = append(statements, result[i][createStmtIdx])
					}
				}
			}
		}
		
	case db.QueryResult:
		// QueryResult format - extract all create_statement values
		if len(result.Data) > 1 {
			// Find the create_statement column index
			createStmtIdx := -1
			for i, header := range result.Data[0] {
				if header == "create_statement" {
					createStmtIdx = i
					break
				}
			}
			
			if createStmtIdx >= 0 {
				// Extract all create statements
				for i := 1; i < len(result.Data); i++ {
					if len(result.Data[i]) > createStmtIdx && result.Data[i][createStmtIdx] != "" {
						statements = append(statements, result.Data[i][createStmtIdx])
					}
				}
			}
		}
		
	case db.StreamingQueryResult:
		// For streaming results, we need to read the data first
		logger.DebugfToFile("extractCreateStatements", "%s: Got StreamingQueryResult, extracting create statements", command)
		
		// Look for create_statement column in either Headers or ColumnNames
		createStmtFound := false
		for _, colName := range result.ColumnNames {
			if colName == "create_statement" {
				createStmtFound = true
				break
			}
		}
		
		if createStmtFound {
			logger.DebugfToFile("extractCreateStatements", "%s: Found create_statement column, reading rows", command)
			// Read all rows and extract create statements
			for {
				rowMap := make(map[string]interface{})
				if !result.Iterator.MapScan(rowMap) {
					break
				}
				
				// Get create_statement from the row
				if stmt, ok := rowMap["create_statement"]; ok && stmt != nil {
					if stmtStr, ok := stmt.(string); ok && stmtStr != "" {
						statements = append(statements, stmtStr)
						logger.DebugfToFile("extractCreateStatements", "%s: Extracted statement: %d chars", command, len(stmtStr))
					}
				}
			}
			
			// Close the iterator - IMPORTANT: we must close it here since we're consuming it
			if result.Iterator != nil {
				_ = result.Iterator.Close()
			}
			
			logger.DebugfToFile("extractCreateStatements", "%s: Extracted %d statements total", command, len(statements))
		} else {
			logger.DebugfToFile("extractCreateStatements", "%s: No create_statement column found, returning as-is", command)
			// No create_statement column, return as-is
			return serverResult
		}
		
	case string:
		// Already a string, return as-is
		logger.DebugfToFile("extractCreateStatements", "%s: Result is already a string", command)
		return result
		
	case error:
		// Error result
		return result
		
	default:
		logger.DebugfToFile("extractCreateStatements", "%s: Unknown result type: %T", command, serverResult)
		// Return as-is for unknown types
		return serverResult
	}
	
	// If we successfully extracted statements, return them as text
	if len(statements) > 0 {
		logger.DebugfToFile("extractCreateStatements", "%s: Returning %d statements as text", command, len(statements))
		return strings.Join(statements, "\n\n")
	}
	
	logger.DebugfToFile("extractCreateStatements", "%s: No statements extracted, returning original result", command)
	// If we couldn't extract the statements, return as-is
	return serverResult
}