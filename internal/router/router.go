package router

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/session"
	"github.com/axonops/cqlai/internal/validation"
)

var metaHandler *MetaCommandHandler
var sessionManager *session.Manager

// InitRouter initializes the router with a session manager
func InitRouter(mgr *session.Manager) {
	sessionManager = mgr
}

// GetMetaHandler returns the current meta command handler
func GetMetaHandler() *MetaCommandHandler {
	return metaHandler
}

// stripComments removes SQL-style comments from a command while respecting quoted strings.
// It handles single-quoted strings, escaped quotes (''), and both line (--) and block (/* */) comments.
func stripComments(input string) string {
	var result strings.Builder
	result.Grow(len(input))

	inSingleQuote := false

	for i := 0; i < len(input); i++ {
		ch := input[i]

		// Handle single quotes
		if ch == '\'' {
			if inSingleQuote {
				// Check for escaped quote ''
				if i+1 < len(input) && input[i+1] == '\'' {
					result.WriteString("''")
					i++
					continue
				}
				// End of quoted string
				inSingleQuote = false
			} else {
				// Start of quoted string
				inSingleQuote = true
			}
			result.WriteByte(ch)
			continue
		}

		// If inside quotes, just copy the character
		if inSingleQuote {
			result.WriteByte(ch)
			continue
		}

		// Not in quotes - check for comments

		// Check for -- line comment
		if ch == '-' && i+1 < len(input) && input[i+1] == '-' {
			// Skip rest of input (single-line command context)
			break
		}

		// Check for // line comment
		if ch == '/' && i+1 < len(input) && input[i+1] == '/' {
			// Skip rest of input
			break
		}

		// Check for /* block comment */
		if ch == '/' && i+1 < len(input) && input[i+1] == '*' {
			i += 2 // Skip /*
			// Find closing */
			for i < len(input) {
				if input[i] == '*' && i+1 < len(input) && input[i+1] == '/' {
					i++ // Will be incremented again by loop
					break
				}
				i++
			}
			continue
		}

		// Regular character
		result.WriteByte(ch)
	}

	return strings.TrimSpace(result.String())
}

// ProcessCommand processes a user command.
func ProcessCommand(command string, session *db.Session, sessionMgr *session.Manager) interface{} {
	// Initialize meta handler if needed
	if metaHandler == nil {
		metaHandler = NewMetaCommandHandler(session, sessionMgr)
	}

	// Strip comments and trim the command
	command = stripComments(command)
	command = strings.TrimSpace(command)
	if command == "" {
		return ""
	}

	// Validate command syntax before processing
	if err := validation.ValidateCommandSyntax(command); err != nil {
		return err.Error()
	}

	// Handle AI command - now handled in UI layer
	// This is kept for backward compatibility but should not be used directly
	if strings.HasPrefix(strings.ToUpper(command), ".AI") {
		return "Please use the .ai command from the main interface for AI-powered CQL generation"
	}

	// Check if it's a meta-command (starts with certain keywords)
	// Trim semicolon for meta-command detection (meta-commands can have optional semicolons)
	trimmedCommand := strings.TrimSuffix(strings.TrimSpace(command), ";")
	upperCommand := strings.ToUpper(trimmedCommand)
	isMetaCommand := false
	metaCommands := []string{"DESCRIBE", "DESC", "CONSISTENCY", "OUTPUT", "PAGING", "AUTOFETCH", "TRACING", "SOURCE", "COPY", "SHOW", "EXPAND", "CAPTURE", "HELP", "SAVE"}

	logger.DebugfToFile("ProcessCommand", "Called with: '%s', trimmed: '%s', upper: '%s'", command, trimmedCommand, upperCommand)

	for _, meta := range metaCommands {
		// Check for word boundary: command equals meta OR starts with "meta "
		if upperCommand == meta || strings.HasPrefix(upperCommand, meta+" ") {
			isMetaCommand = true
			logger.DebugfToFile("ProcessCommand", "Detected meta-command (matched %s)", meta)
			break
		}
	}

	// Special handling for SHOW commands that might be CQL
	if (upperCommand == "SHOW" || strings.HasPrefix(upperCommand, "SHOW ")) &&
		!strings.Contains(upperCommand, "VERSION") &&
		!strings.Contains(upperCommand, "HOST") &&
		!strings.Contains(upperCommand, "SESSION") {
		// SHOW commands that aren't meta-commands should be treated as CQL
		isMetaCommand = false
	}

	if isMetaCommand {
		// Parse as meta-command
		logger.DebugToFile("ProcessCommand", "Routing to parseMetaCommand")
		return parseMetaCommand(command, session, sessionMgr)
	} else {
		// Check if we need to transform SELECT to SELECT JSON
		if sessionManager != nil && sessionManager.GetOutputFormat() == config.OutputFormatJSON {
			// Check if it's a SELECT query that should be transformed (with word boundary)
			if (upperCommand == "SELECT" || strings.HasPrefix(upperCommand, "SELECT ")) && !strings.Contains(upperCommand, "SELECT JSON") {
				// Use db.ConvertToJSONQuery which properly handles SELECT DISTINCT
				modifiedCommand := db.ConvertToJSONQuery(command)
				if modifiedCommand != command {
					logger.DebugfToFile("ProcessCommand", "Transformed query to: %s", modifiedCommand)
					return session.ExecuteCQLQuery(modifiedCommand)
				}
			}
		}
		// Execute as regular CQL query
		logger.DebugToFile("ProcessCommand", "Routing to executeCQLQuery")
		result := session.ExecuteCQLQuery(command)

		// Refresh schema cache if this was a DDL command
		refreshSchemaCacheIfNeeded(command, session)

		// Check if we should capture the result
		if metaHandler != nil && metaHandler.IsCapturing() {
			logger.DebugfToFile("ProcessCommand", "Capturing enabled, checking result type: %T", result)
			switch v := result.(type) {
			case db.QueryResult:
				if len(v.Data) > 0 {
					headers := v.Data[0]
					rows := [][]string{}
					if len(v.Data) > 1 {
						rows = v.Data[1:]
					}
					switch {
					case len(v.ColumnTypes) > 0:
						_ = metaHandler.WriteCaptureResultWithTypes(command, headers, v.ColumnTypes, rows, v.RawData)
					case len(v.RawData) > 0:
						_ = metaHandler.WriteCaptureResultWithRawData(command, headers, rows, v.RawData)
					default:
						_ = metaHandler.WriteCaptureResult(command, headers, rows)
					}
				}
			case db.StreamingQueryResult:
				// For streaming results, we need to fetch all rows
				defer v.Iterator.Close()

				rows := [][]string{}
				rawRows := []map[string]interface{}{}

				// Fetch rows from iterator
				for {
					row := make(map[string]interface{})
					if !v.Iterator.MapScan(row) {
						break
					}

					// Debug: log the first row's types
					if len(rawRows) == 0 && len(row) > 0 {
						for key, val := range row {
							logger.DebugfToFile("ProcessCommand", "Column %s: value type=%T, value=%v", key, val, val)
						}
					}

					// Convert to string array
					stringRow := make([]string, len(v.ColumnNames))
					for i, col := range v.ColumnNames {
						if val, ok := row[col]; ok {
							stringRow[i] = fmt.Sprintf("%v", val)
						} else {
							stringRow[i] = "null"
						}
					}
					rows = append(rows, stringRow)
					rawRows = append(rawRows, row)
				}

				// Write to capture file
				if len(rows) > 0 {
					switch {
					case len(v.ColumnTypes) > 0:
						_ = metaHandler.WriteCaptureResultWithTypes(command, v.Headers, v.ColumnTypes, rows, rawRows)
					default:
						_ = metaHandler.WriteCaptureResultWithRawData(command, v.Headers, rows, rawRows)
					}
				}

				// Convert streaming result to regular QueryResult so it can be displayed
				// Build data with headers as first row
				data := [][]string{v.Headers}
				data = append(data, rows...)
				logger.DebugfToFile("ProcessCommand", "Converting StreamingQueryResult to QueryResult: headers=%d, data rows=%d, total data=%d", len(v.Headers), len(rows), len(data))
				result = db.QueryResult{
					Data:        data,
					ColumnTypes: v.ColumnTypes,
					RawData:     rawRows,
				}
				logger.DebugfToFile("ProcessCommand", "Converted result type: %T, Data length: %d", result, len(result.(db.QueryResult).Data))
			}
		}

		return result
	}
}

// parseMetaCommand parses and executes meta-commands
func parseMetaCommand(command string, session *db.Session, sessionMgr *session.Manager) interface{} {
	// Strip trailing semicolon if present (meta-commands don't need them)
	command = strings.TrimSpace(command)
	command = strings.TrimSuffix(command, ";")
	upperCommand := strings.ToUpper(strings.TrimSpace(command))

	// Handle OUTPUT command with simple string parsing
	if strings.HasPrefix(upperCommand, "OUTPUT") {
		return handleOutputCommand(command, session)
	}

	// Handle SAVE command
	if strings.HasPrefix(upperCommand, "SAVE") {
		logger.DebugfToFile("parseMetaCommand", "SAVE command detected: '%s', upperCommand: '%s'", command, upperCommand)
		// Parse the SAVE command (command already has semicolon stripped)
		cmd, err := ParseSaveCommand(command)
		if err != nil {
			logger.DebugfToFile("parseMetaCommand", "SAVE command error: %v", err)
			return fmt.Sprintf("Error: %v", err)
		}
		logger.DebugfToFile("parseMetaCommand", "SAVE command parsed: interactive=%v", cmd.Interactive)

		// Return the parsed command to the UI for execution
		return cmd
	}

	// Handle simple meta commands with the meta handler
	if strings.HasPrefix(upperCommand, "SHOW") ||
		strings.HasPrefix(upperCommand, "TRACING") ||
		strings.HasPrefix(upperCommand, "PAGING") ||
		strings.HasPrefix(upperCommand, "AUTOFETCH") ||
		strings.HasPrefix(upperCommand, "EXPAND") ||
		strings.HasPrefix(upperCommand, "SOURCE") ||
		strings.HasPrefix(upperCommand, "CAPTURE") ||
		strings.HasPrefix(upperCommand, "COPY") ||
		strings.HasPrefix(upperCommand, "HELP") ||
		strings.HasPrefix(upperCommand, "CONSISTENCY") {
		return metaHandler.HandleMetaCommand(command)
	}

	// Use the new CommandParser for DESCRIBE, LIST, GRANT, REVOKE commands
	parser := NewCommandParser(session, metaHandler, sessionMgr)
	return parser.ParseCommand(command)
}

// handleOutputCommand handles the OUTPUT command with simple string parsing
func handleOutputCommand(command string, session *db.Session) interface{} {
	if sessionManager == nil {
		return "Session manager not initialized"
	}

	parts := strings.Fields(strings.ToUpper(strings.TrimSpace(command)))

	// If just "OUTPUT", show current format
	if len(parts) == 1 {
		currentFormat := sessionManager.GetOutputFormat()
		return fmt.Sprintf("Current output format is %s", currentFormat)
	}

	// If "OUTPUT <format>", set the format
	if len(parts) == 2 {
		formatStr := parts[1]
		format, err := config.ParseOutputFormat(formatStr)
		if err != nil {
			return fmt.Sprintf("Invalid output format '%s'. Valid formats are: TABLE, ASCII, EXPAND, JSON", formatStr)
		}
		if err := sessionManager.SetOutputFormat(format); err != nil {
			return fmt.Sprintf("Failed to set output format: %v", err)
		}
		return fmt.Sprintf("Now using %s output format", formatStr)
	}

	return "Usage: OUTPUT [TABLE|ASCII|EXPAND|JSON]"
}

// IsDangerousCommand checks if a command requires confirmation
// Delegates to the validation package
func IsDangerousCommand(command string) bool {
	return validation.IsDangerousCommand(command)
}

// isDDLCommand checks if a command is a DDL statement that modifies schema
func isDDLCommand(command string) bool {
	upperCommand := strings.ToUpper(strings.TrimSpace(command))

	// Check for schema-modifying DDL commands
	ddlKeywords := []string{
		"CREATE KEYSPACE",
		"CREATE TABLE",
		"CREATE TYPE",
		"CREATE INDEX",
		"CREATE MATERIALIZED VIEW",
		"ALTER KEYSPACE",
		"ALTER TABLE",
		"ALTER TYPE",
		"DROP KEYSPACE",
		"DROP TABLE",
		"DROP TYPE",
		"DROP INDEX",
		"DROP MATERIALIZED VIEW",
		"TRUNCATE",
	}

	for _, keyword := range ddlKeywords {
		if strings.HasPrefix(upperCommand, keyword) {
			return true
		}
	}

	return false
}

// refreshSchemaCacheIfNeeded refreshes the schema cache if the command modified schema
func refreshSchemaCacheIfNeeded(command string, session *db.Session) {
	if session == nil {
		return
	}

	if isDDLCommand(command) {
		logger.DebugfToFile("ProcessCommand", "DDL command detected, refreshing schema cache")
		if cache := session.GetSchemaCache(); cache != nil {
			if err := cache.Refresh(); err != nil {
				logger.DebugfToFile("ProcessCommand", "Failed to refresh schema cache: %v", err)
			} else {
				logger.DebugfToFile("ProcessCommand", "Schema cache refreshed successfully")
			}
		}
	}
}
