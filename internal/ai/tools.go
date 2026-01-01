package ai

import (
	"fmt"
	
	"github.com/axonops/cqlai/internal/logger"
)

// Note: System prompts have been moved to prompts.go
// Use SystemPrompt from that file for all AI interactions

// ExecuteToolCallTyped executes a tool with typed parameters
func ExecuteToolCallTyped(toolName ToolName, params ToolParams) *CommandResult {
	// Validate parameters
	if err := params.Validate(); err != nil {
		return &CommandResult{
			Error: fmt.Errorf("invalid parameters for %s: %w", toolName, err),
		}
	}

	// Execute based on tool type with typed parameters
	switch toolName {
	case ToolFuzzySearch:
		p := params.(FuzzySearchParams)
		return ExecuteCommand(ToolFuzzySearch, p.Query)

	case ToolGetSchema:
		p := params.(GetSchemaParams)
		return ExecuteCommand(ToolGetSchema, fmt.Sprintf("%s.%s", p.Keyspace, p.Table))

	case ToolListKeyspaces:
		return ExecuteCommand(ToolListKeyspaces, "")

	case ToolListTables:
		p := params.(ListTablesParams)
		return ExecuteCommand(ToolListTables, p.Keyspace)

	case ToolUserSelection:
		p := params.(UserSelectionParams)
		return &CommandResult{
			NeedsUserSelection: true,
			SelectionType:      p.Type,
			SelectionOptions:   p.Options,
		}

	case ToolNotEnoughInfo:
		p := params.(InfoMessageParams)
		return &CommandResult{
			NeedsMoreInfo: true,
			InfoMessage:   p.Message,
		}

	case ToolNotRelevant:
		p := params.(InfoMessageParams)
		return &CommandResult{
			NotRelevant: true,
			InfoMessage: p.Message,
		}

	case ToolSubmitQueryPlan:
		p := params.(SubmitQueryPlanParams)
		// Convert to QueryPlan and mark as successful submission
		plan := p.ToQueryPlan()
		// We'll return the plan as a special success case
		return &CommandResult{
			Success: true,
			Data:    "QUERY_PLAN_SUBMITTED", // Special marker
			// Store the plan in a way that can be retrieved
			QueryPlan: plan,
		}

	case ToolInfo:
		p := params.(*InfoResponseParams)
		logger.DebugfToFile("Tools", "Info tool called with response_type=%s, title=%s, content=%s", 
			p.ResponseType, p.Title, p.Content)
		// Return info response as a special success case
		return &CommandResult{
			Success:      true,
			Data:         fmt.Sprintf("Informational response provided: %s", p.Title), // More meaningful message
			InfoResponse: p, // Store the info response
		}

	default:
		return &CommandResult{
			Error: fmt.Errorf("unknown tool: %s", toolName),
		}
	}
}

// ExecuteToolCall is a common function to execute a tool based on its name and arguments
// This function is shared across all AI clients to ensure consistent behavior
// Deprecated: Use ExecuteToolCallTyped with proper typed parameters instead
func ExecuteToolCall(toolName string, args map[string]any) *CommandResult {
	// Convert string to ToolName
	tool := ParseToolName(toolName)
	if !tool.IsValid() {
		return &CommandResult{
			Error: fmt.Errorf("invalid tool name: %s", toolName),
		}
	}
	
	// Parse the arguments into typed parameters
	params, err := ParseToolParamsFromMap(tool, args)
	if err != nil {
		return &CommandResult{
			Error: fmt.Errorf("failed to parse tool parameters: %w", err),
		}
	}

	return ExecuteToolCallTyped(tool, params)
}

// ToolDefinition represents a common tool definition structure
type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]any
	Required    []string
}

// GetCommonToolDefinitions returns the standard tool definitions used across all AI providers
func GetCommonToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        ToolFuzzySearch.String(),
			Description: "Search for tables or keyspaces matching a search term",
			Parameters: map[string]any{
				"query": map[string]any{
					"type":        "string",
					"description": "The search term to find matching tables or keyspaces",
				},
			},
			Required: []string{"query"},
		},
		{
			Name:        ToolGetSchema.String(),
			Description: "Get the complete schema of a specific table including columns and their types",
			Parameters: map[string]any{
				"keyspace": map[string]any{
					"type":        "string",
					"description": "The keyspace name",
				},
				"table": map[string]any{
					"type":        "string",
					"description": "The table name",
				},
			},
			Required: []string{"keyspace", "table"},
		},
		{
			Name:        ToolListKeyspaces.String(),
			Description: "List all available keyspaces in the Cassandra cluster",
			Parameters:  map[string]any{},
			Required:    []string{},
		},
		{
			Name:        ToolListTables.String(),
			Description: "List all tables in a specific keyspace",
			Parameters: map[string]any{
				"keyspace": map[string]any{
					"type":        "string",
					"description": "The keyspace name to list tables from",
				},
			},
			Required: []string{"keyspace"},
		},
		{
			Name:        ToolUserSelection.String(),
			Description: "Ask the user to select from a list of options when there's ambiguity",
			Parameters: map[string]any{
				"type": map[string]any{
					"type":        "string",
					"description": "The type of selection",
					"enum":        []string{"keyspace", "table", "column", "index", "type", "function", "aggregate", "role"},
				},
				"options": map[string]any{
					"type":        "array",
					"description": "List of options for the user to select from",
					"items": map[string]any{
						"type": "string",
					},
				},
			},
			Required: []string{"type", "options"},
		},
		{
			Name:        ToolNotEnoughInfo.String(),
			Description: "Request more information from the user when the request is too vague",
			Parameters: map[string]any{
				"message": map[string]any{
					"type":        "string",
					"description": "Message asking the user for more specific details about their request",
				},
			},
			Required: []string{"message"},
		},
		{
			Name:        ToolNotRelevant.String(),
			Description: "Indicate that the request is not related to CQL or Cassandra",
			Parameters: map[string]any{
				"message": map[string]any{
					"type":        "string",
					"description": "Message explaining why the request is not relevant to Cassandra/CQL",
				},
			},
			Required: []string{"message"},
		},
		{
			Name:        ToolSubmitQueryPlan.String(),
			Description: "Submit the final CQL query plan based on gathered information. Supports ALL CQL features including collections, counters, LWTs, BATCH, etc.",
			Parameters: map[string]any{
				"operation": map[string]any{
					"type":        "string",
					"description": "The CQL operation type",
					"enum": []string{
						// DQL - Data Query Language
						"SELECT", "DESCRIBE", "LIST", "SHOW",
						// DML - Data Manipulation Language
						"INSERT", "UPDATE", "DELETE", "BATCH",
						// DDL - Data Definition Language
						"CREATE", "ALTER", "DROP", "TRUNCATE", "USE",
						// DCL - Data Control Language
						"GRANT", "REVOKE",
					},
				},
				"keyspace": map[string]any{
					"type":        "string",
					"description": "The keyspace name (optional)",
				},
				"table": map[string]any{
					"type":        "string",
					"description": "The table name",
				},
				"columns": map[string]any{
					"type":        "array",
					"description": "Column names for SELECT or INSERT. Supports functions: COUNT(*), WRITETIME(col), TTL(col), CAST(x AS type), col[key], col.field",
					"items":       map[string]any{"type": "string"},
				},
				"values": map[string]any{
					"type":        "object",
					"description": "Values for INSERT or UPDATE operations. Use uuid(), now() for functions. Use lists [], sets {}, maps {'k':'v'}",
				},
				"value_types": map[string]any{
					"type":        "object",
					"description": "Type hints for values (col -> 'list<text>', 'set<int>', 'map<text,int>') - helps format collections correctly",
				},
				"counter_ops": map[string]any{
					"type":        "object",
					"description": "Counter operations (col -> '+5' or '-2') for counter increment/decrement",
				},
				"collection_ops": map[string]any{
					"type":        "object",
					"description": "Collection operations: append, prepend, add, remove, merge, set_element, set_index, set_field",
				},
				"where": map[string]any{
					"type":        "array",
					"description": "WHERE conditions. Supports CONTAINS, CONTAINS KEY, TOKEN(), tuple notation",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"column":    map[string]any{"type": "string"},
							"operator":  map[string]any{"type": "string", "description": "=, <, >, <=, >=, IN, CONTAINS, CONTAINS KEY"},
							"value":     map[string]any{},
							"is_token":  map[string]any{"type": "boolean", "description": "Wrap column in TOKEN()"},
							"columns":   map[string]any{"type": "array", "description": "For tuple notation"},
							"values":    map[string]any{"type": "array", "description": "For tuple notation"},
						},
					},
				},
				"if_not_exists": map[string]any{
					"type":        "boolean",
					"description": "INSERT IF NOT EXISTS or CREATE IF NOT EXISTS",
				},
				"if_exists": map[string]any{
					"type":        "boolean",
					"description": "UPDATE/DELETE/ALTER/DROP IF EXISTS",
				},
				"if_conditions": map[string]any{
					"type":        "array",
					"description": "LWT conditions: UPDATE/DELETE IF col = val",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"column":   map[string]any{"type": "string"},
							"operator": map[string]any{"type": "string"},
							"value":    map[string]any{},
						},
					},
				},
				"using_ttl": map[string]any{
					"type":        "integer",
					"description": "TTL in seconds for INSERT/UPDATE USING TTL",
				},
				"using_timestamp": map[string]any{
					"type":        "integer",
					"description": "Timestamp in microseconds for INSERT/UPDATE/DELETE USING TIMESTAMP",
				},
				"distinct": map[string]any{
					"type":        "boolean",
					"description": "SELECT DISTINCT (partition key columns only)",
				},
				"select_json": map[string]any{
					"type":        "boolean",
					"description": "SELECT JSON output format",
				},
				"per_partition_limit": map[string]any{
					"type":        "integer",
					"description": "PER PARTITION LIMIT for SELECT",
				},
				"insert_json": map[string]any{
					"type":        "boolean",
					"description": "INSERT JSON mode",
				},
				"json_value": map[string]any{
					"type":        "string",
					"description": "JSON string for INSERT JSON",
				},
				"batch_type": map[string]any{
					"type":        "string",
					"description": "BATCH type: LOGGED (default), UNLOGGED, or COUNTER",
					"enum":        []string{"LOGGED", "UNLOGGED", "COUNTER"},
				},
				"batch_statements": map[string]any{
					"type":        "array",
					"description": "Statements in the BATCH (array of query plans)",
					"items":       map[string]any{"type": "object"},
				},
				"order_by": map[string]any{
					"type":        "array",
					"description": "ORDER BY clauses",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"column": map[string]any{"type": "string"},
							"order":  map[string]any{"type": "string", "enum": []string{"ASC", "DESC"}},
						},
					},
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "LIMIT for SELECT queries",
				},
				"allow_filtering": map[string]any{
					"type":        "boolean",
					"description": "Whether to use ALLOW FILTERING",
				},
				"schema": map[string]any{
					"type":        "object",
					"description": "Column definitions for CREATE TABLE (key: column name, value: column type with constraints)",
				},
				"options": map[string]any{
					"type":        "object",
					"description": "Operation-specific parameters. Supports: if_not_exists, if_exists, custom_index, using_class, or_replace, and more",
				},
				"confidence": map[string]any{
					"type":        "number",
					"description": "Confidence level (0.0-1.0) in the query plan",
					"minimum":     0.0,
					"maximum":     1.0,
				},
				"warning": map[string]any{
					"type":        "string",
					"description": "Optional warning about the query",
				},
				"read_only": map[string]any{
					"type":        "boolean",
					"description": "Whether this is a read-only operation",
				},
			},
			Required: []string{"operation"},
		},
		{
			Name:        ToolInfo.String(),
			Description: "Submit an informational response (no CQL execution)",
			Parameters: map[string]any{
				"response_type": map[string]any{
					"type":        "string",
					"description": "Type of response: 'text' (default) or 'schema_info'",
					"enum":        []string{"text", "schema_info"},
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Optional title for the response",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "The text content to display",
				},
				"schema_info": map[string]any{
					"type":        "object",
					"description": "Structured schema information if response_type is 'schema_info'",
				},
				"confidence": map[string]any{
					"type":        "number",
					"description": "Confidence level (0.0-1.0) in the response",
					"minimum":     0.0,
					"maximum":     1.0,
				},
			},
			Required: []string{"content"},
		},
	}
}
