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

// ProcessCommand processes a user command.
func ProcessCommand(command string, session *db.Session) interface{} {
	// Initialize meta handler if needed
	if metaHandler == nil {
		metaHandler = NewMetaCommandHandler(session)
	}

	// Trim the command
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
	upperCommand := strings.ToUpper(command)
	isMetaCommand := false
	metaCommands := []string{"DESCRIBE", "DESC", "CONSISTENCY", "OUTPUT", "PAGING", "TRACING", "SOURCE", "COPY", "SHOW", "EXPAND", "CAPTURE", "HELP"}

	logger.DebugfToFile("ProcessCommand", "Called with: '%s'", command)

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
		return parseMetaCommand(command, session)
	} else {
		// Execute as regular CQL query
		logger.DebugToFile("ProcessCommand", "Routing to executeCQLQuery")
		return session.ExecuteCQLQuery(command)
	}
}

// parseMetaCommand parses and executes meta-commands
func parseMetaCommand(command string, session *db.Session) interface{} {
	// Strip trailing semicolon if present (meta-commands don't need them)
	command = strings.TrimSpace(command)
	command = strings.TrimSuffix(command, ";")
	upperCommand := strings.ToUpper(strings.TrimSpace(command))

	// Handle OUTPUT command with simple string parsing
	if strings.HasPrefix(upperCommand, "OUTPUT") {
		return handleOutputCommand(command, session)
	}

	// Handle non-DESCRIBE meta commands with the meta handler
	if strings.HasPrefix(upperCommand, "SHOW") ||
		strings.HasPrefix(upperCommand, "TRACING") ||
		strings.HasPrefix(upperCommand, "PAGING") ||
		strings.HasPrefix(upperCommand, "EXPAND") ||
		strings.HasPrefix(upperCommand, "SOURCE") ||
		strings.HasPrefix(upperCommand, "CAPTURE") ||
		strings.HasPrefix(upperCommand, "HELP") ||
		strings.HasPrefix(upperCommand, "CONSISTENCY") {
		return metaHandler.HandleMetaCommand(command)
	}

	// DESCRIBE commands use the ANTLR parser
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
