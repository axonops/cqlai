package router

import (
	"fmt"
	"strings"
)

// describeIndex shows detailed information about a specific index
func (v *CqlCommandVisitorImpl) describeIndex(indexName string) interface{} {
	serverResult, indexInfo, err := v.session.DBDescribeIndex(sessionManager, indexName)

	if err != nil {
		if err.Error() == "no keyspace selected" {
			return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
		}
		if strings.Contains(err.Error(), "not found") {
			return err.Error()
		}
		return fmt.Sprintf("Error: %v", err)
	}

	if serverResult != nil {
		// Server-side DESCRIBE result, return as-is
		return serverResult
	}

	// Manual query result - format it
	if indexInfo == nil {
		currentKeyspace := ""
		if sessionManager != nil {
			currentKeyspace = sessionManager.CurrentKeyspace()
		}
		return fmt.Sprintf("Index '%s' not found in keyspace '%s'", indexName, currentKeyspace)
	}

	currentKeyspace := ""
	if sessionManager != nil {
		currentKeyspace = sessionManager.CurrentKeyspace()
	}
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Index: %s.%s\n\n", currentKeyspace, indexName))

	// Format CREATE INDEX statement
	target := ""
	if t, ok := indexInfo.Options["target"]; ok {
		target = t
	}

	result.WriteString(fmt.Sprintf("CREATE INDEX %s ON %s.%s (%s);\n",
		indexName, currentKeyspace, indexInfo.TableName, target))

	result.WriteString(fmt.Sprintf("\nType: %s\n", indexInfo.Kind))
	result.WriteString(fmt.Sprintf("Table: %s\n", indexInfo.TableName))
	result.WriteString(fmt.Sprintf("Target: %s\n", target))

	// Show any additional options
	if len(indexInfo.Options) > 1 {
		result.WriteString("\nOptions:\n")
		for k, v := range indexInfo.Options {
			if k != "target" {
				result.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
			}
		}
	}

	return result.String()
}
