package router

import (
	"fmt"
	"strings"
)

// describeMaterializedView shows detailed information about a materialized view
func (v *CqlCommandVisitorImpl) describeMaterializedView(viewName string) interface{} {
	serverResult, mvInfo, err := v.session.DBDescribeMaterializedView(viewName)
	
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
	if mvInfo == nil {
		currentKeyspace := v.session.CurrentKeyspace()
		return fmt.Sprintf("Materialized view '%s' not found in keyspace '%s'", viewName, currentKeyspace)
	}
	
	currentKeyspace := v.session.CurrentKeyspace()
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Materialized View: %s.%s\n\n", currentKeyspace, viewName))
	result.WriteString(fmt.Sprintf("Base Table: %s\n", mvInfo.BaseTable))
	if mvInfo.WhereClause != "" {
		result.WriteString(fmt.Sprintf("Where Clause: %s\n", mvInfo.WhereClause))
	}
	result.WriteString("\n")

	// Show CREATE MATERIALIZED VIEW statement
	result.WriteString(fmt.Sprintf("CREATE MATERIALIZED VIEW %s.%s AS\n", currentKeyspace, viewName))
	result.WriteString(fmt.Sprintf("  SELECT * FROM %s.%s\n", currentKeyspace, mvInfo.BaseTable))
	if mvInfo.WhereClause != "" {
		result.WriteString(fmt.Sprintf("  WHERE %s\n", mvInfo.WhereClause))
	}

	// Format primary key
	if len(mvInfo.PartitionKeys) > 0 {
		result.WriteString("  PRIMARY KEY (")
		if len(mvInfo.PartitionKeys) == 1 {
			result.WriteString(mvInfo.PartitionKeys[0])
		} else {
			result.WriteString(fmt.Sprintf("(%s)", strings.Join(mvInfo.PartitionKeys, ", ")))
		}
		if len(mvInfo.ClusteringKeys) > 0 {
			result.WriteString(", " + strings.Join(mvInfo.ClusteringKeys, ", "))
		}
		result.WriteString(")")
	}

	result.WriteString(";")

	return result.String()
}
