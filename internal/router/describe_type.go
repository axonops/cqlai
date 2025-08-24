package router

import (
	"fmt"
	"strings"
)

// describeTypes lists all user-defined types
func (v *CqlCommandVisitorImpl) describeTypes() interface{} {
	serverResult, types, err := v.session.DBDescribeTypes()
	
	if err != nil {
		if err.Error() == "no keyspace selected" {
			return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
		}
		return fmt.Sprintf("Error: %v", err)
	}
	
	if serverResult != nil {
		// Server-side DESCRIBE result, return as-is
		return serverResult
	}
	
	// Manual query result - format it
	currentKeyspace := v.session.CurrentKeyspace()
	results := [][]string{{"Type Name"}}
	
	for _, t := range types {
		results = append(results, []string{t.Name})
	}

	if len(results) == 1 {
		return fmt.Sprintf("No user-defined types in keyspace %s", currentKeyspace)
	}

	return results
}

// describeType shows detailed information about a user-defined type
func (v *CqlCommandVisitorImpl) describeType(typeName string) interface{} {
	serverResult, typeInfo, err := v.session.DBDescribeType(typeName)
	
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
	if typeInfo == nil {
		currentKeyspace := v.session.CurrentKeyspace()
		return fmt.Sprintf("Type '%s' not found in keyspace '%s'", typeName, currentKeyspace)
	}
	
	currentKeyspace := v.session.CurrentKeyspace()
	var result strings.Builder
	result.WriteString(fmt.Sprintf("CREATE TYPE %s.%s (\n", currentKeyspace, typeName))

	for i := 0; i < len(typeInfo.FieldNames) && i < len(typeInfo.FieldTypes); i++ {
		result.WriteString(fmt.Sprintf("    %s %s", typeInfo.FieldNames[i], typeInfo.FieldTypes[i]))
		if i < len(typeInfo.FieldNames)-1 {
			result.WriteString(",")
		}
		result.WriteString("\n")
	}

	result.WriteString(");")

	return result.String()
}
