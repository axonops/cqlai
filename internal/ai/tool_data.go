package ai

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/db"
)

// getToolData retrieves raw data for a tool (shared by REPL and MCP).
// This function contains only the data retrieval logic, not formatting.
// REPL formats results as terminal strings, MCP returns raw data as JSON.
//
// Parameters:
//   - resolver: Fuzzy search resolver (can be nil for tools that don't need it)
//   - cache: Schema cache (required for most tools)
//   - toolName: Which tool to execute
//   - arg: Tool-specific argument string
//
// Returns:
//   - data: Raw data (type depends on tool)
//   - error: Error if operation failed
func getToolData(resolver *Resolver, cache *db.SchemaCache, toolName ToolName, arg string) (any, error) {
	switch toolName {
	case ToolFuzzySearch:
		if resolver == nil {
			return nil, fmt.Errorf("resolver not available")
		}
		// Returns []TableCandidate (defined in fuzzy_search.go)
		candidates := resolver.FindTablesWithFuzzy(arg, 10)
		return candidates, nil

	case ToolGetSchema:
		if cache == nil {
			return nil, fmt.Errorf("schema cache not available")
		}

		parts := strings.Split(arg, ".")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid table reference: %s (expected keyspace.table)", arg)
		}

		// Returns *db.TableSchema
		schema, err := cache.GetTableSchema(parts[0], parts[1])
		if err != nil {
			return nil, fmt.Errorf("schema not found for %s: %w", arg, err)
		}
		return schema, nil

	case ToolListKeyspaces:
		if cache == nil {
			return nil, fmt.Errorf("schema cache not available")
		}

		// Returns []string
		keyspaces := cache.Keyspaces
		return keyspaces, nil

	case ToolListTables:
		if cache == nil {
			return nil, fmt.Errorf("schema cache not available")
		}

		// Returns []db.CachedTableInfo
		tables := cache.Tables[arg]
		if tables == nil {
			// Return empty slice instead of nil for consistent behavior
			return []db.CachedTableInfo{}, nil
		}
		return tables, nil

	case ToolUserSelection, ToolNotEnoughInfo, ToolNotRelevant:
		// These are UI/control flow tools, not data retrieval
		// They don't access Cassandra or cache, just return the argument
		return arg, nil

	case ToolSubmitQueryPlan, ToolInfo:
		// These are handled in ExecuteToolCallTyped, not ExecuteCommand
		return nil, fmt.Errorf("tool %s should be handled by ExecuteToolCallTyped", toolName)

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}
