package db

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/session"
)

// MaterializedViewInfo holds materialized view information for manual describe
type MaterializedViewInfo struct {
	Name                    string
	BaseTable               string
	WhereClause             string
	BloomFilterFpChance     float64
	Caching                 string
	Comment                 string
	Compaction              map[string]string
	Compression             map[string]string
	CrcCheckChance          float64
	DclocalReadRepairChance float64
	DefaultTTL              int
	GcGrace                 int
	MaxIndexInterval        int
	MemtableFlushPeriod     int
	MinIndexInterval        int
	ReadRepairChance        float64
	SpeculativeRetry        string
	PartitionKeys           []string
	ClusteringKeys          []string
}

// DescribeMaterializedViewQuery executes the query to get materialized view information (for pre-4.0)
func (s *Session) DescribeMaterializedViewQuery(keyspace string, viewName string) (*MaterializedViewInfo, error) {
	query := `SELECT view_name, base_table_name, where_clause, 
	                bloom_filter_fp_chance, caching, comment, compaction, compression,
	                crc_check_chance, dclocal_read_repair_chance, default_time_to_live,
	                gc_grace_seconds, max_index_interval, memtable_flush_period_in_ms,
	                min_index_interval, read_repair_chance, speculative_retry
	          FROM system_schema.views 
	          WHERE keyspace_name = ? AND view_name = ?`

	iter := s.Query(query, keyspace, viewName).Iter()

	var name, baseTable, whereClause, comment, caching, speculativeRetry string
	var bloomFilterFpChance, crcCheckChance, dclocalReadRepairChance, readRepairChance float64
	var defaultTTL, gcGrace, maxIndexInterval, memtableFlushPeriod, minIndexInterval int
	var compaction, compression map[string]string

	if !iter.Scan(&name, &baseTable, &whereClause, &bloomFilterFpChance, &caching, &comment,
		&compaction, &compression, &crcCheckChance, &dclocalReadRepairChance, &defaultTTL,
		&gcGrace, &maxIndexInterval, &memtableFlushPeriod, &minIndexInterval,
		&readRepairChance, &speculativeRetry) {
		_ = iter.Close()
		return nil, fmt.Errorf("materialized view '%s' not found in keyspace '%s'", viewName, keyspace)
	}
	_ = iter.Close()

	// Get view columns for primary key info
	colQuery := `SELECT column_name, type, kind 
	            FROM system_schema.columns 
	            WHERE keyspace_name = ? AND table_name = ?`

	colIter := s.Query(colQuery, keyspace, viewName).Iter()

	var partitionKeys, clusteringKeys []string
	var colName, colType, colKind string

	for colIter.Scan(&colName, &colType, &colKind) {
		switch colKind {
		case "partition_key":
			partitionKeys = append(partitionKeys, colName)
		case "clustering":
			clusteringKeys = append(clusteringKeys, colName)
		}
	}
	_ = colIter.Close()

	return &MaterializedViewInfo{
		Name:                    name,
		BaseTable:               baseTable,
		WhereClause:             whereClause,
		BloomFilterFpChance:     bloomFilterFpChance,
		Caching:                 caching,
		Comment:                 comment,
		Compaction:              compaction,
		Compression:             compression,
		CrcCheckChance:          crcCheckChance,
		DclocalReadRepairChance: dclocalReadRepairChance,
		DefaultTTL:              defaultTTL,
		GcGrace:                 gcGrace,
		MaxIndexInterval:        maxIndexInterval,
		MemtableFlushPeriod:     memtableFlushPeriod,
		MinIndexInterval:        minIndexInterval,
		ReadRepairChance:        readRepairChance,
		SpeculativeRetry:        speculativeRetry,
		PartitionKeys:           partitionKeys,
		ClusteringKeys:          clusteringKeys,
	}, nil
}

// DBDescribeMaterializedView handles version detection and returns appropriate data
func (s *Session) DBDescribeMaterializedView(sessionMgr *session.Manager, viewName string) (interface{}, *MaterializedViewInfo, error) {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	if s.IsVersion4OrHigher() {
		// Parse keyspace.view or just view
		var describeCmd string
		if strings.Contains(viewName, ".") {
			describeCmd = fmt.Sprintf("DESCRIBE MATERIALIZED VIEW %s", viewName)
		} else {
			currentKeyspace := ""
			if sessionMgr != nil {
				currentKeyspace = sessionMgr.CurrentKeyspace()
			}
			if currentKeyspace == "" {
				return nil, nil, fmt.Errorf("no keyspace selected")
			}
			describeCmd = fmt.Sprintf("DESCRIBE MATERIALIZED VIEW %s.%s", currentKeyspace, viewName)
		}

		result := s.ExecuteCQLQuery(describeCmd)
		return result, nil, nil // Server-side result, no MaterializedViewInfo needed
	}

	// Fall back to manual construction for pre-4.0
	currentKeyspace := ""
	if sessionMgr != nil {
		currentKeyspace = sessionMgr.CurrentKeyspace()
	}
	if currentKeyspace == "" {
		return nil, nil, fmt.Errorf("no keyspace selected")
	}

	mvInfo, err := s.DescribeMaterializedViewQuery(currentKeyspace, viewName)
	if err != nil {
		return nil, nil, err
	}

	return nil, mvInfo, nil // Manual query result, return MaterializedViewInfo for formatting
}
