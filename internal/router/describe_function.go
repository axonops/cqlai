package router

import (
	"fmt"
	"strings"
)

// describeFunctions lists all functions
func (v *CqlCommandVisitorImpl) describeFunctions() interface{} {
	result, isServerSide, err := v.session.DBDescribeFunctions(sessionManager)

	if err != nil {
		if err.Error() == "no keyspace selected" {
			return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
		}
		return fmt.Sprintf("Error: %v", err)
	}

	if isServerSide {
		// Server-side DESCRIBE result, return as-is
		return result
	}

	// Manual query result, check if empty
	if results, ok := result.([][]string); ok {
		if len(results) == 1 {
			currentKeyspace := ""
			if sessionManager != nil {
				currentKeyspace = sessionManager.CurrentKeyspace()
			}
			return fmt.Sprintf("No functions in keyspace %s", currentKeyspace)
		}
		return results
	}

	return result
}

// describeFunction shows detailed information about a specific function
func (v *CqlCommandVisitorImpl) describeFunction(functionName string) interface{} {
	serverResult, functions, err := v.session.DBDescribeFunction(sessionManager, functionName)

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
	if len(functions) == 0 {
		currentKeyspace := ""
		if sessionManager != nil {
			currentKeyspace = sessionManager.CurrentKeyspace()
		}
		return fmt.Sprintf("Function '%s' not found in keyspace '%s'", functionName, currentKeyspace)
	}

	currentKeyspace := ""
	if sessionManager != nil {
		currentKeyspace = sessionManager.CurrentKeyspace()
	}
	var result strings.Builder
	for i, fn := range functions {
		if i > 0 {
			result.WriteString("\n\n")
		}

		result.WriteString(fmt.Sprintf("CREATE FUNCTION %s.%s(", currentKeyspace, functionName))

		// Format arguments
		for j := 0; j < len(fn.ArgumentNames) && j < len(fn.ArgumentTypes); j++ {
			if j > 0 {
				result.WriteString(", ")
			}
			result.WriteString(fmt.Sprintf("%s %s", fn.ArgumentNames[j], fn.ArgumentTypes[j]))
		}

		result.WriteString(")\n")

		if fn.CalledOnNull {
			result.WriteString("    CALLED ON NULL INPUT\n")
		} else {
			result.WriteString("    RETURNS NULL ON NULL INPUT\n")
		}

		result.WriteString(fmt.Sprintf("    RETURNS %s\n", fn.ReturnType))
		result.WriteString(fmt.Sprintf("    LANGUAGE %s\n", fn.Language))
		result.WriteString(fmt.Sprintf("    AS '%s';", fn.Body))
	}

	return result.String()
}
