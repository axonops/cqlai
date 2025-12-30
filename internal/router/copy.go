package router

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/axonops/cqlai/internal/db"
	"github.com/axonops/cqlai/internal/logger"
	"github.com/axonops/cqlai/internal/parquet"
)

// nopCloser wraps an io.Writer to add a no-op Close method
type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }

// formatCSVValue formats a value for CSV export, handling complex types like UDTs
func formatCSVValue(val interface{}) string {
	switch v := val.(type) {
	case []byte:
		// Format as hex string with 0x prefix (standard for BLOBs)
		return fmt.Sprintf("0x%x", v)
	case time.Time:
		// Use RFC3339 format for timestamps
		return v.Format(time.RFC3339)
	case gocql.UUID:
		return v.String()
	case map[string]interface{}, []interface{}, map[interface{}]interface{}:
		// For UDTs, collections, and other complex types, use JSON encoding
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(jsonBytes)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// handleCopy handles COPY TO/FROM commands
func (h *MetaCommandHandler) handleCopy(command string) interface{} {
	// Parse the COPY command
	// Format: COPY table [(col1, col2, ...)] TO 'filename' [WITH options]

	upperCommand := strings.ToUpper(command)

	// Check if it's COPY TO or COPY FROM
	switch {
	case strings.Contains(upperCommand, " TO "):
		return h.handleCopyTo(command)
	case strings.Contains(upperCommand, " FROM "):
		return h.handleCopyFrom(command)
	default:
		return "Invalid COPY syntax. Use: COPY table TO 'file' or COPY table FROM 'file'\nSupported formats: CSV (.csv) and Parquet (.parquet)\nFormat is auto-detected from file extension or can be specified with WITH FORMAT='csv|parquet'"
	}
}

// handleCopyTo handles COPY TO command for exporting data to CSV
func (h *MetaCommandHandler) handleCopyTo(command string) interface{} {
	// Parse command using regex
	// Pattern: COPY table [(columns)] TO 'filename' or STDOUT [WITH options]
	pattern := `(?i)COPY\s+(\S+)(?:\s*\(([^)]+)\))?\s+TO\s+(?:'([^']+)'|(\S+))(?:\s+WITH\s+(.+))?`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(command)

	if len(matches) < 5 {
		return "Invalid COPY TO syntax. Use: COPY table [(col1, col2)] TO 'file.csv|.parquet' [WITH FORMAT='csv|parquet' AND options]\nNote: Format is auto-detected from file extension (.csv, .parquet) if not specified"
	}

	table := matches[1]
	columnsStr := matches[2]
	filename := matches[3]
	if filename == "" {
		filename = matches[4] // Check for unquoted filename (e.g., STDOUT)
	}
	optionsStr := matches[5]

	// Parse columns if specified
	var columns []string
	if columnsStr != "" {
		// Split by comma and trim spaces
		for _, col := range strings.Split(columnsStr, ",") {
			columns = append(columns, strings.TrimSpace(col))
		}
	}

	// Parse options
	options := parseCopyOptions(optionsStr)

	// Execute the copy
	return h.executeCopyTo(table, columns, filename, options)
}

// parseCopyOptions parses COPY command options
func parseCopyOptions(optionsStr string) map[string]string {
	options := map[string]string{
		"HEADER":          "false",
		"NULLVAL":         "null",
		// Don't set FORMAT default here - let executeCopyTo detect from extension
		"DELIMITER":       ",",
		"QUOTE":           "\"",
		"ESCAPE":          "\\",
		"ENCODING":        "utf8",
		"PAGESIZE":        "1000",
		"MAXREQUESTS":     "6",
		"CHUNKSIZE":       "5000",
		"MAXROWS":         "-1",   // -1 means unlimited
		"SKIPROWS":        "0",    // Number of rows to skip at start
		"MAXPARSEERRORS":  "-1",   // -1 means unlimited
		"MAXINSERTERRORS": "1000", // Default max insert errors
		"MAXBATCHSIZE":    "20",   // Max rows per batch
		"MINBATCHSIZE":    "2",    // Min rows per batch
	}

	if optionsStr == "" {
		return options
	}

	// Parse key=value pairs
	// Handle both key=value and key='value' formats
	pattern := `(\w+)\s*=\s*(?:'([^']*)'|"([^"]*)"|([^,\s]+))`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(optionsStr, -1)

	for _, match := range matches {
		key := strings.ToUpper(match[1])
		// Get value from whichever group matched
		value := match[2]
		if value == "" {
			value = match[3]
		}
		if value == "" {
			value = match[4]
		}
		options[key] = value
	}

	return options
}

// executeCopyTo executes the COPY TO operation
func (h *MetaCommandHandler) executeCopyTo(table string, columns []string, filename string, options map[string]string) interface{} {
	logger.DebugfToFile("CopyTo", "Called with filename: %s, options: %+v", filename, options)

	// Determine format from option or filename extension
	format := strings.ToLower(options["FORMAT"])
	if format == "" {
		// Check file extension if format not explicitly specified
		ext := strings.ToLower(filepath.Ext(filename))
		logger.DebugfToFile("CopyTo", "Detecting format from filename: %s, extension: %s", filename, ext)
		switch ext {
		case ".parquet":
			format = "parquet"
		case ".json":
			format = "json"
		default:
			format = "csv" // Default to CSV
		}
	}
	logger.DebugfToFile("CopyTo", "Selected format: %s", format)

	// Route to appropriate handler
	if format == "parquet" {
		logger.DebugToFile("CopyTo", "Routing to Parquet handler")
		return h.executeCopyToParquet(table, columns, filename, options)
	}

	// Default to CSV format
	// Build SELECT query
	var query string
	if len(columns) > 0 {
		query = fmt.Sprintf("SELECT %s FROM %s", strings.Join(columns, ", "), table)
	} else {
		query = fmt.Sprintf("SELECT * FROM %s", table)
	}

	// Check if output is STDOUT
	isStdout := strings.ToUpper(filename) == "STDOUT"

	// Open file for writing (unless STDOUT)
	var writer io.WriteCloser
	var err error

	if isStdout {
		writer = nopCloser{os.Stdout}
	} else {
		// Create writer for local file
		writer, err = parquet.CreateWriter(context.Background(), filename)
		if err != nil {
			return fmt.Sprintf("Error creating file: %v", err)
		}
		defer writer.Close()
	}

	// Create CSV writer
	csvWriter := csv.NewWriter(writer)

	// Set delimiter
	if delimiter := options["DELIMITER"]; delimiter != "" && len(delimiter) > 0 {
		csvWriter.Comma = rune(delimiter[0])
	}

	// Execute query
	result := h.session.ExecuteCQLQuery(query)

	// Handle different result types
	switch v := result.(type) {
	case db.QueryResult:
		// Write header if requested
		if strings.ToLower(options["HEADER"]) == "true" && len(v.Headers) > 0 {
			if err := csvWriter.Write(v.Headers); err != nil {
				return fmt.Sprintf("Error writing header: %v", err)
			}
		}

		// Write data rows
		rowCount := 0
		for _, row := range v.Data {
			// Replace nulls with NULLVAL option if specified
			processedRow := make([]string, len(row))
			nullVal := options["NULLVAL"]
			for i, cell := range row {
				// Check if this is a null value and we have a NULLVAL option
				if nullVal != "" && (cell == "null" || cell == "<null>") {
					processedRow[i] = nullVal
				} else {
					processedRow[i] = cell
				}
			}

			if err := csvWriter.Write(processedRow); err != nil {
				return fmt.Sprintf("Error writing row: %v", err)
			}
			rowCount++
		}

		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			return fmt.Sprintf("Error flushing CSV: %v", err)
		}

		if isStdout {
			return nil // Don't print message when outputting to STDOUT
		}
		return fmt.Sprintf("Exported %d rows to %s", rowCount, filename)

	case db.StreamingQueryResult:
		// For streaming results, we need to iterate through the data
		defer v.Iterator.Close()

		// Get headers
		headers := v.Headers

		// Write header if requested
		if strings.ToLower(options["HEADER"]) == "true" && len(headers) > 0 {
			if err := csvWriter.Write(headers); err != nil {
				return fmt.Sprintf("Error writing header: %v", err)
			}
		}

		// Process rows
		rowCount := 0
		pageSize, _ := strconv.Atoi(options["PAGESIZE"])
		if pageSize <= 0 {
			pageSize = 1000
		}

		for {
			rowMap := make(map[string]interface{})
			if !v.Iterator.MapScan(rowMap) {
				break
			}

			// Convert to string array
			row := make([]string, len(v.ColumnNames))
			for i, colName := range v.ColumnNames {
				if val, ok := rowMap[colName]; ok {
					if val == nil {
						row[i] = options["NULLVAL"]
					} else {
						// Handle different data types specially
						row[i] = formatCSVValue(val)
					}
				} else {
					row[i] = options["NULLVAL"]
				}
			}

			if err := csvWriter.Write(row); err != nil {
				return fmt.Sprintf("Error writing row: %v", err)
			}
			rowCount++

			// Flush periodically
			if rowCount%pageSize == 0 {
				csvWriter.Flush()
			}
		}

		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			return fmt.Sprintf("Error flushing CSV: %v", err)
		}

		if isStdout {
			return nil
		}
		return fmt.Sprintf("Exported %d rows to %s", rowCount, filename)

	case error:
		return fmt.Sprintf("Query error: %v", v)

	default:
		return fmt.Sprintf("Unexpected result type: %T", result)
	}
}

// handleCopyFrom handles COPY FROM command for importing data from CSV
func (h *MetaCommandHandler) handleCopyFrom(command string) interface{} {
	// Parse command using regex
	// Pattern: COPY table [(columns)] FROM 'filename' or STDIN [WITH options]
	pattern := `(?i)COPY\s+(\S+)(?:\s*\(([^)]+)\))?\s+FROM\s+(?:'([^']+)'|(\S+))(?:\s+WITH\s+(.+))?`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(command)

	if len(matches) < 5 {
		return "Invalid COPY FROM syntax. Use: COPY table [(col1, col2)] FROM 'file.csv|.parquet' [WITH FORMAT='csv|parquet' AND options]\nNote: Format is auto-detected from file extension (.csv, .parquet) if not specified"
	}

	table := matches[1]
	columnsStr := matches[2]
	filename := matches[3]
	if filename == "" {
		filename = matches[4] // Check for unquoted filename (e.g., STDIN)
	}
	optionsStr := matches[5]

	// Parse columns if specified
	var columns []string
	if columnsStr != "" {
		// Split by comma and trim spaces
		for _, col := range strings.Split(columnsStr, ",") {
			columns = append(columns, strings.TrimSpace(col))
		}
	}

	// Parse options
	options := parseCopyOptions(optionsStr)

	// Check format option and delegate to appropriate handler
	format := strings.ToLower(options["FORMAT"])

	// If format not specified, detect from file extension
	if format == "" && !strings.EqualFold(filename, "STDIN") {
		ext := strings.ToLower(filepath.Ext(filename))
		logger.DebugfToFile("CopyFrom", "Detecting format from filename: %s, extension: %s", filename, ext)
		switch ext {
		case ".parquet":
			format = "parquet"
		case ".json":
			format = "json"
		default:
			format = "csv" // Default to CSV
		}
	}
	logger.DebugfToFile("CopyFrom", "Selected format: %s", format)

	if format == "parquet" {
		logger.DebugToFile("CopyFrom", "Routing to Parquet handler")
		return h.executeCopyFromParquet(table, columns, filename, options)
	}

	// Default to CSV format
	// Set defaults
	if options["DELIMITER"] == "" {
		options["DELIMITER"] = ","
	}
	if options["NULLVAL"] == "" {
		options["NULLVAL"] = ""
	}
	if options["HEADER"] == "" {
		options["HEADER"] = "false"
	}
	if options["CHUNKSIZE"] == "" {
		options["CHUNKSIZE"] = "5000"
	}
	if options["ENCODING"] == "" {
		options["ENCODING"] = "utf8"
	}

	// Open the CSV file
	var reader io.Reader
	var file *os.File
	var err error

	isStdin := strings.ToUpper(filename) == "STDIN"
	if isStdin {
		reader = os.Stdin
	} else {
		// Clean the filename to prevent path traversal
		cleanPath := filepath.Clean(filename)
		file, err = os.Open(cleanPath) // #nosec G304 - file path is user input but cleaned
		if err != nil {
			return fmt.Sprintf("Error opening file: %v", err)
		}
		defer file.Close()
		reader = file
	}

	// Create CSV reader
	csvReader := csv.NewReader(reader)

	// Set delimiter
	if delimiter := options["DELIMITER"]; delimiter != "" && len(delimiter) > 0 {
		csvReader.Comma = rune(delimiter[0])
	}

	// Handle QUOTE option
	if quote := options["QUOTE"]; quote != "" && len(quote) > 0 {
		csvReader.LazyQuotes = true
	}

	// Read header if present
	hasHeader := strings.ToLower(options["HEADER"]) == "true"
	var headerColumns []string
	if hasHeader {
		headerRow, err := csvReader.Read()
		if err != nil {
			return fmt.Sprintf("Error reading header: %v", err)
		}
		// Clean column names - remove (PK) and (C) suffixes that may be present
		// from COPY TO exports with HEADER=TRUE
		headerColumns = make([]string, len(headerRow))
		for i, col := range headerRow {
			cleanCol := strings.TrimSpace(col)
			// Remove (PK) suffix for primary key columns
			if idx := strings.Index(cleanCol, " (PK)"); idx != -1 {
				cleanCol = cleanCol[:idx]
			}
			// Remove (C) suffix for clustering columns
			if idx := strings.Index(cleanCol, " (C)"); idx != -1 {
				cleanCol = cleanCol[:idx]
			}
			headerColumns[i] = strings.TrimSpace(cleanCol)
		}
	}

	// If no columns specified, try to get them from the table schema or header
	if len(columns) == 0 {
		if hasHeader && len(headerColumns) > 0 {
			columns = headerColumns
		} else {
			// Get all columns from the table schema
			schemaColumns := h.getTableColumns(table)
			if len(schemaColumns) == 0 {
				return fmt.Sprintf("Cannot determine columns for table %s. Please specify columns explicitly.", table)
			}
			columns = schemaColumns
		}
	}

	// Prepare for building INSERT statements
	columnList := strings.Join(columns, ", ")

	// Process rows - use atomic counters for thread safety with concurrent workers
	var rowCount int64
	var insertErrorCount int64
	processedRows := 0 // Only accessed from main goroutine
	parseErrorCount := 0
	skippedRows := 0

	// Parse numeric options
	chunkSize, _ := strconv.Atoi(options["CHUNKSIZE"])
	maxRows, _ := strconv.Atoi(options["MAXROWS"])
	skipRows, _ := strconv.Atoi(options["SKIPROWS"])
	maxParseErrors, _ := strconv.Atoi(options["MAXPARSEERRORS"])
	maxInsertErrors, _ := strconv.Atoi(options["MAXINSERTERRORS"])
	maxBatchSize, _ := strconv.Atoi(options["MAXBATCHSIZE"])
	maxRequests, _ := strconv.Atoi(options["MAXREQUESTS"])
	nullVal := options["NULLVAL"]

	// Ensure reasonable defaults
	if maxRequests < 1 {
		maxRequests = 6
	}

	// Skip initial rows if specified
	for i := 0; i < skipRows; i++ {
		_, err := csvReader.Read()
		if err != nil {
			break
		}
		skippedRows++
	}

	// Create batch channel and wait group for concurrent execution
	batchChan := make(chan []string, maxRequests*2)
	var wg sync.WaitGroup

	// Start worker goroutines for concurrent batch execution
	for i := 0; i < maxRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batchChan {
				errors := h.executeBatch(batch)
				atomic.AddInt64(&insertErrorCount, int64(errors))
				atomic.AddInt64(&rowCount, int64(len(batch)-errors))
			}
		}()
	}

	// Prepare batch for inserts
	batch := make([]string, 0, maxBatchSize)
	lastProgress := int64(0)

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			parseErrorCount++
			if maxParseErrors != -1 && parseErrorCount > maxParseErrors {
				close(batchChan)
				wg.Wait()
				return fmt.Sprintf("Too many parse errors. Imported %d rows, failed after %d parse errors", atomic.LoadInt64(&rowCount), parseErrorCount)
			}
			continue
		}

		// Check if we've reached maxRows before processing this row
		if maxRows != -1 && processedRows >= maxRows {
			break
		}
		processedRows++

		// Check column count
		if len(record) != len(columns) {
			parseErrorCount++
			if maxParseErrors != -1 && parseErrorCount > maxParseErrors {
				close(batchChan)
				wg.Wait()
				return fmt.Sprintf("Too many parse errors. Imported %d rows, failed after %d parse errors", atomic.LoadInt64(&rowCount), parseErrorCount)
			}
			continue
		}

		// Convert values and build INSERT query
		valueStrings := make([]string, len(record))
		for i, val := range record {
			// Handle NULL values
			if val == nullVal {
				valueStrings[i] = "NULL"
			} else {
				valueStrings[i] = h.formatValueForInsert(val, columns[i], table)
			}
		}

		// Build INSERT query
		insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			table, columnList, strings.Join(valueStrings, ", "))

		// Add to batch
		batch = append(batch, insertQuery)

		// Send batch to workers if it reaches maxBatchSize
		if len(batch) >= maxBatchSize {
			// Check for too many insert errors
			if maxInsertErrors != -1 && atomic.LoadInt64(&insertErrorCount) > int64(maxInsertErrors) {
				close(batchChan)
				wg.Wait()
				return fmt.Sprintf("Too many insert errors. Imported %d rows, failed after %d insert errors", atomic.LoadInt64(&rowCount), atomic.LoadInt64(&insertErrorCount))
			}
			// Send batch to workers (make a copy since we reuse the slice)
			batchCopy := make([]string, len(batch))
			copy(batchCopy, batch)
			batchChan <- batchCopy
			batch = batch[:0] // Clear batch
		}

		// Progress update for large imports
		currentRows := atomic.LoadInt64(&rowCount)
		if currentRows-lastProgress >= int64(chunkSize) && !isStdin {
			fmt.Printf("\rImported %d rows...", currentRows)
			lastProgress = currentRows
		}
	}

	// Send any remaining batch
	if len(batch) > 0 {
		batchCopy := make([]string, len(batch))
		copy(batchCopy, batch)
		batchChan <- batchCopy
	}

	// Close channel and wait for all workers to finish
	close(batchChan)
	wg.Wait()

	finalRowCount := atomic.LoadInt64(&rowCount)
	finalInsertErrors := atomic.LoadInt64(&insertErrorCount)

	if !isStdin && finalRowCount > int64(chunkSize) {
		fmt.Println() // New line after progress updates
	}

	totalErrors := int64(parseErrorCount) + finalInsertErrors
	if totalErrors > 0 {
		details := fmt.Sprintf("Imported %d rows from %s", finalRowCount, filename)
		if skipRows > 0 {
			details += fmt.Sprintf(" (skipped %d rows)", skippedRows)
		}
		if parseErrorCount > 0 {
			details += fmt.Sprintf(" (%d parse errors)", parseErrorCount)
		}
		if finalInsertErrors > 0 {
			details += fmt.Sprintf(" (%d insert errors)", finalInsertErrors)
		}
		return details
	}
	details := fmt.Sprintf("Imported %d rows from %s", finalRowCount, filename)
	if skipRows > 0 {
		details += fmt.Sprintf(" (skipped %d rows)", skippedRows)
	}
	return details
}
