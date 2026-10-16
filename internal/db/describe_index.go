package db

import (
	"fmt"

	"github.com/axonops/cqlai/internal/session"
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
	// Need ALLOW FILTERING since we're not specifying table_name
	query := `SELECT table_name, index_name, kind, options
	          FROM system_schema.indexes
	          WHERE keyspace_name = ? AND index_name = ? ALLOW FILTERING`

	iter := s.Query(query, keyspace, indexName).Iter()

	var tableName, idxName, kind string
	var options map[string]string

	if !iter.Scan(&tableName, &idxName, &kind, &options) {
		_ = iter.Close()
		return nil, fmt.Errorf("index '%s' not found in keyspace '%s'", indexName, keyspace)
	}
	_ = iter.Close()

	return &IndexInfo{
		TableName: tableName,
		IndexName: idxName,
		Kind:      kind,
		Options:   options,
	}, nil
}

// DescribeIndexQueryWithTable executes the query to get index information when table is known
func (s *Session) DescribeIndexQueryWithTable(keyspace string, tableName string, indexName string) (*IndexInfo, error) {
	query := `SELECT table_name, index_name, kind, options
	          FROM system_schema.indexes
	          WHERE keyspace_name = ? AND table_name = ? AND index_name = ?`

	iter := s.Query(query, keyspace, tableName, indexName).Iter()

	var tblName, idxName, kind string
	var options map[string]string

	if !iter.Scan(&tblName, &idxName, &kind, &options) {
		_ = iter.Close()
		return nil, fmt.Errorf("index '%s' not found in keyspace '%s' table '%s'", indexName, keyspace, tableName)
	}
	_ = iter.Close()

	return &IndexInfo{
		TableName: tblName,
		IndexName: idxName,
		Kind:      kind,
		Options:   options,
	}, nil
}

// DBDescribeIndex handles version detection and returns appropriate data
func (s *Session) DBDescribeIndex(sessionMgr *session.Manager, indexName string) (interface{}, *IndexInfo, error) {
	keyspace, name, err := qualify(sessionMgr, indexName)
	if err != nil {
		return nil, nil, err
	}

	// Cassandra 4.0 and later describe it themselves.
	if s.IsVersion4OrHigher() {
		return s.ExecuteCQLQuery("DESCRIBE INDEX " + quoteQualified(keyspace, name)), nil, nil
	}

	indexInfo, err := s.DescribeIndexQuery(keyspace, name)
	if err != nil {
		return nil, nil, err
	}
	return nil, indexInfo, nil
}
