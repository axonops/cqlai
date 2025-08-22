package router

import (
	"fmt"
	"strings"
	"github.com/axonops/cqlai/internal/logger"
)

// describeAggregates lists all aggregates in the current keyspace
func (v *CqlCommandVisitorImpl) describeAggregates() interface{} {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeAggregates", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Use server-side DESCRIBE AGGREGATES
		logger.DebugToFile("describeAggregates", "Using server-side DESCRIBE AGGREGATES")
		return v.session.ExecuteCQLQuery("DESCRIBE AGGREGATES")
	}
	
	// Fall back to manual construction for pre-4.0
	currentKeyspace := v.session.CurrentKeyspace()
	if currentKeyspace == "" {
		return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
	}

	query := `SELECT aggregate_name, argument_types, state_type, return_type 
	          FROM system_schema.aggregates 
	          WHERE keyspace_name = ?`

	iter := v.session.Query(query, currentKeyspace).Iter()

	results := [][]string{{"Aggregate", "Arguments", "State Type", "Return Type"}}
	var aggregateName, stateType, returnType string
	var argumentTypes []string

	for iter.Scan(&aggregateName, &argumentTypes, &stateType, &returnType) {
		args := strings.Join(argumentTypes, ", ")
		results = append(results, []string{aggregateName, args, stateType, returnType})
	}
	iter.Close()

	if len(results) == 1 {
		return fmt.Sprintf("No aggregates in keyspace %s", currentKeyspace)
	}

	return results
}

// describeAggregate shows detailed information about a specific aggregate
func (v *CqlCommandVisitorImpl) describeAggregate(aggregateName string) interface{} {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeAggregate", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Parse keyspace.aggregate or just aggregate
		var describeCmd string
		if strings.Contains(aggregateName, ".") {
			describeCmd = fmt.Sprintf("DESCRIBE AGGREGATE %s", aggregateName)
		} else {
			currentKeyspace := v.session.CurrentKeyspace()
			if currentKeyspace == "" {
				return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
			}
			describeCmd = fmt.Sprintf("DESCRIBE AGGREGATE %s.%s", currentKeyspace, aggregateName)
		}
		
		logger.DebugToFile("describeAggregate", fmt.Sprintf("Using server-side: %s", describeCmd))
		return v.session.ExecuteCQLQuery(describeCmd)
	}
	
	// Fall back to manual construction for pre-4.0
	currentKeyspace := v.session.CurrentKeyspace()
	if currentKeyspace == "" {
		return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
	}

	query := `SELECT aggregate_name, argument_types, state_func, state_type, 
	                final_func, initcond, return_type
	          FROM system_schema.aggregates 
	          WHERE keyspace_name = ? AND aggregate_name = ?`

	iter := v.session.Query(query, currentKeyspace, aggregateName).Iter()

	var name, stateFunc, stateType, finalFunc, initCond, returnType string
	var argumentTypes []string

	if !iter.Scan(&name, &argumentTypes, &stateFunc, &stateType, &finalFunc, &initCond, &returnType) {
		iter.Close()
		return fmt.Sprintf("Aggregate '%s' not found in keyspace '%s'", aggregateName, currentKeyspace)
	}
	iter.Close()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("CREATE AGGREGATE %s.%s(", currentKeyspace, aggregateName))

	if len(argumentTypes) > 0 {
		result.WriteString(strings.Join(argumentTypes, ", "))
	}

	result.WriteString(")\n")
	result.WriteString(fmt.Sprintf("    SFUNC %s\n", stateFunc))
	result.WriteString(fmt.Sprintf("    STYPE %s\n", stateType))

	if finalFunc != "" {
		result.WriteString(fmt.Sprintf("    FINALFUNC %s\n", finalFunc))
	}

	if initCond != "" {
		result.WriteString(fmt.Sprintf("    INITCOND %s", initCond))
	}

	result.WriteString(";")

	return result.String()
}
