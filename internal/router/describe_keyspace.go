package router

import (
	"fmt"
	"sort"
	"strings"
)

// describeKeyspace shows detailed information about a specific keyspace
func (v *CqlCommandVisitorImpl) describeKeyspace(keyspaceName string) interface{} {
	serverResult, keyspaceInfo, err := v.session.DBDescribeKeyspace(keyspaceName)

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Sprintf("Keyspace '%s' not found", keyspaceName)
		}
		return fmt.Sprintf("Error: %v", err)
	}

	if serverResult != nil {
		// Server-side DESCRIBE result in Cassandra 4.0+ returns table data
		// Extract the create_statement column for proper CQL view display
		return v.extractCreateStatements(serverResult, "DESCRIBE KEYSPACE")
	}

	// Manual query result - format it
	if keyspaceInfo == nil {
		return fmt.Sprintf("Keyspace '%s' not found", keyspaceName)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("CREATE KEYSPACE %s WITH replication = {\n", keyspaceInfo.Name))

	first := true
	for k, v := range keyspaceInfo.Replication {
		if !first {
			result.WriteString(",\n")
		}
		result.WriteString(fmt.Sprintf("    '%s': '%s'", k, v))
		first = false
	}

	result.WriteString(fmt.Sprintf("\n} AND durable_writes = %v;", keyspaceInfo.DurableWrites))

	return result.String()
}

// describeKeyspaces lists all keyspaces in the cluster.
func (v *CqlCommandVisitorImpl) describeKeyspaces() interface{} {
	serverResult, keyspaces, err := v.session.DBDescribeKeyspaces()

	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	if serverResult != nil {
		// Server-side DESCRIBE result, return as-is
		return serverResult
	}

	// Manual query result - format it
	if len(keyspaces) == 0 {
		return "No keyspaces found"
	}

	// Build results showing keyspace and replication strategy
	results := [][]string{{"keyspace_name", "replication_strategy", "replication_factor"}}
	for _, ks := range keyspaces {
		strategy := "Unknown"
		factor := "N/A"

		if class, ok := ks.Replication["class"]; ok {
			// Extract the strategy name (last part after dot)
			parts := strings.Split(class, ".")
			strategy = parts[len(parts)-1]
		}

		switch strategy {
		case "SimpleStrategy":
			if rf, ok := ks.Replication["replication_factor"]; ok {
				factor = rf
			}
		case "NetworkTopologyStrategy":
			// Show datacenter replication settings
			var dcs []string
			for k, val := range ks.Replication {
				if k != "class" {
					dcs = append(dcs, fmt.Sprintf("%s:%s", k, val))
				}
			}
			if len(dcs) > 0 {
				sort.Strings(dcs)
				factor = strings.Join(dcs, ", ")
			}
		}

		results = append(results, []string{ks.Name, strategy, factor})
	}

	return results
}
