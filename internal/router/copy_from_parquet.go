package router

import (
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

	// Process the file
	stats := &copyStats{}
	if err := h.processParquetFile(table, processColumns, reader, opts, stats); err != nil {
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
func (h *MetaCommandHandler) processParquetFile(table string, columns []string, reader *parquet.ParquetReader, opts *copyOptions, stats *copyStats) error {
	// Skip initial rows if specified
	if opts.skipRows > 0 {
		if err := h.skipRows(reader, opts, stats); err != nil {
			return err
		}
	}

	// Get schema for type information
	_, parquetTypes := reader.GetSchema()

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

			if err := h.insertRow(table, columns, row, parquetTypes, opts, stats); err != nil {
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
func (h *MetaCommandHandler) insertRow(table string, columns []string, row map[string]interface{}, parquetTypes []string, opts *copyOptions, stats *copyStats) error {
	// Build values array for the INSERT
	values := h.extractRowValues(columns, row)

	// Format values for CQL
	valueStrings := h.formatValuesForCQL(values, columns, parquetTypes)

	// Build and execute the INSERT query
	query := h.buildInsertQuery(table, columns, valueStrings)

	logger.DebugfToFile("CopyFromParquet", "INSERT query: %s", query)

	// Execute the query
	result := h.session.ExecuteCQLQuery(query)

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

// formatValuesForCQL formats values for use in CQL INSERT statement
func (h *MetaCommandHandler) formatValuesForCQL(values []interface{}, columns []string, parquetTypes []string) []string {
	valueStrings := make([]string, len(values))
	for i, val := range values {
		valueStrings[i] = h.formatValue(val, columns[i], getColumnType(i, parquetTypes))
	}
	return valueStrings
}

// getColumnType safely gets column type from parquetTypes array
func getColumnType(index int, parquetTypes []string) string {
	if index < len(parquetTypes) {
		return parquetTypes[index]
	}
	return ""
}

// formatValue formats a single value for CQL based on its type
func (h *MetaCommandHandler) formatValue(val interface{}, columnName string, parquetType string) string {
	if val == nil {
		return "null"
	}

	switch v := val.(type) {
	case string:
		return h.formatStringValue(v, columnName, parquetType)
	case time.Time:
		return fmt.Sprintf("'%s'", v.Format(time.RFC3339Nano))
	case bool:
		return fmt.Sprintf("%t", v)
	case []byte:
		return fmt.Sprintf("0x%x", v)
	case []interface{}:
		return h.formatListValue(v, columnName)
	case map[string]interface{}:
		return h.formatUDTValue(v)
	case map[interface{}]interface{}:
		return h.formatMapValue(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatStringValue handles formatting of string values
func (h *MetaCommandHandler) formatStringValue(value string, columnName string, parquetType string) string {
	trimmed := strings.TrimSpace(value)

	// Check if this is a UUID
	if h.isUUIDColumn(columnName, parquetType) && isUUIDFormat(trimmed) {
		return trimmed // No quotes for UUIDs
	}

	// Check for collections and special formats
	if formatted, isSpecial := h.formatSpecialStringValue(trimmed, columnName); isSpecial {
		return formatted
	}

	// Regular string value
	return fmt.Sprintf("'%s'", strings.ReplaceAll(trimmed, "'", "''"))
}

// isUUIDColumn checks if a column should be treated as UUID
func (h *MetaCommandHandler) isUUIDColumn(columnName string, parquetType string) bool {
	lowerType := strings.ToLower(parquetType)
	lowerCol := strings.ToLower(columnName)

	return strings.Contains(lowerType, "uuid") ||
		strings.Contains(lowerCol, "uuid") ||
		strings.Contains(lowerCol, "id")
}

// isUUIDFormat checks if a string is a valid UUID using the google/uuid library
func isUUIDFormat(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// formatSpecialStringValue handles special string formats (collections, UDTs, etc)
func (h *MetaCommandHandler) formatSpecialStringValue(value string, columnName string) (string, bool) {
	switch {
	case strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]"):
		return h.formatListString(value, columnName), true
	case strings.HasPrefix(value, "map[") && strings.HasSuffix(value, "]"):
		return h.formatMapString(value), true
	case strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") &&
		strings.Contains(value, ":") && strings.Contains(value, "\""):
		return h.formatJSONUDTString(value), true
	default:
		return "", false
	}
}

// formatListString formats a list/set string value
func (h *MetaCommandHandler) formatListString(value string, columnName string) string {
	inner := strings.Trim(value, "[]")
	if inner == "" {
		return h.getEmptyCollectionSyntax(columnName)
	}

	parts := strings.Fields(inner)
	quotedParts := make([]string, len(parts))
	for i, part := range parts {
		quotedParts[i] = h.quoteValueIfNeeded(part)
	}

	if h.isSetColumn(columnName) {
		return "{" + strings.Join(quotedParts, ", ") + "}"
	}
	return "[" + strings.Join(quotedParts, ", ") + "]"
}

// formatMapString formats a map string value
func (h *MetaCommandHandler) formatMapString(value string) string {
	inner := strings.TrimPrefix(value, "map[")
	inner = strings.TrimSuffix(inner, "]")

	if inner == "" {
		return "{}"
	}

	pairs := strings.Fields(inner)
	mapPairs := make([]string, 0, len(pairs))

	for _, pair := range pairs {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			key := fmt.Sprintf("'%s'", strings.ReplaceAll(kv[0], "'", "''"))
			val := h.quoteValueIfNeeded(kv[1])
			mapPairs = append(mapPairs, fmt.Sprintf("%s: %s", key, val))
		}
	}

	return "{" + strings.Join(mapPairs, ", ") + "}"
}

// formatJSONUDTString formats a JSON-like UDT string
func (h *MetaCommandHandler) formatJSONUDTString(value string) string {
	udtValue := value
	// Remove quotes from field names
	udtValue = strings.ReplaceAll(udtValue, "\"street\":", "street:")
	udtValue = strings.ReplaceAll(udtValue, "\"city\":", "city:")
	udtValue = strings.ReplaceAll(udtValue, "\"zip\":", "zip:")
	// Replace double quotes with single quotes for string values
	udtValue = strings.ReplaceAll(udtValue, "\"", "'")
	return udtValue
}

// formatListValue formats a list/array value
func (h *MetaCommandHandler) formatListValue(v []interface{}, columnName string) string {
	if len(v) == 0 {
		return "[]"
	}

	quotedParts := make([]string, len(v))
	for i, item := range v {
		quotedParts[i] = h.formatListItem(item)
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

// formatUDTValue formats a UDT (User-Defined Type) value
func (h *MetaCommandHandler) formatUDTValue(v map[string]interface{}) string {
	if len(v) == 0 {
		return "{}"
	}

	udtPairs := make([]string, 0, len(v))
	for fieldName, fieldValue := range v {
		formattedValue := h.formatUDTField(fieldValue)
		udtPairs = append(udtPairs, fmt.Sprintf("%s: %s", fieldName, formattedValue))
	}

	return "{" + strings.Join(udtPairs, ", ") + "}"
}

// formatUDTField formats a single field in a UDT
func (h *MetaCommandHandler) formatUDTField(fieldValue interface{}) string {
	if fieldValue == nil {
		return "null"
	}

	switch fv := fieldValue.(type) {
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(fv, "'", "''"))
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", fv)
	case float32, float64:
		return fmt.Sprintf("%v", fv)
	case bool:
		return fmt.Sprintf("%t", fv)
	default:
		return fmt.Sprintf("'%v'", fv)
	}
}

// formatMapValue formats a map value
func (h *MetaCommandHandler) formatMapValue(v map[interface{}]interface{}) string {
	if len(v) == 0 {
		return "{}"
	}

	mapPairs := make([]string, 0, len(v))
	for key, val := range v {
		quotedKey := h.formatMapKey(key)
		quotedVal := h.formatMapValue2(val)
		mapPairs = append(mapPairs, fmt.Sprintf("%s: %s", quotedKey, quotedVal))
	}

	return "{" + strings.Join(mapPairs, ", ") + "}"
}

// formatMapKey formats a map key
func (h *MetaCommandHandler) formatMapKey(key interface{}) string {
	switch k := key.(type) {
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(k, "'", "''"))
	default:
		return fmt.Sprintf("'%v'", k)
	}
}

// formatMapValue2 formats a map value (named to avoid conflict with main function)
func (h *MetaCommandHandler) formatMapValue2(val interface{}) string {
	switch vt := val.(type) {
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(vt, "'", "''"))
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", vt)
	default:
		return fmt.Sprintf("'%v'", vt)
	}
}

// Helper functions

// isSetColumn determines if a column is a set type based on naming conventions
func (h *MetaCommandHandler) isSetColumn(columnName string) bool {
	return strings.Contains(columnName, "unique") ||
		strings.HasSuffix(columnName, "_set") ||
		strings.HasSuffix(columnName, "_nums")
}

// getEmptyCollectionSyntax returns the appropriate empty collection syntax
func (h *MetaCommandHandler) getEmptyCollectionSyntax(columnName string) string {
	if h.isSetColumn(columnName) {
		return "{}"
	}
	return "[]"
}

// quoteValueIfNeeded quotes a value if it's not a number
func (h *MetaCommandHandler) quoteValueIfNeeded(value string) string {
	// Check if it's a number
	if _, err := strconv.Atoi(value); err == nil {
		return value
	}
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return value
	}
	return fmt.Sprintf("'%s'", strings.ReplaceAll(value, "'", "''"))
}

// buildInsertQuery builds the INSERT query
func (h *MetaCommandHandler) buildInsertQuery(table string, columns []string, valueStrings []string) string {
	fullyQualifiedTable := h.getFullyQualifiedTableName(table)
	columnList := strings.Join(columns, ", ")
	valueList := strings.Join(valueStrings, ", ")

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		fullyQualifiedTable, columnList, valueList)
}

// getFullyQualifiedTableName returns the fully qualified table name
func (h *MetaCommandHandler) getFullyQualifiedTableName(table string) string {
	if h.session.Keyspace() != "" && !strings.Contains(table, ".") {
		return fmt.Sprintf("%s.%s", h.session.Keyspace(), table)
	}
	return table
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