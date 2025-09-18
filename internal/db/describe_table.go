package db

import (
	"fmt"
	"sort"
	"strings"

	"github.com/axonops/cqlai/internal/session"
)

// TableInfo holds table information for manual describe
type TableInfo struct {
	KeyspaceName   string
	TableName      string
	TableProps     map[string]interface{}
	Columns        []ColumnInfo
	PartitionKeys  []string
	ClusteringKeys []string
}

// ColumnInfo holds column information
type ColumnInfo struct {
	Name     string
	DataType string
	Kind     string
	Position int
}

// TableListInfo holds table list information for manual describe
type TableListInfo struct {
	Name           string
	GcGrace        int
	Compaction     map[string]string
	Compression    map[string]string
	PartitionKeys  []string
	ClusteringKeys []string
}

// DescribeTableQuery executes queries to get table information (for pre-4.0)
func (s *Session) DescribeTableQuery(keyspace string, tableName string) (*TableInfo, error) {
	// First check if table exists
	checkQuery := `SELECT table_name FROM system_schema.tables WHERE keyspace_name = ? AND table_name = ?`
	checkIter := s.Query(checkQuery, keyspace, tableName).Iter()
	var checkName string
	if !checkIter.Scan(&checkName) {
		_ = checkIter.Close()
		
		// Get available tables for better error message
		availQuery := `SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?`
		availIter := s.Query(availQuery, keyspace).Iter()
		var availableTables []string
		var availName string
		for availIter.Scan(&availName) {
			availableTables = append(availableTables, availName)
		}
		_ = availIter.Close()
		
		availableStr := "none"
		if len(availableTables) > 0 {
			availableStr = strings.Join(availableTables, ", ")
		}
		return nil, fmt.Errorf("table '%s' not found in keyspace '%s'. Available tables: %s", tableName, keyspace, availableStr)
	}
	_ = checkIter.Close()

	// Get table properties
	tableQuery := `SELECT * FROM system_schema.tables WHERE keyspace_name = ? AND table_name = ?`
	iter := s.Query(tableQuery, keyspace, tableName).Iter()
	
	tableProps := make(map[string]interface{})
	if !iter.MapScan(tableProps) {
		_ = iter.Close()
		return nil, fmt.Errorf("could not retrieve table properties")
	}
	_ = iter.Close()

	// Get columns
	colQuery := `SELECT column_name, type, kind, position 
	            FROM system_schema.columns 
	            WHERE keyspace_name = ? AND table_name = ?`
	
	colIter := s.Query(colQuery, keyspace, tableName).Iter()
	
	var columns []ColumnInfo
	var partitionKeys []string
	var clusteringKeys []string
	
	var colName, colType, colKind string
	var colPosition int
	
	for colIter.Scan(&colName, &colType, &colKind, &colPosition) {
		columns = append(columns, ColumnInfo{
			Name:     colName,
			DataType: colType,
			Kind:     colKind,
			Position: colPosition,
		})
		
		switch colKind {
		case "partition_key":
			partitionKeys = append(partitionKeys, colName)
		case "clustering":
			clusteringKeys = append(clusteringKeys, colName)
		}
	}
	_ = colIter.Close()

	// Sort columns: partition keys first, then clustering keys, then regular columns
	sort.Slice(columns, func(i, j int) bool {
		kindPriority := map[string]int{
			"partition_key": 0,
			"clustering":    1,
			"regular":       2,
		}
		
		iPriority := kindPriority[columns[i].Kind]
		jPriority := kindPriority[columns[j].Kind]
		
		if iPriority != jPriority {
			return iPriority < jPriority
		}
		
		// Within same kind, sort by position
		return columns[i].Position < columns[j].Position
	})

	return &TableInfo{
		KeyspaceName:   keyspace,
		TableName:      tableName,
		TableProps:     tableProps,
		Columns:        columns,
		PartitionKeys:  partitionKeys,
		ClusteringKeys: clusteringKeys,
	}, nil
}

// DescribeAllTablesQuery executes queries to list all tables from all keyspaces
func (s *Session) DescribeAllTablesQuery() ([]TableListInfo, error) {
	// Query all tables from all keyspaces
	tableQuery := `SELECT keyspace_name, table_name, gc_grace_seconds, compaction, compression
	               FROM system_schema.tables`
	iter := s.Query(tableQuery).Iter()

	tableMap := make(map[string]*TableListInfo) // keyspace.table -> TableListInfo

	for {
		var keyspaceName, tableName string
		var gcGrace int
		var compaction, compression map[string]string

		if !iter.Scan(&keyspaceName, &tableName, &gcGrace, &compaction, &compression) {
			break
		}

		// Format table name with keyspace prefix
		fullTableName := fmt.Sprintf("%s.%s", keyspaceName, tableName)

		tableInfo := &TableListInfo{
			Name:           fullTableName,
			GcGrace:        gcGrace,
			Compaction:     compaction,
			Compression:    compression,
			PartitionKeys:  []string{},
			ClusteringKeys: []string{},
		}

		tableMap[fullTableName] = tableInfo
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("failed to list tables: %v", err)
	}

	// Now query all columns and filter for primary keys in the application
	// We can't use WHERE kind IN (...) as it requires ALLOW FILTERING
	columnQuery := `SELECT keyspace_name, table_name, column_name, kind, position
	                FROM system_schema.columns`

	colIter := s.Query(columnQuery).Iter()
	for {
		var keyspaceName, tableName, columnName, kind string
		var position int

		if !colIter.Scan(&keyspaceName, &tableName, &columnName, &kind, &position) {
			break
		}

		// Only process partition_key and clustering columns
		if kind != "partition_key" && kind != "clustering" {
			continue
		}

		fullTableName := fmt.Sprintf("%s.%s", keyspaceName, tableName)
		if tableInfo, ok := tableMap[fullTableName]; ok {
			switch kind {
			case "partition_key":
				// Ensure slice is big enough
				for len(tableInfo.PartitionKeys) <= position {
					tableInfo.PartitionKeys = append(tableInfo.PartitionKeys, "")
				}
				tableInfo.PartitionKeys[position] = columnName
			case "clustering":
				// Ensure slice is big enough
				for len(tableInfo.ClusteringKeys) <= position {
					tableInfo.ClusteringKeys = append(tableInfo.ClusteringKeys, "")
				}
				tableInfo.ClusteringKeys[position] = columnName
			}
		}
	}

	if err := colIter.Close(); err != nil {
		return nil, fmt.Errorf("failed to query column information: %v", err)
	}

	// Build the result slice from the map
	var tables []TableListInfo
	for _, tableInfo := range tableMap {
		tables = append(tables, *tableInfo)
	}

	// Sort by keyspace and table name
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].Name < tables[j].Name
	})

	return tables, nil
}

// DescribeTablesQuery executes queries to list all tables (for pre-4.0)
func (s *Session) DescribeTablesQuery(keyspace string) ([]TableListInfo, error) {
	// Query table details
	tableQuery := `SELECT table_name, gc_grace_seconds, compaction, compression 
	               FROM system_schema.tables 
	               WHERE keyspace_name = ?`
	iter := s.Query(tableQuery, keyspace).Iter()

	var tables []TableListInfo
	var tableName string
	var gcGrace int
	var compaction, compression map[string]string

	// Collect table details
	for iter.Scan(&tableName, &gcGrace, &compaction, &compression) {
		tables = append(tables, TableListInfo{
			Name:        tableName,
			GcGrace:     gcGrace,
			Compaction:  compaction,
			Compression: compression,
		})
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("error listing tables: %v", err)
	}

	// Get primary keys for each table
	for i := range tables {
		colQuery := `SELECT column_name, kind 
		            FROM system_schema.columns 
		            WHERE keyspace_name = ? AND table_name = ?`

		colIter := s.Query(colQuery, keyspace, tables[i].Name).Iter()

		var colName, colKind string
		var pkNames []string
		var ckNames []string

		for colIter.Scan(&colName, &colKind) {
			switch colKind {
			case "partition_key":
				pkNames = append(pkNames, colName)
			case "clustering":
				ckNames = append(ckNames, colName)
			}
		}
		_ = colIter.Close()

		tables[i].PartitionKeys = pkNames
		tables[i].ClusteringKeys = ckNames
	}

	// Sort by table name
	sort.Slice(tables, func(i, j int) bool {
		return tables[i].Name < tables[j].Name
	})

	return tables, nil
}

// DBDescribeTable handles version detection and returns appropriate data
func (s *Session) DBDescribeTable(sessionMgr *session.Manager, tableName string) (interface{}, *TableInfo, error) {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	if s.IsVersion4OrHigher() {
		// Try server-side DESCRIBE
		describeQuery := fmt.Sprintf("DESCRIBE TABLE %s", tableName)
		
		iter := s.Query(describeQuery).Iter()
		
		// The server returns a result set with columns like 'keyspace_name', 'type', 'name', 'create_statement'
		result := make(map[string]interface{})
		if iter.MapScan(result) {
			_ = iter.Close()
			
			if createStmt, ok := result["create_statement"]; ok {
				return fmt.Sprintf("%v", createStmt), nil, nil
			}
		}
		
		_ = iter.Close()
		// Server-side DESCRIBE returned no results, fall back to manual
	}
	
	// Fall back to manual construction for pre-4.0 or if server-side failed
	// Check if table name includes keyspace qualification
	keyspaceName := ""
	if sessionMgr != nil {
		keyspaceName = sessionMgr.CurrentKeyspace()
	}
	actualTableName := tableName
	
	if strings.Contains(tableName, ".") {
		parts := strings.Split(tableName, ".")
		if len(parts) == 2 {
			keyspaceName = parts[0]
			actualTableName = parts[1]
		}
	} else if keyspaceName == "" {
		return nil, nil, fmt.Errorf("no keyspace selected")
	}

	tableInfo, err := s.DescribeTableQuery(keyspaceName, actualTableName)
	if err != nil {
		return nil, nil, err
	}

	return nil, tableInfo, nil // Manual query result, return TableInfo for formatting
}

// DBDescribeTables handles version detection and returns appropriate data
func (s *Session) DBDescribeTables(sessionMgr *session.Manager) (interface{}, []TableListInfo, error) {
	// Get current keyspace
	currentKeyspace := ""
	if sessionMgr != nil {
		currentKeyspace = sessionMgr.CurrentKeyspace()
	}

	// Always use manual construction - DESCRIBE is not a real CQL query even in 4.0+
	// DESCRIBE commands are meta-commands that need to be handled specially

	if currentKeyspace != "" {
		// If keyspace is selected, show tables from that keyspace
		tables, err := s.DescribeTablesQuery(currentKeyspace)
		if err != nil {
			return nil, nil, err
		}
		return nil, tables, nil // Manual query result, return TableListInfo for formatting
	} else {
		// If no keyspace selected, show all tables from all keyspaces
		tables, err := s.DescribeAllTablesQuery()
		if err != nil {
			return nil, nil, err
		}
		return nil, tables, nil // Manual query result, return TableListInfo for formatting
	}
}