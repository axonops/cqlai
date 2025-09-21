package router

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/axonops/cqlai/internal/config"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/parser/grammar"
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

// stripComments removes SQL-style comments from a command
func stripComments(input string) string {
	// First handle line comments (-- and //)
	// These take precedence and terminate the line
	if idx := strings.Index(input, "--"); idx >= 0 {
		input = input[:idx]
	}
	if idx := strings.Index(input, "//"); idx >= 0 {
		input = input[:idx]
	}
	
	// Then handle block comments /* ... */
	for {
		startIdx := strings.Index(input, "/*")
		if startIdx < 0 {
			break
		}
		endIdx := strings.Index(input[startIdx:], "*/")
		if endIdx < 0 {
			// Block comment starts but doesn't end
			input = input[:startIdx]
			break
		}
		// Remove the block comment
		input = input[:startIdx] + input[startIdx+endIdx+2:]
	}
	
	return strings.TrimSpace(input)
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
	metaCommands := []string{"DESCRIBE", "DESC", "CONSISTENCY", "OUTPUT", "PAGING", "AUTOFETCH", "TRACING", "SOURCE", "COPY", "SHOW", "EXPAND", "CAPTURE", "HELP"}

	logger.DebugfToFile("ProcessCommand", "Called with: '%s', trimmed: '%s', upper: '%s'", command, trimmedCommand, upperCommand)

	for _, meta := range metaCommands {
		if strings.HasPrefix(upperCommand, meta) {
			isMetaCommand = true
			logger.DebugfToFile("ProcessCommand", "Detected meta-command (matched %s)", meta)
			break
		}
	}

	// Special handling for SHOW commands that might be CQL
	if strings.HasPrefix(upperCommand, "SHOW") && !strings.Contains(upperCommand, "VERSION") && !strings.Contains(upperCommand, "HOST") && !strings.Contains(upperCommand, "SESSION") {
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
			// Check if it's a SELECT query that should be transformed
			if strings.HasPrefix(upperCommand, "SELECT") && !strings.Contains(upperCommand, "SELECT JSON") {
				// Transform SELECT to SELECT JSON
				// Find the position of SELECT and insert JSON after it
				selectPos := strings.Index(upperCommand, "SELECT")
				if selectPos >= 0 {
					// Insert JSON after SELECT
					modifiedCommand := command[:selectPos+6] + " JSON" + command[selectPos+6:]
					logger.DebugfToFile("ProcessCommand", "Transformed query to: %s", modifiedCommand)
					return session.ExecuteCQLQuery(modifiedCommand)
				}
			}
		}
		// Execute as regular CQL query
		logger.DebugToFile("ProcessCommand", "Routing to executeCQLQuery")
		result := session.ExecuteCQLQuery(command)

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

	// Special handling for DESCRIBE shortcuts that cqlsh supports
	if strings.HasPrefix(upperCommand, "DESCRIBE ") || strings.HasPrefix(upperCommand, "DESC ") {
		// Extract the part after DESCRIBE/DESC
		parts := strings.Fields(command)
		if len(parts) == 2 {
			identifier := parts[1]
			upperIdentifier := strings.ToUpper(identifier)

			// Check if this is a special DESCRIBE command (KEYSPACES, TABLES, TYPES, etc.)
			// These should NOT be transformed
			switch {
			case upperIdentifier == "KEYSPACES" || upperIdentifier == "TABLES" ||
				upperIdentifier == "TYPES" || upperIdentifier == "FUNCTIONS" ||
				upperIdentifier == "AGGREGATES" || upperIdentifier == "CLUSTER":
				// Keep the command as-is - these are special DESCRIBE commands
				logger.DebugfToFile("parseMetaCommand", "Keeping '%s' as-is (special DESCRIBE command)", command)
			case strings.Contains(identifier, "."):
				// This looks like "DESCRIBE keyspace.table" or "DESC keyspace.table"
				// Transform it to "DESCRIBE TABLE keyspace.table"
				transformedCommand := parts[0] + " TABLE " + identifier
				logger.DebugfToFile("parseMetaCommand", "Transformed '%s' to '%s'", command, transformedCommand)
				command = transformedCommand
			default:
				// Single identifier - could be keyspace or table in current keyspace
				// For now, assume it's a keyspace (cqlsh behavior)
				// Transform it to "DESCRIBE KEYSPACE <identifier>"
				transformedCommand := parts[0] + " KEYSPACE " + identifier
				logger.DebugfToFile("parseMetaCommand", "Transformed '%s' to '%s' (assuming keyspace)", command, transformedCommand)
				command = transformedCommand
			}
		}
	}

	// DESCRIBE, LIST, and other complex commands use the ANTLR parser
	logger.DebugfToFile("parseMetaCommand", "Called with: '%s'", command)
	is := antlr.NewInputStream(command)
	lexer := grammar.NewCqlLexer(is)
	lexer.RemoveErrorListeners()
	lexerErrors := NewCustomErrorListener()
	lexer.AddErrorListener(lexerErrors)

	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := grammar.NewCqlParser(stream)
	p.RemoveErrorListeners()
	parserErrors := NewCustomErrorListener()
	p.AddErrorListener(parserErrors)

	tree := p.Root()

	if len(lexerErrors.Errors) > 0 || len(parserErrors.Errors) > 0 {
		// Debug: Log parsing errors
		if len(lexerErrors.Errors) > 0 {
			logger.DebugfToFile("parseMetaCommand", "Lexer errors: %v", lexerErrors.Errors)
		}
		if len(parserErrors.Errors) > 0 {
			logger.DebugfToFile("parseMetaCommand", "Parser errors: %v", parserErrors.Errors)
		}
		logger.DebugToFile("parseMetaCommand", "Parsing failed, falling back to CQL execution")
		// Fall back to CQL execution if parsing fails
		return session.ExecuteCQLQuery(command)
	}

	logger.DebugToFile("parseMetaCommand", "Parsing successful, visiting parse tree")
	commands := NewCqlCommandVisitorImpl(session)
	visitor := NewVisitor(commands)
	result := tree.Accept(visitor)

	logger.DebugfToFile("parseMetaCommand", "Parse tree visit completed, result type: %T", result)
	return result
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
		sessionManager.SetOutputFormat(format)
		return fmt.Sprintf("Now using %s output format", formatStr)
	}

	return "Usage: OUTPUT [TABLE|ASCII|EXPAND|JSON]"
}

// IsDangerousCommand checks if a command requires confirmation
// Delegates to the validation package
func IsDangerousCommand(command string) bool {
	return validation.IsDangerousCommand(command)
}
