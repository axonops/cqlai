package router

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/parquet"
	"github.com/google/uuid"
)

// copyOptions holds parsed options for COPY operation
type copyOptions struct {
	batchSize       int
	maxRows         int
	skipRows        int
	maxInsertErrors int
}

// copyStats tracks statistics for the COPY operation
type copyStats struct {
	rowCount         int
	processedRows    int
	insertErrorCount int
	skippedRows      int
	errorMessages    []string // Store first few error messages for user display
}

// executeCopyFromParquet executes COPY FROM operation for Parquet format
func (h *MetaCommandHandler) executeCopyFromParquet(table string, columns []string, filename string, options map[string]string) interface{} {
	// Validate input
	if err := h.validateParquetInput(filename); err != nil {
		return err.Error()
	}

	cleanPath := filepath.Clean(filename)
	fileInfo, err := parquet.GetFileInfo(cleanPath)
	if err != nil {
		return fmt.Sprintf("Error accessing path: %v", err)
	}

	// Handle directory (partitioned dataset) vs single file
	if fileInfo.IsDir() {
		return h.handlePartitionedParquet(table, columns, cleanPath, options)
	}

	return h.handleSingleParquetFile(table, columns, cleanPath, options)
}

// validateParquetInput validates the input parameters for Parquet COPY
func (h *MetaCommandHandler) validateParquetInput(filename string) error {
	if strings.ToUpper(filename) == "STDIN" {
		return fmt.Errorf("COPY FROM STDIN is not supported for Parquet format. Please provide a file path")
	}
	return nil
}

// handlePartitionedParquet handles COPY from a partitioned Parquet dataset
func (h *MetaCommandHandler) handlePartitionedParquet(table string, columns []string, path string, options map[string]string) interface{} {
	reader, err := parquet.NewPartitionedParquetReader(path)
	if err != nil {
		return fmt.Sprintf("Error opening partitioned Parquet dataset: %v", err)
	}
	defer reader.Close()
	return h.executeCopyFromParquetPartitioned(table, columns, reader, options)
}

// handleSingleParquetFile handles COPY from a single Parquet file
func (h *MetaCommandHandler) handleSingleParquetFile(table string, columns []string, path string, options map[string]string) interface{} {
	reader, err := parquet.NewParquetReader(path)
	if err != nil {
		return fmt.Sprintf("Error opening Parquet file: %v", err)
	}
	defer reader.Close()

	// Prepare columns and validate
	processColumns, err := h.prepareColumns(columns, reader)
	if err != nil {
		return err.Error()
	}

	// Parse options
	opts := parseOptions(options)

	// Load destination table column types so list<...> vs set<...> decisions
	// come from the schema, not column-name heuristics.
	columnTypes := h.getTableColumnTypes(table)

	// Process the file
	stats := &copyStats{}
	if err := h.processParquetFile(table, processColumns, reader, opts, stats, columnTypes); err != nil {
		return err.Error()
	}

	// Log debug info
	h.logParquetDebugInfo(reader)

	return h.formatCopyResult(stats)
}

// prepareColumns prepares and validates the column list for import
func (h *MetaCommandHandler) prepareColumns(columns []string, reader *parquet.ParquetReader) ([]string, error) {
	parquetColumns, _ := reader.GetSchema()

	// If no columns specified, use all columns from the Parquet file
	if len(columns) == 0 {
		return parquetColumns, nil
	}

	// Validate that specified columns exist in the Parquet file
	columnMap := make(map[string]bool)
	for _, col := range parquetColumns {
		columnMap[col] = true
	}

	for _, col := range columns {
		if !columnMap[col] {
			return nil, fmt.Errorf("column '%s' not found in Parquet file", col)
		}
	}

	return columns, nil
}

// parseOptions parses the COPY options into a structured format
func parseOptions(options map[string]string) *copyOptions {
	opts := &copyOptions{
		batchSize: 1000, // Default
	}

	if val, _ := strconv.Atoi(options["CHUNKSIZE"]); val > 0 {
		opts.batchSize = val
	}
	opts.maxRows, _ = strconv.Atoi(options["MAXROWS"])
	opts.skipRows, _ = strconv.Atoi(options["SKIPROWS"])
	opts.maxInsertErrors, _ = strconv.Atoi(options["MAXINSERTERRORS"])

	return opts
}

// processParquetFile processes the Parquet file and inserts data
func (h *MetaCommandHandler) processParquetFile(table string, columns []string, reader *parquet.ParquetReader, opts *copyOptions, stats *copyStats, columnTypes map[string]string) error {
	// Skip initial rows if specified
	if opts.skipRows > 0 {
		if err := h.skipRows(reader, opts, stats); err != nil {
			return err
		}
	}

	// Process data in batches
	for opts.maxRows <= 0 || stats.processedRows < opts.maxRows {
		batch, err := reader.ReadBatch(opts.batchSize)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading Parquet batch: %v", err)
		}

		logger.DebugfToFile("CopyFromParquet", "Read batch of %d rows", len(batch))

		// Process each row in the batch
		for _, row := range batch {
			if opts.maxRows > 0 && stats.processedRows >= opts.maxRows {
				break
			}

			stats.processedRows++

			if err := h.insertRow(table, columns, row, opts, stats, columnTypes); err != nil {
				// Error already handled in insertRow
				if opts.maxInsertErrors > 0 && stats.insertErrorCount >= opts.maxInsertErrors {
					return fmt.Errorf("aborted after %d insert errors. Successfully imported %d rows",
						stats.insertErrorCount, stats.rowCount)
				}
			}
		}
	}

	return nil
}

// skipRows skips the specified number of rows
func (h *MetaCommandHandler) skipRows(reader *parquet.ParquetReader, opts *copyOptions, stats *copyStats) error {
	for i := 0; i < opts.skipRows; i += opts.batchSize {
		batch, err := reader.ReadBatch(opts.batchSize)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading Parquet file: %v", err)
		}

		toSkip := opts.skipRows - i
		if toSkip > len(batch) {
			stats.skippedRows += len(batch)
		} else {
			stats.skippedRows += toSkip
			break
		}
	}
	return nil
}

// insertRow inserts a single row into Cassandra
func (h *MetaCommandHandler) insertRow(table string, columns []string, row map[string]interface{}, opts *copyOptions, stats *copyStats, columnTypes map[string]string) error {
	// Build values array for the INSERT
	values := h.extractRowValues(columns, row)

	// Bind the values rather than formatting them into the statement. Data in a
	// Parquet file is not trusted input: pasted into CQL text it can alter the
	// statement, and it round-trips badly for blobs, floats and timestamps.
	query := buildInsertTemplate(table, columns)
	bound := bindValuesForInsert(columns, values, columnTypes)

	logger.DebugfToFile("CopyFromParquet", "INSERT query: %s", query)

	// Execute the query
	result := h.session.ExecuteCQLQueryWithValues(query, bound...)

	// Check for errors
	if err, isError := result.(error); isError {
		stats.insertErrorCount++
		logger.DebugfToFile("CopyFromParquet", "Insert error: %v", err)
		logger.DebugfToFile("CopyFromParquet", "Failed query: %s", query)

		// Store first few error messages for user display (limit to 5)
		if len(stats.errorMessages) < 5 {
			stats.errorMessages = append(stats.errorMessages, fmt.Sprintf("Row %d: %v", stats.processedRows, err))
		}

		return err
	}

	stats.rowCount++
	return nil
}

// extractRowValues extracts values from row for specified columns
func (h *MetaCommandHandler) extractRowValues(columns []string, row map[string]interface{}) []interface{} {
	values := make([]interface{}, len(columns))
	for i, colName := range columns {
		if val, ok := row[colName]; ok {
			values[i] = val
			logger.DebugfToFile("CopyFromParquet", "Column %s: value=%v, type=%T", colName, val, val)
		} else {
			values[i] = nil
		}
	}
	return values
}

// isUUIDFormat checks if a string is a valid UUID using the google/uuid library
func isUUIDFormat(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// formatListValue formats a list/array value
func (h *MetaCommandHandler) formatListValue(v []interface{}, columnName string, columnTypes map[string]string) string {
	if len(v) == 0 {
		return h.getEmptyCollectionSyntax(columnName, columnTypes)
	}

	quotedParts := make([]string, len(v))
	for i, item := range v {
		quotedParts[i] = h.formatListItem(item)
	}

	// Use curly braces for sets, square brackets for lists
	if h.isSetColumn(columnName, columnTypes) {
		return "{" + strings.Join(quotedParts, ", ") + "}"
	}
	return "[" + strings.Join(quotedParts, ", ") + "]"
}

// formatListItem formats a single item in a list
func (h *MetaCommandHandler) formatListItem(item interface{}) string {
	switch it := item.(type) {
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(it, "'", "''"))
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", it)
	default:
		return fmt.Sprintf("'%v'", it)
	}
}

// Helper functions

// isSetColumn determines if a column is a CQL set type by looking up the
// destination table's schema. columnTypes maps column name to its CQL type
// (e.g. "list<text>", "set<int>", "map<text,text>"); when the column is
// missing from the map (e.g. schema lookup failed or the table is unknown
// during unit tests), the function returns false so the formatter falls back
// to list syntax — matching the most common collection shape.
func (h *MetaCommandHandler) isSetColumn(columnName string, columnTypes map[string]string) bool {
	t, ok := columnTypes[columnName]
	if !ok {
		return false
	}
	return strings.HasPrefix(t, "set<") || t == "set"
}

// getEmptyCollectionSyntax returns the appropriate empty collection syntax
// for the given column, based on the destination table schema.
func (h *MetaCommandHandler) getEmptyCollectionSyntax(columnName string, columnTypes map[string]string) string {
	if h.isSetColumn(columnName, columnTypes) {
		return "{}"
	}
	return "[]"
}

// formatCopyResult formats the final result message
func (h *MetaCommandHandler) formatCopyResult(stats *copyStats) string {
	summary := fmt.Sprintf("Imported %d rows from Parquet file", stats.rowCount)

	if stats.skippedRows > 0 {
		summary += fmt.Sprintf(" (skipped %d rows)", stats.skippedRows)
	}

	if stats.insertErrorCount > 0 {
		summary += fmt.Sprintf(" with %d errors", stats.insertErrorCount)

		// Show first few error messages to help user diagnose issues
		if len(stats.errorMessages) > 0 {
			summary += "\n\nFirst errors encountered:"
			for _, errMsg := range stats.errorMessages {
				summary += "\n  - " + errMsg
			}
			if stats.insertErrorCount > len(stats.errorMessages) {
				summary += fmt.Sprintf("\n  ... and %d more errors", stats.insertErrorCount-len(stats.errorMessages))
			}
		}
	}

	return summary
}

// logParquetDebugInfo logs debug information about the Parquet file
func (h *MetaCommandHandler) logParquetDebugInfo(reader *parquet.ParquetReader) {
	parquetColumns, parquetTypes := reader.GetSchema()
	logger.DebugfToFile("CopyFromParquet", "Parquet columns: %v", parquetColumns)
	logger.DebugfToFile("CopyFromParquet", "Parquet types: %v", parquetTypes)
	logger.DebugfToFile("CopyFromParquet", "Total rows in file: %d", reader.GetRowCount())
}
