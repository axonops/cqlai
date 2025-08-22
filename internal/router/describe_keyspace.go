package router

import (
	"fmt"
	"sort"
	"strings"
	"github.com/axonops/cqlai/internal/logger"
)

// describeKeyspace shows detailed information about a specific keyspace
func (v *CqlCommandVisitorImpl) describeKeyspace(keyspaceName string) interface{} {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeKeyspace", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Use server-side DESCRIBE KEYSPACE
		describeCmd := fmt.Sprintf("DESCRIBE KEYSPACE %s", keyspaceName)
		logger.DebugToFile("describeKeyspace", fmt.Sprintf("Using server-side: %s", describeCmd))
		return v.session.ExecuteCQLQuery(describeCmd)
	}
	
	// Fall back to manual construction for pre-4.0
	query := `SELECT keyspace_name, durable_writes, replication 
	          FROM system_schema.keyspaces 
	          WHERE keyspace_name = ?`

	iter := v.session.Query(query, keyspaceName).Iter()

	var name string
	var durableWrites bool
	var replication map[string]string

	if !iter.Scan(&name, &durableWrites, &replication) {
		iter.Close()
		return fmt.Sprintf("Keyspace '%s' not found", keyspaceName)
	}
	iter.Close()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("CREATE KEYSPACE %s WITH replication = {\n", name))

	first := true
	for k, v := range replication {
		if !first {
			result.WriteString(",\n")
		}
		result.WriteString(fmt.Sprintf("    '%s': '%s'", k, v))
		first = false
	}

	result.WriteString(fmt.Sprintf("\n} AND durable_writes = %v;", durableWrites))

	return result.String()
}

// describeKeyspaces lists all keyspaces in the cluster.
func (v *CqlCommandVisitorImpl) describeKeyspaces() interface{} {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeKeyspaces", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Use server-side DESCRIBE KEYSPACES
		logger.DebugToFile("describeKeyspaces", "Using server-side DESCRIBE KEYSPACES")
		return v.session.ExecuteCQLQuery("DESCRIBE KEYSPACES")
	}
	
	// Fall back to manual construction for pre-4.0
	// Equivalent to: SELECT keyspace_name FROM system_schema.keyspaces
	iter := v.session.Query("SELECT keyspace_name, replication FROM system_schema.keyspaces").Iter()

	// Prepare result as table data
	type ksInfo struct {
		name        string
		replication map[string]string
	}

	var keyspaces []ksInfo
	var keyspaceName string
	var replication map[string]string

	// Collect all keyspace info
	for iter.Scan(&keyspaceName, &replication) {
		keyspaces = append(keyspaces, ksInfo{name: keyspaceName, replication: replication})
	}
	if err := iter.Close(); err != nil {
		return fmt.Errorf("error listing keyspaces: %v", err)
	}

	// Sort by name
	sort.Slice(keyspaces, func(i, j int) bool {
		return keyspaces[i].name < keyspaces[j].name
	})

	// Build results showing keyspace and replication strategy
	results := [][]string{{"keyspace_name", "replication_strategy", "replication_factor"}}
	for _, ks := range keyspaces {
		strategy := "Unknown"
		factor := "N/A"

		if class, ok := ks.replication["class"]; ok {
			// Extract the strategy name (last part after dot)
			parts := strings.Split(class, ".")
			strategy = parts[len(parts)-1]
		}

		if strategy == "SimpleStrategy" {
			if rf, ok := ks.replication["replication_factor"]; ok {
				factor = rf
			}
		} else if strategy == "NetworkTopologyStrategy" {
			// Show datacenter replication settings
			var dcs []string
			for k, val := range ks.replication {
				if k != "class" {
					dcs = append(dcs, fmt.Sprintf("%s:%s", k, val))
				}
			}
			if len(dcs) > 0 {
				sort.Strings(dcs)
				factor = strings.Join(dcs, ", ")
			}
		}

		results = append(results, []string{ks.name, strategy, factor})
	}

	return results
}