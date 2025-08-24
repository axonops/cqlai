package db

import (
	"fmt"
	"strings"
)

// IndexInfo holds index information for manual describe
type IndexInfo struct {
	TableName string
	IndexName string
	Kind      string
	Options   map[string]string
}

// DescribeIndexQuery executes the query to get index information (for pre-4.0)
func (s *Session) DescribeIndexQuery(keyspace string, indexName string) (*IndexInfo, error) {
	query := `SELECT table_name, index_name, kind, options 
	          FROM system_schema.indexes 
	          WHERE keyspace_name = ? AND index_name = ?`

	iter := s.Query(query, keyspace, indexName).Iter()

	var tableName, idxName, kind string
	var options map[string]string

	if !iter.Scan(&tableName, &idxName, &kind, &options) {
		iter.Close()
		return nil, fmt.Errorf("index '%s' not found in keyspace '%s'", indexName, keyspace)
	}
	iter.Close()

	return &IndexInfo{
		TableName: tableName,
		IndexName: idxName,
		Kind:      kind,
		Options:   options,
	}, nil
}

// DBDescribeIndex handles version detection and returns appropriate data
func (s *Session) DBDescribeIndex(indexName string) (interface{}, *IndexInfo, error) {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	if s.IsVersion4OrHigher() {
		// Parse keyspace.index or just index
		var describeCmd string
		if strings.Contains(indexName, ".") {
			describeCmd = fmt.Sprintf("DESCRIBE INDEX %s", indexName)
		} else {
			currentKeyspace := s.CurrentKeyspace()
			if currentKeyspace == "" {
				return nil, nil, fmt.Errorf("no keyspace selected")
			}
			describeCmd = fmt.Sprintf("DESCRIBE INDEX %s.%s", currentKeyspace, indexName)
		}

		result := s.ExecuteCQLQuery(describeCmd)
		return result, nil, nil // Server-side result, no IndexInfo needed
	}

	// Fall back to manual construction for pre-4.0
	currentKeyspace := s.CurrentKeyspace()
	if currentKeyspace == "" {
		return nil, nil, fmt.Errorf("no keyspace selected")
	}

	indexInfo, err := s.DescribeIndexQuery(currentKeyspace, indexName)
	if err != nil {
		return nil, nil, err
	}

	return nil, indexInfo, nil // Manual query result, return IndexInfo for formatting
}
