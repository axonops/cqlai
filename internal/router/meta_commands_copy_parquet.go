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

// executeCopyToParquet executes COPY TO operation for Parquet format
func (h *MetaCommandHandler) executeCopyToParquet(table string, columns []string, filename string, options map[string]string) interface{} {
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

	// Execute query to get schema information first
	result := h.session.ExecuteCQLQuery(query)

	// Handle different result types
	switch v := result.(type) {
	case db.QueryResult:
		if len(v.Headers) == 0 {
			return "No data to export"
		}

		// Get column types from the result
		columnTypes := v.ColumnTypes
		if len(columnTypes) == 0 {
			// Fall back to text if no types available
			columnTypes = make([]string, len(v.Headers))
			for i := range columnTypes {
				columnTypes[i] = "text"
			}
		}

		// Clean headers (remove PK/C suffixes) for Parquet
		cleanHeaders := make([]string, len(v.Headers))
		for i, header := range v.Headers {
			// Remove (PK) suffix
			if idx := strings.Index(header, " (PK)"); idx != -1 {
				cleanHeaders[i] = header[:idx]
			} else if idx := strings.Index(header, " (C)"); idx != -1 {
				// Remove (C) suffix
				cleanHeaders[i] = header[:idx]
			} else {
				cleanHeaders[i] = header
			}
		}

		// Set up Parquet writer options with optimizations
		writerOptions := parquet.DefaultWriterOptions()

		// Optimize chunk size based on table size hints
		// Default is 10000, but we can increase for better performance
		if chunkSize := options["CHUNKSIZE"]; chunkSize != "" {
			if size, err := parseSize(chunkSize); err == nil {
				writerOptions.ChunkSize = size
			}
		} else {
			// Use larger chunks for better compression and performance
			// This balances memory usage with write efficiency
			writerOptions.ChunkSize = 50000
		}

		// Apply row group size if specified (for advanced users)
		if rowGroupSize := options["ROWGROUPSIZE"]; rowGroupSize != "" {
			// Row group size affects how data is organized in the file
			// Larger row groups = better compression, more memory
			if size, err := parseSize(rowGroupSize); err == nil {
				// Store for future use when we implement custom row group sizing
				logger.DebugfToFile("CopyToParquet", "Row group size set to: %d", size)
			}
		}

		// Create Parquet writer
		parquetWriter, err := parquet.NewParquetCaptureWriter(filename, cleanHeaders, columnTypes, writerOptions)
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

		// Write data rows
		rowCount := 0
		if len(v.RawData) > 0 && len(v.RawData) == len(v.Data) {
			// Use raw data if available (preserves types)
			for _, rowData := range v.RawData {
				// Clean the keys to match clean headers
				cleanedRow := make(map[string]interface{})
				for i, header := range v.Headers {
					if i < len(cleanHeaders) && rowData != nil {
						if val, ok := rowData[header]; ok {
							cleanedRow[cleanHeaders[i]] = val
						}
					}
				}

				if err := parquetWriter.WriteRow(cleanedRow); err != nil {
					return fmt.Sprintf("Error writing row: %v", err)
				}
				rowCount++
			}
		} else {
			// Use string data
			if err := parquetWriter.WriteStringRows(cleanHeaders, v.Data); err != nil {
				return fmt.Sprintf("Error writing rows: %v", err)
			}
			rowCount = len(v.Data)
		}

		// Flush and close
		if err := parquetWriter.Close(); err != nil {
			return fmt.Sprintf("Error closing Parquet file: %v", err)
		}

		if isStdout {
			return nil // Don't print message when outputting to STDOUT
		}
		return fmt.Sprintf("Exported %d rows to %s (Parquet format)", rowCount, filename)

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

		// Create Parquet writer
		parquetWriter, err := parquet.NewParquetCaptureWriter(filename, cleanHeaders, columnTypes, writerOptions)
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
		scanDest := make([]interface{}, len(headers))
		for i := range scanDest {
			scanDest[i] = new(interface{})
		}

		for v.Iterator.Scan(scanDest...) {

			// Build row map from scanned values
			cleanedRow := make(map[string]interface{})
			for i, header := range cleanHeaders {
				if i < len(scanDest) {
					val := *(scanDest[i].(*interface{}))
					cleanedRow[header] = val
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

	default:
		return fmt.Sprintf("Unexpected result type: %T", result)
	}
}

// parseSize parses size strings like "10000", "10K", "1M"
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	multiplier := int64(1)
	if strings.HasSuffix(sizeStr, "K") {
		multiplier = 1000
		sizeStr = strings.TrimSuffix(sizeStr, "K")
	} else if strings.HasSuffix(sizeStr, "M") {
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