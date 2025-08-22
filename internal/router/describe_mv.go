package router

import (
	"fmt"
	"strings"
	"github.com/axonops/cqlai/internal/logger"
)

// describeMaterializedView shows detailed information about a materialized view
func (v *CqlCommandVisitorImpl) describeMaterializedView(viewName string) interface{} {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	cassandraVersion := v.session.CassandraVersion()
	logger.DebugToFile("describeMaterializedView", fmt.Sprintf("Cassandra version: %s", cassandraVersion))
	
	if isVersion4OrHigher(cassandraVersion) {
		// Parse keyspace.view or just view
		var describeCmd string
		if strings.Contains(viewName, ".") {
			describeCmd = fmt.Sprintf("DESCRIBE MATERIALIZED VIEW %s", viewName)
		} else {
			currentKeyspace := v.session.CurrentKeyspace()
			if currentKeyspace == "" {
				return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
			}
			describeCmd = fmt.Sprintf("DESCRIBE MATERIALIZED VIEW %s.%s", currentKeyspace, viewName)
		}
		
		logger.DebugToFile("describeMaterializedView", fmt.Sprintf("Using server-side: %s", describeCmd))
		return v.session.ExecuteCQLQuery(describeCmd)
	}
	
	// Fall back to manual construction for pre-4.0
	currentKeyspace := v.session.CurrentKeyspace()
	if currentKeyspace == "" {
		return "No keyspace selected. Use 'USE keyspace_name' to select a keyspace."
	}

	query := `SELECT view_name, base_table_name, where_clause, 
	                bloom_filter_fp_chance, caching, comment, compaction, compression,
	                crc_check_chance, dclocal_read_repair_chance, default_time_to_live,
	                gc_grace_seconds, max_index_interval, memtable_flush_period_in_ms,
	                min_index_interval, read_repair_chance, speculative_retry
	          FROM system_schema.views 
	          WHERE keyspace_name = ? AND view_name = ?`

	iter := v.session.Query(query, currentKeyspace, viewName).Iter()

	var name, baseTable, whereClause, comment, caching, speculativeRetry string
	var bloomFilterFpChance, crcCheckChance, dclocalReadRepairChance, readRepairChance float64
	var defaultTTL, gcGrace, maxIndexInterval, memtableFlushPeriod, minIndexInterval int
	var compaction, compression map[string]string

	if !iter.Scan(&name, &baseTable, &whereClause, &bloomFilterFpChance, &caching, &comment,
		&compaction, &compression, &crcCheckChance, &dclocalReadRepairChance, &defaultTTL,
		&gcGrace, &maxIndexInterval, &memtableFlushPeriod, &minIndexInterval,
		&readRepairChance, &speculativeRetry) {
		iter.Close()
		return fmt.Sprintf("Materialized view '%s' not found in keyspace '%s'", viewName, currentKeyspace)
	}
	iter.Close()

	// Get view columns
	colQuery := `SELECT column_name, type, kind 
	            FROM system_schema.columns 
	            WHERE keyspace_name = ? AND table_name = ?`

	colIter := v.session.Query(colQuery, currentKeyspace, viewName).Iter()

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Materialized View: %s.%s\n\n", currentKeyspace, viewName))
	result.WriteString(fmt.Sprintf("Base Table: %s\n", baseTable))
	if whereClause != "" {
		result.WriteString(fmt.Sprintf("Where Clause: %s\n", whereClause))
	}
	result.WriteString("\n")

	// Show CREATE MATERIALIZED VIEW statement
	result.WriteString(fmt.Sprintf("CREATE MATERIALIZED VIEW %s.%s AS\n", currentKeyspace, viewName))
	result.WriteString(fmt.Sprintf("  SELECT * FROM %s.%s\n", currentKeyspace, baseTable))
	if whereClause != "" {
		result.WriteString(fmt.Sprintf("  WHERE %s\n", whereClause))
	}

	// Get primary key info
	var partitionKeys, clusteringKeys []string
	var colName, colType, colKind string

	for colIter.Scan(&colName, &colType, &colKind) {
		if colKind == "partition_key" {
			partitionKeys = append(partitionKeys, colName)
		} else if colKind == "clustering" {
			clusteringKeys = append(clusteringKeys, colName)
		}
	}
	colIter.Close()

	if len(partitionKeys) > 0 {
		result.WriteString("  PRIMARY KEY (")
		if len(partitionKeys) == 1 {
			result.WriteString(partitionKeys[0])
		} else {
			result.WriteString(fmt.Sprintf("(%s)", strings.Join(partitionKeys, ", ")))
		}
		if len(clusteringKeys) > 0 {
			result.WriteString(", " + strings.Join(clusteringKeys, ", "))
		}
		result.WriteString(")")
	}

	result.WriteString(";")

	return result.String()
}
