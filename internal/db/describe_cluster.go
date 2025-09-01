package db

import (
	"fmt"
)

// ClusterInfo holds cluster information for manual describe
type ClusterInfo struct {
	ClusterName string
	Partitioner string
	Version     string
}

// DescribeClusterQuery executes the query to get cluster information (for pre-4.0)
func (s *Session) DescribeClusterQuery() (*ClusterInfo, error) {
	iter := s.Query("SELECT cluster_name, partitioner, release_version FROM system.local").Iter()
	
	var clusterName, partitioner, version string
	if iter.Scan(&clusterName, &partitioner, &version) {
		_ = iter.Close()
		return &ClusterInfo{
			ClusterName: clusterName,
			Partitioner: partitioner,
			Version:     version,
		}, nil
	}
	
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("error describing cluster: %v", err)
	}
	
	return nil, fmt.Errorf("could not retrieve cluster information")
}

// DBDescribeCluster handles version detection and returns appropriate data
func (s *Session) DBDescribeCluster() (interface{}, *ClusterInfo, error) {
	// Check if we can use server-side DESCRIBE (Cassandra 4.0+)
	if s.IsVersion4OrHigher() {
		// Use server-side DESCRIBE CLUSTER
		result := s.ExecuteCQLQuery("DESCRIBE CLUSTER")
		return result, nil, nil // Server-side result, no ClusterInfo needed
	}
	
	// Fall back to manual construction for pre-4.0
	clusterInfo, err := s.DescribeClusterQuery()
	if err != nil {
		return nil, nil, err
	}
	
	return nil, clusterInfo, nil // Manual query result, return ClusterInfo for formatting
}