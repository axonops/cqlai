package router

import (
	"fmt"
	"strings"

	"github.com/axonops/cqlai/internal/db"
)

// formatPrimaryKey formats partition and clustering keys for display
func (p *CommandParser) formatPrimaryKey(partitionKeys []string, clusteringKeys []string) string {
	if len(partitionKeys) == 0 {
		return "?"
	}

	var primaryKeys []string
	switch {
	case len(partitionKeys) == 1 && len(clusteringKeys) == 0:
		// Single partition key only
		primaryKeys = partitionKeys
	case len(partitionKeys) == 1 && len(clusteringKeys) > 0:
		// Single partition key with clustering keys
		primaryKeys = make([]string, 0, len(partitionKeys)+len(clusteringKeys))
		primaryKeys = append(primaryKeys, partitionKeys...)
		primaryKeys = append(primaryKeys, clusteringKeys...)
	default:
		// Composite partition key
		primaryKeys = append([]string{fmt.Sprintf("(%s)", strings.Join(partitionKeys, ","))}, clusteringKeys...)
	}

	return strings.Join(primaryKeys, ", ")
}

// formatCompaction formats compaction strategy for display
func (p *CommandParser) formatCompaction(compaction map[string]string) string {
	compactionStr := "Unknown"
	if class, ok := compaction["class"]; ok {
		// Get just the strategy name without the full class path
		parts := strings.Split(class, ".")
		strategy := parts[len(parts)-1]
		// Use abbreviation if available, otherwise use the strategy name
		if abbrev, found := compactionStrategyAbbrev[strategy]; found {
			compactionStr = abbrev
		} else {
			compactionStr = strategy
		}
	}
	return compactionStr
}

// formatCompression formats compression strategy for display
func (p *CommandParser) formatCompression(compression map[string]string) string {
	compressionStr := "None"
	if class, ok := compression["class"]; ok {
		parts := strings.Split(class, ".")
		algorithm := parts[len(parts)-1]
		// Use abbreviation if available, otherwise use the algorithm name
		if abbrev, found := compressionAbbrev[algorithm]; found {
			compressionStr = abbrev
		} else {
			compressionStr = algorithm
		}
	}
	return compressionStr
}

// formatGCGrace formats GC grace seconds for display
func (p *CommandParser) formatGCGrace(gcGrace int) string {
	gcGraceStr := fmt.Sprintf("%d", gcGrace)
	switch {
	case gcGrace >= 86400:
		days := gcGrace / 86400
		gcGraceStr = fmt.Sprintf("%dd", days)
	case gcGrace >= 3600:
		hours := gcGrace / 3600
		gcGraceStr = fmt.Sprintf("%dh", hours)
	case gcGrace > 0:
		gcGraceStr = fmt.Sprintf("%ds", gcGrace)
	}
	return gcGraceStr
}


// filterStreamingTablesResult filters streaming results for current keyspace
func (p *CommandParser) filterStreamingTablesResult(streamResult db.StreamingQueryResult, currentKeyspace string) interface{} {
	// Find keyspace_name column index
	keyspaceColIdx := -1
	nameColIdx := -1
	for idx, colName := range streamResult.ColumnNames {
		switch colName {
		case "keyspace_name":
			keyspaceColIdx = idx
		case "name":
			nameColIdx = idx
		}
	}

	if keyspaceColIdx >= 0 && nameColIdx >= 0 && currentKeyspace != "" {
		// Consume the iterator and collect filtered table names
		filteredResults := [][]string{{"name"}}

		for {
			rowMap := make(map[string]interface{})
			if !streamResult.Iterator.MapScan(rowMap) {
				break
			}

			// Check if this row is for current keyspace
			if ks, ok := rowMap["keyspace_name"].(string); ok && ks == currentKeyspace {
				if name, ok := rowMap["name"].(string); ok {
					filteredResults = append(filteredResults, []string{name})
				}
			}
		}

		return db.QueryResult{
			Data:        filteredResults,
			ColumnTypes: []string{"text"},
			RowCount:    len(filteredResults) - 1,
		}
	}

	// Return original if we can't filter
	return streamResult
}

// filterQueryTablesResult filters query results for current keyspace
func (p *CommandParser) filterQueryTablesResult(queryResult db.QueryResult, currentKeyspace string) interface{} {
	if currentKeyspace == "" || len(queryResult.Data) == 0 {
		return queryResult
	}

	// Find column indices
	headers := queryResult.Data[0]
	keyspaceColIdx := -1
	nameColIdx := -1

	for idx, header := range headers {
		switch header {
		case "keyspace_name":
			keyspaceColIdx = idx
		case "name":
			nameColIdx = idx
		}
	}

	if keyspaceColIdx >= 0 && nameColIdx >= 0 {
		// Filter results
		filteredResults := [][]string{{"name"}}

		for i := 1; i < len(queryResult.Data); i++ {
			row := queryResult.Data[i]
			if len(row) > keyspaceColIdx && row[keyspaceColIdx] == currentKeyspace {
				if len(row) > nameColIdx {
					filteredResults = append(filteredResults, []string{row[nameColIdx]})
				}
			}
		}

		return db.QueryResult{
			Data:        filteredResults,
			ColumnTypes: []string{"text"},
			RowCount:    len(filteredResults) - 1,
		}
	}

	return queryResult
}