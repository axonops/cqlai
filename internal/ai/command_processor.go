package ai

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/logger"
)

// CommandType represents the type of tool command
type CommandType string

const (
	CommandFuzzySearch   CommandType = "FUZZY_SEARCH"
	CommandGetSchema     CommandType = "GET_SCHEMA"
	CommandListKeyspaces CommandType = "LIST_KEYSPACES"
	CommandListTables    CommandType = "LIST_TABLES"
	CommandUserSelection CommandType = "USER_SELECTION"
	CommandNotEnoughInfo CommandType = "NOT_ENOUGH_INFO"
)

// ParseCommand parses a response to extract tool commands
func ParseCommand(response string) (CommandType, string, bool) {
	response = strings.TrimSpace(response)
	lines := strings.Split(response, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "FUZZY_SEARCH:") {
			query := strings.TrimPrefix(line, "FUZZY_SEARCH:")
			return CommandFuzzySearch, query, true
		}
		
		if strings.HasPrefix(line, "GET_SCHEMA:") {
			tableRef := strings.TrimPrefix(line, "GET_SCHEMA:")
			return CommandGetSchema, tableRef, true
		}
		
		if line == "LIST_KEYSPACES" {
			return CommandListKeyspaces, "", true
		}
		
		if strings.HasPrefix(line, "LIST_TABLES:") {
			keyspace := strings.TrimPrefix(line, "LIST_TABLES:")
			return CommandListTables, keyspace, true
		}
		
		if strings.HasPrefix(line, "USER_SELECTION:") {
			selectionInfo := strings.TrimPrefix(line, "USER_SELECTION:")
			return CommandUserSelection, selectionInfo, true
		}
		
		if strings.HasPrefix(line, "NOT_ENOUGH_INFO:") {
			reason := strings.TrimPrefix(line, "NOT_ENOUGH_INFO:")
			return CommandNotEnoughInfo, reason, true
		}
	}
	
	return "", "", false
}

// CommandResult represents the result of executing a command
type CommandResult struct {
	// Success case - command executed and returned data
	Success bool
	Data    string
	
	// User interaction needed
	NeedsUserSelection bool
	SelectionType      string
	SelectionOptions   []string
	
	NeedsMoreInfo bool
	InfoMessage   string
	
	// Error case
	Error error
}

// IsInteractionNeeded returns true if user interaction is required
func (r CommandResult) IsInteractionNeeded() bool {
	return r.NeedsUserSelection || r.NeedsMoreInfo
}

// InteractionRequest represents a request for user interaction
type InteractionRequest struct {
	Type             string   // "selection" or "info"
	SelectionType    string   // For selection: type of item to select
	SelectionOptions []string // For selection: available options
	InfoMessage      string   // For info: message to show user
	ConversationID   string   // ID of the conversation to resume
}

// Error implements the error interface for compatibility
func (i *InteractionRequest) Error() string {
	if i.Type == "selection" {
		return fmt.Sprintf("User selection needed for %s", i.SelectionType)
	}
	return i.InfoMessage
}

// ExecuteCommand executes a tool command and returns the result
func ExecuteCommand(cmd CommandType, arg string) CommandResult {
	if globalAI == nil || globalAI.cache == nil {
		return CommandResult{
			Success: false,
			Error:   fmt.Errorf("AI system not initialized"),
		}
	}
	
	switch cmd {
	case CommandFuzzySearch:
		logger.DebugfToFile("CommandProcessor", "Executing fuzzy search for: %s", arg)
		
		if globalAI.resolver != nil {
			candidates := globalAI.resolver.FindTablesWithFuzzy(arg, 10)
			logger.DebugfToFile("CommandProcessor", "Fuzzy search returned %d candidates", len(candidates))
			
			if len(candidates) > 0 {
				result := fmt.Sprintf("Found %d tables matching '%s':\n", len(candidates), arg)
				for _, c := range candidates {
					result += fmt.Sprintf("- %s.%s (score: %.2f, columns: %v)\n",
						c.Keyspace, c.Table, c.Score, c.Columns)
				}
				return CommandResult{Success: true, Data: result}
			}
			
			// No direct matches, show available keyspaces
			if len(globalAI.cache.Keyspaces) > 0 {
				result := fmt.Sprintf("No tables found matching '%s'. Available keyspaces: %s\n",
					arg, strings.Join(globalAI.cache.Keyspaces[:min(10, len(globalAI.cache.Keyspaces))], ", "))
				result += "\nTry searching with a different term or use LIST_TABLES:<keyspace> to see tables in a specific keyspace."
				return CommandResult{Success: true, Data: result}
			}
			
			return CommandResult{
				Success: true,
				Data:    fmt.Sprintf("No tables found matching '%s' and no keyspaces available.", arg),
			}
		}
		return CommandResult{
			Success: false,
			Error:   fmt.Errorf("resolver not available"),
		}
		
	case CommandGetSchema:
		parts := strings.Split(arg, ".")
		if len(parts) != 2 {
			return CommandResult{
				Success: false,
				Error:   fmt.Errorf("invalid table reference: %s (expected keyspace.table)", arg),
			}
		}
		
		logger.DebugfToFile("CommandProcessor", "Getting schema for: %s", arg)
		schemaInfo, err := globalAI.cache.GetTableSchema(parts[0], parts[1])
		if err != nil {
			return CommandResult{Success: true, Data: "Schema not found"}
		}
		
		result := fmt.Sprintf("Table %s.%s schema:\n", parts[0], parts[1])
		result += fmt.Sprintf("Partition Keys: %v\n", schemaInfo.PartitionKeys)
		result += fmt.Sprintf("Clustering Keys: %v\n", schemaInfo.ClusteringKeys)
		result += fmt.Sprintf("Columns: %v", schemaInfo.Columns)
		return CommandResult{Success: true, Data: result}
		
	case CommandListKeyspaces:
		logger.DebugfToFile("CommandProcessor", "Listing keyspaces")
		keyspaces := globalAI.cache.Keyspaces
		return CommandResult{
			Success: true,
			Data:    fmt.Sprintf("Keyspaces: %s", strings.Join(keyspaces, ", ")),
		}
		
	case CommandListTables:
		logger.DebugfToFile("CommandProcessor", "Listing tables for keyspace: %s", arg)
		tables := globalAI.cache.Tables[arg]
		tableNames := []string{}
		for _, t := range tables {
			tableNames = append(tableNames, t.TableName)
		}
		return CommandResult{
			Success: true,
			Data:    fmt.Sprintf("Tables in %s: %s", arg, strings.Join(tableNames, ", ")),
		}
		
	case CommandUserSelection:
		logger.DebugfToFile("CommandProcessor", "User selection requested: %s", arg)
		
		// Parse the type and values from arg (format: type:value1,value2,value3 or type:["value1","value2"])
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 {
			return CommandResult{
				Success: false,
				Error:   fmt.Errorf("invalid USER_SELECTION format: %s (expected type:values)", arg),
			}
		}
		
		selectionType := parts[0]
		values := parts[1]
		
		// Parse options - handle both comma-separated and JSON array formats
		var options []string
		if strings.HasPrefix(values, "[") && strings.HasSuffix(values, "]") {
			// JSON array format: ["option1", "option2"]
			values = strings.TrimPrefix(values, "[")
			values = strings.TrimSuffix(values, "]")
			// Split and clean up quotes
			parts := strings.Split(values, ",")
			for _, part := range parts {
				cleaned := strings.TrimSpace(part)
				cleaned = strings.Trim(cleaned, `"`)
				if cleaned != "" {
					options = append(options, cleaned)
				}
			}
		} else {
			// Simple comma-separated format: option1,option2,option3
			options = strings.Split(values, ",")
		}
		
		// Return result indicating user selection is needed
		return CommandResult{
			Success:            false, // Not a success yet - needs user input
			NeedsUserSelection: true,
			SelectionType:      selectionType,
			SelectionOptions:   options,
		}
		
	case CommandNotEnoughInfo:
		logger.DebugfToFile("CommandProcessor", "Not enough information: %s", arg)
		
		// The arg contains the message from AI requesting more information
		return CommandResult{
			Success:       false, // Not a success yet - needs user input
			NeedsMoreInfo: true,
			InfoMessage:   arg,
		}
		
	default:
		return CommandResult{
			Success: false,
			Error:   fmt.Errorf("unknown command type: %s", cmd),
		}
	}
}