package router

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/parquet"
)

// isNullValue checks if a value from gocql scanning represents a NULL value.
// LIMITATION: When scanning into interface{}, gocql doesn't preserve NULL information.
// It returns zero values (empty string, 0, empty slice, etc.) for NULL columns.
// There's no way to distinguish between a NULL and an actual zero value when using interface{}.
// This would require scanning into typed pointers instead of interface{}.
func isNullValue(val interface{}) bool {
	// Only return true for actual nil values
	// We cannot reliably detect other NULL values when scanning into interface{}
	return val == nil
}

// executeCopyToParquet executes COPY TO operation for Parquet format
func (h *MetaCommandHandler) executeCopyToParquet(table string, columns []string, filename string, options map[string]string) interface{} {
	logger.DebugfToFile("CopyToParquet", "Starting Parquet export for table: %s, filename: %s", table, filename)

	// Check if partitioning is requested
	partitionColumns := options["PARTITION"]
	if partitionColumns != "" {
		// Use partitioned writer for directory output
		return h.executeCopyToParquetPartitioned(table, columns, filename, options)
	}

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

	// Optimize query execution with page size for better streaming
	// Larger page sizes reduce round trips to Cassandra
	if pageSize := options["PAGESIZE"]; pageSize == "" {
		// Use a larger default page size for Parquet exports
		// This improves throughput significantly
		options["PAGESIZE"] = "5000"
	}

	// Check if output is STDOUT
	isStdout := strings.ToUpper(filename) == "STDOUT"

	// Set up file path
	var file *os.File
	var err error

	if isStdout {
		filename = "-"
	} else {
		// Clean the filename to prevent path traversal
		cleanPath := filepath.Clean(filename)

		// Add .parquet extension if not present
		if !strings.HasSuffix(cleanPath, ".parquet") {
			cleanPath += ".parquet"
		}

		file, err = os.Create(cleanPath) // #nosec G304 - file path is user input but cleaned
		if err != nil {
			return fmt.Sprintf("Error creating file: %v", err)
		}
		defer file.Close()
		filename = cleanPath
	}

	// Execute query - use streaming for better UDT handling
	// Streaming mode properly scans UDT data into maps instead of strings
	result := h.session.ExecuteStreamingQuery(query)

	// Handle streaming result (always use streaming for UDT support)
	switch v := result.(type) {
	case db.StreamingQueryResult:
		// For streaming results, we need to iterate through the data
		defer v.Iterator.Close()

		// Get headers and column types from the streaming result
		if len(v.Headers) == 0 {
			return "No columns found in result"
		}

		headers := v.Headers
		columnTypes := v.ColumnTypes
		cleanHeaders := make([]string, len(headers))

		for i, header := range headers {
			// Clean header for Parquet
			if idx := strings.Index(header, " (PK)"); idx != -1 {
				cleanHeaders[i] = header[:idx]
			} else if idx := strings.Index(header, " (C)"); idx != -1 {
				cleanHeaders[i] = header[:idx]
			} else {
				cleanHeaders[i] = header
			}
		}

		// Set up Parquet writer options with optimizations
		writerOptions := parquet.DefaultWriterOptions()

		// Apply same optimizations as non-streaming case
		if chunkSize := options["CHUNKSIZE"]; chunkSize != "" {
			if size, err := parseSize(chunkSize); err == nil {
				writerOptions.ChunkSize = size
			}
		} else {
			// Use larger chunks for streaming data
			writerOptions.ChunkSize = 50000
		}

		// Create Parquet writer - use streaming result's TypeInfo if available
		var parquetWriter *parquet.ParquetCaptureWriter
		if len(v.ColumnTypeInfos) > 0 {
			// Use the new writer with TypeInfo for proper UDT support
			parquetWriter, err = parquet.NewParquetCaptureWriterWithTypeInfo(filename, cleanHeaders, columnTypes, v.ColumnTypeInfos, writerOptions)
		} else {
			// Fall back to standard writer
			parquetWriter, err = parquet.NewParquetCaptureWriter(filename, cleanHeaders, columnTypes, writerOptions)
		}
		if err != nil {
			return fmt.Sprintf("Error creating Parquet writer: %v", err)
		}
		defer parquetWriter.Close()

		// Set compression if specified
		if compression := options["COMPRESSION"]; compression != "" {
			if err := parquetWriter.SetCompression(strings.ToLower(compression)); err != nil {
				logger.DebugfToFile("CopyToParquet", "Failed to set compression %s: %v", compression, err)
			}
		}

		// Stream and write data
		rowCount := 0

		// Create scan destinations for each column
		// Use cleanHeaders since that's what we'll iterate with
		// Get column info from iterator to detect UDT columns
		columns := v.Iterator.Columns()
		scanDest := make([]interface{}, len(cleanHeaders))
		for i := range scanDest {
			// Check if this column is a UDT
			if i < len(columns) && columns[i].TypeInfo != nil &&
				columns[i].TypeInfo.Type() == gocql.TypeUDT {
				// Use map[string]interface{} for UDT columns to get populated data
				scanDest[i] = new(map[string]interface{})
			} else {
				scanDest[i] = new(interface{})
			}
		}

		for v.Iterator.Scan(scanDest...) {

			// Build row map from scanned values
			cleanedRow := make(map[string]interface{})
			for i := range cleanHeaders {
				if i < len(scanDest) {
					var val interface{}

					// Extract value based on how it was scanned
					if i < len(columns) && columns[i].TypeInfo != nil &&
						columns[i].TypeInfo.Type() == gocql.TypeUDT {
						// For UDT columns, we scanned into *map[string]interface{}
						udtMap := scanDest[i].(*map[string]interface{})
						if udtMap != nil && *udtMap != nil {
							val = *udtMap
						} else {
							val = nil
						}
					} else {
						// Regular column
						val = *(scanDest[i].(*interface{}))
					}

					// Check if the value is a "zero" value that should be NULL
					// gocql returns typed zero values for NULL columns
					if isNullValue(val) {
						cleanedRow[cleanHeaders[i]] = nil
					} else {
						cleanedRow[cleanHeaders[i]] = val
					}
				}
			}

			if err := parquetWriter.WriteRow(cleanedRow); err != nil {
				return fmt.Sprintf("Error writing row: %v", err)
			}
			rowCount++
		}

		if err := v.Iterator.Close(); err != nil {
			return fmt.Sprintf("Error during query execution: %v", err)
		}

		// Flush and close
		if err := parquetWriter.Close(); err != nil {
			return fmt.Sprintf("Error closing Parquet file: %v", err)
		}

		if isStdout {
			return nil
		}
		return fmt.Sprintf("Exported %d rows to %s (Parquet format)", rowCount, filename)

	case db.QueryResult:
		// Handle non-streaming query results (primarily for testing)
		if len(v.Data) == 0 {
			return "No data to export"
		}

		headers := v.Headers
		cleanHeaders := make([]string, len(headers))
		for i, header := range headers {
			// Clean header for Parquet
			if idx := strings.Index(header, " (PK)"); idx != -1 {
				cleanHeaders[i] = header[:idx]
			} else if idx := strings.Index(header, " (C)"); idx != -1 {
				cleanHeaders[i] = header[:idx]
			} else {
				cleanHeaders[i] = header
			}
		}

		// Set up Parquet writer options
		writerOptions := parquet.DefaultWriterOptions()
		if chunkSize := options["CHUNKSIZE"]; chunkSize != "" {
			if size, err := parseSize(chunkSize); err == nil {
				writerOptions.ChunkSize = size
			}
		}

		// Create Parquet writer
		parquetWriter, err := parquet.NewParquetCaptureWriter(filename, cleanHeaders, v.ColumnTypes, writerOptions)
		if err != nil {
			return fmt.Sprintf("Error creating Parquet writer: %v", err)
		}
		defer parquetWriter.Close()

		// Set compression if specified
		if compression := options["COMPRESSION"]; compression != "" {
			if err := parquetWriter.SetCompression(strings.ToLower(compression)); err != nil {
				logger.DebugfToFile("CopyToParquet", "Failed to set compression %s: %v", compression, err)
			}
		}

		// Write data
		if v.RawData != nil {
			// Use raw data if available
			if err := parquetWriter.WriteRows(v.RawData); err != nil {
				return fmt.Sprintf("Error writing data: %v", err)
			}
		} else {
			// Fall back to string data
			if err := parquetWriter.WriteStringRows(cleanHeaders, v.Data); err != nil {
				return fmt.Sprintf("Error writing data: %v", err)
			}
		}

		// Close the writer to flush data
		if err := parquetWriter.Close(); err != nil {
			return fmt.Sprintf("Error closing Parquet file: %v", err)
		}

		rowCount := len(v.Data)
		if isStdout {
			return nil
		}
		return fmt.Sprintf("Exported %d rows to %s (Parquet format)", rowCount, filename)

	default:
		return fmt.Sprintf("Unexpected result type: %T", result)
	}
}

// parseSize parses size strings like "10000", "10K", "1M"
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	multiplier := int64(1)
	switch {
	case strings.HasSuffix(sizeStr, "K"):
		multiplier = 1000
		sizeStr = strings.TrimSuffix(sizeStr, "K")
	case strings.HasSuffix(sizeStr, "M"):
		multiplier = 1000000
		sizeStr = strings.TrimSuffix(sizeStr, "M")
	}

	var base int64
	_, err := fmt.Sscanf(sizeStr, "%d", &base)
	if err != nil {
		return 0, err
	}

	return base * multiplier, nil
}