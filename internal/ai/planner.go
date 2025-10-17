package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ValidatePlan checks if a plan is valid against the schema
func (v *PlanValidator) ValidatePlan(plan *AIResult) error {
	// Basic validation
	if plan.Operation == "" {
		return fmt.Errorf("operation is required")
	}

	// Operation-specific validation
	switch strings.ToUpper(plan.Operation) {
	case "SELECT", "UPDATE", "DELETE":
		if plan.Table == "" {
			return fmt.Errorf("table is required for %s operation", plan.Operation)
		}
	case "INSERT":
		if plan.Table == "" {
			return fmt.Errorf("table is required for INSERT operation")
		}
		if len(plan.Values) == 0 {
			return fmt.Errorf("values are required for INSERT operation")
		}
	case "CREATE", "ALTER", "DROP":
		// DDL operations have different requirements
		if plan.Table == "" && plan.Keyspace == "" {
			return fmt.Errorf("table or keyspace required for DDL operation")
		}
	}

	// Check for dangerous operations
	if !plan.ReadOnly {
		switch strings.ToUpper(plan.Operation) {
		case "DROP", "TRUNCATE", "DELETE":
			if plan.Warning == "" {
				plan.Warning = "This is a destructive operation that will permanently delete data"
			}
		case "ALTER":
			if plan.Warning == "" {
				plan.Warning = "This operation will modify the schema"
			}
		}
	}

	return nil
}

// RenderCQL converts a validated plan to CQL
func RenderCQL(plan *AIResult) (string, error) {
	switch strings.ToUpper(plan.Operation) {
	case "SELECT":
		return renderSelect(plan)
	case "INSERT":
		return renderInsert(plan)
	case "UPDATE":
		return renderUpdate(plan)
	case "DELETE":
		return renderDelete(plan)
	case "CREATE":
		return renderCreate(plan)
	case "DROP":
		return renderDrop(plan)
	case "DESCRIBE":
		return renderDescribe(plan)
	default:
		return "", fmt.Errorf("unsupported operation: %s", plan.Operation)
	}
}

func renderSelect(plan *AIResult) (string, error) {
	// Validate table name - prevent wildcards
	if plan.Table == "" || plan.Table == "*" || strings.Contains(plan.Table, "*") {
		return "", fmt.Errorf("invalid table name: must specify an exact table name, not '%s'", plan.Table)
	}

	var sb strings.Builder

	// SELECT clause
	sb.WriteString("SELECT ")
	if len(plan.Columns) == 0 {
		sb.WriteString("*")
	} else {
		sb.WriteString(strings.Join(plan.Columns, ", "))
	}

	// FROM clause
	sb.WriteString(" FROM ")
	if plan.Keyspace != "" {
		sb.WriteString(fmt.Sprintf("%s.%s", plan.Keyspace, plan.Table))
	} else {
		sb.WriteString(plan.Table)
	}

	// WHERE clause
	if len(plan.Where) > 0 {
		sb.WriteString(" WHERE ")
		conditions := make([]string, 0, len(plan.Where))
		for _, w := range plan.Where {
			conditions = append(conditions, renderWhereClause(w))
		}
		sb.WriteString(strings.Join(conditions, " AND "))
	}

	// GROUP BY clause
	if len(plan.GroupBy) > 0 {
		sb.WriteString(" GROUP BY ")
		sb.WriteString(strings.Join(plan.GroupBy, ", "))
	}

	// ORDER BY clause
	if len(plan.OrderBy) > 0 {
		sb.WriteString(" ORDER BY ")
		orderClauses := make([]string, 0, len(plan.OrderBy))
		for _, o := range plan.OrderBy {
			orderClauses = append(orderClauses, fmt.Sprintf("%s %s", o.Column, o.Order))
		}
		sb.WriteString(strings.Join(orderClauses, ", "))
	}

	// LIMIT clause
	if plan.Limit > 0 {
		sb.WriteString(fmt.Sprintf(" LIMIT %d", plan.Limit))
	}

	// ALLOW FILTERING
	if plan.AllowFiltering {
		sb.WriteString(" ALLOW FILTERING")
	}

	sb.WriteString(";")
	return sb.String(), nil
}

func renderInsert(plan *AIResult) (string, error) {
	if len(plan.Values) == 0 {
		return "", fmt.Errorf("no values to insert")
	}

	var sb strings.Builder
	sb.WriteString("INSERT INTO ")

	if plan.Keyspace != "" {
		sb.WriteString(fmt.Sprintf("%s.%s", plan.Keyspace, plan.Table))
	} else {
		sb.WriteString(plan.Table)
	}

	// Column names
	columns := make([]string, 0, len(plan.Values))
	values := make([]string, 0, len(plan.Values))

	for col, val := range plan.Values {
		columns = append(columns, col)
		values = append(values, formatValue(val))
	}

	sb.WriteString(fmt.Sprintf(" (%s) VALUES (%s);",
		strings.Join(columns, ", "),
		strings.Join(values, ", ")))

	return sb.String(), nil
}

func renderUpdate(plan *AIResult) (string, error) {
	var sb strings.Builder
	sb.WriteString("UPDATE ")

	if plan.Keyspace != "" {
		sb.WriteString(fmt.Sprintf("%s.%s", plan.Keyspace, plan.Table))
	} else {
		sb.WriteString(plan.Table)
	}

	// SET clause
	if len(plan.Values) > 0 {
		sb.WriteString(" SET ")
		setClauses := make([]string, 0, len(plan.Values))
		for col, val := range plan.Values {
			setClauses = append(setClauses, fmt.Sprintf("%s = %s", col, formatValue(val)))
		}
		sb.WriteString(strings.Join(setClauses, ", "))
	}

	// WHERE clause (required for UPDATE)
	if len(plan.Where) == 0 {
		return "", fmt.Errorf("WHERE clause is required for UPDATE")
	}

	sb.WriteString(" WHERE ")
	conditions := make([]string, 0, len(plan.Where))
	for _, w := range plan.Where {
		conditions = append(conditions, renderWhereClause(w))
	}
	sb.WriteString(strings.Join(conditions, " AND "))

	sb.WriteString(";")
	return sb.String(), nil
}

func renderDelete(plan *AIResult) (string, error) {
	var sb strings.Builder
	sb.WriteString("DELETE ")

	// Optional column specification
	if len(plan.Columns) > 0 {
		sb.WriteString(strings.Join(plan.Columns, ", "))
		sb.WriteString(" ")
	}

	sb.WriteString("FROM ")
	if plan.Keyspace != "" {
		sb.WriteString(fmt.Sprintf("%s.%s", plan.Keyspace, plan.Table))
	} else {
		sb.WriteString(plan.Table)
	}

	// WHERE clause (required for DELETE)
	if len(plan.Where) == 0 {
		return "", fmt.Errorf("WHERE clause is required for DELETE")
	}

	sb.WriteString(" WHERE ")
	conditions := make([]string, 0, len(plan.Where))
	for _, w := range plan.Where {
		conditions = append(conditions, renderWhereClause(w))
	}
	sb.WriteString(strings.Join(conditions, " AND "))

	sb.WriteString(";")
	return sb.String(), nil
}

func renderCreate(plan *AIResult) (string, error) {
	var sb strings.Builder

	// Handle CREATE KEYSPACE
	if plan.Table == "" && plan.Keyspace != "" {
		sb.WriteString(fmt.Sprintf("CREATE KEYSPACE %s", plan.Keyspace))

		// Add WITH REPLICATION clause if options are provided
		if plan.Options != nil {
			if replication, ok := plan.Options["replication"]; ok {
				sb.WriteString(" WITH REPLICATION = ")
				// Convert replication object to map syntax
				if replMap, ok := replication.(map[string]any); ok {
					sb.WriteString(formatMapValue(replMap))
				} else {
					sb.WriteString(fmt.Sprintf("%v", replication))
				}
			}
		}

		sb.WriteString(";")
		return sb.String(), nil
	}

	// Handle CREATE TABLE
	if plan.Table == "" {
		return "", fmt.Errorf("table or keyspace name required for CREATE")
	}

	sb.WriteString("CREATE TABLE ")

	if plan.Keyspace != "" {
		sb.WriteString(fmt.Sprintf("%s.%s", plan.Keyspace, plan.Table))
	} else {
		sb.WriteString(plan.Table)
	}

	if len(plan.Schema) > 0 {
		sb.WriteString(" (")
		cols := make([]string, 0, len(plan.Schema))
		for name, typ := range plan.Schema {
			cols = append(cols, fmt.Sprintf("%s %s", name, typ))
		}
		sb.WriteString(strings.Join(cols, ", "))
		sb.WriteString(")")
	}

	sb.WriteString(";")
	return sb.String(), nil
}

func renderDrop(plan *AIResult) (string, error) {
	var sb strings.Builder
	sb.WriteString("DROP ")

	if plan.Table != "" {
		sb.WriteString("TABLE ")
		if plan.Keyspace != "" {
			sb.WriteString(fmt.Sprintf("%s.%s", plan.Keyspace, plan.Table))
		} else {
			sb.WriteString(plan.Table)
		}
	} else if plan.Keyspace != "" {
		sb.WriteString(fmt.Sprintf("KEYSPACE %s", plan.Keyspace))
	}

	sb.WriteString(";")
	return sb.String(), nil
}

func renderDescribe(plan *AIResult) (string, error) {
	var sb strings.Builder
	sb.WriteString("DESCRIBE ")

	// Handle different DESCRIBE targets
	tableUpper := strings.ToUpper(plan.Table)
	switch tableUpper {
	case "KEYSPACES":
		sb.WriteString("KEYSPACES")
	case "TABLES":
		sb.WriteString("TABLES")
		// If keyspace is specified, it will be handled by the outer logic
	case "CLUSTER":
		sb.WriteString("CLUSTER")
	case "SCHEMA":
		sb.WriteString("SCHEMA")
	default:
		// Describing a specific table or keyspace
		if plan.Table != "" {
			if plan.Keyspace != "" {
				sb.WriteString(fmt.Sprintf("TABLE %s.%s", plan.Keyspace, plan.Table))
			} else {
				sb.WriteString(fmt.Sprintf("TABLE %s", plan.Table))
			}
		} else if plan.Keyspace != "" {
			sb.WriteString(fmt.Sprintf("KEYSPACE %s", plan.Keyspace))
		}
	}

	sb.WriteString(";")
	return sb.String(), nil
}

func renderWhereClause(w WhereClause) string {
	return fmt.Sprintf("%s %s %s", w.Column, w.Operator, formatValue(w.Value))
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		// Escape single quotes in strings
		escaped := strings.ReplaceAll(val, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case []any:
		// Handle IN clause values
		values := make([]string, len(val))
		for i, item := range val {
			values[i] = formatValue(item)
		}
		return fmt.Sprintf("(%s)", strings.Join(values, ", "))
	default:
		return fmt.Sprintf("%v", val)
	}
}

// formatMapValue formats a map as CQL map literal: {'key': 'value', 'key2': value2}
func formatMapValue(m map[string]any) string {
	if len(m) == 0 {
		return "{}"
	}

	pairs := make([]string, 0, len(m))
	for k, v := range m {
		var valueStr string
		switch val := v.(type) {
		case string:
			valueStr = fmt.Sprintf("'%s'", val)
		case map[string]any:
			// Nested map
			valueStr = formatMapValue(val)
		default:
			valueStr = fmt.Sprintf("%v", val)
		}
		pairs = append(pairs, fmt.Sprintf("'%s': %s", k, valueStr))
	}

	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

// ParsePlanFromJSON parses a JSON plan from LLM response
func ParsePlanFromJSON(jsonStr string) (*AIResult, error) {
	var plan AIResult
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan: %v", err)
	}
	return &plan, nil
}
