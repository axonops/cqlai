package router

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/parquet"
)

// executeCopyFromParquetPartitioned handles COPY FROM for partitioned Parquet datasets
func (h *MetaCommandHandler) executeCopyFromParquetPartitioned(table string, columns []string, reader *parquet.PartitionedParquetReader, options map[string]string) interface{} {

	// Get schema from partitioned dataset
	parquetColumns, _ := reader.GetSchema()
	partitionColumns := reader.GetPartitionColumns()

	logger.DebugfToFile("CopyFromPartitioned", "Found %d partition columns: %v",
		len(partitionColumns), partitionColumns)
	logger.DebugfToFile("CopyFromPartitioned", "Parquet columns from files: %v", parquetColumns)

	// If no columns specified, filter to only columns that don't have dots (virtual columns)
	// and let Cassandra handle validation of which columns exist in the table
	if len(columns) == 0 {
		columns = []string{}
		for _, col := range parquetColumns {
			// Skip virtual partition columns and partition columns that won't exist in table
			if !strings.Contains(col, ".") {
				// Also skip partition columns that are likely from Hive-style paths
				isPartitionCol := false
				for _, partCol := range partitionColumns {
					if col == partCol && !strings.Contains(partCol, ".") {
						// This is a regular partition column (like "year", "month")
						// Skip it unless we know the table has it
						isPartitionCol = true
						break
					}
				}
				if !isPartitionCol {
					columns = append(columns, col)
				}
			}
		}

		if len(columns) == 0 {
			return "No importable columns found in Parquet file"
		}
		logger.DebugfToFile("CopyFromPartitioned", "Auto-selected columns for import: %v", columns)
	} else {
		// Validate that specified columns exist in the Parquet files
		columnMap := make(map[string]bool)
		for _, col := range parquetColumns {
			columnMap[col] = true
		}

		for _, col := range columns {
			if !columnMap[col] {
				return fmt.Sprintf("Column '%s' not found in Parquet dataset", col)
			}
		}
	}

	// Build column list for INSERT
	columnList := strings.Join(columns, ", ")

	// Parse numeric options
	batchSize, _ := strconv.Atoi(options["CHUNKSIZE"])
	if batchSize <= 0 {
		batchSize = 1000 // Default batch size
	}

	maxRows, _ := strconv.Atoi(options["MAXROWS"])
	skipRows, _ := strconv.Atoi(options["SKIPROWS"])
	maxInsertErrors, _ := strconv.Atoi(options["MAXINSERTERRORS"])

	// Get partition filter if specified
	partitionFilter := options["PARTITION_FILTER"]

	// Process rows
	rowCount := 0
	processedRows := 0
	insertErrorCount := 0
	skippedRows := 0
	errorMessages := make([]string, 0, 5) // Store first few error messages

	batch := make([]string, 0, batchSize)

	for {
		// Read a batch of rows
		rows, err := reader.ReadBatch(batchSize)
		if err != nil {
			if err.Error() == "EOF" || len(rows) == 0 {
				break
			}
			return fmt.Sprintf("Error reading partitioned dataset: %v", err)
		}

		for _, row := range rows {
			rowCount++

			// Skip initial rows if specified
			if skipRows > 0 && skippedRows < skipRows {
				skippedRows++
				continue
			}

			// Apply partition filter if specified
			if partitionFilter != "" && !matchesPartitionFilter(row, partitionFilter) {
				continue
			}

			// Check max rows limit
			if maxRows > 0 && processedRows >= maxRows {
				goto done
			}

			// Build INSERT statement
			values := make([]string, len(columns))
			for i, col := range columns {
				if val, ok := row[col]; ok {
					values[i] = h.formatParquetValueForInsert(val, col)
				} else {
					values[i] = "NULL"
				}
			}

			insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				table, columnList, strings.Join(values, ", "))
			logger.DebugfToFile("CopyFromPartitioned", "Insert query: %s", insertQuery)

			batch = append(batch, insertQuery)

			// Execute batch when full
			if len(batch) >= batchSize {
				errors := h.executePartitionedBatch(batch, processedRows, &errorMessages)
				insertErrorCount += errors

				if maxInsertErrors > 0 && insertErrorCount > maxInsertErrors {
					return fmt.Sprintf("Too many insert errors: %d (max: %d)",
						insertErrorCount, maxInsertErrors)
				}

				processedRows += len(batch) - errors
				batch = batch[:0]
			}
		}

		if len(rows) < batchSize {
			// No more rows available
			break
		}
	}

done:
	// Execute remaining batch
	if len(batch) > 0 {
		errors := h.executePartitionedBatch(batch, processedRows, &errorMessages)
		insertErrorCount += errors
		processedRows += len(batch) - errors
	}

	// Build result message
	partitionInfo := reader.GetPartitionFiles()
	result := fmt.Sprintf("Imported %d rows from partitioned dataset (%d files, %d partitions)",
		processedRows, len(partitionInfo), len(partitionColumns))

	if insertErrorCount > 0 {
		result += fmt.Sprintf(" with %d errors", insertErrorCount)

		// Show first few error messages to help user diagnose issues
		if len(errorMessages) > 0 {
			result += "\n\nFirst errors encountered:"
			for _, errMsg := range errorMessages {
				result += "\n  - " + errMsg
			}
			if insertErrorCount > len(errorMessages) {
				result += fmt.Sprintf("\n  ... and %d more errors", insertErrorCount-len(errorMessages))
			}
		}
	}

	if skipRows > 0 {
		result += fmt.Sprintf(" (skipped %d rows)", skipRows)
	}

	return result
}

// matchesPartitionFilter checks if a row matches the partition filter
// Filter format: "year=2024,month=01" or "year=2024 AND month IN (01,02,03)"
func matchesPartitionFilter(row map[string]interface{}, filter string) bool {
	// Simple implementation for key=value filters
	parts := strings.Split(filter, ",")

	for _, part := range parts {
		kv := strings.Split(strings.TrimSpace(part), "=")
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		expectedValue := strings.TrimSpace(kv[1])

		if actualValue, ok := row[key]; ok {
			if fmt.Sprintf("%v", actualValue) != expectedValue {
				return false
			}
		} else {
			return false
		}
	}

	return true
}

// formatParquetValueForInsert formats a Parquet value for use in an INSERT statement
func (h *MetaCommandHandler) formatParquetValueForInsert(value interface{}, columnName string) string {
	if value == nil {
		return "NULL"
	}

	switch v := value.(type) {
	case string:
		trimmed := strings.TrimSpace(v)

		// Check if this is a UUID (for UUID/TIMEUUID columns)
		if strings.Contains(strings.ToLower(columnName), "uuid") ||
		   strings.Contains(strings.ToLower(columnName), "id") {
			// Check if it looks like a UUID (8-4-4-4-12 format)
			if isUUIDFormat(trimmed) {
				return trimmed // No quotes for UUIDs
			}
		}

		// Regular string - escape single quotes
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case time.Time:
		// Format timestamp for CQL (use RFC3339 format with quotes)
		return fmt.Sprintf("'%s'", v.Format(time.RFC3339Nano))
	case bool:
		return fmt.Sprintf("%t", v)
	case float32, float64:
		return fmt.Sprintf("%f", v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case []byte:
		// Format as hex blob
		return fmt.Sprintf("0x%x", v)
	case map[string]interface{}:
		// Format as JSON for complex types
		// This is a simplified version - actual implementation would need proper JSON encoding
		return fmt.Sprintf("'%v'", v)
	case []interface{}:
		// Format as list
		items := make([]string, len(v))
		for i, item := range v {
			items[i] = h.formatParquetValueForInsert(item, columnName)
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ", "))
	default:
		// Default to string representation
		return fmt.Sprintf("'%v'", v)
	}
}

// executePartitionedBatch executes a batch of INSERT statements for partitioned data
func (h *MetaCommandHandler) executePartitionedBatch(batch []string, rowIndex int, errorMessages *[]string) int {
	errorCount := 0
	for i, query := range batch {
		result := h.session.ExecuteCQLQuery(query)
		if err, ok := result.(error); ok {
			logger.DebugfToFile("CopyFromPartitioned", "Insert error: %v", err)
			logger.DebugfToFile("CopyFromPartitioned", "Failed query: %s", query)
			errorCount++

			// Store first few error messages for user display (limit to 5)
			if len(*errorMessages) < 5 {
				*errorMessages = append(*errorMessages, fmt.Sprintf("Row %d: %v", rowIndex+i, err))
			}
		}
	}
	return errorCount
}