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
	case "GRANT":
		return renderGrant(plan)
	case "REVOKE":
		return renderRevoke(plan)
	case "TRUNCATE":
		return renderTruncate(plan)
	case "ALTER":
		return renderAlter(plan)
	case "LIST":
		return renderList(plan)
	case "SHOW":
		return renderShow(plan)
	case "USE":
		return renderUse(plan)
	case "BATCH":
		return renderBatch(plan)
	case "DESC":
		// DESC is an alias for DESCRIBE
		return renderDescribe(plan)
	case "CONSISTENCY", "PAGING", "TRACING", "COPY", "SOURCE":
		// SESSION and FILE operations - handled by buildRawCommand at MCP layer
		// If we get here, it means they weren't intercepted at MCP layer
		// Build the command string directly
		return buildRawCommandForPlanner(plan)
	case "EXPAND", "OUTPUT", "CAPTURE", "SAVE", "AUTOFETCH":
		// Display-only commands - return the command string, will be handled by handleShellCommand
		return buildRawCommandForPlanner(plan)
	case "ADD":
		// Handle ADD IDENTITY
		return renderAddIdentity(plan)
	case "BEGIN", "APPLY":
		// BEGIN BATCH and APPLY BATCH - not supported yet
		return "", fmt.Errorf("BATCH operations not yet supported in query builder - use raw CQL")
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

	// Phase 1: SELECT modifiers (DISTINCT, JSON)
	if plan.SelectJSON {
		sb.WriteString("JSON ")
	}
	if plan.Distinct {
		sb.WriteString("DISTINCT ")
	}

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

	// Phase 1: PER PARTITION LIMIT clause (before regular LIMIT)
	if plan.PerPartitionLimit > 0 {
		sb.WriteString(fmt.Sprintf(" PER PARTITION LIMIT %d", plan.PerPartitionLimit))
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
	// Phase 1: INSERT JSON uses different syntax
	if plan.InsertJSON {
		if plan.JSONValue == "" {
			return "", fmt.Errorf("json_value required for INSERT JSON")
		}

		var sb strings.Builder
		sb.WriteString("INSERT INTO ")

		if plan.Keyspace != "" {
			sb.WriteString(fmt.Sprintf("%s.%s", plan.Keyspace, plan.Table))
		} else {
			sb.WriteString(plan.Table)
		}

		// Escape single quotes in JSON string
		escapedJSON := strings.ReplaceAll(plan.JSONValue, "'", "''")
		sb.WriteString(fmt.Sprintf(" JSON '%s'", escapedJSON))

		// USING clause works with INSERT JSON
		usingClauses := []string{}
		if plan.UsingTTL > 0 {
			usingClauses = append(usingClauses, fmt.Sprintf("TTL %d", plan.UsingTTL))
		}
		if plan.UsingTimestamp > 0 {
			usingClauses = append(usingClauses, fmt.Sprintf("TIMESTAMP %d", plan.UsingTimestamp))
		}
		if len(usingClauses) > 0 {
			sb.WriteString(" USING ")
			sb.WriteString(strings.Join(usingClauses, " AND "))
		}

		sb.WriteString(";")
		return sb.String(), nil
	}

	// Regular INSERT with column/value pairs
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
		// TODO: Get type hint from plan.ValueTypes or schema
		typeHint := ""
		if plan.ValueTypes != nil {
			typeHint = plan.ValueTypes[col]
		}
		values = append(values, formatValue(val, typeHint))
	}

	sb.WriteString(fmt.Sprintf(" (%s) VALUES (%s)",
		strings.Join(columns, ", "),
		strings.Join(values, ", ")))

	// Phase 1: USING clause (TTL and/or TIMESTAMP)
	usingClauses := []string{}
	if plan.UsingTTL > 0 {
		usingClauses = append(usingClauses, fmt.Sprintf("TTL %d", plan.UsingTTL))
	}
	if plan.UsingTimestamp > 0 {
		usingClauses = append(usingClauses, fmt.Sprintf("TIMESTAMP %d", plan.UsingTimestamp))
	}
	if len(usingClauses) > 0 {
		sb.WriteString(" USING ")
		sb.WriteString(strings.Join(usingClauses, " AND "))
	}

	sb.WriteString(";")
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

	// Phase 1: USING clause (appears after table name, before SET in UPDATE)
	usingClauses := []string{}
	if plan.UsingTTL > 0 {
		usingClauses = append(usingClauses, fmt.Sprintf("TTL %d", plan.UsingTTL))
	}
	if plan.UsingTimestamp > 0 {
		usingClauses = append(usingClauses, fmt.Sprintf("TIMESTAMP %d", plan.UsingTimestamp))
	}
	if len(usingClauses) > 0 {
		sb.WriteString(" USING ")
		sb.WriteString(strings.Join(usingClauses, " AND "))
	}

	// SET clause
	sb.WriteString(" SET ")
	setClauses := make([]string, 0)

	// Phase 2: Counter operations (counter = counter + N)
	if len(plan.CounterOps) > 0 {
		for col, delta := range plan.CounterOps {
			// Validate delta format (+N or -N)
			if !strings.HasPrefix(delta, "+") && !strings.HasPrefix(delta, "-") {
				return "", fmt.Errorf("counter operation must be increment (+N) or decrement (-N), got: %s", delta)
			}
			// Parse operator and number for proper spacing
			operator := delta[0:1]     // "+" or "-"
			number := delta[1:]        // "5", "1", etc.
			// Render as: counter_col = counter_col + N (with spaces)
			setClauses = append(setClauses, fmt.Sprintf("%s = %s %s %s", col, col, operator, number))
		}
	}

	// Regular value updates (non-counters)
	if len(plan.Values) > 0 {
		for col, val := range plan.Values {
			// TODO: Get type hint from plan.ValueTypes or schema
			typeHint := ""
			if plan.ValueTypes != nil {
				typeHint = plan.ValueTypes[col]
			}
			setClauses = append(setClauses, fmt.Sprintf("%s = %s", col, formatValue(val, typeHint)))
		}
	}

	if len(setClauses) == 0 {
		return "", fmt.Errorf("UPDATE requires SET clause (values or counter operations)")
	}

	sb.WriteString(strings.Join(setClauses, ", "))

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

	// Phase 1: USING TIMESTAMP clause (DELETE only supports TIMESTAMP, not TTL)
	if plan.UsingTimestamp > 0 {
		sb.WriteString(fmt.Sprintf(" USING TIMESTAMP %d", plan.UsingTimestamp))
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
	// Check if this is a specialized CREATE (INDEX, TYPE, FUNCTION, etc.)
	if plan.Options != nil {
		if objectType, ok := plan.Options["object_type"].(string); ok && objectType != "" {
			switch strings.ToUpper(objectType) {
			case "INDEX":
				return renderCreateIndex(plan)
			case "TYPE":
				return renderCreateType(plan)
			case "FUNCTION":
				return renderCreateFunction(plan)
			case "AGGREGATE":
				return renderCreateAggregate(plan)
			case "TRIGGER":
				return renderCreateTrigger(plan)
			case "MATERIALIZED VIEW", "MATERIALIZED_VIEW", "MV":
				return renderCreateMaterializedView(plan)
			case "ROLE":
				return renderCreateRole(plan)
			case "USER":
				return renderCreateUser(plan)
			}
		}
	}

	var sb strings.Builder

	// Check for IF NOT EXISTS option
	ifNotExists := false
	if plan.Options != nil {
		if ine, ok := plan.Options["if_not_exists"].(bool); ok {
			ifNotExists = ine
		}
	}

	// Handle CREATE KEYSPACE
	if plan.Table == "" && plan.Keyspace != "" {
		if ifNotExists {
			sb.WriteString(fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s", plan.Keyspace))
		} else {
			sb.WriteString(fmt.Sprintf("CREATE KEYSPACE %s", plan.Keyspace))
		}

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

	if ifNotExists {
		sb.WriteString("CREATE TABLE IF NOT EXISTS ")
	} else {
		sb.WriteString("CREATE TABLE ")
	}

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
	// Check if this is a specialized DROP (INDEX, TYPE, FUNCTION, etc.)
	if plan.Options != nil {
		if objectType, ok := plan.Options["object_type"].(string); ok && objectType != "" {
			switch strings.ToUpper(objectType) {
			case "INDEX":
				return renderDropIndex(plan)
			case "TYPE":
				return renderDropType(plan)
			case "FUNCTION":
				return renderDropFunction(plan)
			case "AGGREGATE":
				return renderDropAggregate(plan)
			case "TRIGGER":
				return renderDropTrigger(plan)
			case "MATERIALIZED VIEW", "MATERIALIZED_VIEW", "MV":
				return renderDropMaterializedView(plan)
			case "ROLE":
				return renderDropRole(plan)
			case "USER":
				return renderDropUser(plan)
			case "IDENTITY":
				return renderAddIdentity(plan) // Reuse renderAddIdentity for DROP IDENTITY
			}
		}
	}

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
	// Handle IS NULL and IS NOT NULL (no value needed)
	opUpper := strings.ToUpper(w.Operator)
	if opUpper == "IS NULL" || opUpper == "IS NOT NULL" {
		return fmt.Sprintf("%s %s", w.Column, w.Operator)
	}
	// TODO: Add ValueType field to WhereClause for type hints
	return fmt.Sprintf("%s %s %s", w.Column, w.Operator, formatValue(w.Value, ""))
}

// formatValue is the main entry point for formatting CQL values with optional type hints
// typeHint examples: "text", "int", "list<text>", "set<int>", "map<text,int>", "tuple<text,int>", "blob"
// typeHint can be "" to infer type from value (backward compatible)
func formatValue(v any, typeHint string) string {
	// Handle nil early
	if v == nil {
		return "null"
	}

	// Parse type hint to determine routing
	baseType, elementType := parseTypeHint(typeHint)

	// Route to appropriate formatter based on type hint
	switch baseType {
	case "list":
		return formatList(v, elementType)
	case "set":
		return formatSet(v, elementType)
	case "map":
		keyType, valueType := parseMapTypes(typeHint)
		return formatMap(v, keyType, valueType)
	case "tuple":
		return formatTuple(v)
	case "blob":
		return formatBlob(v)
	case "udt", "frozen": // UDT or frozen UDT
		return formatUDT(v, typeHint)
	default:
		// No type hint or primitive type - infer from value
		return formatPrimitive(v)
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

// renderGrant generates GRANT statements (permission or role)
func renderGrant(plan *AIResult) (string, error) {
	if plan.Options == nil {
		return "", fmt.Errorf("options required for GRANT")
	}

	// Check if this is GRANT ROLE vs GRANT permission
	grantType, _ := plan.Options["grant_type"].(string)

	if strings.ToUpper(grantType) == "ROLE" {
		// GRANT role TO another_role
		role, ok := plan.Options["role"].(string)
		if !ok || role == "" {
			return "", fmt.Errorf("'role' required in options for GRANT ROLE")
		}
		toRole, ok := plan.Options["to_role"].(string)
		if !ok || toRole == "" {
			return "", fmt.Errorf("'to_role' required in options for GRANT ROLE")
		}
		return fmt.Sprintf("GRANT %s TO %s;", role, toRole), nil
	}

	// GRANT permission ON resource TO role
	permission, ok := plan.Options["permission"].(string)
	if !ok || permission == "" {
		return "", fmt.Errorf("'permission' required in options for GRANT (e.g., SELECT, MODIFY, ALL)")
	}

	// Validate permission is one of the Cassandra-supported types
	validPermissions := []string{"CREATE", "ALTER", "DROP", "SELECT", "MODIFY", "AUTHORIZE", "DESCRIBE", "EXECUTE", "UNMASK", "SELECT_MASKED", "ALL", "ALL PERMISSIONS"}
	permUpper := strings.ToUpper(permission)
	isValid := false
	for _, vp := range validPermissions {
		if permUpper == vp {
			isValid = true
			break
		}
	}
	if !isValid {
		return "", fmt.Errorf("invalid permission '%s' - must be one of: CREATE, ALTER, DROP, SELECT, MODIFY, AUTHORIZE, DESCRIBE, EXECUTE, UNMASK, SELECT_MASKED, ALL", permission)
	}

	role, ok := plan.Options["role"].(string)
	if !ok || role == "" {
		return "", fmt.Errorf("'role' required in options for GRANT")
	}

	// Determine resource scope - supports all Cassandra resource types
	resourceScope, ok := plan.Options["resource_scope"].(string)
	if !ok || resourceScope == "" {
		resourceScope = "KEYSPACE" // Default to keyspace level
	}

	var resource string
	switch strings.ToUpper(resourceScope) {
	case "ALL KEYSPACES", "ALL_KEYSPACES":
		resource = "ALL KEYSPACES"

	case "KEYSPACE":
		if plan.Keyspace == "" {
			return "", fmt.Errorf("keyspace required for GRANT ... ON KEYSPACE")
		}
		resource = fmt.Sprintf("KEYSPACE %s", plan.Keyspace)

	case "TABLE":
		if plan.Keyspace == "" || plan.Table == "" {
			return "", fmt.Errorf("keyspace and table required for GRANT ... ON TABLE")
		}
		resource = fmt.Sprintf("TABLE %s.%s", plan.Keyspace, plan.Table)

	case "ALL ROLES", "ALL_ROLES":
		resource = "ALL ROLES"

	case "ROLE":
		targetRole, ok := plan.Options["target_role"].(string)
		if !ok || targetRole == "" {
			return "", fmt.Errorf("'target_role' required for GRANT ... ON ROLE")
		}
		resource = fmt.Sprintf("ROLE %s", targetRole)

	case "ALL FUNCTIONS", "ALL_FUNCTIONS":
		// ALL FUNCTIONS [IN KEYSPACE ks]
		if plan.Keyspace != "" {
			resource = fmt.Sprintf("ALL FUNCTIONS IN KEYSPACE %s", plan.Keyspace)
		} else {
			resource = "ALL FUNCTIONS"
		}

	case "FUNCTION":
		if plan.Keyspace == "" {
			return "", fmt.Errorf("keyspace required for GRANT ... ON FUNCTION")
		}
		functionName, ok := plan.Options["function_name"].(string)
		if !ok || functionName == "" {
			return "", fmt.Errorf("'function_name' required for GRANT ... ON FUNCTION")
		}
		resource = fmt.Sprintf("FUNCTION %s.%s", plan.Keyspace, functionName)

	case "ALL MBEANS", "ALL_MBEANS":
		resource = "ALL MBEANS"

	case "MBEAN", "MBEANS":
		mbeanName, ok := plan.Options["mbean_name"].(string)
		if !ok || mbeanName == "" {
			return "", fmt.Errorf("'mbean_name' required for GRANT ... ON MBEAN")
		}
		resource = fmt.Sprintf("MBEAN %s", mbeanName)

	default:
		return "", fmt.Errorf("unsupported resource_scope: %s (see Cassandra docs for valid resource types)", resourceScope)
	}

	return fmt.Sprintf("GRANT %s ON %s TO %s;", permission, resource, role), nil
}

// renderRevoke generates REVOKE statements (permission or role)
func renderRevoke(plan *AIResult) (string, error) {
	if plan.Options == nil {
		return "", fmt.Errorf("options required for REVOKE")
	}

	// Check if this is REVOKE ROLE vs REVOKE permission
	grantType, _ := plan.Options["grant_type"].(string)

	if strings.ToUpper(grantType) == "ROLE" {
		// REVOKE role FROM another_role
		role, ok := plan.Options["role"].(string)
		if !ok || role == "" {
			return "", fmt.Errorf("'role' required in options for REVOKE ROLE")
		}
		fromRole, ok := plan.Options["from_role"].(string)
		if !ok || fromRole == "" {
			return "", fmt.Errorf("'from_role' required in options for REVOKE ROLE")
		}
		return fmt.Sprintf("REVOKE %s FROM %s;", role, fromRole), nil
	}

	// REVOKE permission ON resource FROM role
	permission, ok := plan.Options["permission"].(string)
	if !ok || permission == "" {
		return "", fmt.Errorf("'permission' required in options for REVOKE")
	}

	// Validate permission is one of the Cassandra-supported types
	validPermissions := []string{"CREATE", "ALTER", "DROP", "SELECT", "MODIFY", "AUTHORIZE", "DESCRIBE", "EXECUTE", "UNMASK", "SELECT_MASKED", "ALL", "ALL PERMISSIONS"}
	permUpper := strings.ToUpper(permission)
	isValid := false
	for _, vp := range validPermissions {
		if permUpper == vp {
			isValid = true
			break
		}
	}
	if !isValid {
		return "", fmt.Errorf("invalid permission '%s' - must be one of: CREATE, ALTER, DROP, SELECT, MODIFY, AUTHORIZE, DESCRIBE, EXECUTE, UNMASK, SELECT_MASKED, ALL", permission)
	}

	role, ok := plan.Options["role"].(string)
	if !ok || role == "" {
		return "", fmt.Errorf("'role' required in options for REVOKE")
	}

	// Determine resource scope - supports all Cassandra resource types
	resourceScope, ok := plan.Options["resource_scope"].(string)
	if !ok || resourceScope == "" {
		resourceScope = "KEYSPACE" // Default to keyspace level for backwards compatibility
	}

	var resource string
	switch strings.ToUpper(resourceScope) {
	case "ALL KEYSPACES", "ALL_KEYSPACES":
		resource = "ALL KEYSPACES"

	case "KEYSPACE":
		if plan.Keyspace == "" {
			return "", fmt.Errorf("keyspace required for REVOKE ... ON KEYSPACE")
		}
		resource = fmt.Sprintf("KEYSPACE %s", plan.Keyspace)

	case "TABLE":
		if plan.Keyspace == "" || plan.Table == "" {
			return "", fmt.Errorf("keyspace and table required for REVOKE ... ON TABLE")
		}
		resource = fmt.Sprintf("TABLE %s.%s", plan.Keyspace, plan.Table)

	case "ALL ROLES", "ALL_ROLES":
		resource = "ALL ROLES"

	case "ROLE":
		targetRole, ok := plan.Options["target_role"].(string)
		if !ok || targetRole == "" {
			return "", fmt.Errorf("'target_role' required for REVOKE ... ON ROLE")
		}
		resource = fmt.Sprintf("ROLE %s", targetRole)

	case "ALL FUNCTIONS", "ALL_FUNCTIONS":
		// ALL FUNCTIONS [IN KEYSPACE ks]
		if plan.Keyspace != "" {
			resource = fmt.Sprintf("ALL FUNCTIONS IN KEYSPACE %s", plan.Keyspace)
		} else {
			resource = "ALL FUNCTIONS"
		}

	case "FUNCTION":
		if plan.Keyspace == "" {
			return "", fmt.Errorf("keyspace required for REVOKE ... ON FUNCTION")
		}
		functionName, ok := plan.Options["function_name"].(string)
		if !ok || functionName == "" {
			return "", fmt.Errorf("'function_name' required for REVOKE ... ON FUNCTION")
		}
		resource = fmt.Sprintf("FUNCTION %s.%s", plan.Keyspace, functionName)

	case "ALL MBEANS", "ALL_MBEANS":
		resource = "ALL MBEANS"

	case "MBEAN", "MBEANS":
		mbeanName, ok := plan.Options["mbean_name"].(string)
		if !ok || mbeanName == "" {
			return "", fmt.Errorf("'mbean_name' required for REVOKE ... ON MBEAN")
		}
		resource = fmt.Sprintf("MBEAN %s", mbeanName)

	default:
		return "", fmt.Errorf("unsupported resource_scope: %s (see Cassandra docs for valid resource types)", resourceScope)
	}

	return fmt.Sprintf("REVOKE %s ON %s FROM %s;", permission, resource, role), nil
}

// renderAddIdentity generates ADD IDENTITY and DROP IDENTITY statements
func renderAddIdentity(plan *AIResult) (string, error) {
	if plan.Options == nil {
		return "", fmt.Errorf("options required for ADD/DROP IDENTITY")
	}

	objectType, _ := plan.Options["object_type"].(string)
	if strings.ToUpper(objectType) != "IDENTITY" {
		return "", fmt.Errorf("object_type must be IDENTITY for ADD/DROP IDENTITY operations")
	}

	identity, ok := plan.Options["identity"].(string)
	if !ok || identity == "" {
		return "", fmt.Errorf("'identity' required in options for ADD/DROP IDENTITY")
	}

	role, ok := plan.Options["role"].(string)
	if !ok || role == "" {
		return "", fmt.Errorf("'role' required in options for ADD/DROP IDENTITY")
	}

	// Determine if this is ADD or DROP
	op := strings.ToUpper(plan.Operation)
	if op == "ADD" {
		// ADD IDENTITY syntax: ADD IDENTITY 'identity' TO role_name
		return fmt.Sprintf("ADD IDENTITY '%s' TO %s;", identity, role), nil
	} else {
		// DROP IDENTITY syntax: DROP IDENTITY 'identity' (role is implicit from context)
		// Cassandra may not support this syntax directly, but generate it for testing
		return fmt.Sprintf("DROP IDENTITY '%s';", identity), nil
	}
}

// renderTruncate generates a TRUNCATE statement
func renderTruncate(plan *AIResult) (string, error) {
	if plan.Keyspace == "" || plan.Table == "" {
		return "", fmt.Errorf("keyspace and table required for TRUNCATE")
	}

	return fmt.Sprintf("TRUNCATE %s.%s", plan.Keyspace, plan.Table), nil
}

// renderAlter generates ALTER statements for various object types
func renderAlter(plan *AIResult) (string, error) {
	if plan.Options == nil {
		return "", fmt.Errorf("options required for ALTER - must specify object_type")
	}

	objectType, ok := plan.Options["object_type"].(string)
	if !ok || objectType == "" {
		return "", fmt.Errorf("'object_type' required in options for ALTER (TABLE, KEYSPACE, TYPE, ROLE, USER)")
	}

	switch strings.ToUpper(objectType) {
	case "TABLE":
		return renderAlterTable(plan)
	case "KEYSPACE":
		return renderAlterKeyspace(plan)
	case "TYPE":
		return renderAlterType(plan)
	case "ROLE":
		return renderAlterRole(plan)
	case "USER":
		return renderAlterUser(plan)
	default:
		return "", fmt.Errorf("unsupported ALTER object_type: %s", objectType)
	}
}

// renderAlterTable generates ALTER TABLE statements
func renderAlterTable(plan *AIResult) (string, error) {
	if plan.Keyspace == "" || plan.Table == "" {
		return "", fmt.Errorf("keyspace and table required for ALTER TABLE")
	}

	action, ok := plan.Options["action"].(string)
	if !ok || action == "" {
		return "", fmt.Errorf("'action' required in options for ALTER TABLE (ADD, DROP, RENAME, WITH)")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ALTER TABLE %s.%s ", plan.Keyspace, plan.Table))

	switch strings.ToUpper(action) {
	case "ADD":
		columnName, ok := plan.Options["column_name"].(string)
		if !ok || columnName == "" {
			return "", fmt.Errorf("'column_name' required for ALTER TABLE ADD")
		}
		columnType, ok := plan.Options["column_type"].(string)
		if !ok || columnType == "" {
			return "", fmt.Errorf("'column_type' required for ALTER TABLE ADD")
		}
		sb.WriteString(fmt.Sprintf("ADD %s %s", columnName, columnType))

	case "DROP":
		columnName, ok := plan.Options["column_name"].(string)
		if !ok || columnName == "" {
			return "", fmt.Errorf("'column_name' required for ALTER TABLE DROP")
		}
		sb.WriteString(fmt.Sprintf("DROP %s", columnName))

	case "RENAME":
		oldName, ok1 := plan.Options["old_column_name"].(string)
		newName, ok2 := plan.Options["new_column_name"].(string)
		if !ok1 || !ok2 || oldName == "" || newName == "" {
			return "", fmt.Errorf("'old_column_name' and 'new_column_name' required for ALTER TABLE RENAME")
		}
		sb.WriteString(fmt.Sprintf("RENAME %s TO %s", oldName, newName))

	case "WITH":
		properties, ok := plan.Options["properties"].(map[string]interface{})
		if !ok || len(properties) == 0 {
			return "", fmt.Errorf("'properties' map required for ALTER TABLE WITH")
		}
		sb.WriteString("WITH ")
		first := true
		for k, v := range properties {
			if !first {
				sb.WriteString(" AND ")
			}
			sb.WriteString(fmt.Sprintf("%s = %v", k, v))
			first = false
		}

	default:
		return "", fmt.Errorf("unsupported ALTER TABLE action: %s (must be ADD, DROP, RENAME, or WITH)", action)
	}

	return sb.String(), nil
}

// renderAlterKeyspace generates ALTER KEYSPACE statements
func renderAlterKeyspace(plan *AIResult) (string, error) {
	if plan.Keyspace == "" {
		return "", fmt.Errorf("keyspace required for ALTER KEYSPACE")
	}

	properties, ok := plan.Options["properties"].(map[string]interface{})
	if !ok || len(properties) == 0 {
		return "", fmt.Errorf("'properties' map required for ALTER KEYSPACE WITH")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ALTER KEYSPACE %s WITH ", plan.Keyspace))

	first := true
	for k, v := range properties {
		if !first {
			sb.WriteString(" AND ")
		}
		// Handle replication specially (it's a map)
		if k == "replication" {
			if repMap, ok := v.(map[string]interface{}); ok {
				sb.WriteString("replication = {")
				firstRep := true
				for rk, rv := range repMap {
					if !firstRep {
						sb.WriteString(", ")
					}
					sb.WriteString(fmt.Sprintf("'%s': '%v'", rk, rv))
					firstRep = false
				}
				sb.WriteString("}")
			}
		} else {
			sb.WriteString(fmt.Sprintf("%s = %v", k, v))
		}
		first = false
	}

	return sb.String(), nil
}

// renderAlterType generates ALTER TYPE statements for UDTs
func renderAlterType(plan *AIResult) (string, error) {
	if plan.Keyspace == "" {
		return "", fmt.Errorf("keyspace required for ALTER TYPE")
	}

	typeName, ok := plan.Options["type_name"].(string)
	if !ok || typeName == "" {
		return "", fmt.Errorf("'type_name' required in options for ALTER TYPE")
	}

	action, ok := plan.Options["action"].(string)
	if !ok || action == "" {
		return "", fmt.Errorf("'action' required for ALTER TYPE (ADD, RENAME)")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ALTER TYPE %s.%s ", plan.Keyspace, typeName))

	switch strings.ToUpper(action) {
	case "ADD":
		fieldName, ok := plan.Options["field_name"].(string)
		if !ok || fieldName == "" {
			return "", fmt.Errorf("'field_name' required for ALTER TYPE ADD")
		}
		fieldType, ok := plan.Options["field_type"].(string)
		if !ok || fieldType == "" {
			return "", fmt.Errorf("'field_type' required for ALTER TYPE ADD")
		}
		sb.WriteString(fmt.Sprintf("ADD %s %s", fieldName, fieldType))

	case "RENAME":
		oldName, ok1 := plan.Options["old_field_name"].(string)
		newName, ok2 := plan.Options["new_field_name"].(string)
		if !ok1 || !ok2 || oldName == "" || newName == "" {
			return "", fmt.Errorf("'old_field_name' and 'new_field_name' required for ALTER TYPE RENAME")
		}
		sb.WriteString(fmt.Sprintf("RENAME %s TO %s", oldName, newName))

	default:
		return "", fmt.Errorf("unsupported ALTER TYPE action: %s (must be ADD or RENAME)", action)
	}

	return sb.String(), nil
}

// renderAlterRole generates ALTER ROLE statements
func renderAlterRole(plan *AIResult) (string, error) {
	roleName, ok := plan.Options["role_name"].(string)
	if !ok || roleName == "" {
		return "", fmt.Errorf("'role_name' required in options for ALTER ROLE")
	}

	// Check if this is ADD_IDENTITY or DROP_IDENTITY
	if action, ok := plan.Options["action"].(string); ok {
		actionUpper := strings.ToUpper(action)
		if actionUpper == "ADD_IDENTITY" || actionUpper == "DROP_IDENTITY" {
			identity, ok := plan.Options["identity"].(string)
			if !ok || identity == "" {
				return "", fmt.Errorf("'identity' required in options for ADD/DROP IDENTITY")
			}

			if actionUpper == "ADD_IDENTITY" {
				return fmt.Sprintf("ALTER ROLE %s ADD IDENTITY '%s';", roleName, identity), nil
			} else {
				return fmt.Sprintf("ALTER ROLE %s DROP IDENTITY '%s';", roleName, identity), nil
			}
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ALTER ROLE %s WITH ", roleName))

	// Build WITH clause from options
	withClauses := []string{}

	if password, ok := plan.Options["password"].(string); ok && password != "" {
		withClauses = append(withClauses, fmt.Sprintf("PASSWORD = '%s'", password))
	}
	if login, ok := plan.Options["login"].(bool); ok {
		withClauses = append(withClauses, fmt.Sprintf("LOGIN = %t", login))
	}
	if superuser, ok := plan.Options["superuser"].(bool); ok {
		withClauses = append(withClauses, fmt.Sprintf("SUPERUSER = %t", superuser))
	}

	if len(withClauses) == 0 {
		return "", fmt.Errorf("at least one property required for ALTER ROLE (password, login, or superuser)")
	}

	sb.WriteString(strings.Join(withClauses, " AND "))
	return sb.String(), nil
}

// renderAlterUser generates ALTER USER statements (deprecated but still supported)
func renderAlterUser(plan *AIResult) (string, error) {
	userName, ok := plan.Options["user_name"].(string)
	if !ok || userName == "" {
		return "", fmt.Errorf("'user_name' required in options for ALTER USER")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ALTER USER %s WITH ", userName))

	// Build WITH clause
	withClauses := []string{}

	if password, ok := plan.Options["password"].(string); ok && password != "" {
		withClauses = append(withClauses, fmt.Sprintf("PASSWORD '%s'", password))
	}
	if superuser, ok := plan.Options["superuser"].(bool); ok {
		withClauses = append(withClauses, fmt.Sprintf("SUPERUSER %t", superuser))
	}

	if len(withClauses) == 0 {
		return "", fmt.Errorf("at least one property required for ALTER USER (password or superuser)")
	}

	sb.WriteString(strings.Join(withClauses, " AND "))
	return sb.String(), nil
}

// buildRawCommandForPlanner converts AIResult to raw shell command for SESSION/FILE operations
func buildRawCommandForPlanner(plan *AIResult) (string, error) {
	opUpper := strings.ToUpper(plan.Operation)

	switch opUpper {
	case "CONSISTENCY":
		if plan.Options != nil {
			if level, ok := plan.Options["level"].(string); ok {
				return fmt.Sprintf("CONSISTENCY %s", strings.ToUpper(level)), nil
			}
		}
		return "CONSISTENCY", nil

	case "PAGING":
		if plan.Options != nil {
			if state, ok := plan.Options["state"].(string); ok {
				return fmt.Sprintf("PAGING %s", state), nil
			}
		}
		return "PAGING", nil

	case "TRACING":
		if plan.Options != nil {
			if state, ok := plan.Options["state"].(string); ok {
				return fmt.Sprintf("TRACING %s", strings.ToUpper(state)), nil
			}
		}
		return "TRACING", nil

	case "COPY":
		direction := "TO"
		if plan.Options != nil {
			if dir, ok := plan.Options["direction"].(string); ok {
				direction = strings.ToUpper(dir)
			}
		}

		filePath := "/tmp/export.csv"
		if plan.Options != nil {
			if fp, ok := plan.Options["file_path"].(string); ok {
				filePath = fp
			}
		}

		tableName := plan.Table
		if plan.Keyspace != "" {
			tableName = fmt.Sprintf("%s.%s", plan.Keyspace, plan.Table)
		}

		return fmt.Sprintf("COPY %s %s '%s'", tableName, direction, filePath), nil

	case "SOURCE":
		filePath := ""
		if plan.Options != nil {
			if fp, ok := plan.Options["file_path"].(string); ok {
				filePath = fp
			}
		}
		return fmt.Sprintf("SOURCE '%s'", filePath), nil

	default:
		return plan.Operation, nil
	}
}

// renderList generates LIST statements (ROLES, USERS, PERMISSIONS)
func renderList(plan *AIResult) (string, error) {
	if plan.Options == nil {
		return "", fmt.Errorf("options required for LIST - must specify object_type")
	}

	objectType, ok := plan.Options["object_type"].(string)
	if !ok || objectType == "" {
		return "", fmt.Errorf("'object_type' required in options for LIST (ROLES, USERS, PERMISSIONS)")
	}

	switch strings.ToUpper(objectType) {
	case "ROLES":
		return "LIST ROLES", nil

	case "USERS":
		return "LIST USERS", nil

	case "PERMISSIONS":
		// LIST PERMISSIONS [OF role]
		if role, ok := plan.Options["role"].(string); ok && role != "" {
			return fmt.Sprintf("LIST PERMISSIONS OF %s", role), nil
		}
		return "LIST PERMISSIONS", nil

	default:
		return "", fmt.Errorf("unsupported LIST object_type: %s (must be ROLES, USERS, or PERMISSIONS)", objectType)
	}
}

// renderShow generates SHOW statements
func renderShow(plan *AIResult) (string, error) {
	if plan.Options == nil {
		return "", fmt.Errorf("options required for SHOW - must specify show_type")
	}

	showType, ok := plan.Options["show_type"].(string)
	if !ok || showType == "" {
		return "", fmt.Errorf("'show_type' required in options for SHOW (VERSION, HOST, SESSION)")
	}

	switch strings.ToUpper(showType) {
	case "VERSION":
		return "SHOW VERSION", nil
	case "HOST":
		return "SHOW HOST", nil
	case "SESSION":
		return "SHOW SESSION", nil
	default:
		return "", fmt.Errorf("unsupported SHOW show_type: %s (must be VERSION, HOST, or SESSION)", showType)
	}
}

// renderUse generates USE statement
func renderUse(plan *AIResult) (string, error) {
	if plan.Keyspace == "" {
		return "", fmt.Errorf("keyspace required for USE")
	}

	return fmt.Sprintf("USE %s", plan.Keyspace), nil
}

// renderBatch generates BATCH statements
func renderBatch(plan *AIResult) (string, error) {
	// BATCH is complex - requires multiple statements
	// For now, return error indicating not yet supported
	return "", fmt.Errorf("BATCH operations not yet supported in query builder - use raw CQL")
}

// renderCreateIndex generates CREATE INDEX statements
func renderCreateIndex(plan *AIResult) (string, error) {
	if plan.Keyspace == "" || plan.Table == "" {
		return "", fmt.Errorf("keyspace and table required for CREATE INDEX")
	}

	indexName, ok := plan.Options["index_name"].(string)
	if !ok || indexName == "" {
		return "", fmt.Errorf("'index_name' required in options for CREATE INDEX")
	}

	column, ok := plan.Options["column"].(string)
	if !ok || column == "" {
		return "", fmt.Errorf("'column' required in options for CREATE INDEX")
	}

	return fmt.Sprintf("CREATE INDEX %s ON %s.%s (%s);", indexName, plan.Keyspace, plan.Table, column), nil
}

// renderDropIndex generates DROP INDEX statements
func renderDropIndex(plan *AIResult) (string, error) {
	if plan.Keyspace == "" {
		return "", fmt.Errorf("keyspace required for DROP INDEX")
	}

	indexName, ok := plan.Options["index_name"].(string)
	if !ok || indexName == "" {
		return "", fmt.Errorf("'index_name' required in options for DROP INDEX")
	}

	return fmt.Sprintf("DROP INDEX %s.%s;", plan.Keyspace, indexName), nil
}

// renderCreateType generates CREATE TYPE statements for UDTs
func renderCreateType(plan *AIResult) (string, error) {
	if plan.Keyspace == "" {
		return "", fmt.Errorf("keyspace required for CREATE TYPE")
	}

	typeName, ok := plan.Options["type_name"].(string)
	if !ok || typeName == "" {
		return "", fmt.Errorf("'type_name' required in options for CREATE TYPE")
	}

	if len(plan.Schema) == 0 {
		return "", fmt.Errorf("schema required for CREATE TYPE (field definitions)")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE TYPE %s.%s (", plan.Keyspace, typeName))

	fields := make([]string, 0, len(plan.Schema))
	for fieldName, fieldType := range plan.Schema {
		fields = append(fields, fmt.Sprintf("%s %s", fieldName, fieldType))
	}
	sb.WriteString(strings.Join(fields, ", "))
	sb.WriteString(");")

	return sb.String(), nil
}

// renderDropType generates DROP TYPE statements
func renderDropType(plan *AIResult) (string, error) {
	if plan.Keyspace == "" {
		return "", fmt.Errorf("keyspace required for DROP TYPE")
	}

	typeName, ok := plan.Options["type_name"].(string)
	if !ok || typeName == "" {
		return "", fmt.Errorf("'type_name' required in options for DROP TYPE")
	}

	return fmt.Sprintf("DROP TYPE %s.%s;", plan.Keyspace, typeName), nil
}

// renderCreateFunction generates CREATE FUNCTION statements
func renderCreateFunction(plan *AIResult) (string, error) {
	if plan.Keyspace == "" {
		return "", fmt.Errorf("keyspace required for CREATE FUNCTION")
	}

	functionName, ok := plan.Options["function_name"].(string)
	if !ok || functionName == "" {
		return "", fmt.Errorf("'function_name' required in options for CREATE FUNCTION")
	}

	returns, ok := plan.Options["returns"].(string)
	if !ok || returns == "" {
		return "", fmt.Errorf("'returns' required in options for CREATE FUNCTION")
	}

	language, ok := plan.Options["language"].(string)
	if !ok || language == "" {
		return "", fmt.Errorf("'language' required in options for CREATE FUNCTION")
	}

	body, ok := plan.Options["body"].(string)
	if !ok || body == "" {
		return "", fmt.Errorf("'body' required in options for CREATE FUNCTION")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE FUNCTION %s.%s (", plan.Keyspace, functionName))

	// Add arguments if provided
	if args, ok := plan.Options["arguments"].(map[string]interface{}); ok && len(args) > 0 {
		argList := make([]string, 0, len(args))
		for argName, argType := range args {
			if argTypeStr, ok := argType.(string); ok {
				argList = append(argList, fmt.Sprintf("%s %s", argName, argTypeStr))
			}
		}
		sb.WriteString(strings.Join(argList, ", "))
	}

	sb.WriteString(")")
	sb.WriteString(fmt.Sprintf(" RETURNS NULL ON NULL INPUT RETURNS %s LANGUAGE %s AS '%s';",
		returns, language, body))

	return sb.String(), nil
}

// renderDropFunction generates DROP FUNCTION statements
func renderDropFunction(plan *AIResult) (string, error) {
	if plan.Keyspace == "" {
		return "", fmt.Errorf("keyspace required for DROP FUNCTION")
	}

	functionName, ok := plan.Options["function_name"].(string)
	if !ok || functionName == "" {
		return "", fmt.Errorf("'function_name' required in options for DROP FUNCTION")
	}

	return fmt.Sprintf("DROP FUNCTION %s.%s;", plan.Keyspace, functionName), nil
}

// renderCreateAggregate generates CREATE AGGREGATE statements
func renderCreateAggregate(plan *AIResult) (string, error) {
	if plan.Keyspace == "" {
		return "", fmt.Errorf("keyspace required for CREATE AGGREGATE")
	}

	aggregateName, ok := plan.Options["aggregate_name"].(string)
	if !ok || aggregateName == "" {
		return "", fmt.Errorf("'aggregate_name' required in options for CREATE AGGREGATE")
	}

	sfunc, ok := plan.Options["sfunc"].(string)
	if !ok || sfunc == "" {
		return "", fmt.Errorf("'sfunc' (state function) required in options for CREATE AGGREGATE")
	}

	stype, ok := plan.Options["stype"].(string)
	if !ok || stype == "" {
		return "", fmt.Errorf("'stype' (state type) required in options for CREATE AGGREGATE")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE AGGREGATE %s.%s (", plan.Keyspace, aggregateName))

	// Add input type if provided
	if inputType, ok := plan.Options["input_type"].(string); ok && inputType != "" {
		sb.WriteString(inputType)
	}

	sb.WriteString(fmt.Sprintf(") SFUNC %s STYPE %s", sfunc, stype))

	// Optional: final function
	if finalFunc, ok := plan.Options["finalfunc"].(string); ok && finalFunc != "" {
		sb.WriteString(fmt.Sprintf(" FINALFUNC %s", finalFunc))
	}

	// Optional: initial condition
	if initCond, ok := plan.Options["initcond"].(string); ok && initCond != "" {
		sb.WriteString(fmt.Sprintf(" INITCOND %s", initCond))
	}

	sb.WriteString(";")
	return sb.String(), nil
}

// renderDropAggregate generates DROP AGGREGATE statements
func renderDropAggregate(plan *AIResult) (string, error) {
	if plan.Keyspace == "" {
		return "", fmt.Errorf("keyspace required for DROP AGGREGATE")
	}

	aggregateName, ok := plan.Options["aggregate_name"].(string)
	if !ok || aggregateName == "" {
		return "", fmt.Errorf("'aggregate_name' required in options for DROP AGGREGATE")
	}

	return fmt.Sprintf("DROP AGGREGATE %s.%s;", plan.Keyspace, aggregateName), nil
}

// renderCreateTrigger generates CREATE TRIGGER statements
func renderCreateTrigger(plan *AIResult) (string, error) {
	if plan.Keyspace == "" || plan.Table == "" {
		return "", fmt.Errorf("keyspace and table required for CREATE TRIGGER")
	}

	triggerName, ok := plan.Options["trigger_name"].(string)
	if !ok || triggerName == "" {
		return "", fmt.Errorf("'trigger_name' required in options for CREATE TRIGGER")
	}

	triggerClass, ok := plan.Options["trigger_class"].(string)
	if !ok || triggerClass == "" {
		return "", fmt.Errorf("'trigger_class' required in options for CREATE TRIGGER")
	}

	return fmt.Sprintf("CREATE TRIGGER %s ON %s.%s USING '%s';",
		triggerName, plan.Keyspace, plan.Table, triggerClass), nil
}

// renderDropTrigger generates DROP TRIGGER statements
func renderDropTrigger(plan *AIResult) (string, error) {
	if plan.Keyspace == "" || plan.Table == "" {
		return "", fmt.Errorf("keyspace and table required for DROP TRIGGER")
	}

	triggerName, ok := plan.Options["trigger_name"].(string)
	if !ok || triggerName == "" {
		return "", fmt.Errorf("'trigger_name' required in options for DROP TRIGGER")
	}

	return fmt.Sprintf("DROP TRIGGER %s ON %s.%s;", triggerName, plan.Keyspace, plan.Table), nil
}

// renderCreateMaterializedView generates CREATE MATERIALIZED VIEW statements
func renderCreateMaterializedView(plan *AIResult) (string, error) {
	if plan.Keyspace == "" {
		return "", fmt.Errorf("keyspace required for CREATE MATERIALIZED VIEW")
	}

	viewName, ok := plan.Options["view_name"].(string)
	if !ok || viewName == "" {
		return "", fmt.Errorf("'view_name' required in options for CREATE MATERIALIZED VIEW")
	}

	baseTable, ok := plan.Options["base_table"].(string)
	if !ok || baseTable == "" {
		return "", fmt.Errorf("'base_table' required in options for CREATE MATERIALIZED VIEW")
	}

	// Columns to select (default to all)
	selectColumns := "*"
	if len(plan.Columns) > 0 {
		selectColumns = strings.Join(plan.Columns, ", ")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE MATERIALIZED VIEW %s.%s AS SELECT %s FROM %s.%s",
		plan.Keyspace, viewName, selectColumns, plan.Keyspace, baseTable))

	// WHERE clause (required for MVs)
	if len(plan.Where) > 0 {
		sb.WriteString(" WHERE ")
		conditions := make([]string, 0, len(plan.Where))
		for _, w := range plan.Where {
			conditions = append(conditions, renderWhereClause(w))
		}
		sb.WriteString(strings.Join(conditions, " AND "))
	}

	// Primary key definition
	if pk, ok := plan.Options["primary_key"].(string); ok && pk != "" {
		sb.WriteString(fmt.Sprintf(" PRIMARY KEY (%s)", pk))
	}

	sb.WriteString(";")
	return sb.String(), nil
}

// renderDropMaterializedView generates DROP MATERIALIZED VIEW statements
func renderDropMaterializedView(plan *AIResult) (string, error) {
	if plan.Keyspace == "" {
		return "", fmt.Errorf("keyspace required for DROP MATERIALIZED VIEW")
	}

	viewName, ok := plan.Options["view_name"].(string)
	if !ok || viewName == "" {
		return "", fmt.Errorf("'view_name' required in options for DROP MATERIALIZED VIEW")
	}

	return fmt.Sprintf("DROP MATERIALIZED VIEW %s.%s;", plan.Keyspace, viewName), nil
}

// renderCreateRole generates CREATE ROLE statements
func renderCreateRole(plan *AIResult) (string, error) {
	roleName, ok := plan.Options["role_name"].(string)
	if !ok || roleName == "" {
		return "", fmt.Errorf("'role_name' required in options for CREATE ROLE")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE ROLE %s", roleName))

	// Build WITH clause if options provided
	withClauses := []string{}

	if password, ok := plan.Options["password"].(string); ok && password != "" {
		withClauses = append(withClauses, fmt.Sprintf("PASSWORD = '%s'", password))
	}
	if login, ok := plan.Options["login"].(bool); ok {
		withClauses = append(withClauses, fmt.Sprintf("LOGIN = %t", login))
	}
	if superuser, ok := plan.Options["superuser"].(bool); ok {
		withClauses = append(withClauses, fmt.Sprintf("SUPERUSER = %t", superuser))
	}

	if len(withClauses) > 0 {
		sb.WriteString(" WITH ")
		sb.WriteString(strings.Join(withClauses, " AND "))
	}

	sb.WriteString(";")
	return sb.String(), nil
}

// renderDropRole generates DROP ROLE statements
func renderDropRole(plan *AIResult) (string, error) {
	roleName, ok := plan.Options["role_name"].(string)
	if !ok || roleName == "" {
		return "", fmt.Errorf("'role_name' required in options for DROP ROLE")
	}

	return fmt.Sprintf("DROP ROLE %s;", roleName), nil
}

// renderCreateUser generates CREATE USER statements (deprecated but still supported)
func renderCreateUser(plan *AIResult) (string, error) {
	userName, ok := plan.Options["user_name"].(string)
	if !ok || userName == "" {
		return "", fmt.Errorf("'user_name' required in options for CREATE USER")
	}

	password, ok := plan.Options["password"].(string)
	if !ok || password == "" {
		return "", fmt.Errorf("'password' required in options for CREATE USER")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'", userName, password))

	// Optional superuser flag
	if superuser, ok := plan.Options["superuser"].(bool); ok && superuser {
		sb.WriteString(" SUPERUSER")
	}

	sb.WriteString(";")
	return sb.String(), nil
}

// renderDropUser generates DROP USER statements
func renderDropUser(plan *AIResult) (string, error) {
	userName, ok := plan.Options["user_name"].(string)
	if !ok || userName == "" {
		return "", fmt.Errorf("'user_name' required in options for DROP USER")
	}

	return fmt.Sprintf("DROP USER %s;", userName), nil
}

// ============================================================================
// Phase 0: Enhanced Data Type Formatting Functions
// ============================================================================
//
// These functions provide proper CQL formatting for all Cassandra data types:
// - Lists use [] (not ())
// - Sets use {} with deduplication
// - Maps use {key: val} with quoted keys
// - UDTs use {field: val} with unquoted field names
// - Functions like uuid() pass through unquoted
//
// Manual testing against Cassandra 5.0.6 confirmed all syntax.
// See: claude-notes/features/phase0_data_types_research.md
// ============================================================================

// formatPrimitive formats primitive CQL values (strings, numbers, UUIDs, booleans, null)
func formatPrimitive(v any) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		// Check if this is a function call (uuid(), now(), etc.)
		if isFunctionCall(val) {
			return val // Pass through without quotes
		}

		// Check if this is a UUID (8-4-4-4-12 format)
		if isUUID(val) {
			return val // UUIDs not quoted
		}

		// Regular string - escape single quotes and wrap
		escaped := strings.ReplaceAll(val, "'", "''")
		return fmt.Sprintf("'%s'", escaped)

	case bool:
		// CQL booleans are lowercase: true, false
		if val {
			return "true"
		}
		return "false"

	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%v", val)

	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", val)

	case float32, float64:
		return fmt.Sprintf("%v", val)

	default:
		// Fallback for unknown types
		return fmt.Sprintf("%v", val)
	}
}

// isFunctionCall detects CQL function calls (uuid(), now(), toTimestamp(), etc.)
func isFunctionCall(s string) bool {
	if s == "" {
		return false
	}

	// Must end with ) and contain (
	if !strings.HasSuffix(s, ")") || !strings.Contains(s, "(") {
		return false
	}

	// Must not start with quotes
	if strings.HasPrefix(s, "'") || strings.HasPrefix(s, "\"") {
		return false
	}

	// Must not start with ( (would be tuple/IN clause)
	if strings.HasPrefix(s, "(") {
		return false
	}

	return true
}

// isUUID checks if string matches UUID format (8-4-4-4-12 hex pattern)
func isUUID(s string) bool {
	return len(s) == 36 &&
		s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-'
}

// convertToSlice converts various slice types to []any
func convertToSlice(v any) ([]any, bool) {
	switch val := v.(type) {
	case []any:
		return val, true
	case []string:
		result := make([]any, len(val))
		for i, s := range val {
			result[i] = s
		}
		return result, true
	case []int:
		result := make([]any, len(val))
		for i, n := range val {
			result[i] = n
		}
		return result, true
	case []int64:
		result := make([]any, len(val))
		for i, n := range val {
			result[i] = n
		}
		return result, true
	case []float64:
		result := make([]any, len(val))
		for i, n := range val {
			result[i] = n
		}
		return result, true
	default:
		return nil, false
	}
}

// deduplicateElements removes duplicates from slice (for sets)
func deduplicateElements(elements []any) []any {
	if len(elements) == 0 {
		return elements
	}

	seen := make(map[string]bool)
	unique := make([]any, 0, len(elements))

	for _, elem := range elements {
		// Use string representation as key
		key := fmt.Sprintf("%v", elem)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, elem)
		}
	}

	return unique
}

// formatList formats list literals using square brackets []
// Example: ['item1', 'item2', 'item3']
func formatList(v any, elementType string) string {
	elements, ok := convertToSlice(v)
	if !ok {
		return "[]" // Invalid input → empty list
	}

	if len(elements) == 0 {
		return "[]" // Empty list
	}

	formatted := make([]string, len(elements))
	for i, elem := range elements {
		formatted[i] = formatValue(elem, elementType) // Recursive for nested types
	}

	return fmt.Sprintf("[%s]", strings.Join(formatted, ", "))
}

// formatSet formats set literals using curly braces {}
// Sets are deduplicated (Cassandra ensures uniqueness)
// Example: {'item1', 'item2', 'item3'}
func formatSet(v any, elementType string) string {
	elements, ok := convertToSlice(v)
	if !ok {
		return "{}" // Invalid input → empty set
	}

	if len(elements) == 0 {
		return "{}" // Empty set
	}

	// Deduplicate (Cassandra does this, but we do it too for correctness)
	unique := deduplicateElements(elements)

	formatted := make([]string, len(unique))
	for i, elem := range unique {
		formatted[i] = formatValue(elem, elementType) // Recursive
	}

	return fmt.Sprintf("{%s}", strings.Join(formatted, ", "))
}

// formatMap formats map literals for DML operations
// Maps use {key: value} with QUOTED keys (unlike UDTs)
// Example: {'key1': 'value1', 'key2': 'value2'}
func formatMap(v any, keyType, valueType string) string {
	m, ok := v.(map[string]any)
	if !ok {
		return "{}" // Invalid input
	}

	if len(m) == 0 {
		return "{}" // Empty map
	}

	pairs := make([]string, 0, len(m))
	for key, value := range m {
		// Format key and value
		formattedKey := formatValue(key, keyType)
		formattedValue := formatValue(value, valueType)

		// Map syntax: 'key': value
		pairs = append(pairs, fmt.Sprintf("%s: %s", formattedKey, formattedValue))
	}

	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

// formatUDT formats User-Defined Type literals
// UDTs use {field: value} with UNQUOTED field names (critical difference from maps!)
// Example: {street: '123 Main', city: 'NYC', zip: '10001'}
func formatUDT(v any, udtTypeName string) string {
	m, ok := v.(map[string]any)
	if !ok {
		return "{}" // Invalid input
	}

	if len(m) == 0 {
		return "{}" // Empty UDT
	}

	pairs := make([]string, 0, len(m))
	for field, value := range m {
		// Format value (will infer type or use primitive)
		formattedValue := formatValue(value, "")

		// UDT syntax: field: value (field name NOT quoted!)
		pairs = append(pairs, fmt.Sprintf("%s: %s", field, formattedValue))
	}

	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

// formatTuple formats tuple literals using parentheses ()
// Example: ('value1', 'value2', 'value3')
func formatTuple(v any) string {
	elements, ok := convertToSlice(v)
	if !ok {
		return "()" // Invalid input
	}

	if len(elements) == 0 {
		return "()" // Empty tuple (may be invalid, but return it)
	}

	formatted := make([]string, len(elements))
	for i, elem := range elements {
		formatted[i] = formatPrimitive(elem)
	}

	return fmt.Sprintf("(%s)", strings.Join(formatted, ", "))
}

// formatBlob formats blob literals with 0x hex prefix
// Example: 0xCAFEBABE
func formatBlob(v any) string {
	var hexStr string

	switch val := v.(type) {
	case string:
		// Remove 0x prefix if present, we'll add it back
		hexStr = strings.TrimPrefix(strings.ToUpper(val), "0X")
		hexStr = strings.TrimPrefix(hexStr, "0x")
	case []byte:
		// Convert byte array to hex string
		hexStr = fmt.Sprintf("%X", val)
	default:
		return "0x" // Invalid
	}

	return fmt.Sprintf("0x%s", hexStr)
}

// parseTypeHint extracts base type and element type from CQL type hint
// Examples:
//   "text" → ("text", "")
//   "list<text>" → ("list", "text")
//   "set<int>" → ("set", "int")
func parseTypeHint(hint string) (baseType, elementType string) {
	if hint == "" {
		return "", ""
	}

	// Simple types (no generics)
	if !strings.Contains(hint, "<") {
		return hint, ""
	}

	// Collection types: list<text>, set<int>, map<text,int>, tuple<text,int>
	openIdx := strings.Index(hint, "<")
	closeIdx := strings.LastIndex(hint, ">")

	if openIdx == -1 || closeIdx == -1 {
		return hint, "" // Malformed, return as-is
	}

	baseType = hint[:openIdx]
	elementType = hint[openIdx+1 : closeIdx]

	return baseType, elementType
}

// parseMapTypes extracts key and value types from map type hint
// Example: "map<text,int>" → ("text", "int")
func parseMapTypes(hint string) (keyType, valueType string) {
	if hint == "" {
		return "", ""
	}

	// Extract content between < >
	openIdx := strings.Index(hint, "<")
	closeIdx := strings.LastIndex(hint, ">")

	if openIdx == -1 || closeIdx == -1 {
		return "", ""
	}

	types := hint[openIdx+1 : closeIdx]

	// Split by comma (handles map<key,value>)
	parts := strings.Split(types, ",")
	if len(parts) != 2 {
		return "", "" // Not a map or malformed
	}

	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}
