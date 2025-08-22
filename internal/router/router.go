package router

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/parser/grammar"
	"github.com/axonops/cqlai/internal/validation"
)


var metaHandler *MetaCommandHandler

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
	metaCommands := []string{"DESCRIBE", "DESC", "CONSISTENCY", "PAGING", "TRACING", "SOURCE", "COPY", "SHOW", "EXPAND", "CAPTURE", "HELP"}
	
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
	upperCommand := strings.ToUpper(strings.TrimSpace(command))
	
	// Handle non-DESCRIBE meta commands with the meta handler
	if strings.HasPrefix(upperCommand, "CONSISTENCY") ||
		strings.HasPrefix(upperCommand, "SHOW") ||
		strings.HasPrefix(upperCommand, "TRACING") ||
		strings.HasPrefix(upperCommand, "PAGING") ||
		strings.HasPrefix(upperCommand, "EXPAND") ||
		strings.HasPrefix(upperCommand, "SOURCE") ||
		strings.HasPrefix(upperCommand, "CAPTURE") ||
		strings.HasPrefix(upperCommand, "HELP") {
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

// IsDangerousCommand checks if a command requires confirmation
// Delegates to the validation package
func IsDangerousCommand(command string) bool {
	return validation.IsDangerousCommand(command)
}
