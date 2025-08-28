package router

import (
	"fmt"
	"strings"
)

// describeAggregates lists all aggregates in the current keyspace
func (v *CqlCommandVisitorImpl) describeAggregates() interface{} {
	serverResult, aggregates, err := v.session.DBDescribeAggregates(sessionManager)
	
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
	currentKeyspace := ""
	if sessionManager != nil {
		currentKeyspace = sessionManager.CurrentKeyspace()
	}
	results := [][]string{{"Aggregate", "Arguments", "State Type", "Return Type"}}
	
	for _, agg := range aggregates {
		args := strings.Join(agg.ArgumentTypes, ", ")
		results = append(results, []string{agg.Name, args, agg.StateType, agg.ReturnType})
	}

	if len(results) == 1 {
		return fmt.Sprintf("No aggregates in keyspace %s", currentKeyspace)
	}

	return results
}

// describeAggregate shows detailed information about a specific aggregate
func (v *CqlCommandVisitorImpl) describeAggregate(aggregateName string) interface{} {
	serverResult, aggregateInfo, err := v.session.DBDescribeAggregate(sessionManager, aggregateName)
	
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
	if aggregateInfo == nil {
		currentKeyspace := ""
		if sessionManager != nil {
			currentKeyspace = sessionManager.CurrentKeyspace()
		}
		return fmt.Sprintf("Aggregate '%s' not found in keyspace '%s'", aggregateName, currentKeyspace)
	}
	
	currentKeyspace := ""
	if sessionManager != nil {
		currentKeyspace = sessionManager.CurrentKeyspace()
	}
	var result strings.Builder
	result.WriteString(fmt.Sprintf("CREATE AGGREGATE %s.%s(", currentKeyspace, aggregateName))

	if len(aggregateInfo.ArgumentTypes) > 0 {
		result.WriteString(strings.Join(aggregateInfo.ArgumentTypes, ", "))
	}

	result.WriteString(")\n")
	result.WriteString(fmt.Sprintf("    SFUNC %s\n", aggregateInfo.StateFunc))
	result.WriteString(fmt.Sprintf("    STYPE %s\n", aggregateInfo.StateType))

	if aggregateInfo.FinalFunc != "" {
		result.WriteString(fmt.Sprintf("    FINALFUNC %s\n", aggregateInfo.FinalFunc))
	}

	if aggregateInfo.InitCond != "" {
		result.WriteString(fmt.Sprintf("    INITCOND %s", aggregateInfo.InitCond))
	}

	result.WriteString(";")

	return result.String()
}
