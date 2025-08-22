package router

import (
	"fmt"
	"strings"
	"github.com/axonops/cqlai/internal/logger"
)

// describeFunctions lists all functions
func (v *CqlCommandVisitorImpl) describeFunctions() interface{} {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeFunctions", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Use server-side DESCRIBE FUNCTIONS
		logger.DebugToFile("describeFunctions", "Using server-side DESCRIBE FUNCTIONS")
		return v.session.ExecuteCQLQuery("DESCRIBE FUNCTIONS")
	}
	
	// Fall back to manual construction for pre-4.0
	currentKeyspace := v.session.CurrentKeyspace()
	if currentKeyspace == "" {
		return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
	}

	query := `SELECT function_name, argument_types, return_type 
	          FROM system_schema.functions 
	          WHERE keyspace_name = ?`

	iter := v.session.Query(query, currentKeyspace).Iter()

	results := [][]string{{"Function", "Arguments", "Return Type"}}
	var functionName, returnType string
	var argumentTypes []string

	for iter.Scan(&functionName, &argumentTypes, &returnType) {
		args := strings.Join(argumentTypes, ", ")
		results = append(results, []string{functionName, args, returnType})
	}
	iter.Close()

	if len(results) == 1 {
		return fmt.Sprintf("No functions in keyspace %s", currentKeyspace)
	}

	return results
}

// describeFunction shows detailed information about a specific function
func (v *CqlCommandVisitorImpl) describeFunction(functionName string) interface{} {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeFunction", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Parse keyspace.function or just function
		var describeCmd string
		if strings.Contains(functionName, ".") {
			describeCmd = fmt.Sprintf("DESCRIBE FUNCTION %s", functionName)
		} else {
			currentKeyspace := v.session.CurrentKeyspace()
			if currentKeyspace == "" {
				return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
			}
			describeCmd = fmt.Sprintf("DESCRIBE FUNCTION %s.%s", currentKeyspace, functionName)
		}
		
		logger.DebugToFile("describeFunction", fmt.Sprintf("Using server-side: %s", describeCmd))
		return v.session.ExecuteCQLQuery(describeCmd)
	}
	
	// Fall back to manual construction for pre-4.0
	currentKeyspace := v.session.CurrentKeyspace()
	if currentKeyspace == "" {
		return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
	}

	// Functions can be overloaded, so we might get multiple results
	query := `SELECT function_name, argument_types, argument_names, return_type, 
	                language, body, called_on_null_input
	          FROM system_schema.functions 
	          WHERE keyspace_name = ? AND function_name = ?`

	iter := v.session.Query(query, currentKeyspace, functionName).Iter()

	var result strings.Builder
	var found bool

	var name, returnType, language, body string
	var argumentTypes, argumentNames []string
	var calledOnNull bool

	for iter.Scan(&name, &argumentTypes, &argumentNames, &returnType, &language, &body, &calledOnNull) {
		found = true

		if result.Len() > 0 {
			result.WriteString("\n\n")
		}

		result.WriteString(fmt.Sprintf("CREATE FUNCTION %s.%s(", currentKeyspace, functionName))

		// Format arguments
		for i := 0; i < len(argumentNames) && i < len(argumentTypes); i++ {
			if i > 0 {
				result.WriteString(", ")
			}
			result.WriteString(fmt.Sprintf("%s %s", argumentNames[i], argumentTypes[i]))
		}

		result.WriteString(")\n")

		if calledOnNull {
			result.WriteString("    CALLED ON NULL INPUT\n")
		} else {
			result.WriteString("    RETURNS NULL ON NULL INPUT\n")
		}

		result.WriteString(fmt.Sprintf("    RETURNS %s\n", returnType))
		result.WriteString(fmt.Sprintf("    LANGUAGE %s\n", language))
		result.WriteString(fmt.Sprintf("    AS '%s';", body))
	}
	iter.Close()

	if !found {
		return fmt.Sprintf("Function '%s' not found in keyspace '%s'", functionName, currentKeyspace)
	}

	return result.String()
}
