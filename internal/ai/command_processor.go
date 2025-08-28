package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/logger"
)

// ToolRequest represents a JSON tool request from the AI
type ToolRequest struct {
	Tool   string            `json:"tool"`
	Params map[string]any `json:"params"`
}

// ParseCommand parses a response to extract tool commands (JSON format only)
func ParseCommand(response string) (ToolName, string, bool) {
	response = strings.TrimSpace(response)
	
	// Parse JSON command
	if toolReq, ok := parseJSONCommand(response); ok {
		return toolReq.Tool, toolReq.Arg, true
	}
	
	return "", "", false
}

// ParsedCommand represents a parsed command with its type and arguments
type ParsedCommand struct {
	Tool ToolName
	Arg  string
}

// parseJSONCommand attempts to parse a JSON tool request
func parseJSONCommand(response string) (*ParsedCommand, bool) {
	// Try to extract JSON from the response
	jsonStr := extractJSONObject(response)
	if jsonStr == "" {
		return nil, false
	}
	
	var toolReq ToolRequest
	if err := json.Unmarshal([]byte(jsonStr), &toolReq); err != nil {
		return nil, false
	}
	
	// Parse and validate tool name
	toolName := ParseToolName(toolReq.Tool)
	if !toolName.IsValid() {
		return nil, false
	}
	
	switch toolName {
	case ToolFuzzySearch:
		if query, ok := toolReq.Params["query"].(string); ok {
			return &ParsedCommand{Tool: ToolFuzzySearch, Arg: query}, true
		}
	
	case ToolGetSchema:
		keyspace, _ := toolReq.Params["keyspace"].(string)
		table, _ := toolReq.Params["table"].(string)
		if keyspace != "" && table != "" {
			return &ParsedCommand{Tool: ToolGetSchema, Arg: fmt.Sprintf("%s.%s", keyspace, table)}, true
		}
	
	case ToolListKeyspaces:
		return &ParsedCommand{Tool: ToolListKeyspaces, Arg: ""}, true
	
	case ToolListTables:
		if keyspace, ok := toolReq.Params["keyspace"].(string); ok {
			return &ParsedCommand{Tool: ToolListTables, Arg: keyspace}, true
		}
	
	case ToolUserSelection:
		selType, _ := toolReq.Params["type"].(string)
		options, _ := toolReq.Params["options"].([]any)
		if selType != "" && len(options) > 0 {
			// Convert options to string array
			strOptions := make([]string, len(options))
			for i, opt := range options {
				strOptions[i] = fmt.Sprintf("%v", opt)
			}
			// Format for existing handler
			arg := fmt.Sprintf("%s:%s", selType, strings.Join(strOptions, ","))
			return &ParsedCommand{Tool: ToolUserSelection, Arg: arg}, true
		}
	
	case ToolNotEnoughInfo:
		if msg, ok := toolReq.Params["message"].(string); ok {
			return &ParsedCommand{Tool: ToolNotEnoughInfo, Arg: msg}, true
		}
	
	case ToolNotRelevant:
		msg, _ := toolReq.Params["message"].(string)
		return &ParsedCommand{Tool: ToolNotRelevant, Arg: msg}, true
	}
	
	return nil, false
}

// extractJSONObject extracts a JSON object from text
func extractJSONObject(text string) string {
	// Look for JSON object starting with {
	start := strings.Index(text, "{")
	if start == -1 {
		return ""
	}
	
	// Find the matching closing brace
	braceCount := 0
	inString := false
	escape := false
	
	for i := start; i < len(text); i++ {
		ch := text[i]
		
		if escape {
			escape = false
			continue
		}
		
		if ch == '\\' {
			escape = true
			continue
		}
		
		if ch == '"' {
			inString = !inString
			continue
		}
		
		if !inString {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
				if braceCount == 0 {
					return text[start : i+1]
				}
			}
		}
	}
	
	return ""
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
	
	NotRelevant bool
	
	// Query plan submission (for submit_query_plan tool)
	QueryPlan *QueryPlan
	
	// Error case
	Error error
}

// IsInteractionNeeded returns true if user interaction is required
func (r CommandResult) IsInteractionNeeded() bool {
	return r.NeedsUserSelection || r.NeedsMoreInfo || r.NotRelevant
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
func ExecuteCommand(toolName ToolName, arg string) *CommandResult {
	if globalAI == nil || globalAI.cache == nil {
		return &CommandResult{
			Success: false,
			Error:   fmt.Errorf("AI system not initialized"),
		}
	}
	
	switch toolName {
	case ToolFuzzySearch:
		logger.DebugfToFile("CommandProcessor", "Executing fuzzy search for: %s", arg)
		
		if globalAI.resolver != nil {
			candidates := globalAI.resolver.FindTablesWithFuzzy(arg, 10)
			logger.DebugfToFile("CommandProcessor", "Fuzzy search returned %d candidates", len(candidates))
			
			if len(candidates) > 0 {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("Found %d tables matching '%s':\n", len(candidates), arg))
				for _, c := range candidates {
					sb.WriteString(fmt.Sprintf("- %s.%s (score: %.2f, columns: %v)\n",
						c.Keyspace, c.Table, c.Score, c.Columns))
				}
				return &CommandResult{Success: true, Data: sb.String()}
			}
			
			// No direct matches, show available keyspaces
			if len(globalAI.cache.Keyspaces) > 0 {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("No tables found matching '%s'. Available keyspaces: %s\n",
					arg, strings.Join(globalAI.cache.Keyspaces[:min(10, len(globalAI.cache.Keyspaces))], ", ")))
				sb.WriteString("\nTry searching with a different term or use LIST_TABLES:<keyspace> to see tables in a specific keyspace.")
				return &CommandResult{Success: true, Data: sb.String()}
			}
			
			return &CommandResult{
				Success: true,
				Data:    fmt.Sprintf("No tables found matching '%s' and no keyspaces available.", arg),
			}
		}
		return &CommandResult{
			Success: false,
			Error:   fmt.Errorf("resolver not available"),
		}
		
	case ToolGetSchema:
		parts := strings.Split(arg, ".")
		if len(parts) != 2 {
			return &CommandResult{
				Success: false,
				Error:   fmt.Errorf("invalid table reference: %s (expected keyspace.table)", arg),
			}
		}
		
		logger.DebugfToFile("CommandProcessor", "Getting schema for: %s", arg)
		schemaInfo, err := globalAI.cache.GetTableSchema(parts[0], parts[1])
		if err != nil {
			return &CommandResult{Success: true, Data: "Schema not found"}
		}
		
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Table %s.%s schema:\n", parts[0], parts[1]))
		sb.WriteString(fmt.Sprintf("Partition Keys: %v\n", schemaInfo.PartitionKeys))
		sb.WriteString(fmt.Sprintf("Clustering Keys: %v\n", schemaInfo.ClusteringKeys))
		sb.WriteString(fmt.Sprintf("Columns: %v", schemaInfo.Columns))
		return &CommandResult{Success: true, Data: sb.String()}
		
	case ToolListKeyspaces:
		logger.DebugfToFile("CommandProcessor", "Listing keyspaces")
		keyspaces := globalAI.cache.Keyspaces
		return &CommandResult{
			Success: true,
			Data:    fmt.Sprintf("Keyspaces: %s", strings.Join(keyspaces, ", ")),
		}
		
	case ToolListTables:
		logger.DebugfToFile("CommandProcessor", "Listing tables for keyspace: %s", arg)
		tables := globalAI.cache.Tables[arg]
		tableNames := []string{}
		for _, t := range tables {
			tableNames = append(tableNames, t.TableName)
		}
		return &CommandResult{
			Success: true,
			Data:    fmt.Sprintf("Tables in %s: %s", arg, strings.Join(tableNames, ", ")),
		}
		
	case ToolUserSelection:
		logger.DebugfToFile("CommandProcessor", "User selection requested: %s", arg)
		
		// Parse the type and values from arg (format: type:value1,value2,value3 or type:["value1","value2"])
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 {
			return &CommandResult{
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
		return &CommandResult{
			Success:            false, // Not a success yet - needs user input
			NeedsUserSelection: true,
			SelectionType:      selectionType,
			SelectionOptions:   options,
		}
		
	case ToolNotEnoughInfo:
		logger.DebugfToFile("CommandProcessor", "Not enough information: %s", arg)
		
		// The arg contains the message from AI requesting more information
		return &CommandResult{
			Success:       false, // Not a success yet - needs user input
			NeedsMoreInfo: true,
			InfoMessage:   arg,
		}
		
	case ToolNotRelevant:
		logger.DebugfToFile("CommandProcessor", "Request not relevant to CQL: %s", arg)
		
		// The request is not relevant to CQL/Cassandra
		message := "This request is not related to Cassandra or CQL."
		if arg != "" {
			message = arg
		}
		return &CommandResult{
			NotRelevant: true,
			InfoMessage: message,
		}
		
	default:
		return &CommandResult{
			Success: false,
			Error:   fmt.Errorf("unknown tool: %s", toolName),
		}
	}
}