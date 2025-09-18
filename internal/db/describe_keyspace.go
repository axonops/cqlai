package db

import (
	"fmt"
	"sort"
)

// KeyspaceInfo holds keyspace information for manual describe
type KeyspaceInfo struct {
	Name          string
	DurableWrites bool
	Replication   map[string]string
}

// KeyspaceListInfo holds keyspace list information for manual describe
type KeyspaceListInfo struct {
	Name        string
	Replication map[string]string
}

// DescribeKeyspaceQuery executes the query to get keyspace information (for pre-4.0)
func (s *Session) DescribeKeyspaceQuery(keyspaceName string) (*KeyspaceInfo, error) {
	query := `SELECT keyspace_name, durable_writes, replication 
	          FROM system_schema.keyspaces 
	          WHERE keyspace_name = ?`

	iter := s.Query(query, keyspaceName).Iter()

	var name string
	var durableWrites bool
	var replication map[string]string

	if !iter.Scan(&name, &durableWrites, &replication) {
		_ = iter.Close()
		return nil, fmt.Errorf("keyspace '%s' not found", keyspaceName)
	}
	_ = iter.Close()

	return &KeyspaceInfo{
		Name:          name,
		DurableWrites: durableWrites,
		Replication:   replication,
	}, nil
}

// DescribeKeyspacesQuery executes the query to list all keyspaces (for pre-4.0)
func (s *Session) DescribeKeyspacesQuery() ([]KeyspaceListInfo, error) {
	iter := s.Query("SELECT keyspace_name, replication FROM system_schema.keyspaces").Iter()

	var keyspaces []KeyspaceListInfo
	var keyspaceName string
	var replication map[string]string

	// Collect all keyspace info
	for iter.Scan(&keyspaceName, &replication) {
		keyspaces = append(keyspaces, KeyspaceListInfo{
			Name:        keyspaceName,
			Replication: replication,
		})
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("error listing keyspaces: %v", err)
	}

	// Sort by name
	sort.Slice(keyspaces, func(i, j int) bool {
		return keyspaces[i].Name < keyspaces[j].Name
	})

	return keyspaces, nil
}

// DBDescribeKeyspace handles version detection and returns appropriate data
func (s *Session) DBDescribeKeyspace(keyspaceName string) (interface{}, *KeyspaceInfo, error) {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	if s.IsVersion4OrHigher() {
		// Use server-side DESCRIBE KEYSPACE
		describeCmd := fmt.Sprintf("DESCRIBE KEYSPACE %s", keyspaceName)
		result := s.ExecuteCQLQuery(describeCmd)
		return result, nil, nil // Server-side result, no KeyspaceInfo needed
	}

	// Fall back to manual construction for pre-4.0
	keyspaceInfo, err := s.DescribeKeyspaceQuery(keyspaceName)
	if err != nil {
		return nil, nil, err
	}

	return nil, keyspaceInfo, nil // Manual query result, return KeyspaceInfo for formatting
}

// DBDescribeKeyspaces handles version detection and returns appropriate data
func (s *Session) DBDescribeKeyspaces() (interface{}, []KeyspaceListInfo, error) {
	// Always use manual construction - DESCRIBE is not a real CQL query even in 4.0+
	// DESCRIBE commands are meta-commands that need to be handled specially
	keyspaces, err := s.DescribeKeyspacesQuery()
	if err != nil {
		return nil, nil, err
	}

	return nil, keyspaces, nil // Manual query result, return KeyspaceListInfo for formatting
}
