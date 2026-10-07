package ai

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/policy"
)

// ToolRequest represents a JSON tool request from the AI
type ToolRequest struct {
	Tool   string         `json:"tool"`
	Params map[string]any `json:"params"`
}

// ParsedCommand represents a parsed command with its type and arguments
type ParsedCommand struct {
	Tool ToolName
	Arg  string
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
			switch ch {
			case '{':
				braceCount++
			case '}':
				braceCount--
				if braceCount == 0 {
					return text[start : i+1]
				}
			}
		}
	}

	return ""
}

// IsInteractionNeeded returns true if user interaction is required
func (r CommandResult) IsInteractionNeeded() bool {
	return r.NeedsUserSelection || r.NeedsMoreInfo || r.NotRelevant
}

// Error implements the error interface for compatibility
func (i *InteractionRequest) Error() string {
	if i.Type == "selection" {
		return fmt.Sprintf("User selection needed for %s", i.SelectionType)
	}
	return i.InfoMessage
}

// ExecuteCommand executes a tool command for the CHAT view, which sees what
// the shell sees.
func ExecuteCommand(toolName ToolName, arg string) *CommandResult {
	return executeCommandFor(policy.Shell(), globalAI, toolName, arg)
}

// executeCommandFor executes a tool command under a policy: what the policy
// hides is left out of every answer, and a hidden table is answered for as one
// that is not there.
//
// a is the schema cache to answer from: the CHAT view's, or an MCP server's
// own, which is on a session of its own.
func executeCommandFor(p policy.Policy, a *AI, toolName ToolName, arg string) *CommandResult {
	if a == nil || a.cache == nil {
		return &CommandResult{
			Success: false,
			Error:   fmt.Errorf("AI system not initialized"),
		}
	}

	switch toolName {
	case ToolFuzzySearch:
		logger.DebugfToFile("CommandProcessor", "Executing fuzzy search for: %s", arg)

		if a.resolver != nil {
			// More than ten are asked for, so ten are left once the hidden
			// ones are taken out.
			var candidates []TableCandidate
			for _, c := range a.resolver.FindTablesWithFuzzy(arg, 50) {
				if p.VisibleTable(c.Keyspace, c.Table) && len(candidates) < 10 {
					candidates = append(candidates, c)
				}
			}
			logger.DebugfToFile("CommandProcessor", "Fuzzy search returned %d candidates", len(candidates))

			if len(candidates) > 0 {
				var sb strings.Builder
				fmt.Fprintf(&sb, "Found %d tables matching '%s':\n", len(candidates), arg)
				for _, c := range candidates {
					fmt.Fprintf(&sb, "- %s.%s (score: %.2f, columns: %v)\n",
						c.Keyspace, c.Table, c.Score, c.Columns)
				}
				return &CommandResult{Success: true, Data: sb.String()}
			}

			// No direct matches, show available keyspaces
			if keyspaces := visibleKeyspaces(p, a); len(keyspaces) > 0 {
				var sb strings.Builder
				fmt.Fprintf(&sb, "No tables found matching '%s'. Available keyspaces: %s\n",
					arg, strings.Join(keyspaces[:min(10, len(keyspaces))], ", "))
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
		if !p.VisibleTable(parts[0], parts[1]) {
			return &CommandResult{Success: true, Data: "Schema not found"}
		}
		schemaInfo, err := a.cache.GetTableSchema(parts[0], parts[1])
		if err != nil {
			return &CommandResult{Success: true, Data: "Schema not found"}
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Table %s.%s schema:\n", parts[0], parts[1])
		fmt.Fprintf(&sb, "Partition Keys: %v\n", schemaInfo.PartitionKeys)
		fmt.Fprintf(&sb, "Clustering Keys: %v\n", schemaInfo.ClusteringKeys)
		fmt.Fprintf(&sb, "Columns: %v", schemaInfo.Columns)
		return &CommandResult{Success: true, Data: sb.String()}

	case ToolListKeyspaces:
		logger.DebugfToFile("CommandProcessor", "Listing keyspaces")
		keyspaces := visibleKeyspaces(p, a)
		return &CommandResult{
			Success: true,
			Data:    fmt.Sprintf("Keyspaces: %s", strings.Join(keyspaces, ", ")),
		}

	case ToolListTables:
		logger.DebugfToFile("CommandProcessor", "Listing tables for keyspace: %s", arg)
		// Each with its key: a virtual table can only be filtered on its key,
		// and for a stored one the key is what says how to read it.
		var lines []string
		if p.Visible(arg) {
			for _, t := range a.cache.Tables[arg] {
				if !p.VisibleTable(arg, t.TableName) {
					continue
				}
				line := "- " + t.TableName
				if len(t.PartitionKeys) > 0 {
					line += " (partition key: " + strings.Join(t.PartitionKeys, ", ")
					if len(t.ClusteringKeys) > 0 {
						line += "; clustering: " + strings.Join(t.ClusteringKeys, ", ")
					}
					line += ")"
				}
				lines = append(lines, line)
			}
		}
		sort.Strings(lines)
		return &CommandResult{
			Success: true,
			Data:    fmt.Sprintf("Tables in %s:\n%s", arg, strings.Join(lines, "\n")),
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

// visibleKeyspaces is the cached keyspaces a policy lets be seen.
func visibleKeyspaces(p policy.Policy, a *AI) []string {
	var keyspaces []string
	for _, ks := range a.cache.Keyspaces {
		if p.Visible(ks) {
			keyspaces = append(keyspaces, ks)
		}
	}
	return keyspaces
}

// refreshIfChanged fetches the schema again when the cluster's schema version
// has moved since it was last fetched, which is one row to ask for. A
// long-running MCP server would otherwise answer from the schema it started
// with.
func (a *AI) refreshIfChanged() {
	if a == nil || a.cache == nil || a.session == nil {
		return
	}
	version, err := a.session.SchemaVersion()
	if err != nil {
		return
	}

	a.versionMu.Lock()
	defer a.versionMu.Unlock()
	if version == a.schemaVersion {
		return
	}
	// A different version, or none recorded - a cache built without saying
	// which schema it was built from - is fetched again.
	if err := a.cache.Refresh(); err != nil {
		logger.DebugfToFile("CommandProcessor", "Schema refresh failed: %v", err)
		return
	}
	a.resolver = NewResolver(a.cache)
	a.schemaVersion = version
}

// recordVersion notes which schema the cache was just built from, so the
// first tool call does not fetch it all again.
func (a *AI) recordVersion() {
	if a == nil || a.session == nil {
		return
	}
	version, err := a.session.SchemaVersion()
	if err != nil {
		return
	}
	a.versionMu.Lock()
	a.schemaVersion = version
	a.versionMu.Unlock()
}

// NewSchemaTools is a schema cache of its own on a session, for an MCP
// server: the CHAT view's is on the shell's session, which the server does not
// share.
func NewSchemaTools(session *db.Session) (*AI, error) {
	a, err := NewAIWithCache(session, &Config{Provider: "local", PrivacyMode: "schema_only"})
	if err != nil {
		return nil, err
	}
	a.recordVersion()
	return a, nil
}
