package router

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/parquet"
)

// executeCopyFromParquet executes COPY FROM operation for Parquet format
func (h *MetaCommandHandler) executeCopyFromParquet(table string, columns []string, filename string, options map[string]string) interface{} {
	// Check if input is STDIN (not supported for Parquet)
	isStdin := strings.ToUpper(filename) == "STDIN"
	if isStdin {
		return "COPY FROM STDIN is not supported for Parquet format. Please provide a file path."
	}

	// Clean the filename to prevent path traversal
	cleanPath := filepath.Clean(filename)

	// Open the Parquet file
	reader, err := parquet.NewParquetReader(cleanPath)
	if err != nil {
		return fmt.Sprintf("Error opening Parquet file: %v", err)
	}
	defer reader.Close()

	// Get schema from Parquet file
	parquetColumns, parquetTypes := reader.GetSchema()

	// If no columns specified, use all columns from the Parquet file
	if len(columns) == 0 {
		columns = parquetColumns
	} else {
		// Validate that specified columns exist in the Parquet file
		columnMap := make(map[string]bool)
		for _, col := range parquetColumns {
			columnMap[col] = true
		}

		for _, col := range columns {
			if !columnMap[col] {
				return fmt.Sprintf("Column '%s' not found in Parquet file", col)
			}
		}
	}

	// Build column list for INSERT
	columnList := strings.Join(columns, ", ")
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	placeholderList := strings.Join(placeholders, ", ")

	// Prepare INSERT statement template (for reference/logging)
	_ = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, columnList, placeholderList)

	// Parse numeric options
	batchSize, _ := strconv.Atoi(options["CHUNKSIZE"])
	if batchSize <= 0 {
		batchSize = 1000 // Default batch size
	}

	maxRows, _ := strconv.Atoi(options["MAXROWS"])
	skipRows, _ := strconv.Atoi(options["SKIPROWS"])
	maxInsertErrors, _ := strconv.Atoi(options["MAXINSERTERRORS"])

	// Process rows
	rowCount := 0
	processedRows := 0
	insertErrorCount := 0
	skippedRows := 0

	// Skip initial rows if specified
	if skipRows > 0 {
		// Read and discard the skip rows
		for i := 0; i < skipRows; i += batchSize {
			batch, err := reader.ReadBatch(batchSize)
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Sprintf("Error reading Parquet file: %v", err)
			}

			toSkip := skipRows - i
			if toSkip > len(batch) {
				skippedRows += len(batch)
			} else {
				skippedRows += toSkip
				break
			}
		}
	}

	// Process data in batches
	for {
		// Check if we've reached the max rows limit
		if maxRows > 0 && processedRows >= maxRows {
			break
		}

		// Read a batch of rows
		batch, err := reader.ReadBatch(batchSize)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Sprintf("Error reading Parquet batch: %v", err)
		}

		logger.DebugfToFile("CopyFromParquet", "Read batch of %d rows", len(batch))

		// Process each row in the batch
		for _, row := range batch {
			processedRows++

			// Check max rows limit
			if maxRows > 0 && processedRows > maxRows {
				break
			}

			// Build values array for the INSERT
			values := make([]any, len(columns))
			for i, colName := range columns {
				if val, ok := row[colName]; ok {
					values[i] = val
				} else {
					values[i] = nil
				}
			}

			// Execute the INSERT
			// For now, we'll build a string representation
			// In a real implementation, this would use prepared statements
			valueStrings := make([]string, len(values))
			for i, val := range values {
				if val == nil {
					valueStrings[i] = "null"
				} else {
					// Format value based on type
					switch v := val.(type) {
					case string:
						valueStrings[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
					case bool:
						valueStrings[i] = fmt.Sprintf("%t", v)
					case []byte:
						valueStrings[i] = fmt.Sprintf("0x%x", v)
					default:
						valueStrings[i] = fmt.Sprintf("%v", v)
					}
				}
			}

			// Build and execute the actual INSERT query
			actualQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				table, columnList, strings.Join(valueStrings, ", "))

			// Execute the query
			result := h.session.ExecuteCQLQuery(actualQuery)

			// Check for errors
			if _, isError := result.(error); isError {
				insertErrorCount++
				logger.DebugfToFile("CopyFromParquet", "Insert error: %v", result)

				// Check if we've exceeded max insert errors
				if maxInsertErrors > 0 && insertErrorCount >= maxInsertErrors {
					return fmt.Sprintf("Aborted after %d insert errors. Successfully imported %d rows.",
						insertErrorCount, rowCount)
				}
			} else {
				rowCount++
			}
		}
	}

	// Return summary
	summary := fmt.Sprintf("Imported %d rows from Parquet file", rowCount)

	if skippedRows > 0 {
		summary += fmt.Sprintf(" (skipped %d rows)", skippedRows)
	}

	if insertErrorCount > 0 {
		summary += fmt.Sprintf(" with %d errors", insertErrorCount)
	}

	// Log type mappings for debugging
	logger.DebugfToFile("CopyFromParquet", "Parquet columns: %v", parquetColumns)
	logger.DebugfToFile("CopyFromParquet", "Parquet types: %v", parquetTypes)
	logger.DebugfToFile("CopyFromParquet", "Total rows in file: %d", reader.GetRowCount())

	return summary
}