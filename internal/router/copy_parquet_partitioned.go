package router

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/parquet"
)

// executeCopyToParquetPartitioned executes COPY TO operation with partitioning
func (h *MetaCommandHandler) executeCopyToParquetPartitioned(table string, columns []string, outputDir string, options map[string]string) interface{} {
	logger.DebugfToFile("CopyToParquetPartitioned", "Starting partitioned export for table: %s, outputDir: %s", table, outputDir)

	// Parse partition columns
	partitionStr := options["PARTITION"]
	var partitionColumns []string
	if partitionStr != "" {
		partitionColumns = strings.Split(partitionStr, ",")
		for i := range partitionColumns {
			partitionColumns[i] = strings.TrimSpace(partitionColumns[i])
		}
	}

	logger.DebugfToFile("CopyToParquetPartitioned", "Partition columns: %v", partitionColumns)

	// Build SELECT query
	var query string
	if len(columns) > 0 {
		query = fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), table)
	} else {
		query = fmt.Sprintf("SELECT * FROM %s", table)
	}

	// Add LIMIT if specified
	if limit := options["LIMIT"]; limit != "" {
		query += " LIMIT " + limit
	}

	// Clean the output directory path
	cleanPath := filepath.Clean(outputDir)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(cleanPath, 0750); err != nil { // #nosec G703 - the directory the user named for their own export
		return fmt.Sprintf("Error creating output directory: %v", err)
	}

	// Execute query with streaming
	result := h.session.ExecuteStreamingQuery(query)

	switch v := result.(type) {
	case db.StreamingQueryResult:
		defer v.Iterator.Close()

		if len(v.Headers) == 0 {
			return "No columns found in result"
		}

		// Clean headers
		cleanHeaders := make([]string, len(v.Headers))
		for i, header := range v.Headers {
			cleanHeaders[i] = db.StripKeyMarker(header)
		}

		// Create compression option
		compressionStr := strings.ToUpper(options["COMPRESSION"])
		compression := parquet.ParseCompression(compressionStr)

		// Parse max file size if specified
		maxFileSizeStr := options["MAX_FILE_SIZE"]
		maxFileSize := int64(100 * 1024 * 1024) // 100MB default
		if maxFileSizeStr != "" {
			// Parse size with units (e.g., "50MB", "1GB")
			maxFileSize = parseFileSize(maxFileSizeStr)
		}

		// Create partitioned writer options
		writerOptions := parquet.PartitionedWriterOptions{
			WriterOptions: parquet.WriterOptions{
				ChunkSize:   10000,
				Compression: compression,
			},
			PartitionColumns: partitionColumns,
			MaxOpenFiles:     10, // Could be configurable
			MaxFileSize:      maxFileSize,
		}

		// Create partitioned writer with TypeInfo if available
		var writer *parquet.PartitionedParquetWriter
		var err error
		if len(v.ColumnTypeInfos) > 0 {
			// Use the TypeInfo-aware constructor
			writer, err = parquet.NewPartitionedParquetWriterWithTypeInfo(cleanPath, cleanHeaders, v.ColumnTypes, v.ColumnTypeInfos, writerOptions)
		} else {
			// Fall back to standard constructor
			writer, err = parquet.NewPartitionedParquetWriter(cleanPath, cleanHeaders, v.ColumnTypes, writerOptions)
		}
		if err != nil {
			return fmt.Sprintf("Error creating partitioned writer: %v", err)
		}
		defer writer.Close()

		// Read and write data in batches
		rowCount := 0
		batchSize := 1000
		batch := make([]map[string]interface{}, 0, batchSize)

		// Prepare scan destinations - use cleanHeaders since that's what we'll iterate with
		// ScanRow: nil is a real NULL, and a tuple is one value.
		cols := v.Iterator.Columns()
		for scanned := map[string]interface{}{}; db.ScanRow(v.Iterator, scanned); scanned = map[string]interface{}{} {
			rowData := make(map[string]interface{})
			for i, colName := range cleanHeaders {
				if i < len(cols) {
					rowData[colName] = scanned[cols[i].Name]
				}
			}

			batch = append(batch, rowData)

			// Write batch when full
			if len(batch) >= batchSize {
				if err := writer.WriteRows(batch); err != nil {
					return fmt.Sprintf("Error writing batch: %v", err)
				}
				rowCount += len(batch)
				batch = batch[:0]
			}
		}

		// Write remaining batch
		if len(batch) > 0 {
			if err := writer.WriteRows(batch); err != nil {
				return fmt.Sprintf("Error writing final batch: %v", err)
			}
			rowCount += len(batch)
		}

		// The rows stop at the end of the table or at an error, and only
		// Close tells which. A partial export is not reported as a whole one.
		if err := v.Iterator.Close(); err != nil {
			return fmt.Sprintf("Error reading rows after %d were exported: %v", rowCount, err)
		}
		// Each file's footer is written as it closes: until then it cannot
		// be read.
		if err := writer.Close(); err != nil {
			return fmt.Sprintf("Error writing the partition files: %v", err)
		}
		return fmt.Sprintf("Exported %d rows to %d partitions in %s", rowCount, writer.PartitionCount(), cleanPath)

	case db.QueryResult:
		// Handle non-streaming result
		if len(v.Data) == 0 {
			return "No data to export"
		}

		// Clean headers
		cleanHeaders := make([]string, len(v.Headers))
		for i, header := range v.Headers {
			cleanHeaders[i] = db.StripKeyMarker(header)
		}

		// Create compression option
		compressionStr := strings.ToUpper(options["COMPRESSION"])
		compression := parquet.ParseCompression(compressionStr)

		// Create partitioned writer options
		writerOptions := parquet.PartitionedWriterOptions{
			WriterOptions: parquet.WriterOptions{
				ChunkSize:   10000,
				Compression: compression,
			},
			PartitionColumns: partitionColumns,
			MaxOpenFiles:     10,
			MaxFileSize:      100 * 1024 * 1024, // 100MB default
		}

		// Create partitioned writer
		writer, err := parquet.NewPartitionedParquetWriter(cleanPath, cleanHeaders, v.ColumnTypes, writerOptions)
		if err != nil {
			return fmt.Sprintf("Error creating partitioned writer: %v", err)
		}
		defer writer.Close()

		// Convert string data to maps
		batch := make([]map[string]interface{}, 0, len(v.Data))
		for _, row := range v.Data {
			rowData := make(map[string]interface{})
			for i, colName := range cleanHeaders {
				if i < len(row) {
					// Parse the string value to appropriate type
					rowData[colName] = parseStringValue(row[i], v.ColumnTypes[i])
				}
			}
			batch = append(batch, rowData)
		}

		// Write all data
		if err := writer.WriteRows(batch); err != nil {
			return fmt.Sprintf("Error writing data: %v", err)
		}

		if err := writer.Close(); err != nil {
			return fmt.Sprintf("Error writing the partition files: %v", err)
		}
		return fmt.Sprintf("Exported %d rows to %d partitions in %s", len(v.Data), writer.PartitionCount(), cleanPath)

	case error:
		return fmt.Sprintf("Query error: %v", v)

	default:
		return fmt.Sprintf("Unexpected result type: %T", result)
	}
}

// parseFileSize parses file size strings like "100MB", "1GB", etc.
func parseFileSize(sizeStr string) int64 {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	multiplier := int64(1)
	var numStr string

	switch {
	case strings.HasSuffix(sizeStr, "GB"):
		multiplier = 1024 * 1024 * 1024
		numStr = sizeStr[:len(sizeStr)-2]
	case strings.HasSuffix(sizeStr, "MB"):
		multiplier = 1024 * 1024
		numStr = sizeStr[:len(sizeStr)-2]
	case strings.HasSuffix(sizeStr, "KB"):
		multiplier = 1024
		numStr = sizeStr[:len(sizeStr)-2]
	default:
		numStr = sizeStr
	}

	var size int64
	if _, err := fmt.Sscanf(numStr, "%d", &size); err != nil {
		return 100 * 1024 * 1024 // Default 100MB on parse error
	}
	if size <= 0 {
		return 100 * 1024 * 1024 // Default 100MB
	}

	return size * multiplier
}

// parseStringValue attempts to parse a string value to the appropriate type
func parseStringValue(value string, columnType string) interface{} {
	// Handle NULL values
	if value == "null" || value == "<null>" || value == "" {
		return nil
	}

	// For now, return as string - the Parquet writer will handle conversion
	// This could be enhanced to parse specific types
	return value
}
